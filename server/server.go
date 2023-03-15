package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/k1LoW/octoslack/config"
	"github.com/k1LoW/octoslack/transformer"
	"golang.org/x/exp/slog"
	"golang.org/x/sync/errgroup"
)

var _ http.Handler = (*Server)(nil)

type Server struct {
	tr *transformer.Transformer
	hc *http.Client
	mu sync.RWMutex
}

func NewUnstartedServer(cfg *config.Config) *Server {
	return &Server{
		tr: transformer.New(cfg),
		hc: &http.Client{
			Timeout:   30 * time.Second,
			Transport: http.DefaultTransport.(*http.Transport).Clone(),
		},
	}
}

func (s *Server) Start(ctx context.Context, port uint64) error {
	addr := fmt.Sprintf(":%d", port)
	slog.Info("Start server", slog.String("addr", addr))
	hs := &http.Server{
		Addr:              addr,
		Handler:           s,
		ReadHeaderTimeout: 30 * time.Second,
	}
	eg := &errgroup.Group{}
	eg.Go(func() error {
		if err := hs.ListenAndServe(); err != nil {
			if errors.Is(http.ErrServerClosed, err) {
				return nil
			}
			return fmt.Errorf("failed to close: %w", err)
		}
		return nil
	})

	<-ctx.Done()
	if err := hs.Shutdown(context.Background()); err != nil {
		slog.Error("failed to shutdown", err)
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func Start(ctx context.Context, cfg *config.Config, port uint64) error {
	s := NewUnstartedServer(cfg)
	return s.Start(ctx, port)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.Info("Request received", slog.String("method", r.Method), slog.String("url", r.URL.String()))
	s.mu.Lock()
	defer s.mu.Unlock()
	req, err := s.tr.Transform(r)
	if err != nil {
		if errors.Is(transformer.ErrNoneOfConditionsMet, err) || errors.Is(transformer.ErrDropAction, err) {
			w.WriteHeader(http.StatusNotFound)
			slog.Info("Request dropped", slog.String("reason", fmt.Sprintf("%s", err)), slog.String("method", r.Method), slog.String("url", r.URL.String()))
			_, _ = w.Write([]byte(fmt.Sprintf("Request dropped, because %s", err)))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("Failed to transform", err)
		_, _ = w.Write([]byte(fmt.Sprintf("%v", err)))
		return
	}
	resp, err := s.hc.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("Failed to request", err)
		_, _ = w.Write([]byte(fmt.Sprintf("%v", err)))
		return
	}
	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("Failed to read response", err)
		_, _ = w.Write([]byte(fmt.Sprintf("%v", err)))
		return
	}
	_, _ = w.Write(buf.Bytes())
}

func (s *Server) UpdateConfig(cfg *config.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tr = transformer.New(cfg)
	return nil
}
