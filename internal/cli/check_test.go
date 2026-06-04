package cli

import (
	"os"
	"testing"

	"github.com/rd-mg/architect-ai/internal/paths"
)

func TestCheckFoundation(t *testing.T) {
	// Setup
	os.MkdirAll(".atl/_generated", 0755)
	defer os.RemoveAll(".atl")
	
	// Create foundation.md
	content := []byte("architect-ai:foundation:start")
	os.WriteFile(".atl/_generated/foundation.md", content, 0644)
	
	ctx := paths.New(".", false)
	c := checkFoundation(ctx)
	err := c.Run()
	if err != nil {
		t.Errorf("Expected foundation check to succeed, got %v", err)
	}
}
