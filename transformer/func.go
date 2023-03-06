package transformer

import (
	"fmt"
	"strings"
)

func quote(v interface{}) string {
	lines := strings.Split(v.(string), "\n")
	quoted := []string{}
	for _, l := range lines {
		ql := fmt.Sprintf("> %s", l)
		if ql == "> " {
			ql = ">"
		}
		quoted = append(quoted, ql)
	}
	return strings.Join(quoted, "\n")
}

func quoteMarkdown(v interface{}) string {
	lines := strings.Split(v.(string), "\n")
	quoted := []string{}
	inBlock := false
	for _, l := range lines {
		if strings.HasPrefix(l, "```") {
			inBlock = !inBlock
			if !inBlock {
				// codeblock end
				quoted = append(quoted, l)
				continue
			}
		}
		if inBlock && !strings.HasPrefix(l, "```") {
			quoted = append(quoted, l)
			continue
		}
		ql := fmt.Sprintf("> %s", l)
		if ql == "> " {
			ql = ">"
		}
		quoted = append(quoted, ql)
	}
	return strings.Join(quoted, "\n")
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
