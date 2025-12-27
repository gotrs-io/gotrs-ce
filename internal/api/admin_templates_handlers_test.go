package api

import (
	"testing"
)

func TestValidTemplateTypes(t *testing.T) {
	types := ValidTemplateTypes()

	if len(types) != 8 {
		t.Errorf("Expected 8 template types, got %d", len(types))
	}

	expected := []string{"Answer", "Create", "Email", "Forward", "Note", "PhoneCall", "ProcessManagement", "Snippet"}
	for i, tt := range types {
		if tt.Key != expected[i] {
			t.Errorf("Expected type %s at position %d, got %s", expected[i], i, tt.Key)
		}
	}
}

func TestParseTemplateTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty string", "", nil},
		{"single type", "Answer", []string{"Answer"}},
		{"multiple types", "Answer,Forward,Note", []string{"Answer", "Forward", "Note"}},
		{"types with spaces", "Answer, Forward, Note", []string{"Answer", "Forward", "Note"}},
		{"unsorted input", "Note,Answer,Forward", []string{"Answer", "Forward", "Note"}},
		{"duplicate types", "Answer,Answer,Note", []string{"Answer", "Answer", "Note"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseTemplateTypes(tc.input)
			if tc.expected == nil && result != nil {
				t.Errorf("Expected nil, got %v", result)
				return
			}
			if tc.expected != nil && result == nil {
				t.Errorf("Expected %v, got nil", tc.expected)
				return
			}
			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d types, got %d", len(tc.expected), len(result))
				return
			}
			for i, exp := range tc.expected {
				if result[i] != exp {
					t.Errorf("Expected %s at position %d, got %s", exp, i, result[i])
				}
			}
		})
	}
}

func TestJoinTemplateTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{"empty slice", []string{}, ""},
		{"single type", []string{"Answer"}, "Answer"},
		{"multiple types sorted", []string{"Answer", "Forward", "Note"}, "Answer,Forward,Note"},
		{"multiple types unsorted", []string{"Note", "Answer", "Forward"}, "Answer,Forward,Note"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := JoinTemplateTypes(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestStandardTemplateStruct(t *testing.T) {
	template := StandardTemplate{
		ID:           1,
		Name:         "Test Template",
		Text:         "Hello, this is a test",
		ContentType:  "text/plain",
		TemplateType: "Answer,Forward",
		Comments:     "Test comment",
		ValidID:      1,
	}

	if template.ID != 1 {
		t.Errorf("Expected ID 1, got %d", template.ID)
	}
	if template.Name != "Test Template" {
		t.Errorf("Expected Name 'Test Template', got %s", template.Name)
	}
	if template.TemplateType != "Answer,Forward" {
		t.Errorf("Expected TemplateType 'Answer,Forward', got %s", template.TemplateType)
	}
}

func TestStandardTemplateWithStatsStruct(t *testing.T) {
	template := StandardTemplateWithStats{
		StandardTemplate: StandardTemplate{
			ID:   1,
			Name: "Test",
		},
		QueueCount: 5,
	}

	if template.QueueCount != 5 {
		t.Errorf("Expected QueueCount 5, got %d", template.QueueCount)
	}
}
