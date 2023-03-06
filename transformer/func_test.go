package transformer

import (
	"fmt"
	"testing"
)

func TestShortenLines(t *testing.T) {
	tests := []struct {
		in   string
		c    int
		want string
	}{
		{"a\nb\nc\n", 5, "a\nb\nc\n"},
		{"a\nb\nc\n", 1, "a"},
		{"a\nb\nc\n", 2, "a\nb"},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := shortenLines(tt.in, tt.c, "")
			if got != tt.want {
				t.Errorf("got %#v\nwant %#v", got, tt.want)
			}
		})
	}
}
