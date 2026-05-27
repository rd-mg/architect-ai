package screens

import (
	"strings"
	"testing"
)

func TestRenderPurge(t *testing.T) {
	choices := []PurgeChoice{
		{Kind: PurgeChoiceManagedConfig, Label: "Managed Config", Selected: true},
		{Kind: PurgeChoiceEngramProject, Label: "Engram", Selected: false},
	}
	
	output := RenderPurge(choices, 0)
	
	if !strings.Contains(output, "Managed Config") {
		t.Errorf("Expected output to contain 'Managed Config'")
	}
	if !strings.Contains(output, "[x]") {
		t.Errorf("Expected [x] for selected choice")
	}
	if !strings.Contains(output, "[ ]") {
		t.Errorf("Expected [ ] for unselected choice")
	}
}
