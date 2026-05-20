package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SkillTier string

const (
	TierFoundation       SkillTier = "foundation"
	TierContextActivated SkillTier = "context_activated"
	TierOnDemand         SkillTier = "on_demand"
)

type SkillEntry struct {
	Name          string
	Path          string
	Tier          SkillTier
	ActivatesWhen string
	CompactLines  int
}

var FoundationSkills = []SkillEntry{
	{Name: "ripgrep", Tier: TierFoundation, CompactLines: 15},
	{Name: "bash-expert", Tier: TierFoundation, CompactLines: 20},
	{Name: "architecture-guardrails", Tier: TierFoundation, CompactLines: 12},
	{Name: "context-guardian", Tier: TierFoundation, CompactLines: 10},
	{Name: "adaptive-reasoning", Tier: TierFoundation, CompactLines: 25},
	{Name: "cognitive-mode", Tier: TierFoundation, CompactLines: 18},
}

var ContextActivatedSkills = []SkillEntry{
	{Name: "go-testing", ActivatesWhen: "*.go files in diff OR task mentions 'test'"},
	{Name: "odoo-development-skill", ActivatesWhen: "__manifest__.py detected"},
	{Name: "branch-pr", ActivatesWhen: "task involves PR creation OR git push"},
	{Name: "work-unit-commits", ActivatesWhen: "sdd-apply running OR task involves commit"},
	{Name: "issue-creation", ActivatesWhen: "task involves creating GitHub/GitLab issue"},
}

// GenerateManifest creates .atl/skill-manifest.yaml based on Tiers
func GenerateManifest(atDir string, extraSkills []SkillEntry) error {
	manifest := buildManifestYAML(extraSkills)
	path := filepath.Join(atDir, "skill-manifest.yaml")
	tmp := path + ".tmp"

	if err := os.WriteFile(tmp, []byte(manifest), 0644); err != nil {
		return fmt.Errorf("write manifest tmp: %w", err)
	}
	return os.Rename(tmp, path)
}

// GenerateFoundationBlock merges compact rules of all Tier 1 skills into one file
func GenerateFoundationBlock(atDir string) error {
	genDir := filepath.Join(atDir, "_generated")
	if err := os.MkdirAll(genDir, 0755); err != nil {
		return fmt.Errorf("create generated dir: %w", err)
	}

	var sections []string
	sections = append(sections, "## Project Foundation Standards [AUTO-GENERATED — do not edit]")
	sections = append(sections, fmt.Sprintf("<!-- generated: %s -->", time.Now().UTC().Format(time.RFC3339)))
	sections = append(sections, "<!-- architect-ai:foundation:start -->")

	for i, skill := range FoundationSkills {
		skillPath := filepath.Join(atDir, "skills", skill.Name, "SKILL.md")
		compact, err := extractCompactRules(skillPath, skill.CompactLines)
		if err != nil {
			compact = fmt.Sprintf("### %d. %s\n[compact rules not available — run skill-registry --refresh]", i+1, skill.Name)
		}
		sections = append(sections, compact)
	}

	sections = append(sections, "<!-- architect-ai:foundation:end -->")

	content := strings.Join(sections, "\n\n")
	outPath := filepath.Join(genDir, "foundation.md")
	tmp := outPath + ".tmp"

	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return fmt.Errorf("write foundation tmp: %w", err)
	}
	return os.Rename(tmp, outPath)
}

func extractCompactRules(skillPath string, maxLines int) (string, error) {
	data, err := os.ReadFile(skillPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	var compact []string
	inCompact := false

	for _, line := range lines {
		if strings.Contains(line, "## Compact Rules") || strings.Contains(line, "## Quick Reference") {
			inCompact = true
			compact = append(compact, line)
			continue
		}
		if inCompact {
			if strings.HasPrefix(line, "## ") && !strings.Contains(line, "Compact") {
				break // next section
			}
			compact = append(compact, line)
			if len(compact) >= maxLines {
				break
			}
		}
	}

	if len(compact) == 0 {
		// Fallback: take first maxLines lines after frontmatter
		inFrontmatter := false
		for _, line := range lines {
			if line == "---" {
				inFrontmatter = !inFrontmatter
				continue
			}
			if !inFrontmatter && len(compact) < maxLines {
				compact = append(compact, line)
			}
		}
	}

	return strings.Join(compact, "\n"), nil
}

func buildManifestYAML(extra []SkillEntry) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# .atl/skill-manifest.yaml\n# AUTO-GENERATED: %s\nversion: \"3.0\"\n\nfoundation:\n", time.Now().UTC().Format(time.RFC3339)))

	for _, s := range FoundationSkills {
		sb.WriteString(fmt.Sprintf("  - name: %s\n    path: \".atl/skills/%s/SKILL.md\"\n", s.Name, s.Name))
	}

	sb.WriteString("\ncontext_activated:\n")
	for _, s := range ContextActivatedSkills {
		sb.WriteString(fmt.Sprintf("  - name: %s\n    path: \".atl/skills/%s/SKILL.md\"\n    activates_when: %q\n", s.Name, s.Name, s.ActivatesWhen))
	}

	sb.WriteString("\non_demand:\n")
	onDemand := []string{"researcher", "solver", "ideator", "generalist", "skill-creator", "mcp-notebooklm-orchestrator"}
	for _, name := range onDemand {
		sb.WriteString(fmt.Sprintf("  - name: %s\n    path: \".atl/skills/%s/SKILL.md\"\n", name, name))
	}

	for _, s := range extra {
		sb.WriteString(fmt.Sprintf("  - name: %s\n    path: %q\n    tier: %s\n", s.Name, s.Path, s.Tier))
	}

	return sb.String()
}
