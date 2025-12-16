package filters

import (
	"regexp"
	"strings"
)

var ticketTokenRegexp = regexp.MustCompile(`(?i)\[\s*ticket\s*#\s*([0-9]{1,20})[^\]]*]`)

func findTicketToken(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}
	matches := ticketTokenRegexp.FindStringSubmatch(input)
	if len(matches) < 2 {
		return ""
	}
	return strings.TrimSpace(matches[1])
}

func stripHTML(input string) string {
	if input == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(input))
	inTag := false
	for _, r := range input {
		switch r {
		case '<':
			inTag = true
		case '>':
			if inTag {
				inTag = false
				b.WriteRune(' ')
			}
		default:
			if !inTag {
				b.WriteRune(r)
			}
		}
	}
	return strings.TrimSpace(b.String())
}
