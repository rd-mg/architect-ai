package cli

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rd-mg/architect-ai/internal/agents"
	"github.com/rd-mg/architect-ai/internal/components/filemerge"
)

type skillEntry struct {
	Name         string
	Trigger      string
	CompactRules string
	Path         string
	Origin       string // "user", "project", "overlay", "system", "shared"
	Kind         string // "System", "User", "Project", "Overlay", "SharedRule"
}

type conventionEntry struct {
	File  string
	Path  string
	Notes string
}

type assetEntry struct {
	Name    string
	Type    string
	Overlay string
	Path    string
}

func RunSkillRegistry(args []string, stdout io.Writer) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve current working directory: %w", err)
	}

	fs := flag.NewFlagSet("skill-registry", flag.ContinueOnError)
	fs.SetOutput(ioDiscard{})
	refreshOverlays := fs.Bool("refresh-overlays", false, "refresh project-local overlays before regenerating the registry")
	enterprisePath := fs.String("enterprise-repo", "", "local Odoo enterprise repository path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *refreshOverlays {
		// If explicit refresh is requested, we bypass the "ensure" check and force a bootstrap.
		_, err := BootstrapProjectLocalOverlays(projectRoot, true, *enterprisePath)
		if err != nil {
			return err
		}
	}

	result, err := EnsureProjectRegistryReady(projectRoot)
	if err != nil {
		return err
	}

	if result.IsOdooProject {
		for _, overlay := range result.Overlays {
			action := "reused"
			switch result.Actions[overlay.Name] {
			case "installed":
				action = "bootstrapped"
			case "refreshed":
				action = "refreshed"
			}
			_, _ = fmt.Fprintf(stdout, "%s overlay %q for this Odoo project\n", strings.Title(action), overlay.Name)
		}
		if len(result.Versions) > 0 {
			_, _ = fmt.Fprintf(stdout, "Regenerated .atl/skill-registry.md for Odoo versions: %s\n", formatVersionSet(result.Versions))
		} else {
			_, _ = fmt.Fprintln(stdout, "Regenerated .atl/skill-registry.md, but no Odoo major version could be extracted from detected __manifest__.py files yet.")
		}
	} else {
		_, _ = fmt.Fprintln(stdout, "Skill registry regenerated successfully at .atl/skill-registry.md")
	}

	return nil
}

func WriteLocalSkillRegistry(projectRoot string) error {
	homeDir, err := osUserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve user home directory: %w", err)
	}

	// 1. Collect all entries
	var skills []skillEntry
	var conventions []conventionEntry
	var assets []assetEntry

	// User skills
	userSkills, err := collectUserSkills(homeDir)
	if err == nil {
		skills = append(skills, userSkills...)
	}

	// Project skills
	projectSkills, err := collectProjectSkills(projectRoot)
	if err == nil {
		skills = append(skills, projectSkills...)
	}

	// Overlay content
	overlaySkills, overlayAssets, err := collectOverlayContent(projectRoot)
	if err == nil {
		skills = append(skills, overlaySkills...)
		assets = append(assets, overlayAssets...)
	}

	// Deduplicate skills by name: project/overlay overrides user
	skills = deduplicateSkills(skills)

	// Project conventions
	conventions = collectProjectConventions(projectRoot)

	// 2. Build Markdown
	markdown := buildRegistryMarkdown(projectRoot, skills, conventions, assets)

	// 3. Write to file
	registryPath := filepath.Join(projectRoot, ".atl", "skill-registry.md")
	if err := os.MkdirAll(filepath.Dir(registryPath), 0o755); err != nil {
		return fmt.Errorf("create local registry directory: %w", err)
	}
	_, err = filemerge.WriteFileAtomic(registryPath, []byte(markdown), 0o644)
	if err != nil {
		return fmt.Errorf("write local skill registry: %w", err)
	}

	return nil
}

func collectUserSkills(homeDir string) ([]skillEntry, error) {
	reg, err := agents.NewDefaultRegistry()
	if err != nil {
		return nil, err
	}

	var entries []skillEntry
	for _, id := range reg.SupportedAgents() {
		adapter, ok := reg.Get(id)
		if !ok || !adapter.SupportsSkills() {
			continue
		}

		dir := adapter.SkillsDir(homeDir)
		if dir == "" {
			continue
		}

		entries = append(entries, scanSkillsDir(dir, "user")...)
	}
	return entries, nil
}

func collectProjectSkills(projectRoot string) ([]skillEntry, error) {
	var entries []skillEntry

	// standard exclusions
	excluded := map[string]bool{
		".git":         true,
		"node_modules": true,
		"vendor":       true,
		".terraform":   true,
		".venv":        true,
		"__pycache__":  true,
		"dist":         true,
		"build":        true,
		".tmp":         true,
		".atl":         true, // Skip .atl to avoid self-referencing registry or redundant overlay scans
		"testdata":     true,
		"tests":        true,
		"e2e":          true,
	}

	err := filepath.WalkDir(projectRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // soft skip
		}

		if d.IsDir() {
			if excluded[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Any SKILL.md is a potential skill container
		if d.Name() == "SKILL.md" {
			info := parseSkillFile(path)
			if info.Name == "" {
				info.Name = filepath.Base(filepath.Dir(path))
			}
			info.Path = path
			info.Origin = "project"

			// Standard classification
			if info.Name == "_shared" {
				info.Kind = "SharedRule"
			} else if strings.HasPrefix(info.Name, "sdd-") || info.Name == "skill-registry" {
				info.Kind = "System"
			} else {
				info.Kind = "Project"
			}

			entries = append(entries, info)
		}

		return nil
	})

	return entries, err
}

func collectOverlayContent(projectRoot string) ([]skillEntry, []assetEntry, error) {
	overlaysRoot := filepath.Join(projectRoot, ".atl", "overlays")
	if _, err := os.Stat(overlaysRoot); err != nil {
		return nil, nil, nil
	}

	dirs, err := os.ReadDir(overlaysRoot)
	if err != nil {
		return nil, nil, err
	}

	// For Odoo version filtering if applicable
	odooVersions, isOdoo, _ := detectOdooMajorVersions(projectRoot)

	var skills []skillEntry
	var assets []assetEntry
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		overlayName := d.Name()
		overlayRoot := filepath.Join(overlaysRoot, overlayName)

		manifestPath := filepath.Join(overlayRoot, "manifest.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}
		var manifest OverlayManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			continue
		}
		if manifest.ActivationState != overlayActivationActive {
			continue
		}

		// Skills
		overlaySkillDir := filepath.Join(overlayRoot, "skills")
		if _, err := os.Stat(overlaySkillDir); err == nil {
			dirs, err := os.ReadDir(overlaySkillDir)
			if err == nil {
				for _, entry := range dirs {
					if isOdoo && len(odooVersions) > 0 && !matchesOverlaySkillVersion(entry.Name(), odooVersions) {
						continue
					}

					skillPath := filepath.Join(overlaySkillDir, entry.Name(), "SKILL.md")
					if _, err := os.Stat(skillPath); err != nil {
						continue
					}

					info := parseSkillFile(skillPath)
					if info.Name == "" {
						info.Name = entry.Name()
					}
					info.Path = skillPath
					info.Origin = "overlay"
					
					if entry.Name() == "_shared" {
						info.Kind = "SharedRule"
					} else if strings.HasPrefix(entry.Name(), "sdd-") || entry.Name() == "skill-registry" {
						info.Kind = "System"
					} else {
						info.Kind = "Overlay"
					}
					
					skills = append(skills, info)
				}
			}
		}

		// Agents
		agentDir := filepath.Join(overlayRoot, "agents")
		if entries, err := os.ReadDir(agentDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
					assets = append(assets, assetEntry{
						Name:    strings.TrimSuffix(e.Name(), ".md"),
						Type:    "agent",
						Overlay: overlayName,
						Path:    filepath.Join(agentDir, e.Name()),
					})
				}
			}
		}

		// Patterns, Instructions, Prompts, Scripts, Assets (Filtered by Odoo version if needed)
		for _, sub := range []string{"patterns", "instructions", "prompts", "scripts", "assets"} {
			dir := filepath.Join(overlayRoot, sub)
			if entries, err := os.ReadDir(dir); err == nil {
				for _, e := range entries {
					if e.IsDir() {
						continue
					}

					// Apply version filtering only to patterns, scripts, and assets for Odoo projects.
					// Instructions and prompts are usually version-agnostic.
					if isOdoo && (sub == "patterns" || sub == "scripts" || sub == "assets") {
						if !matchesOdooVersion(filepath.Join(dir, e.Name()), odooVersions) {
							continue
						}
					}

					entryType := strings.TrimSuffix(sub, "s")

					assets = append(assets, assetEntry{
						Name:    strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())),
						Type:    entryType,
						Overlay: overlayName,
						Path:    filepath.Join(dir, e.Name()),
					})
				}
			}
		}
	}
	return skills, assets, nil
}

func scanSkillsDir(dir string, origin string) []skillEntry {
	if _, err := os.Stat(dir); err != nil {
		return nil
	}

	subdirs, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var entries []skillEntry
	for _, d := range subdirs {
		if !d.IsDir() {
			continue
		}

		kind := "User"
		if d.Name() == "_shared" {
			kind = "SharedRule"
		} else if strings.HasPrefix(d.Name(), "sdd-") || d.Name() == "skill-registry" {
			kind = "System"
		}

		skillPath := filepath.Join(dir, d.Name(), "SKILL.md")
		if _, err := os.Stat(skillPath); err == nil {
			info := parseSkillFile(skillPath)
			if info.Name == "" {
				info.Name = d.Name()
			}
			info.Path = skillPath
			info.Origin = origin
			info.Kind = kind
			entries = append(entries, info)
		}
	}
	return entries
}

func parseSkillFile(path string) skillEntry {
	f, err := os.Open(path)
	if err != nil {
		return skillEntry{}
	}
	defer f.Close()

	var entry skillEntry
	scanner := bufio.NewScanner(f)

	inFrontmatter := false
	frontmatterDone := false

	var descriptionBuffer strings.Builder
	inDescription := false

	var rulesLines []string
	inRules := false

	lineCount := 0
	for scanner.Scan() {
		lineCount++
		if lineCount > 200 { // Safety guard
			break
		}
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Simple frontmatter parsing
		if trimmedLine == "---" {
			if !inFrontmatter && !frontmatterDone {
				inFrontmatter = true
			} else if inFrontmatter {
				inFrontmatter = false
				frontmatterDone = true
				// Extract Trigger from description buffer if not found as separate field
				if entry.Trigger == "" && descriptionBuffer.Len() > 0 {
					descText := descriptionBuffer.String()
					if idx := strings.Index(descText, "Trigger:"); idx != -1 {
						entry.Trigger = strings.TrimSpace(descText[idx+len("Trigger:"):])
					}
				}
			}
			continue
		}

		if inFrontmatter {
			if strings.HasPrefix(trimmedLine, "name:") {
				entry.Name = strings.TrimSpace(strings.TrimPrefix(trimmedLine, "name:"))
			} else if strings.HasPrefix(trimmedLine, "Trigger:") {
				entry.Trigger = strings.TrimSpace(strings.TrimPrefix(trimmedLine, "Trigger:"))
			} else if strings.HasPrefix(trimmedLine, "description:") {
				// Handle multiline description with > operator
				descValue := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "description:"))
				if strings.HasPrefix(descValue, ">") {
					inDescription = true
					// Check if there's text after > on the same line
					remaining := strings.TrimSpace(strings.TrimPrefix(descValue, ">"))
					if remaining != "" {
						descriptionBuffer.WriteString(remaining)
						descriptionBuffer.WriteString(" ")
					}
				} else {
					// Single-line description
					descriptionBuffer.WriteString(descValue)
					descriptionBuffer.WriteString(" ")
				}
			} else if inDescription {
				// Continue reading multiline description
				if trimmedLine == "" || (!strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "\t")) {
					// Description block ended
					inDescription = false
				} else {
					descriptionBuffer.WriteString(trimmedLine)
					descriptionBuffer.WriteString(" ")
				}
			}
			continue
		}

		// Look for name in first H1 if not in frontmatter
		if entry.Name == "" && strings.HasPrefix(trimmedLine, "# ") {
			entry.Name = strings.TrimPrefix(trimmedLine, "# ")
		}

		// Extract compact rules
		if strings.HasPrefix(trimmedLine, "## ") {
			lower := strings.ToLower(trimmedLine)
			if strings.Contains(lower, "rules") || strings.Contains(lower, "patterns") || strings.Contains(lower, "critical") {
				inRules = true
				continue
			} else {
				inRules = false
			}
		}

		if inRules && len(rulesLines) < 15 {
			if trimmedLine != "" {
				rulesLines = append(rulesLines, line)
			}
		}
	}

	entry.CompactRules = strings.Join(rulesLines, "\n")
	return entry
}

func deduplicateSkills(skills []skillEntry) []skillEntry {
	m := make(map[string]skillEntry)
	var names []string

	for _, s := range skills {
		existing, exists := m[s.Name]
		if !exists {
			m[s.Name] = s
			names = append(names, s.Name)
			continue
		}

		// project > overlay > user > system > shared
		priority := map[string]int{"project": 5, "overlay": 4, "user": 3, "system": 2, "shared": 1}
		if priority[s.Origin] > priority[existing.Origin] {
			m[s.Name] = s
		}
	}

	sort.Strings(names)
	var result []skillEntry
	for _, name := range names {
		result = append(result, m[name])
	}
	return result
}

func collectProjectConventions(projectRoot string) []conventionEntry {
	files := []string{"agents.md", "AGENTS.md", "CLAUDE.md", ".cursorrules", "GEMINI.md", "copilot-instructions.md"}
	var entries []conventionEntry
	for _, f := range files {
		path := filepath.Join(projectRoot, f)
		if _, err := os.Stat(path); err == nil {
			entries = append(entries, conventionEntry{
				File: f,
				Path: path,
			})
		}
	}
	return entries
}

func buildRegistryMarkdown(projectRoot string, skills []skillEntry, conventions []conventionEntry, assets []assetEntry) string {
	var b strings.Builder
	b.WriteString("# Skill Registry\n\n")
	b.WriteString("**Delegator use only.** Any agent that launches sub-agents reads this registry to resolve compact rules, then injects them directly into sub-agent prompts. Sub-agents do NOT read this registry or individual SKILL.md files.\n\n")

	// Group skills by Kind
	kinds := []string{"System", "SharedRule", "Project", "Overlay", "User"}
	skillsByKind := make(map[string][]skillEntry)
	for _, s := range skills {
		skillsByKind[s.Kind] = append(skillsByKind[s.Kind], s)
	}

	for _, kind := range kinds {
		entries := skillsByKind[kind]
		if len(entries) == 0 {
			continue
		}

		b.WriteString(fmt.Sprintf("## %s Skills\n\n", kind))
		b.WriteString("| Trigger | Skill | Path |\n")
		b.WriteString("|---------|-------|------|\n")
		for _, s := range entries {
			relPath := s.Path
			if rel, err := filepath.Rel(projectRoot, s.Path); err == nil && !strings.HasPrefix(rel, "..") {
				relPath = rel
			}
			b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", escapeTable(s.Trigger), s.Name, filepath.ToSlash(relPath)))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Compact Rules\n\n")
	b.WriteString("Pre-digested rules per skill. Delegators copy matching blocks into sub-agent prompts as `## Project Standards (auto-resolved)`.\n\n")
	for _, kind := range kinds {
		entries := skillsByKind[kind]
		for _, s := range entries {
			b.WriteString(fmt.Sprintf("### %s\n%s\n\n", s.Name, s.CompactRules))
		}
	}

	b.WriteString("## Project Conventions\n\n")
	b.WriteString("| File | Path | Notes |\n")
	b.WriteString("|------|------|-------|\n")
	for _, c := range conventions {
		relPath := c.Path
		if rel, err := filepath.Rel(projectRoot, c.Path); err == nil && !strings.HasPrefix(rel, "..") {
			relPath = rel
		}
		b.WriteString(fmt.Sprintf("| %s | %s | |\n", c.File, filepath.ToSlash(relPath)))
	}
	b.WriteString("\n")

	if len(assets) > 0 {
		b.WriteString("## Specialist Overlay Resources\n\n")
		b.WriteString("| Name | Type | Overlay | Path |\n")
		b.WriteString("|------|------|---------|------|\n")
		sort.Slice(assets, func(i, j int) bool {
			if assets[i].Overlay == assets[j].Overlay {
				return assets[i].Name < assets[j].Name
			}
			return assets[i].Overlay < assets[j].Overlay
		})
		for _, a := range assets {
			relPath := a.Path
			if rel, err := filepath.Rel(projectRoot, a.Path); err == nil && !strings.HasPrefix(rel, "..") {
				relPath = rel
			}
			b.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", a.Name, a.Type, a.Overlay, filepath.ToSlash(relPath)))
		}
	}

	return b.String()
}

func escapeTable(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}

// EnsureProjectRegistryReady performs the base initialization of a project for ATL/SDD.
// It creates the .atl directory, bootstraps project-local overlays, builds the skill registry,
// and ensures core project conventions (AGENTS.md, GEMINI.md) are present.
func EnsureProjectRegistryReady(projectRoot string) (OverlayBootstrapResult, error) {
	atlDir := filepath.Join(projectRoot, ".atl")
	if err := os.MkdirAll(atlDir, 0o755); err != nil {
		return OverlayBootstrapResult{}, fmt.Errorf("create .atl directory: %w", err)
	}

	// 1. Bootstrap project-local overlays (Odoo, etc.)
	// We pass refresh=false by default for the "ensure" check.
	result, err := BootstrapProjectLocalOverlays(projectRoot, false, "")
	if err != nil {
		return OverlayBootstrapResult{}, fmt.Errorf("bootstrap local overlays: %w", err)
	}

	// 2. Build/Update the registry markdown
	if err := WriteLocalSkillRegistry(projectRoot); err != nil {
		return OverlayBootstrapResult{}, fmt.Errorf("write skill registry: %w", err)
	}

	// 3. Bootstrap core project conventions (AGENTS.md, GEMINI.md)
	if err := bootstrapProjectConventions(projectRoot); err != nil {
		return OverlayBootstrapResult{}, fmt.Errorf("bootstrap project conventions: %w", err)
	}

	return result, nil
}

func bootstrapProjectConventions(projectRoot string) error {
	absProjectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return fmt.Errorf("resolve absolute project root: %w", err)
	}

	conventions := []struct {
		filename string
		content  string
	}{
		{"AGENTS.md", agentsTemplate},
		{"GEMINI.md", geminiTemplate},
	}

	for _, conv := range conventions {
		path := filepath.Join(absProjectRoot, conv.filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			_, err = filemerge.WriteFileAtomicWithOptions(path, []byte(conv.content), filemerge.WriteOptions{
				Perm:  0o644,
				Force: true,
			})
			if err != nil {
				return fmt.Errorf("create %s: %w", conv.filename, err)
			}
		}
	}

	return nil
}

const agentsTemplate = `# Project Agents

This file documents the specialized agents allowed to operate in this repository.

## Architect
- **Role**: Technical lead, design patterns, and architectural integrity.
- **Rules**: Must follow SOLID and Hexagonal patterns as defined in GEMINI.md.

## Developer
- **Role**: Implementation and bug fixing.
- **Rules**: Must write tests before implementation (Strict TDD).
`

const geminiTemplate = `# Project Rules (Gemini)

This file defines the technical mandates for AI agents working in this repository.

## General
- Use conventional commits.
- Always use 'rg' for searching.
- Never improvise architecture.

## Architecture
- Prefer composition over inheritance.
- External dependencies must be wrapped in adapters.
`

// EnsureSDDReady validates that the project is ready for SDD operations.
// It checks for the existence of the CLI bootstrap marker (.atl/state/bootstrap.json).
// If missing, it fails with an instruction to run sdd-init.
func EnsureSDDReady(projectRoot string) error {
	bootstrapPath := filepath.Join(projectRoot, ".atl", "state", "bootstrap.json")

	if _, err := os.Stat(bootstrapPath); err != nil {
		return fmt.Errorf("sdd guard: project not bootstrapped. Please run 'architect-ai sdd-init' first.")
	}

	return nil
}