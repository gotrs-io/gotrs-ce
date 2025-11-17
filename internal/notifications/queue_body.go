package notifications

import (
	"html"
	"strings"

	"github.com/gotrs-io/gotrs-ce/internal/utils"
)

// ApplyBranding stitches salutation and signature around the base body.
func ApplyBranding(base string, baseIsHTML bool, identity *QueueIdentity) string {
	if identity == nil {
		return strings.TrimSpace(base)
	}
	return composeBody(base, baseIsHTML, identity.SalutationSnippet(), identity.SignatureSnippet())
}

func composeBody(base string, baseIsHTML bool, salutation, signature *Snippet) string {
	trimmed := strings.TrimSpace(base)
	finalIsHTML := baseIsHTML || snippetIsHTML(salutation) || snippetIsHTML(signature)
	if finalIsHTML {
		var parts []string
		if salutation != nil {
			parts = append(parts, snippetAsHTML(salutation))
		}
		if trimmed != "" {
			parts = append(parts, textAsHTML(trimmed, baseIsHTML))
		}
		if signature != nil {
			parts = append(parts, snippetAsHTML(signature))
		}
		return strings.Join(filterEmpty(parts), "\n")
	}

	var blocks []string
	if salutation != nil {
		blocks = append(blocks, snippetAsText(salutation))
	}
	if trimmed != "" {
		blocks = append(blocks, trimmed)
	}
	if signature != nil {
		blocks = append(blocks, snippetAsText(signature))
	}
	return strings.Join(filterEmpty(blocks), "\n\n")
}

func snippetIsHTML(snippet *Snippet) bool {
	return snippet != nil && strings.Contains(snippet.ContentType, "html")
}

func snippetAsText(snippet *Snippet) string {
	if snippet == nil {
		return ""
	}
	if snippetIsHTML(snippet) {
		return strings.TrimSpace(utils.StripHTML(snippet.Text))
	}
	return strings.TrimSpace(snippet.Text)
}

func snippetAsHTML(snippet *Snippet) string {
	if snippet == nil {
		return ""
	}
	if snippetIsHTML(snippet) {
		return strings.TrimSpace(snippet.Text)
	}
	return wrapPlainText(snippet.Text)
}

func textAsHTML(content string, alreadyHTML bool) string {
	if alreadyHTML {
		return strings.TrimSpace(content)
	}
	return wrapPlainText(content)
}

func wrapPlainText(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	escaped := html.EscapeString(trimmed)
	escaped = strings.ReplaceAll(escaped, "\n", "<br>")
	return "<p>" + escaped + "</p>"
}

func filterEmpty(values []string) []string {
	result := make([]string, 0, len(values))
	for _, v := range values {
		if strings.TrimSpace(v) == "" {
			continue
		}
		result = append(result, v)
	}
	return result
}
