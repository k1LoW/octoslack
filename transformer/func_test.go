package transformer

import (
	"fmt"
	"strings"
	"testing"
)

func TestQuote(t *testing.T) {
	tests := []struct {
		in   []string
		want []string
	}{
		{
			[]string{"a", "b", "c"},
			[]string{">>> ", "a", "b", "c"},
		},
		{
			[]string{"a", "", "c"},
			[]string{">>> ", "a", "", "c"},
		},
		{
			[]string{"following code", "``` sh", "go run main.go", "```", "ok"},
			[]string{">>> ", "following code", "``` sh", "go run main.go", "```", "ok"},
		},
		{
			[]string{"a"},
			[]string{">>> ", "a"},
		},
		{
			[]string{"> a", "> b", "c"},
			[]string{">>> ", "> a", "> b", "c"},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := quote(strings.Join(tt.in, "\n"))
			if got != strings.Join(tt.want, "\n") {
				t.Errorf("got %#v\nwant %#v", got, strings.Join(tt.want, "\n"))
			}
		})
	}
}

func TestShortenLines(t *testing.T) {
	tests := []struct {
		in      string
		message string
		c       int
		want    string
	}{
		{"a\nb\nc\n", "", 5, "a\nb\nc\n"},
		{"a\nb\nc\n", "", 1, "a"},
		{"a\nb\nc\n", "", 2, "a\nb"},
		{"a\nb\nc\n", "(snip)", 2, "a\nb\n(snip)"},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := shortenLines(tt.in, tt.c, tt.message)
			if got != tt.want {
				t.Errorf("got %#v\nwant %#v", got, tt.want)
			}
		})
	}
}
