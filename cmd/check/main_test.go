package main

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
	// checkFoundation returns error if file exists (returns age/size error), 
	// which is how it signals success ("not found" is the failure condition)
	if err == nil {
		t.Errorf("Expected foundation check to return error (age/size info)")
	}
}
