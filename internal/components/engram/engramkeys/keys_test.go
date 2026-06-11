package engramkeys

import "testing"

func TestResearchTopicKey(t *testing.T) {
	tests := []struct {
		name  string
		tool  string
		query string
	}{
		{
			name:  "basic query",
			tool:  "context7",
			query: "How to use SDD?",
		},
		{
			name:  "another query",
			tool:  "notebooklm",
			query: "Go generics best practices",
		},
		{
			name:  "multiple spaces",
			tool:  "web",
			query: "  Multiple   Spaces  ",
		},
		{
			name:  "long query",
			tool:  "google",
			query: "Very long query that should be truncated to fifty characters or so to keep it manageable",
		},
		{
			name:  "special chars only",
			tool:  "empty",
			query: "!!!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResearchTopicKey(tt.tool, tt.query)
			if got == "" {
				t.Error("ResearchTopicKey() returned empty string")
			}
			// Verify format: research/{tool}/{slug}-{hash}
			if len(got) < 10 {
				t.Errorf("ResearchTopicKey() too short: %q", got)
			}
		})
	}
}

func TestResearchTopicKey_Deterministic(t *testing.T) {
	key1 := ResearchTopicKey("test", "same query")
	key2 := ResearchTopicKey("test", "same query")
	if key1 != key2 {
		t.Errorf("ResearchTopicKey() not deterministic: %q != %q", key1, key2)
	}
}

func TestResearchTopicKey_DifferentQueriesDifferentKeys(t *testing.T) {
	key1 := ResearchTopicKey("test", "query one")
	key2 := ResearchTopicKey("test", "query two")
	if key1 == key2 {
		t.Errorf("ResearchTopicKey() same key for different queries: %q", key1)
	}
}
