package engramkeys

import "testing"

func TestResearchTopicKey(t *testing.T) {
	tests := []struct {
		tool     string
		query    string
		expected string
	}{
		{
			tool:     "context7",
			query:    "How to use SDD?",
			expected: "research/context7/how-to-use-sdd-len15",
		},
		{
			tool:     "notebooklm",
			query:    "Go generics best practices",
			expected: "research/notebooklm/go-generics-best-practices-len26",
		},
		{
			tool:     "web",
			query:    "  Multiple   Spaces  ",
			expected: "research/web/multiple-spaces-len21",
		},
		{
			tool:     "google",
			query:    "Very long query that should be truncated to fifty characters or so to keep it manageable",
			expected: "research/google/very-long-query-that-should-be-truncated-to-fifty-len88",
		},
		{
			tool:     "empty",
			query:    "!!!",
			expected: "research/empty/query-len3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := ResearchTopicKey(tt.tool, tt.query)
			if got != tt.expected {
				t.Errorf("ResearchTopicKey(%q, %q) = %q; want %q", tt.tool, tt.query, got, tt.expected)
			}
		})
	}
}
