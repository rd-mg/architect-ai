package skill

import (
	"strings"
	"testing"

	"github.com/rd-mg/architect-ai/internal/assets"
)

func TestBashExpert_FishPatterns(t *testing.T) {
	content := assets.MustRead("skills/bash-expert/SKILL.md")
	if len(content) == 0 {
		t.Fatal("bash-expert/SKILL.md is empty")
	}

	requiredFishSnippets := []string{
		"fish_add_path",
		"set -x MY_VAR",
		"test -f file.txt",
		"command -q",
	}

	for _, snippet := range requiredFishSnippets {
		if !strings.Contains(content, snippet) {
			t.Errorf("bash-expert/SKILL.md missing required fish snippet: %q", snippet)
		}
	}
}

func TestRipgrepPatterns_GoFunctionSearch(t *testing.T) {
	content := assets.MustRead("skills/ripgrep/SKILL.md")
	if len(content) == 0 {
		t.Fatal("ripgrep/SKILL.md is empty")
	}

	requiredPatterns := []string{
		`func\s+\([^)]+\)\s+\w+\s*\([^)]*\)`,
		`func \w+`,
		`class \w+(\(\w+\))?:`,
		`def \w+\([^)]*\):`,
	}

	for _, pattern := range requiredPatterns {
		if !strings.Contains(content, pattern) {
			t.Errorf("ripgrep/SKILL.md missing required pattern: %q", pattern)
		}
	}
}
