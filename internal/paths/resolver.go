package paths

import (
	"path/filepath"
)

// Context encapsulates the path resolution strategy.
type Context struct {
	ProjectRoot string
	IsDevMode   bool
}

// New returns a new path context. If devMode is true, it resolves
// paths relative to the architect-ai development repository. Otherwise,
// it resolves paths for a target project deployment.
func New(projectRoot string, devMode bool) Context {
	if projectRoot == "" {
		projectRoot = "."
	}
	return Context{
		ProjectRoot: projectRoot,
		IsDevMode:   devMode,
	}
}

// SkillsDir returns the path to the skills directory.
// DevMode: internal/assets/skills
// TargetMode: .atl/skills
func (c Context) SkillsDir() string {
	if c.IsDevMode {
		return filepath.Join(c.ProjectRoot, "internal/assets/skills")
	}
	return filepath.Join(c.ProjectRoot, ".atl/skills")
}

// L2SkillGlob returns the glob pattern for matching L2 skills.
// DevMode: internal/assets/skills/*/SKILL.md
// TargetMode: .agent/skills/*/SKILL.md
func (c Context) L2SkillGlob() string {
	if c.IsDevMode {
		return filepath.Join(c.ProjectRoot, "internal/assets/skills/*/SKILL.md")
	}
	return filepath.Join(c.ProjectRoot, ".agent/skills/*/SKILL.md")
}

// RegistryPath returns the path to the skill registry.
// DevMode: empty string (not applicable)
// TargetMode: .atl/skill-registry.md
func (c Context) RegistryPath() string {
	if c.IsDevMode {
		return ""
	}
	return filepath.Join(c.ProjectRoot, ".atl/skill-registry.md")
}

// FoundationPath returns the path to the foundation markdown file.
func (c Context) FoundationPath() string {
	return filepath.Join(c.ProjectRoot, ".atl/_generated/foundation.md")
}

// SddApplySkillPath returns the path to the sdd-apply skill.
func (c Context) SddApplySkillPath() string {
	if c.IsDevMode {
		return filepath.Join(c.ProjectRoot, "internal/assets/skills/sdd-apply/SKILL.md")
	}
	return filepath.Join(c.ProjectRoot, ".agent/skills/sdd-apply/SKILL.md")
}
