package config

import "testing"

func TestLoad(t *testing.T) {
	tests := []struct {
		path    string
		wantErr bool
	}{
		{"../testdata/config.yml", false},
		{"../testdata/config.empty.yml", false},
		{"../testdata/notexist.yml", true},
		{"github://k1LoW/octoslack/config.example.yml", false},
		{"github://k1LoW/octoslack/notexist.yml", true},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			_, err := Load(tt.path)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("got error: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Error("want error")
			}
		})
	}
}
