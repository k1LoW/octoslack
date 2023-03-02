package transformer

import (
	"bytes"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/octoslack/config"
	"github.com/tenntenn/golden"
)

func TestTransform(t *testing.T) {
	tests := []struct {
		config  string
		url     string
		event   string
		payload string
		wantErr bool
	}{
		{"../testdata/config.empty.yml", "", "", "../testdata/empty.json", true},
		{
			"../testdata/config.yml",
			"https://octoslack.example.com/services/XXXXXxxxxxXXXXXX/XXXxxxxXXXXXXxxxxXXXXXX",
			"discussion",
			"../testdata/discussion_create.json",
			false,
		},
		{
			"../testdata/config.yml",
			"https://octoslack.example.com/services/XXXXXxxxxxXXXXXX/XXXxxxxXXXXXXxxxxXXXXXX",
			"discussion",
			"../testdata/empty.json",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.payload, func(t *testing.T) {
			cfg, err := config.Load(tt.config)
			if err != nil {
				t.Fatal(err)
			}
			tr := New(cfg)

			// create http request
			b, err := os.ReadFile(tt.payload)
			if err != nil {
				t.Fatal(err)
			}
			req, err := http.NewRequest(http.MethodPost, tt.url, bytes.NewReader(b))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Add("X-GitHub-Event", tt.event)

			treq, err := tr.Transform(req)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("got error: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Error("want error")
			}
			got, err := httputil.DumpRequest(treq, true)
			if err != nil {
				t.Error(err)
			}
			d, f := filepath.Split(tt.payload)
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
