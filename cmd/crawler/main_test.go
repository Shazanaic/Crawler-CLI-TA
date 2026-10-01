package main

import (
	"testing"
)

// только эту функцию покрыл, а то мейн слишком много всего делает, что уже покрыто тестами
func TestSplitURLs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single URL",
			input: "https://example.com",
			want:  []string{"https://example.com"},
		},
		{
			name:  "multiple URLs",
			input: "https://example.com,https://google.com",
			want:  []string{"https://example.com", "https://google.com"},
		},
		{
			name:  "URLs with spaces",
			input: "https://example.com, https://google.com",
			want:  []string{"https://example.com", "https://google.com"},
		},
		{
			name:  "empty elements",
			input: "https://example.com,, ,https://google.com",
			want:  []string{"https://example.com", "https://google.com"},
		},
		{
			name:  "empty string",
			input: "",
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.input

			got := splitURLs(input)

			if len(got) != len(tt.want) {
				t.Fatalf("expected %v URLs, got %v", len(tt.want), len(got))
			}

			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("URL %v: expected %q, got %q", i, tt.want[i], got[i])
				}
			}
		})
	}
}
