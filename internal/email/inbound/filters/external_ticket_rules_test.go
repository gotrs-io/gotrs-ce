package filters

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadExternalTicketRules(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "rules.yaml")
	content := []byte(`rules:
  - name: vendor
    pattern: "Case #([0-9]+)"
    search_subject: true
    headers:
      - X-Vendor-Case
`)
	if err := os.WriteFile(file, content, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	rules, err := LoadExternalTicketRules(file)
	if err != nil {
		t.Fatalf("LoadExternalTicketRules returned error: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if !rules[0].SearchSubject || len(rules[0].Headers) != 1 {
		t.Fatalf("rule not parsed correctly: %+v", rules[0])
	}
}

func TestLoadExternalTicketRulesMissingFile(t *testing.T) {
	if rules, err := LoadExternalTicketRules(filepath.Join(t.TempDir(), "missing.yaml")); err != nil || len(rules) != 0 {
		t.Fatalf("expected no rules for missing file, got %v %v", rules, err)
	}
}

func TestLoadExternalTicketRulesRequiresValidEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.yaml")
	if err := os.WriteFile(path, []byte("rules: []"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := LoadExternalTicketRules(path); err == nil {
		t.Fatalf("expected error for empty rules")
	}
}
