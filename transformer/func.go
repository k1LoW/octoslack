package transformer

import (
	"strings"
)

func quote(v interface{}) string {
	lines := strings.Split(v.(string), "\n")
	quoted := []string{">>> "}
	for _, l := range lines {
		quoted = append(quoted, l)
	}
	return strings.Join(quoted, "\n")
}

func quoteMarkdown(v interface{}) string {
	// deprecated
	return quote(v)
}

func shortenLines(v interface{}, c int, shortenMessage string) string {
	lines := strings.Split(v.(string), "\n")
	if len(lines) < c {
		return strings.Join(lines, "\n")
	}
	if shortenMessage == "" {
		return strings.Join(lines[:c], "\n")
	}
	return strings.Join(append(lines[:c], shortenMessage), "\n")
}

func shortenLinesMarkdown(v interface{}, c int, shortenMessage string) string {
	lines := strings.Split(v.(string), "\n")
	if len(lines) < c {
		return strings.Join(lines, "\n")
	}
	shortened := lines[:c]
	inBlock := false
	for _, l := range shortened {
		if strings.HasPrefix(l, "```") {
			inBlock = !inBlock
		}
	}
	if inBlock {
		shortened = append(shortened, "```")
	}
	if shortenMessage == "" {
		return strings.Join(shortened, "\n")
	}
	return strings.Join(append(shortened, shortenMessage), "\n")
}
