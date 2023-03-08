package server

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/httpstub"
	"github.com/k1LoW/octoslack/config"
	"github.com/tenntenn/golden"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		config         string
		event          string
		payload        string
		wantStatusCode int
	}{
		{
			"../testdata/config.yml",
			"discussion",
			"../testdata/discussion_create.json",
			http.StatusOK,
		},
		{
			"../testdata/config.false.yml",
			"discussion",
			"../testdata/discussion_create.json",
			http.StatusNotFound,
		},
	}

	slk := httpstub.NewServer(t)
	t.Cleanup(func() {
		slk.Close()
	})
	slk.Method(http.MethodPost).ResponseString(http.StatusOK, "ok")
	sc := slk.Client()

	for _, tt := range tests {
		t.Run(tt.payload, func(t *testing.T) {
			t.Cleanup(func() {
				slk.ClearRequests()
			})
			cfg, err := config.Load(tt.config)
			if err != nil {
				t.Fatal(err)
			}
			s := NewUnstartedServer(cfg)
			s.hc = sc
			ts := httptest.NewServer(s)
			t.Cleanup(func() {
				ts.Close()
			})
			hc := ts.Client()

			// create http request
			b, err := os.ReadFile(tt.payload)
			if err != nil {
				t.Fatal(err)
			}
			u := fmt.Sprintf("%s/services/XXXXXxxxxxXXXXXX/XXXxxxxXXXXXXxxxxXXXXXX", ts.URL)
			req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(b))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("X-GitHub-Event", tt.event)

			resp, err := hc.Do(req)
			if err != nil {
				t.Error(err)
			}
			t.Cleanup(func() {
				resp.Body.Close()
			})

			if got := resp.StatusCode; got != tt.wantStatusCode {
				t.Errorf("got %v\nwant %v", got, tt.wantStatusCode)
			}
			if resp.StatusCode >= 300 {
				return
			}
			sreq := slk.Requests()[0]
			t.Cleanup(func() {
				sreq.Body.Close()
			})

			got, err := httputil.DumpRequest(sreq, true)
			if err != nil {
				t.Error(err)
			}

			d, f := filepath.Split(tt.payload)
			f = fmt.Sprintf("%s.server", f)
			if os.Getenv("UPDATE_GOLDEN") != "" {
				golden.Update(t, d, f, got)
				return
			}
			if diff := golden.Diff(t, d, f, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestUpdateConfig(t *testing.T) {
	tests := []struct {
		config         string
		event          string
		payload        string
		wantStatusCode int
	}{
		{
			"../testdata/config.false.yml",
			"discussion",
			"../testdata/discussion_create.json",
			http.StatusNotFound,
		},
		{
			"../testdata/config.yml",
			"discussion",
			"../testdata/discussion_create.json",
			http.StatusOK,
		},
	}

	slk := httpstub.NewServer(t)
	t.Cleanup(func() {
		slk.Close()
	})
	slk.Method(http.MethodPost).ResponseString(http.StatusOK, "ok")
	sc := slk.Client()
	cfg, err := config.Load("../testdata/config.false.yml")
	if err != nil {
		t.Fatal(err)
	}
	s := NewUnstartedServer(cfg)
	s.hc = sc
	ts := httptest.NewServer(s)
	t.Cleanup(func() {
		ts.Close()
	})
	hc := ts.Client()

	for _, tt := range tests {
		t.Run(tt.payload, func(t *testing.T) {
			t.Cleanup(func() {
				slk.ClearRequests()
			})
			cfg, err := config.Load(tt.config)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.UpdateConfig(cfg); err != nil {
				t.Fatal(err)
			}
			// create http request
			b, err := os.ReadFile(tt.payload)
			if err != nil {
				t.Fatal(err)
			}
			u := fmt.Sprintf("%s/services/XXXXXxxxxxXXXXXX/XXXxxxxXXXXXXxxxxXXXXXX", ts.URL)
			req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(b))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("X-GitHub-Event", tt.event)

			resp, err := hc.Do(req)
			if err != nil {
				t.Error(err)
			}
			t.Cleanup(func() {
				resp.Body.Close()
			})

			if got := resp.StatusCode; got != tt.wantStatusCode {
				t.Errorf("got %v\nwant %v", got, tt.wantStatusCode)
			}
			if resp.StatusCode >= 300 {
				return
			}
			sreq := slk.Requests()[0]
			t.Cleanup(func() {
				sreq.Body.Close()
			})

			got, err := httputil.DumpRequest(sreq, true)
			if err != nil {
				t.Error(err)
			}

			d, f := filepath.Split(tt.payload)
			f = fmt.Sprintf("%s.server", f)
			if os.Getenv("UPDATE_GOLDEN") != "" {
				golden.Update(t, d, f, got)
				return
			}
			if diff := golden.Diff(t, d, f, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
