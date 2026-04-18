package antigravity

import (
	"testing"
)

func TestExtractUsage_GoogleNative(t *testing.T) {
	raw := []byte(`{
		"usageMetadata": {
			"promptTokenCount": 100,
			"candidatesTokenCount": 50
		},
		"model": "gemini-1.5-pro"
	}`)
	
	delta, err := ExtractUsage(raw)
	if err != nil {
		t.Fatalf("Failed to extract Google-native usage: %v", err)
	}
	
	if delta.PromptTokens != 100 {
		t.Errorf("Expected 100 prompt tokens, got %d", delta.PromptTokens)
	}
	if delta.CompletionTokens != 50 {
		t.Errorf("Expected 50 completion tokens, got %d", delta.CompletionTokens)
	}
}

func TestExtractUsage_OpenAI(t *testing.T) {
	raw := []byte(`{
		"usage": {
			"prompt_tokens": 80,
			"completion_tokens": 40
		},
		"model": "gpt-4o"
	}`)
	
	delta, err := ExtractUsage(raw)
	if err != nil {
		t.Fatalf("Failed to extract OpenAI usage: %v", err)
	}
	
	if delta.PromptTokens != 80 {
		t.Errorf("Expected 80 prompt tokens, got %d", delta.PromptTokens)
	}
}

func TestRecordResponse_Safety(t *testing.T) {
	// Should not panic
	RecordResponse(nil)
	RecordResponse([]byte("garbage"))
}
