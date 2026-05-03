package skills

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	indexStart  = "<!-- architect-ai:skills-index:start -->"
	indexEnd    = "<!-- architect-ai:skills-index:end -->"
	indexHeader = `# Architect AI — Agent Skills Index

When working on this project, load the relevant skill(s) BEFORE writing any code.

## How to Use

1. Check the trigger column to find skills that match your current task
2. Load the skill by reading the SKILL.md at the listed path
3. Follow ALL patterns and rules from the loaded skill
4. Multiple skills can apply simultaneously

## Skills

`
)

// SkillEntry is one row in the generated AGENTS.md skills table.
type SkillEntry struct {
	Name        string // frontmatter name: field
	Description string // first sentence of frontmatter description:
	Path        string // relative path to SKILL.md
}

// WriteAgentsIndex writes or refreshes the skills table in AGENTS.md.
// Creates the file if absent. Replaces only the marker block if present.
// Appends to the end if the file exists but has no markers.
func WriteAgentsIndex(projectDir string, entries []SkillEntry) error {
	target := filepath.Join(projectDir, "AGENTS.md")

	table := buildTable(entries)
	block := indexStart + "\n" + table + indexEnd

	existing, err := os.ReadFile(target)
	if errors.Is(err, os.ErrNotExist) {
		return os.WriteFile(target, []byte(indexHeader+block+"\n"), 0o644)
	}
	if err != nil {
		return err
	}

	content := string(existing)
	si := strings.Index(content, indexStart)
	ei := strings.Index(content, indexEnd)

	var out string
	if si >= 0 && ei >= 0 && ei > si {
		out = content[:si] + block + content[ei+len(indexEnd):]
	} else {
		out = strings.TrimRight(content, "\n") + "\n\n" + block + "\n"
	}
	return os.WriteFile(target, []byte(out), 0o644)
}

func buildTable(entries []SkillEntry) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "| Skill | Trigger | Path |\n")
	fmt.Fprintf(&b, "|-------|---------|------|\n")
	for _, e := range entries {
		fmt.Fprintf(&b, "| %s | %s | `%s` |\n", e.Name, e.Description, e.Path)
	}
	b.WriteString("\n")
	return b.String()
}

// ParseFrontmatter extracts name and description from a SKILL.md YAML frontmatter block.
func ParseFrontmatter(content []byte) (name, description string) {
	lines := strings.Split(string(content), "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return "", ""
	}

	inDescription := false
	var descLines []string

	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			break
		}

		if inDescription {
			// If line is indented, it's part of the multiline string
			if strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t") {
				descLines = append(descLines, trimmed)
				continue
			}
			inDescription = false
		}

		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)

		switch k {
		case "name":
			name = strings.Trim(v, `"'`)
		case "description":
			if v == ">" || v == "|" {
				inDescription = true
			} else {
				description = strings.Trim(v, `"'`)
				if idx := strings.IndexAny(description, ".\n"); idx > 0 {
					description = description[:idx]
				}
			}
		}
	}

	if len(descLines) > 0 {
		description = strings.Join(descLines, " ")
		if idx := strings.IndexAny(description, ".\n"); idx > 0 {
			description = description[:idx]
		}
	}

	return
}
