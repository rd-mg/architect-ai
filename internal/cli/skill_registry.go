package cli

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rd-mg/architect-ai/internal/agents"
	"github.com/rd-mg/architect-ai/internal/components/filemerge"
	"github.com/rd-mg/architect-ai/internal/components/skills"
	skillslib "github.com/rd-mg/architect-ai/internal/skills"
	"github.com/rd-mg/architect-ai/internal/scope"
	"golang.org/x/sync/errgroup"
)

type skillEntry struct {
	Name         string
	Trigger      string
	CompactRules string
	Path         string
	Origin       string // "user", "project", "overlay", "system", "shared", "community"
	Source       string // "project", "community", "overlay/{name}", "builtin"
	Kind         string // "System", "User", "Project", "Overlay", "SharedRule", "Community"
	SHA256       string
	Deprecated   bool
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

// registryVersionMarker identifies the registry format version.
const registryVersionMarker = "<!-- architect-ai:registry:version:2 -->"

// buildQuickIndex creates a machine-readable Quick Index section that allows
// agents to resolve skills by trigger without reading the full registry.
func buildQuickIndex(skillsByKind map[string][]skillEntry) string {
	var b strings.Builder
	b.WriteString("## Quick Index\n\n")
	b.WriteString("| Skill | Kind | Trigger | Anchor |\n")
	b.WriteString("|-------|------|---------|--------|\n")
	for _, kind := range []string{"System", "SharedRule", "Project", "Overlay", "User", "Community"} {
		for _, s := range skillsByKind[kind] {
			anchor := "#skill-" + strings.ToLower(strings.ReplaceAll(s.Name, " ", "-"))
			b.WriteString(fmt.Sprintf("| %s | %s | %s | [link](%s) |\n", s.Name, kind, escapeTable(s.Trigger), anchor))
		}
	}
	return b.String()
}

// buildKeywordIndex builds a JSON keyword index block embedded in an HTML comment.
// The Dynamic Context Assembler reads this to filter skills by phase/task keywords
// without parsing the full markdown registry.
func buildKeywordIndex(allSkills []skillEntry) string {
	var b strings.Builder
	b.WriteString("<!-- SKILL-REGISTRY-INDEX-V4\n")
	b.WriteString("Format: skill_id: [keyword1, keyword2, ...]\n")
	b.WriteString("Used by: Dynamic Context Assembler for keyword-filtered injection\n")

	index := make(map[string][]string)
	for _, s := range allSkills {
		keywords := extractKeywords(s.Trigger)
		index[s.Name] = keywords
	}

	indexJSON, _ := json.MarshalIndent(index, "", "  ")
	b.Write(indexJSON)
	b.WriteString("\nSKILL-REGISTRY-INDEX-END -->")
	return b.String()
}

// extractKeywords splits a trigger string into individual keywords.
func extractKeywords(trigger string) []string {
	if trigger == "" {
		return nil
	}
	parts := strings.FieldsFunc(trigger, func(r rune) bool {
		return r == ',' || r == ';' || r == '/' || r == '|'
	})
	var keywords []string
	seen := make(map[string]bool)
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		if p == "" || seen[p] {
			continue
		}
		// Split whitespace-separated words within each part
		words := strings.Fields(p)
		for _, w := range words {
			w = strings.Trim(w, "\"'")
			if w == "" || seen[w] || len(w) < 2 {
				continue
			}
			seen[w] = true
			keywords = append(keywords, w)
		}
	}
	return keywords
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
	force := fs.Bool("force", false, "force regeneration even if SHAs are unchanged")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *refreshOverlays {
		_, err := BootstrapProjectLocalOverlays(projectRoot, true, *enterprisePath)
		if err != nil {
			return err
		}
	}

	result, err := EnsureProjectRegistryReady(projectRoot, *force || os.Getenv("ARCHITECT_AI_FORCE_REGISTRY") == "1")
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

func WriteLocalSkillRegistry(projectRoot string, force bool) error {
	homeDir, err := osUserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve user home directory: %w", err)
	}

	// 1. Collect all entries in parallel using errgroup
	type collectionResult struct {
		skills []skillEntry
		assets []assetEntry
		err    error
		origin string
	}

	results := make([]collectionResult, 4)
	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		skills, err := collectUserSkills(homeDir)
		results[0] = collectionResult{skills: skills, err: err, origin: "user"}
		return nil // non-fatal: log and continue
	})

	g.Go(func() error {
		skills, err := collectProjectSkills(projectRoot)
		results[1] = collectionResult{skills: skills, err: err, origin: "project"}
		return nil
	})

	g.Go(func() error {
		skills, assets, err := collectOverlayContent(projectRoot)
		results[2] = collectionResult{skills: skills, assets: assets, err: err, origin: "overlay"}
		return nil
	})

	g.Go(func() error {
		skills, err := collectCommunitySkills(projectRoot, homeDir)
		results[3] = collectionResult{skills: skills, err: err, origin: "community"}
		return nil
	})

	_ = g.Wait()

	// Fan-in: merge results (now safe — all goroutines done)
	var allSkills []skillEntry
	var assets []assetEntry
	for _, r := range results {
		if r.err == nil {
			allSkills = append(allSkills, r.skills...)
			assets = append(assets, r.assets...)
		}
	}

	// Deduplicate skills by name: project/overlay overrides user
	allSkills = deduplicateSkills(allSkills)

	// Hash-based invalidation: compare current SHAs with existing registry.
	if !force {
		registryPath := filepath.Join(projectRoot, ".atl", "skill-registry.md")
		oldHashes := parseExistingRegistryHashes(registryPath)
		allMatch := len(oldHashes) > 0
		for _, s := range allSkills {
			if s.SHA256 == "" {
				continue
			}
			oldHash, exists := oldHashes[s.Name]
			if !exists || oldHash != s.SHA256 {
				allMatch = false
				break
			}
		}
		if allMatch {
			return nil // No changes — skip writing.
		}
	}

	// Group skills by Kind
	skillsByKind := make(map[string][]skillEntry)
	for _, s := range allSkills {
		skillsByKind[s.Kind] = append(skillsByKind[s.Kind], s)
	}

	// Project conventions
	conventions := collectProjectConventions(projectRoot)

	// 2. Build Markdown sections with v2 section markers
	registryPath := filepath.Join(projectRoot, ".atl", "skill-registry.md")
	existingContent, _ := os.ReadFile(registryPath)
	content := string(existingContent)
	if content == "" {
		content = registryVersionMarker + "\n\n# Skill Registry\n\n**Delegator use only.** Any agent that launches sub-agents reads this registry to resolve compact rules, then injects them directly into sub-agent prompts. Sub-agents do NOT read this registry or individual SKILL.md files.\n"
	}

	// Keyword Index — machine-readable JSON block for Dynamic Context Assembler
	content = filemerge.InjectMarkdownSection(content, "registry:keyword-index", buildKeywordIndex(allSkills))

	// Quick Index — machine-readable section for agent resolution
	content = filemerge.InjectMarkdownSection(content, "registry:index", buildQuickIndex(skillsByKind))

	kinds := []string{"System", "SharedRule", "Project", "Overlay", "User", "Community"}
	for _, kind := range kinds {
		sectionID := "registry:" + strings.ToLower(kind)
		entries := skillsByKind[kind]
		
		var b strings.Builder
		if len(entries) > 0 {
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
		}
		content = filemerge.InjectMarkdownSection(content, sectionID, b.String())
	}

	// Compact Rules
	var rulesBuilder strings.Builder
	rulesBuilder.WriteString("## Compact Rules\n\nPre-digested rules per skill. Delegators copy matching blocks into sub-agent prompts as `## Project Standards (auto-resolved)`.\n\n")
	for _, kind := range kinds {
		entries := skillsByKind[kind]
		for _, s := range entries {
			compact := strings.TrimSpace(s.CompactRules)
			if compact == "" {
				continue
			}
			rulesBuilder.WriteString(fmt.Sprintf("### %s\n%s\n\n", s.Name, compact))
		}
	}
	content = filemerge.InjectMarkdownSection(content, "registry:compact-rules", rulesBuilder.String())

	// Registry Hashes — for hash-based invalidation on next run.
	var hashesBuilder strings.Builder
	hashesBuilder.WriteString("## Registry Hashes\n\n")
	hashesBuilder.WriteString("| Skill | SHA256 |\n")
	hashesBuilder.WriteString("|-------|--------|\n")
	for _, s := range allSkills {
		if s.SHA256 != "" {
			hashesBuilder.WriteString(fmt.Sprintf("| %s | %s |\n", s.Name, s.SHA256))
		}
	}
	content = filemerge.InjectMarkdownSection(content, "registry:hashes", hashesBuilder.String())

	// Project Conventions
	var convBuilder strings.Builder
	convBuilder.WriteString("## Project Conventions\n\n")
	convBuilder.WriteString("| File | Path | Notes |\n")
	convBuilder.WriteString("|------|------|-------|\n")
	for _, c := range conventions {
		relPath := c.Path
		if rel, err := filepath.Rel(projectRoot, c.Path); err == nil && !strings.HasPrefix(rel, "..") {
			relPath = rel
		}
		convBuilder.WriteString(fmt.Sprintf("| %s | %s | |\n", c.File, filepath.ToSlash(relPath)))
	}
	content = filemerge.InjectMarkdownSection(content, "registry:conventions", convBuilder.String())

	// Specialist Overlay Resources
	var assetBuilder strings.Builder
	if len(assets) > 0 {
		assetBuilder.WriteString("## Specialist Overlay Resources\n\n")
		assetBuilder.WriteString("| Name | Type | Overlay | Path |\n")
		assetBuilder.WriteString("|------|------|---------|------|\n")
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
			assetBuilder.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", a.Name, a.Type, a.Overlay, filepath.ToSlash(relPath)))
		}
	}
	content = filemerge.InjectMarkdownSection(content, "registry:specialist-resources", assetBuilder.String())

	// Research Status (NotebookLM)
	var rb strings.Builder
	rb.WriteString("## Research Status\n\n")
	status, count, _ := probeNotebookLMState()
	rb.WriteString(fmt.Sprintf("- **NotebookLM**: %s", status))
	if count > 0 {
		rb.WriteString(fmt.Sprintf(" (%d notebooks found)", count))
	}
	rb.WriteString("\n")
	content = filemerge.InjectMarkdownSection(content, "registry:research-status", rb.String())

	// 3. Write to file
	if err := os.MkdirAll(filepath.Dir(registryPath), 0o755); err != nil {
		return fmt.Errorf("create local registry directory: %w", err)
	}
	_, err = filemerge.WriteFileAtomic(registryPath, []byte(content), 0o644)
	if err != nil {
		return fmt.Errorf("write local skill registry: %w", err)
	}

	// 4. Update AGENTS.md skills index
	var indexEntries []skills.SkillEntry
	for _, s := range allSkills {
		// Only include high-signal skills in the index
		if s.Kind == "System" || s.Kind == "Project" || s.Kind == "Overlay" || s.Kind == "Community" {
			relPath := s.Path
			if rel, err := filepath.Rel(projectRoot, s.Path); err == nil && !strings.HasPrefix(rel, "..") {
				relPath = rel
			}
			indexEntries = append(indexEntries, skills.SkillEntry{
				Name:        s.Name,
				Description: s.Trigger, // Use Trigger as a proxy for concise description
				Path:        filepath.ToSlash(relPath),
			})
		}
	}
	if err := skills.WriteAgentsIndex(projectRoot, indexEntries); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not update AGENTS.md skills index: %v\n", err)
	}

	return nil
}

func collectCommunitySkills(projectRoot, homeDir string) ([]skillEntry, error) {
	lockfile, err := skillslib.LoadLockfile(projectRoot)
	if err != nil {
		lockfile = skillslib.SkillManifest{}
	}

	cmEntries, err := skillslib.LoadCommunityManifestAsEntries(homeDir)
	if err != nil {
		cmEntries = nil
	}

	idSet := make(map[string]bool)
	var lookups []struct {
		id     string
		source string
		sha    string
	}

	for _, e := range lockfile.Skills {
		if e.Source == "community" && !idSet[e.ID] {
			idSet[e.ID] = true
			lookups = append(lookups, struct {
				id     string
				source string
				sha    string
			}{e.ID, e.Source, e.SHA})
		}
	}
	for _, e := range cmEntries {
		if !idSet[e.ID] {
			idSet[e.ID] = true
			lookups = append(lookups, struct {
				id     string
				source string
				sha    string
			}{e.ID, "community", e.SHA})
		}
	}

	if len(lookups) == 0 {
		return nil, nil
	}

	reg, err := agents.NewDefaultRegistry()
	if err != nil {
		return nil, nil
	}

	var entries []skillEntry
	for _, l := range lookups {
		found := false
		for _, agentID := range reg.SupportedAgents() {
			adapter, ok := reg.Get(agentID)
			if !ok || !adapter.SupportsSkills() {
				continue
			}
			dir := adapter.SkillsDir(homeDir)
			if dir == "" {
				continue
			}
			skillPath := filepath.Join(dir, l.id, "SKILL.md")
			if _, err := os.Stat(skillPath); err == nil {
				info := parseSkillFile(skillPath)
				if info.Name == "" {
					info.Name = l.id
				}
				info.Path = skillPath
				info.Origin = "community"
				info.Source = "community"
				info.Kind = "Community"
				entries = append(entries, info)
				found = true
				break
			}
		}
		if !found {
			entries = append(entries, skillEntry{
				Name:   l.id,
				Origin: "community",
				Source: "community",
				Kind:   "Community",
				Path:   "(not downloaded — run architect-ai skills update)",
				Trigger: fmt.Sprintf("Community skill %q (not yet downloaded)", l.id),
			})
		}
	}
	return entries, nil
}

// MergeLockfileIntoRegistry merges lockfile-managed community skills into registry scan results.
func MergeLockfileIntoRegistry(projectRoot string, existing []skillEntry) []skillEntry {
	homeDir, err := osUserHomeDir()
	if err != nil {
		return existing
	}

	community, err := collectCommunitySkills(projectRoot, homeDir)
	if err != nil {
		return existing
	}

	if len(community) == 0 {
		return existing
	}

	return append(existing, community...)
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

	// For Odoo version filtering if applicable
	odooVersions, _, _ := detectOdooMajorVersions(projectRoot)

	err := filepath.WalkDir(projectRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}

		rel, _ := filepath.Rel(projectRoot, path)
		relSlash := filepath.ToSlash(rel)

		// Explicitly allow .agent and its subdirectories
		if strings.Contains(relSlash, "/.agent") || strings.HasPrefix(relSlash, ".agent") {
			// continue
		} else if scope.ShouldSkipRefactorPath(rel) || strings.HasPrefix(relSlash, "internal/assets/overlays") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if d.Name() == "SKILL.md" {
			skillName := filepath.Base(filepath.Dir(path))

			// Apply Odoo version filtering if applicable
			if !matchesOverlaySkillVersion(skillName, odooVersions) {
				return nil
			}

			info := parseSkillFile(path)
			if info.Name == "" {
				info.Name = skillName
			}
			// Skip deprecated and archived skills.
			if info.Deprecated || strings.Contains(path, "_archived") {
				return nil
			}
			info.Path = path
			info.Origin = "project"

			// Standard classification
			if info.Name == "_shared" {
				info.Kind = "SharedRule"
			} else if strings.HasPrefix(info.Name, "sdd-") || info.Name == "skill-registry" || info.Name == "cognitive-mode" || info.Name == "adaptive-reasoning" {
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
		if len(manifest.RegistryEntries) > 0 {
			for _, re := range manifest.RegistryEntries {
				// We still apply version filtering if it's an Odoo project,
				// though RegistryEntries should already be version-filtered at install time.
				// This is a safety check for when a project version changes after install.
				if !matchesOverlaySkillVersion(re.Skill, odooVersions) {
					continue
				}

				info := parseSkillFile(re.Path)
				if info.Name == "" {
					info.Name = re.Skill
				}
				if info.Deprecated || strings.Contains(re.Path, "_archived") {
					continue
				}
				// Prioritize manifest trigger if present
				if re.Trigger != "" {
					info.Trigger = re.Trigger
				}
				info.Path = re.Path
				info.Origin = "overlay"
				info.Kind = "Overlay"

				skills = append(skills, info)
			}
		} else {
			overlaySkillDir := filepath.Join(overlayRoot, "skills")
			if _, err := os.Stat(overlaySkillDir); err == nil {
				dirs, err := os.ReadDir(overlaySkillDir)
				if err == nil {
					for _, entry := range dirs {
						if !matchesOverlaySkillVersion(entry.Name(), odooVersions) {
							continue
						}

						skillPath := filepath.Join(overlaySkillDir, entry.Name(), "SKILL.md")
						if _, err := os.Stat(skillPath); err != nil {
							continue
						}

						// Prefer the .agent/skills/ symlink path when it exists.
						agentSkillPath := filepath.Join(projectRoot, ".agent", "skills", entry.Name(), "SKILL.md")
						if _, err := os.Stat(agentSkillPath); err == nil {
							skillPath = agentSkillPath
						}

						info := parseSkillFile(skillPath)
						if info.Name == "" {
							info.Name = entry.Name()
						}
						if info.Deprecated || strings.Contains(skillPath, "_archived") || entry.Name() == "_archived" {
							continue
						}
						info.Path = skillPath
						info.Origin = "overlay"

						if entry.Name() == "_shared" {
							info.Kind = "SharedRule"
						} else if strings.HasPrefix(entry.Name(), "sdd-") || entry.Name() == "skill-registry" || entry.Name() == "cognitive-mode" || entry.Name() == "adaptive-reasoning" {
							info.Kind = "System"
						} else {
							info.Kind = "Overlay"
						}

						skills = append(skills, info)
					}
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
		} else if strings.HasPrefix(d.Name(), "sdd-") || d.Name() == "skill-registry" || d.Name() == "cognitive-mode" || d.Name() == "adaptive-reasoning" {
			kind = "System"
		}

		skillPath := filepath.Join(dir, d.Name(), "SKILL.md")
		if _, err := os.Stat(skillPath); err == nil {
			info := parseSkillFile(skillPath)
			if info.Name == "" {
				info.Name = d.Name()
			}
			if info.Deprecated || d.Name() == "_archived" || strings.Contains(skillPath, "_archived") {
				continue
			}
			info.Path = skillPath
			info.Origin = origin
			info.Kind = kind
			entries = append(entries, info)
		}
	}
	return entries
}

// canonicalRuleHeadings is the ordered list of H2 headings whose body
// we extract as compact rules. Earlier entries have higher priority.
// Only lower-case normalized H2 text is compared against this list.
var canonicalRuleHeadings = []string{
	"compact rules",
	"rules",
	"patterns",
	"critical rules",
	"contract",
	"convention",
	"guidelines",
	"mandates",
	"standards",
	"workflow",
	"procedure",
	"core behavior",
	"essential patterns",
	"required conventions",
}

func parseSkillFile(path string) skillEntry {
	data, err := os.ReadFile(path)
	if err != nil {
		return skillEntry{}
	}

	var entry skillEntry

	// Compute SHA-256 of the entire file for hash-based invalidation.
	h := sha256.Sum256(data)
	entry.SHA256 = hex.EncodeToString(h[:])

	scanner := bufio.NewScanner(strings.NewReader(string(data)))

	inFrontmatter := false
	frontmatterDone := false

	var descriptionBuffer strings.Builder
	inDescription := false

	var rulesSections []string
	var currentSection []string
	inRules := false
	bestHeadingMatch := -1 // -1 = no match, 0 = "compact rules", 1 = "rules", etc.

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Simple frontmatter parsing
		if trimmedLine == "---" {
			if !inFrontmatter && !frontmatterDone {
				inFrontmatter = true
			} else if inFrontmatter {
				inFrontmatter = false
				frontmatterDone = true
			}
			continue
		}

		if inFrontmatter {
			if _, val, ok := parseYAMLField(trimmedLine, "name"); ok {
				entry.Name = val
			} else if _, val, ok := parseYAMLField(trimmedLine, "trigger"); ok {
				entry.Trigger = val
			} else if _, val, ok := parseYAMLField(trimmedLine, "description"); ok {
				// Handle multiline description with > operator
				if strings.HasPrefix(val, ">") {
					inDescription = true
					remaining := strings.TrimSpace(strings.TrimPrefix(val, ">"))
					if remaining != "" {
						descriptionBuffer.WriteString(remaining)
						descriptionBuffer.WriteString(" ")
					}
				} else {
					descriptionBuffer.WriteString(val)
					descriptionBuffer.WriteString(" ")
				}
			} else if inDescription {
				if trimmedLine == "" || (!strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "\t")) {
					inDescription = false
				} else {
					descriptionBuffer.WriteString(trimmedLine)
					descriptionBuffer.WriteString(" ")
				}
			} else if strings.TrimLeft(trimmedLine, "  \t") == "deprecated: true" || strings.TrimLeft(trimmedLine, "  \t") == "deprecated: True" {
				entry.Deprecated = true
			}
			continue
		}

		// Look for name in first H1 if not in frontmatter
		if entry.Name == "" && strings.HasPrefix(trimmedLine, "# ") {
			entry.Name = strings.TrimPrefix(trimmedLine, "# ")
		}

		// Heading-based section detection.
		// When we encounter any heading (##, ###), we:
		//  1. Save any previously active rules section.
		//  2. Check if this heading matches a canonical one.
		if strings.HasPrefix(trimmedLine, "## ") || strings.HasPrefix(trimmedLine, "### ") {
			// Save previous section if active.
			if inRules && len(currentSection) > 0 {
				rulesSections = append(rulesSections, strings.Join(currentSection, "\n"))
				currentSection = nil
			}
			inRules = false

			if strings.HasPrefix(trimmedLine, "## ") {
				lower := strings.TrimPrefix(strings.ToLower(trimmedLine), "## ")
				// Check against canonical headings.
				for idx, want := range canonicalRuleHeadings {
					if strings.Contains(lower, want) {
						// Only consider this a match if it improves on current best.
						// Lower index = higher priority ("compact rules" at 0 beats "rules" at 1).
						if bestHeadingMatch < 0 || idx < bestHeadingMatch {
							bestHeadingMatch = idx
							if len(currentSection) > 0 {
								rulesSections = append(rulesSections, strings.Join(currentSection, "\n"))
								currentSection = nil
							}
						}
						inRules = true
						break
					}
				}
			}
			continue
		}

		if inRules && len(currentSection) < 40 {
			currentSection = append(currentSection, line)
		}
	}

	// Save last section.
	if inRules && len(currentSection) > 0 {
		rulesSections = append(rulesSections, strings.Join(currentSection, "\n"))
	}

	// Join multiple sections with blank line separator.
	entry.CompactRules = strings.TrimSpace(strings.Join(rulesSections, "\n\n"))

	// Post-processing for triggers if missing from frontmatter
	if entry.Trigger == "" && descriptionBuffer.Len() > 0 {
		descText := strings.TrimSpace(descriptionBuffer.String())
		if idx := strings.Index(descText, "Trigger:"); idx != -1 {
			raw := strings.TrimSpace(descText[idx+len("Trigger:"):])
			if nl := strings.IndexAny(raw, "\n\r"); nl != -1 {
				raw = strings.TrimSpace(raw[:nl])
			}
			entry.Trigger = raw
		} else {
			trigger := descText
			if dot := strings.IndexAny(trigger, ".。"); dot > 0 && dot < 300 {
				trigger = strings.TrimSpace(trigger[:dot])
			} else if len(trigger) > 300 {
				trigger = trigger[:300]
				if lastSpace := strings.LastIndexAny(trigger, " \t\n\r"); lastSpace > 250 {
					trigger = strings.TrimSpace(trigger[:lastSpace])
				}
			}
			entry.Trigger = trigger
		}
	}

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

		// community > project > overlay > user > system > shared
		priority := map[string]int{"community": 6, "project": 5, "overlay": 4, "user": 3, "system": 2, "shared": 1}
		if priority[s.Origin] > priority[existing.Origin] {
			if s.Origin == "project" && existing.Origin == "overlay" {
				fmt.Fprintf(os.Stderr, "Warning: project skill %q overrides overlay skill at %s\n", s.Name, existing.Path)
			}
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

func escapeTable(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}

// probeNotebookLMState checks the status of the NotebookLM MCP server.
// Returns one of: "NOT FOUND", "FOUND", "READY".
func probeNotebookLMState() (status string, count int, err error) {
	// First check if 'nlm' is on path
	p, err := exec.LookPath("nlm")
	if err != nil {
		return "NOT FOUND", 0, nil
	}

	// Check if any notebooks are found
	out, err := exec.Command(p, "list", "--json").CombinedOutput()
	if err != nil {
		// Found the binary but it might not be configured or fails
		return "FOUND", 0, nil
	}

	var notebooks []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	if err := json.Unmarshal(out, &notebooks); err != nil {
		return "FOUND", 0, nil
	}

	if len(notebooks) > 0 {
		return "READY", len(notebooks), nil
	}

	return "FOUND", 0, nil
}

// parseExistingRegistryHashes reads the existing registry file and extracts
// skill name → SHA256 mappings from the "## Registry Hashes" section.
func parseExistingRegistryHashes(registryPath string) map[string]string {
	data, err := os.ReadFile(registryPath)
	if err != nil {
		return nil
	}
	content := string(data)
	// Locate the hashes section between the markers.
	open := "<!-- architect-ai:registry:hashes -->"
	close := "<!-- /architect-ai:registry:hashes -->"
	openIdx := strings.Index(content, open)
	closeIdx := strings.Index(content, close)
	if openIdx < 0 || closeIdx < 0 || closeIdx <= openIdx {
		return nil
	}
	section := content[openIdx+len(open) : closeIdx]
	hashes := make(map[string]string)
	for _, line := range strings.Split(section, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "| "))
		line = strings.TrimSuffix(line, " |")
		parts := strings.SplitN(line, " | ", 2)
		if len(parts) == 2 && parts[0] != "Skill" && parts[0] != "" {
			hashes[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return hashes
}

func parseYAMLField(line, fieldName string) (key, value string, ok bool) {
	lower := strings.ToLower(line)
	prefix := fieldName + ":"
	if !strings.HasPrefix(lower, prefix) {
		return "", "", false
	}
	return fieldName, strings.TrimSpace(line[len(prefix):]), true
}

// EnsureProjectRegistryReady performs the base initialization of a project for ATL/SDD.
// It creates the .atl directory, bootstraps project-local overlays, builds the skill registry,
// and ensures core project conventions (AGENTS.md, GEMINI.md) are present.
func EnsureProjectRegistryReady(projectRoot string, force bool) (OverlayBootstrapResult, error) {
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

	// 1.1 Cleanup false positives: if not Odoo but Odoo overlay exists, remove it.
	if !result.IsOdooProject {
		baseOverlayRoot := filepath.Join(projectRoot, ".atl", "overlays", defaultOverlayName)
		if _, err := os.Stat(baseOverlayRoot); err == nil {
			// We use a simplified removal to avoid circular dependency with WriteLocalSkillRegistry
			manifestPath := filepath.Join(baseOverlayRoot, "manifest.json")
			if manifest, err := readOverlayManifest(manifestPath); err == nil {
				_ = unbridgeOverlaySkills(projectRoot, manifest)
			}
			_ = os.RemoveAll(baseOverlayRoot)
		}
	}

	// 2. Build/Update the registry markdown
	if err := WriteLocalSkillRegistry(projectRoot, force); err != nil {
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