package config

import "testing"

func TestLoad(t *testing.T) {
	tests := []struct {
		path    string
		want    int
		wantErr bool
	}{
		{"../testdata/config.yml", 1, false},
		{"../testdata/config.empty.yml", 0, false},
		{"../testdata/notexist.yml", 0, true},
		{"github://k1LoW/octoslack/config.example.yml", 3, false},
		{"github://k1LoW/octoslack/notexist.yml", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			cfg, err := Load(tt.path)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("got error: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Error("want error")
			}
			if got := len(cfg.Requests); got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
		})
	}
}
