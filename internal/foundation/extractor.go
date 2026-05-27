// Package foundation extracts compact rules from skill SKILL.md files and generates
// the merged foundation block used for project-wide standards injection.
package foundation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SkillRef identifies a skill whose compact rules are to be extracted.
type SkillRef struct {
	Name         string
	CompactLines int // max lines of compact rules to extract
}

// ExtractResult holds the extracted content for one skill.
type ExtractResult struct {
	Name    string
	Content string
	Err     error
}

// Extractor reads SKILL.md files and extracts compact rules sections.
type Extractor struct {
	skillsDir string
}

// NewExtractor creates an Extractor that reads skills from the given directory.
func NewExtractor(skillsDir string) *Extractor {
	return &Extractor{skillsDir: skillsDir}
}

// Extract reads one skill's SKILL.md and extracts its compact rules content.
// On read failure the result carries the error and empty content.
func (e *Extractor) Extract(skill SkillRef) ExtractResult {
	path := filepath.Join(e.skillsDir, skill.Name, "SKILL.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return ExtractResult{
			Name: skill.Name,
			Err:  fmt.Errorf("read skill %s SKILL.md: %w", skill.Name, err),
		}
	}

	return ExtractResult{
		Name:    skill.Name,
		Content: extractCompactRules(data, skill.CompactLines),
	}
}

// ExtractAll extracts compact rules for multiple skills in order.
func (e *Extractor) ExtractAll(skills []SkillRef) []ExtractResult {
	results := make([]ExtractResult, len(skills))
	for i, skill := range skills {
		results[i] = e.Extract(skill)
	}
	return results
}

// extractCompactRules extracts the compact rules section from a SKILL.md byte slice.
//
// It first searches for "## Compact Rules" or "## Quick Reference" headings and
// captures content until the next top-level heading or maxLines is reached.
// If no such section is found it falls back to the first maxLines lines after
// the YAML frontmatter.
func extractCompactRules(data []byte, maxLines int) string {
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
			if strings.HasPrefix(line, "## ") &&
				!strings.Contains(line, "Compact Rules") &&
				!strings.Contains(line, "Compact") &&
				!strings.Contains(line, "Quick Reference") {
				break
			}
			compact = append(compact, line)
			if len(compact) >= maxLines {
				break
			}
		}
	}

	if len(compact) == 0 {
		// Fallback: take first maxLines lines after frontmatter.
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

	return strings.Join(compact, "\n")
}
