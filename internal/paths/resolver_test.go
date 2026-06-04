package paths

import (
	"path/filepath"
	"testing"
)

func TestNewContext(t *testing.T) {
	c := New("", false)
	if c.ProjectRoot != "." {
		t.Errorf("expected '.', got %q", c.ProjectRoot)
	}

	c = New("/tmp", true)
	if c.ProjectRoot != "/tmp" {
		t.Errorf("expected '/tmp', got %q", c.ProjectRoot)
	}
}

func TestSkillsDir(t *testing.T) {
	c := New(".", true)
	if got := c.SkillsDir(); got != filepath.Join(".", "internal/assets/skills") {
		t.Errorf("DevMode got %q", got)
	}

	c = New(".", false)
	if got := c.SkillsDir(); got != filepath.Join(".", ".atl/skills") {
		t.Errorf("TargetMode got %q", got)
	}
}

func TestL2SkillGlob(t *testing.T) {
	c := New(".", true)
	if got := c.L2SkillGlob(); got != filepath.Join(".", "internal/assets/skills/*/SKILL.md") {
		t.Errorf("DevMode got %q", got)
	}

	c = New(".", false)
	if got := c.L2SkillGlob(); got != filepath.Join(".", ".agent/skills/*/SKILL.md") {
		t.Errorf("TargetMode got %q", got)
	}
}

func TestRegistryPath(t *testing.T) {
	c := New(".", true)
	if got := c.RegistryPath(); got != "" {
		t.Errorf("DevMode got %q", got)
	}

	c = New(".", false)
	if got := c.RegistryPath(); got != filepath.Join(".", ".atl/skill-registry.md") {
		t.Errorf("TargetMode got %q", got)
	}
}

func TestSddApplySkillPath(t *testing.T) {
	c := New(".", true)
	if got := c.SddApplySkillPath(); got != filepath.Join(".", "internal/assets/skills/sdd-apply/SKILL.md") {
		t.Errorf("DevMode got %q", got)
	}

	c = New(".", false)
	if got := c.SddApplySkillPath(); got != filepath.Join(".", ".agent/skills/sdd-apply/SKILL.md") {
		t.Errorf("TargetMode got %q", got)
	}
}
