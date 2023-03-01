package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/k1LoW/octoslack/config"
	"github.com/k1LoW/octoslack/transformer"
	"golang.org/x/sync/errgroup"
)

var _ http.Handler = (*Server)(nil)

type Server struct {
	tr *transformer.Transformer
	hc *http.Client
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

func (s *Server) Start(ctx context.Context, port uint) error {
	hs := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
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
		log.Printf("failed to shutdown: %v", err)
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func Start(ctx context.Context, cfg *config.Config, port uint) error {
	s := NewUnstartedServer(cfg)
	return s.Start(ctx, port)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := s.tr.Transform(r)
	if err != nil {
		if errors.Is(transformer.ErrNoneOfConditionsMet, err) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("Request dropped, because %s", err)))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(fmt.Sprintf("%v", err)))
		return
	}
	resp, err := s.hc.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(fmt.Sprintf("%v", err)))
		return
	}
	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(fmt.Sprintf("%v", err)))
		return
	}
	_, _ = w.Write(buf.Bytes())
}
