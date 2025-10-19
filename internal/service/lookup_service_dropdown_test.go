package service

import (
	"encoding/json"
	"testing"
	"unicode"
)

func TestLookupServiceDropdownData(t *testing.T) {
	lookupService := NewLookupService()

	// Test English dropdown data
	t.Run("English Dropdown Data", func(t *testing.T) {
		formData := lookupService.GetTicketFormDataWithLang("en")

		// Check if we have priorities
		if len(formData.Priorities) == 0 {
			t.Error("No priorities returned for English")
		} else {
			t.Logf("English Priorities (%d):", len(formData.Priorities))
			for _, p := range formData.Priorities {
				t.Logf("  - Value: %q, Label: %q, Active: %v", p.Value, p.Label, p.Active)
				if p.Label == "" {
					t.Errorf("Empty label for priority with value %q", p.Value)
				}
			}
		}

		// Check if we have statuses
		if len(formData.Statuses) == 0 {
			t.Error("No statuses returned for English")
		} else {
			t.Logf("English Statuses (%d):", len(formData.Statuses))
			for _, s := range formData.Statuses {
				t.Logf("  - Value: %q, Label: %q, Active: %v", s.Value, s.Label, s.Active)
				if s.Label == "" {
					t.Errorf("Empty label for status with value %q", s.Value)
				}
			}
		}

		// Check if we have queues
		if len(formData.Queues) == 0 {
			t.Error("No queues returned for English")
		} else {
			t.Logf("English Queues (%d):", len(formData.Queues))
			for _, q := range formData.Queues {
				t.Logf("  - ID: %d, Name: %q, Active: %v", q.ID, q.Name, q.Active)
				if q.Name == "" {
					t.Errorf("Empty name for queue with ID %d", q.ID)
				}
			}
		}

		// Output as JSON for debugging
		jsonBytes, _ := json.MarshalIndent(formData, "", "  ")
		t.Logf("Full English form data:\n%s", string(jsonBytes))
	})

	// Test German dropdown data
	t.Run("German Dropdown Data", func(t *testing.T) {
		formData := lookupService.GetTicketFormDataWithLang("de")

		// Check if we have priorities
		if len(formData.Priorities) == 0 {
			t.Error("No priorities returned for German")
		} else {
			t.Logf("German Priorities (%d):", len(formData.Priorities))
			for _, p := range formData.Priorities {
				t.Logf("  - Value: %q, Label: %q, Active: %v", p.Value, p.Label, p.Active)
				if p.Label == "" {
					t.Errorf("Empty label for priority with value %q", p.Value)
				}
				if startsWithDigit(p.Value) && p.Label == p.Value {
					t.Errorf("Priority label not translated: %q", p.Label)
				}
			}
		}

		// Check if "Alles" (All) option exists
		hasAlles := false
		for _, p := range formData.Priorities {
			if p.Label == "Alles" || p.Label == "Alle" {
				hasAlles = true
				break
			}
		}
		if hasAlles {
			t.Error("'Alles' should not be in the priority list - it's added by the template")
		}
	})
}

func startsWithDigit(s string) bool {
	for _, r := range s {
		return unicode.IsDigit(r)
	}
	return false
}
