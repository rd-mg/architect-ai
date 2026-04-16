package agents

import (
	"testing"

	"github.com/rd-mg/architect-ai/internal/catalog"
	"github.com/rd-mg/architect-ai/internal/model"
)

func TestSupportedAgentsHaveCatalogAndRegistryParity(t *testing.T) {
	reg, err := NewDefaultRegistry()
	if err != nil {
		t.Fatalf("failed to create default registry: %v", err)
	}

	supported := reg.SupportedAgents()
	
	// Check that all supported agents are in the factory/registry
	for _, id := range supported {
		_, ok := reg.Get(id)
		if !ok {
			t.Errorf("agent %q is in SupportedAgents() but not in Registry", id)
		}
	}

	// Check for specific critical agents
	critical := []model.AgentID{
		model.AgentClaudeCode,
		model.AgentOpenCode,
		model.AgentGeminiCLI,
		model.AgentCursor,
		model.AgentKiroIDE,
	}

	for _, id := range critical {
		found := false
		for _, s := range supported {
			if s == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("critical agent %q is missing from Registry", id)
		}
	}
}

func TestMVPSkillsParity(t *testing.T) {
	skills := catalog.MVPSkills()
	if len(skills) == 0 {
		t.Fatal("MVPSkills() is empty")
	}

	// Ensure sdd group is correctly represented
	sddCount := 0
	for _, s := range skills {
		if s.Category == "sdd" {
			sddCount++
		}
	}
	if sddCount < 10 {
		t.Errorf("expected at least 10 SDD skills, found %d", sddCount)
	}
}
