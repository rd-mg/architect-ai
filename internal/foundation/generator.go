package foundation

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// FoundationBlock holds the assembled foundation block content.
type FoundationBlock struct {
	Content    string
	SkillNames []string
}

// Generator assembles foundation markdown blocks from extracted skill rules.
type Generator struct{}

// NewGenerator creates a new Generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate builds the full foundation markdown block from a set of extract results.
// Each result is numbered in order. Results with an empty Content (failed or missing
// extraction) produce a "not available" placeholder.
func (g *Generator) Generate(results []ExtractResult) FoundationBlock {
	var sb strings.Builder
	names := make([]string, 0, len(results))

	sb.WriteString("## Project Foundation Standards [AUTO-GENERATED — do not edit]\n")
	sb.WriteString(fmt.Sprintf("<!-- generated: %s -->\n", time.Now().UTC().Format(time.RFC3339)))
	sb.WriteString("<!-- architect-ai:foundation:start -->\n")

	for i, r := range results {
		names = append(names, r.Name)
		content := r.Content
		if content == "" {
			content = fmt.Sprintf("### %d. %s\n[compact rules not available]", i+1, r.Name)
		}
		sb.WriteString("\n")
		sb.WriteString(content)
	}

	sb.WriteString("\n<!-- architect-ai:foundation:end -->\n")

	return FoundationBlock{
		Content:    sb.String(),
		SkillNames: names,
	}
}

// WriteToFile writes the foundation block content to the given path atomically:
// it writes to a .tmp sibling first, then renames to the final path.
// Parent directories are created as needed.
func (b FoundationBlock) WriteToFile(path string) error {
	dir := path + "."
	// Use the parent directory for MkdirAll; path could be nested.
	parentDir := path
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		parentDir = path[:idx]
	}
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return fmt.Errorf("create parent directory for %s: %w", path, err)
	}

	// Remove dir placeholder if it was set (path without a slash means CWD).
	_ = dir

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(b.Content), 0o644); err != nil {
		return fmt.Errorf("write foundation block tmp: %w", err)
	}
	return os.Rename(tmp, path)
}
