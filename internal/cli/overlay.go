package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	embeddedassets "github.com/rd-mg/architect-ai/internal/assets"
)

const (
	defaultOverlayName      = "odoo-development-skill"
	overlayActivationActive = "active"
	bundledOverlayRoot      = "overlays"
	embeddedOverlayPrefix   = "embedded://"
)

type OverlayInstallOptions struct {
	OverlayName     string
	ProjectRoot     string
	SourcePath      string
	VersionIntent   string
	EnterprisePath  string
	ExplicitRequest bool
}

type RegistryEntry struct {
	Skill   string `json:"skill"`
	Trigger string `json:"trigger"`
	Path    string `json:"path"`
}

type OverlayManifest struct {
	Name             string          `json:"name"`
	SourcePath       string          `json:"source_path"`
	EnterprisePath   string          `json:"enterprise_path,omitempty"`
	InstalledAtUTC   string          `json:"installed_at_utc"`
	VersionIntent    string          `json:"version_intent,omitempty"`
	ActivationState  string          `json:"activation_state"`
	DetectedVersions []int           `json:"detected_versions,omitempty"`
	Skills           []string        `json:"skills"`
	OptionalSkills   []string        `json:"optional_skills,omitempty"`
	EnabledOptionals []string        `json:"enabled_optionals,omitempty"`
	Agents           []string        `json:"agents"`
	Patterns         []string        `json:"patterns,omitempty"`
	Instructions     []string        `json:"instructions,omitempty"`
	Prompts          []string        `json:"prompts,omitempty"`
	Scripts          []string        `json:"scripts,omitempty"`
	StaticAssets     []string        `json:"static_assets,omitempty"`
	Assets           []string        `json:"assets"`
	RegistryEntries  []RegistryEntry `json:"registry_entries,omitempty"`
}

var odooManifestVersionPattern = regexp.MustCompile(`(?m)["']version["']\s*:\s*["']([0-9]{1,2})\.[^"']*["']`)
var versionSuffixPattern = regexp.MustCompile(`-(\d{2})(?:-(\d{2}))?$`)
var overlaySkillRangePattern = regexp.MustCompile(`-(\d{2})-(\d{2})(?:\.0)?$`)
var overlaySkillSinglePattern = regexp.MustCompile(`-(\d{2})(?:\.0)?$`)

func InstallOverlay(opts OverlayInstallOptions) (OverlayManifest, error) {
	overlayName := strings.TrimSpace(opts.OverlayName)
	if overlayName == "" {
		overlayName = defaultOverlayName
	}
	if overlayName != defaultOverlayName && !strings.HasPrefix(overlayName, "odoo-") {
		return OverlayManifest{}, fmt.Errorf("unsupported overlay %q", overlayName)
	}

	projectRoot := strings.TrimSpace(opts.ProjectRoot)
	if projectRoot == "" {
		return OverlayManifest{}, errors.New("project root is required")
	}

	if !opts.ExplicitRequest && !hasStrongOverlayRepoEvidence(projectRoot, overlayName) {
		return OverlayManifest{}, errors.New("overlay installation requires explicit request or strong repository evidence")
	}

	sourceFS, manifestSourcePath, err := resolveOverlaySource(overlayName, strings.TrimSpace(opts.SourcePath))
	if err != nil {
		return OverlayManifest{}, err
	}

	overlayRoot := filepath.Join(projectRoot, ".atl", "overlays", overlayName)
	if err := os.RemoveAll(overlayRoot); err != nil {
		return OverlayManifest{}, fmt.Errorf("remove existing overlay root: %w", err)
	}
	if err := os.MkdirAll(overlayRoot, 0o755); err != nil {
		return OverlayManifest{}, fmt.Errorf("create overlay root: %w", err)
	}

	manifest := OverlayManifest{
		Name:            overlayName,
		SourcePath:      manifestSourcePath,
		EnterprisePath:  resolveEnterpriseRepoPath(strings.TrimSpace(opts.EnterprisePath)),
		InstalledAtUTC:  time.Now().UTC().Format(time.RFC3339),
		VersionIntent:   strings.TrimSpace(opts.VersionIntent),
		ActivationState: overlayActivationActive,
		Skills:          []string{},
		Agents:          []string{},
		Assets:          []string{},
	}

	skillDest := filepath.Join(overlayRoot, "skills", overlayName, "SKILL.md")
	if err := copyFSFile(sourceFS, "SKILL.md", skillDest); err != nil {
		return OverlayManifest{}, err
	}
	// Root skill goes into Assets only — it is the overlay meta-identifier, not a
	// consumable skill for agents. Do NOT add it to manifest.Skills.
	manifest.Assets = append(manifest.Assets, filepath.ToSlash(filepath.Join("skills", overlayName, "SKILL.md")))

	// Parse VersionIntent into a filter set so non-matching skill bundles are
	// skipped at install time (not only at bridge time). Empty intent = no filter.
	versionFilter := parseVersionIntent(opts.VersionIntent)

	// Optional skills: heavy or rarely-needed skills are moved to OptionalSkills so they
	// are not bridged by default. Enable with: atl overlay enable <skill-name>
	optionalSkillSet := map[string]bool{
		"odoo-minimax-xlsx-o-spreadsheets": true,
		"odoo-quote-calculator":            true,
	}

	skillBundleNames, skillBundleAssets, err := copyOverlaySkillBundles(sourceFS, filepath.Join(overlayRoot, "skills"), overlayName, versionFilter)
	if err != nil {
		return OverlayManifest{}, err
	}
	for _, name := range skillBundleNames {
		if optionalSkillSet[name] {
			manifest.OptionalSkills = append(manifest.OptionalSkills, name)
		} else {
			manifest.Skills = append(manifest.Skills, name)
		}
	}
	manifest.Assets = append(manifest.Assets, skillBundleAssets...)
	manifest.Skills = uniqueStrings(manifest.Skills)
	manifest.OptionalSkills = uniqueStrings(manifest.OptionalSkills)
	sort.Strings(manifest.Skills)
	sort.Strings(manifest.OptionalSkills)

	agentFiles, err := copyFSTree(sourceFS, "agents", filepath.Join(overlayRoot, "agents"), nil)
	if err != nil {
		return OverlayManifest{}, err
	}
	for _, rel := range agentFiles {
		manifest.Assets = append(manifest.Assets, filepath.ToSlash(filepath.Join("agents", rel)))
		name := strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))
		if name != "" {
			manifest.Agents = append(manifest.Agents, name)
		}
	}
	manifest.Agents = uniqueStrings(manifest.Agents)
	sort.Strings(manifest.Agents)

	// REMOVED in V3: loose-pattern copy bypass version filtering.
	// All patterns are now inside version-gated bundle directories
	// (patterns-agnostic/, patterns-18/, patterns-19/, etc.) which are
	// handled by copyOverlaySkillBundles() and version-filtered by
	// bridgeOverlaySkills() → matchesOverlaySkillVersion().
	manifest.Patterns = []string{}

	instructionFiles, err := copyFSTree(sourceFS, "instructions", filepath.Join(overlayRoot, "instructions"), nil)
	if err != nil {
		return OverlayManifest{}, err
	}
	sort.Strings(instructionFiles)
	manifest.Instructions = append(manifest.Instructions, instructionFiles...)
	for _, rel := range instructionFiles {
		manifest.Assets = append(manifest.Assets, filepath.ToSlash(filepath.Join("instructions", rel)))
	}

	promptFiles, err := copyFSTree(sourceFS, "prompts", filepath.Join(overlayRoot, "prompts"), nil)
	if err != nil {
		return OverlayManifest{}, err
	}
	sort.Strings(promptFiles)
	manifest.Prompts = append(manifest.Prompts, promptFiles...)
	for _, rel := range promptFiles {
		manifest.Assets = append(manifest.Assets, filepath.ToSlash(filepath.Join("prompts", rel)))
	}

	scriptFiles, err := copyFSTree(sourceFS, "scripts", filepath.Join(overlayRoot, "scripts"), nil)
	if err != nil {
		return OverlayManifest{}, err
	}
	sort.Strings(scriptFiles)
	manifest.Scripts = append(manifest.Scripts, scriptFiles...)
	for _, rel := range scriptFiles {
		manifest.Assets = append(manifest.Assets, filepath.ToSlash(filepath.Join("scripts", rel)))
	}

	staticAssetFiles, err := copyFSTree(sourceFS, "assets", filepath.Join(overlayRoot, "assets"), nil)
	if err != nil {
		return OverlayManifest{}, err
	}
	sort.Strings(staticAssetFiles)
	manifest.StaticAssets = append(manifest.StaticAssets, staticAssetFiles...)
	for _, rel := range staticAssetFiles {
		manifest.Assets = append(manifest.Assets, filepath.ToSlash(filepath.Join("assets", rel)))
	}

	// sdd-supplements: phase-specific Odoo context injected by the SDD orchestrator
	// into each sub-agent delegation. All orchestrator assets reference:
	//   .atl/overlays/odoo-*/sdd-supplements/{phase}-odoo.md
	// Without this copy those files never reach disk and agents silently skip them.
	sddSuppFiles, err := copyFSTree(sourceFS, "sdd-supplements", filepath.Join(overlayRoot, "sdd-supplements"), nil)
	if err != nil {
		return OverlayManifest{}, err
	}
	sort.Strings(sddSuppFiles)
	for _, rel := range sddSuppFiles {
		manifest.Assets = append(manifest.Assets, filepath.ToSlash(filepath.Join("sdd-supplements", rel)))
	}

	// rules: coding-style, security and naming rules consumed by the skill-registry
	// and referenced in the overlay SKILL.md as "compact rules for coding style,
	// security, and manifest conventions". Without this copy they are unreachable.
	rulesFiles, err := copyFSTree(sourceFS, "rules", filepath.Join(overlayRoot, "rules"), nil)
	if err != nil {
		return OverlayManifest{}, err
	}
	sort.Strings(rulesFiles)
	for _, rel := range rulesFiles {
		manifest.Assets = append(manifest.Assets, filepath.ToSlash(filepath.Join("rules", rel)))
	}

	if manifest.EnterprisePath != "" {
		if err := injectEnterpriseContext(overlayRoot, manifest.EnterprisePath); err != nil {
			return OverlayManifest{}, err
		}
	}

	populateRegistryEntries(&manifest, projectRoot)

	// Cache detected Odoo versions to avoid O(n) walk in bridgeOverlaySkills.
	if detVersions, isOdoo, _ := detectOdooMajorVersions(projectRoot); isOdoo && len(detVersions) > 0 {
		vers := make([]int, 0, len(detVersions))
		for v := range detVersions {
			vers = append(vers, v)
		}
		sort.Ints(vers)
		manifest.DetectedVersions = vers
	}

	manifestPath := filepath.Join(overlayRoot, "manifest.json")
	if err := writeOverlayManifest(manifestPath, manifest); err != nil {
		return OverlayManifest{}, err
	}

	if err := WriteLocalSkillRegistry(projectRoot); err != nil {
		return OverlayManifest{}, err
	}

	if err := bridgeOverlaySkills(projectRoot, manifest); err != nil {
		fmt.Printf("Warning: Failed to bridge overlay skills: %v\n", err)
	}

	return manifest, nil
}

func RefreshOverlay(projectRoot string, overlayName string, enterprisePath string) (OverlayManifest, error) {
	if strings.TrimSpace(projectRoot) == "" {
		return OverlayManifest{}, errors.New("project root is required")
	}

	overlayName = strings.TrimSpace(overlayName)
	if overlayName == "" {
		return OverlayManifest{}, errors.New("overlay name is required")
	}

	manifest, err := readOverlayManifest(filepath.Join(projectRoot, ".atl", "overlays", overlayName, "manifest.json"))
	if err != nil {
		return OverlayManifest{}, err
	}

	sourcePath := manifest.SourcePath
	if isEmbeddedOverlaySource(sourcePath) {
		sourcePath = ""
	}

	resolvedEnterprisePath := strings.TrimSpace(enterprisePath)
	if resolvedEnterprisePath == "" {
		resolvedEnterprisePath = manifest.EnterprisePath
	}

	return InstallOverlay(OverlayInstallOptions{
		OverlayName:     overlayName,
		ProjectRoot:     projectRoot,
		SourcePath:      sourcePath,
		VersionIntent:   manifest.VersionIntent,
		EnterprisePath:  resolvedEnterprisePath,
		ExplicitRequest: true,
	})
}

func RemoveOverlay(projectRoot string, overlayName string) error {
	if strings.TrimSpace(projectRoot) == "" {
		return errors.New("project root is required")
	}
	overlayName = strings.TrimSpace(overlayName)
	if overlayName == "" {
		return errors.New("overlay name is required")
	}

	overlayRoot := filepath.Join(projectRoot, ".atl", "overlays", overlayName)
	manifest, err := readOverlayManifest(filepath.Join(overlayRoot, "manifest.json"))
	if err == nil {
		_ = unbridgeOverlaySkills(projectRoot, manifest)
	}

	if err := os.RemoveAll(overlayRoot); err != nil {
		return fmt.Errorf("remove overlay root: %w", err)
	}
	return WriteLocalSkillRegistry(projectRoot)
}

type OverlayBootstrapResult struct {
	Overlays      []OverlayManifest
	Actions       map[string]string
	Versions      map[int]struct{}
	IsOdooProject bool
}

// NOTE: AutoDeployVersionOverlay was removed in the single-overlay refactor.
// It used to install both an "agnostic" and a version-specific overlay ("odoo-N"),
// producing byte-identical duplicates. The logic is now folded into
// BootstrapProjectLocalOverlays, which installs only the base overlay
// (odoo-development-skill) with the detected versions as VersionIntent.

func BootstrapProjectLocalOverlays(projectRoot string, refresh bool, enterprisePath string) (OverlayBootstrapResult, error) {
	// 1. Resolve absolute project root for reliable path handling.
	absProjectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return OverlayBootstrapResult{}, fmt.Errorf("resolve absolute project root: %w", err)
	}

	versions, isOdooProject, err := detectOdooMajorVersions(absProjectRoot)
	if err != nil {
		return OverlayBootstrapResult{}, err
	}

	result := OverlayBootstrapResult{
		Overlays:      []OverlayManifest{},
		Actions:       make(map[string]string),
		Versions:      versions,
		IsOdooProject: isOdooProject,
	}

	if !isOdooProject {
		return result, nil
	}

	// SINGLE OVERLAY MODEL:
	// There is only ONE embedded source (overlays/odoo-development-skill). Previously
	// we installed BOTH a versioned overlay (odoo-19) AND the base — both resolving to
	// the SAME embedded source, producing 100% byte-identical duplicates on disk.
	//
	// Now we install only the base overlay `odoo-development-skill`, passing the set of
	// detected versions as VersionIntent. bridgeOverlaySkills + copyOverlaySkillBundles
	// then filter by those versions, so disk and .agent/skills/ stay version-correct.
	baseOverlayName := defaultOverlayName
	baseManifestPath := filepath.Join(absProjectRoot, ".atl", "overlays", baseOverlayName, "manifest.json")
	baseOverlayExists := overlayManifestExists(baseManifestPath)

	versionIntent := formatVersionSet(versions)

	action := "reused"
	var manifest OverlayManifest

	switch {
	case refresh && baseOverlayExists:
		manifest, err = RefreshOverlay(absProjectRoot, baseOverlayName, enterprisePath)
		action = "refreshed"
	case !baseOverlayExists:
		manifest, err = InstallOverlay(OverlayInstallOptions{
			OverlayName:     baseOverlayName,
			ProjectRoot:     absProjectRoot,
			EnterprisePath:  enterprisePath,
			ExplicitRequest: true,
			VersionIntent:   versionIntent,
		})
		action = "installed"
	default:
		manifest, err = readOverlayManifest(baseManifestPath)
	}

	if err != nil {
		fmt.Printf("Warning: Base overlay deployment failed for %s: %v\n", baseOverlayName, err)
	} else {
		result.Overlays = append(result.Overlays, manifest)
		result.Actions[manifest.Name] = action
	}

	// Clean up any stale versioned overlay directories from previous installs that
	// used the dual-overlay model. They are identical duplicates of the base and
	// leaving them around re-introduces the confusion we just fixed.
	for version := range versions {
		staleDir := filepath.Join(absProjectRoot, ".atl", "overlays", fmt.Sprintf("odoo-%d", version))
		if info, err := os.Stat(staleDir); err == nil && info.IsDir() {
			_ = os.RemoveAll(staleDir)
		}
	}

	return result, nil
}

func overlayManifestExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readOverlayManifest(path string) (OverlayManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return OverlayManifest{}, fmt.Errorf("overlay manifest not found: %s", path)
		}
		return OverlayManifest{}, fmt.Errorf("read overlay manifest %q: %w", path, err)
	}

	var manifest OverlayManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return OverlayManifest{}, fmt.Errorf("decode overlay manifest %q: %w", path, err)
	}
	return manifest, nil
}

func writeOverlayManifest(path string, manifest OverlayManifest) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create manifest directory: %w", err)
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal overlay manifest: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write overlay manifest: %w", err)
	}
	return nil
}

func resolveOverlaySource(overlayName string, sourcePath string) (fs.FS, string, error) {
	if sourcePath == "" {
		targetOverlayName := overlayName
		// Fallback for odoo-X versioned overlays to use the main development skill as source
		if strings.HasPrefix(overlayName, "odoo-") && overlayName != defaultOverlayName {
			targetOverlayName = defaultOverlayName
		}

		bundledPath := path.Join(bundledOverlayRoot, targetOverlayName)
		sourceFS, err := fs.Sub(embeddedassets.FS, bundledPath)
		if err != nil {
			return nil, "", fmt.Errorf("bundled overlay %q not available: %w", targetOverlayName, err)
		}
		if _, err := fs.Stat(sourceFS, "SKILL.md"); err != nil {
			return nil, "", fmt.Errorf("bundled overlay %q missing root SKILL.md: %w", targetOverlayName, err)
		}
		return sourceFS, embeddedOverlayPrefix + bundledPath, nil
	}

	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return nil, "", fmt.Errorf("overlay source not accessible: %w", err)
	}
	if !sourceInfo.IsDir() {
		return nil, "", fmt.Errorf("overlay source path %q is not a directory", sourcePath)
	}

	sourceFS := os.DirFS(sourcePath)
	if _, err := fs.Stat(sourceFS, "SKILL.md"); err != nil {
		return nil, "", fmt.Errorf("overlay source missing root SKILL.md: %w", err)
	}

	return sourceFS, sourcePath, nil
}

func isEmbeddedOverlaySource(sourcePath string) bool {
	return strings.HasPrefix(strings.TrimSpace(sourcePath), embeddedOverlayPrefix)
}

func copyOverlaySkillBundles(sourceFS fs.FS, destRoot string, rootOverlayName string, versionFilter map[int]struct{}) ([]string, []string, error) {
	entries, err := fs.ReadDir(sourceFS, "skills")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("read skill bundle directory: %w", err)
	}

	bundleNames := make([]string, 0)
	bundleAssets := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		bundleName := entry.Name()
		if bundleName == rootOverlayName {
			continue
		}
		// Single-overlay model: filter at install time so non-matching versions
		// never touch disk. Prevents the "patterns-14..18 copied for a v19 project"
		// contamination that existed under the dual-overlay design.
		if len(versionFilter) > 0 && !matchesOverlaySkillVersion(bundleName, versionFilter) {
			continue
		}
		sourceDir := path.Join("skills", bundleName)
		files, err := copyFSTree(sourceFS, sourceDir, filepath.Join(destRoot, bundleName), nil)
		if err != nil {
			return nil, nil, err
		}
		bundleNames = append(bundleNames, bundleName)
		for _, rel := range files {
			bundleAssets = append(bundleAssets, filepath.ToSlash(filepath.Join("skills", bundleName, rel)))
		}
	}

	return bundleNames, bundleAssets, nil
}

func copyFSDirFiles(sourceFS fs.FS, sourceDir string, destDir string, include func(path string, d fs.DirEntry) bool) ([]string, error) {
	entries, err := fs.ReadDir(sourceFS, sourceDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read source directory %q: %w", sourceDir, err)
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, fmt.Errorf("create destination directory %q: %w", destDir, err)
	}

	files := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		assetPath := path.Join(sourceDir, entry.Name())
		if include != nil && !include(assetPath, entry) {
			continue
		}
		if err := copyFSFile(sourceFS, assetPath, filepath.Join(destDir, entry.Name())); err != nil {
			return nil, err
		}
		files = append(files, entry.Name())
	}
	return files, nil
}

func copyFSTree(sourceFS fs.FS, sourceDir string, destDir string, include func(path string, d fs.DirEntry) bool) ([]string, error) {
	if _, err := fs.Stat(sourceFS, sourceDir); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat source directory %q: %w", sourceDir, err)
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, fmt.Errorf("create destination directory %q: %w", destDir, err)
	}

	names := make([]string, 0)
	err := fs.WalkDir(sourceFS, sourceDir, func(assetPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if include != nil && !include(assetPath, d) {
			return nil
		}

		rel, err := filepath.Rel(sourceDir, assetPath)
		if err != nil {
			return fmt.Errorf("resolve relative path: %w", err)
		}
		destPath := filepath.Join(destDir, filepath.FromSlash(rel))
		if err := copyFSFile(sourceFS, assetPath, destPath); err != nil {
			return err
		}

		names = append(names, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("copy tree: %w", err)
	}

	return names, nil
}

func detectOdooMajorVersions(projectRoot string) (map[int]struct{}, bool, error) {
	versions := make(map[int]struct{})
	seenManifest := false

	err := filepath.WalkDir(projectRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := d.Name()
			if base == ".git" || base == ".atl" || base == "node_modules" || base == ".venv" {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "__manifest__.py" {
			return nil
		}

		seenManifest = true
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		major, ok := extractOdooMajorVersion(string(data))
		if ok {
			versions[major] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, false, fmt.Errorf("scan manifests for Odoo version: %w", err)
	}

	return versions, seenManifest, nil
}

func extractOdooMajorVersion(content string) (int, bool) {
	match := odooManifestVersionPattern.FindStringSubmatch(content)
	if len(match) < 2 {
		return 0, false
	}
	major, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, false
	}
	if major < 10 || major > 99 {
		return 0, false
	}
	return major, true
}

func matchesOdooVersion(path string, targetVersions map[int]struct{}) bool {
	if len(targetVersions) == 0 {
		return false
	}

	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	lower := strings.ToLower(base)
	if strings.HasSuffix(lower, "-all") {
		return false
	}

	match := versionSuffixPattern.FindStringSubmatch(base)
	if len(match) < 2 {
		// No explicit version suffix means agnostic content; skip by policy.
		return false
	}

	from, err := strconv.Atoi(match[1])
	if err != nil {
		return false
	}
	to := from
	if len(match) >= 3 && match[2] != "" {
		to, err = strconv.Atoi(match[2])
		if err != nil {
			return false
		}
	}
	if to < from {
		from, to = to, from
	}

	for v := range targetVersions {
		if v >= from && v <= to {
			return true
		}
	}
	return false
}

func matchesOverlaySkillVersion(name string, targetVersions map[int]struct{}) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	if lower == "" {
		return false
	}

	// No explicit version suffix: treat as agnostic bundle.
	// We allow these even if targetVersions is empty (Day 0 project).
	isAgnostic := !overlaySkillRangePattern.MatchString(lower) && !overlaySkillSinglePattern.MatchString(lower)

	if len(targetVersions) == 0 {
		return isAgnostic
	}

	if strings.HasSuffix(lower, "-all") {
		return true
	}

	if match := overlaySkillRangePattern.FindStringSubmatch(lower); len(match) == 3 {
		from, errFrom := strconv.Atoi(match[1])
		to, errTo := strconv.Atoi(match[2])
		if errFrom != nil || errTo != nil {
			return false
		}
		if to < from {
			from, to = to, from
		}
		for v := range targetVersions {
			if v >= from && v <= to {
				return true
			}
		}
		return false
	}

	if match := overlaySkillSinglePattern.FindStringSubmatch(lower); len(match) == 2 {
		major, err := strconv.Atoi(match[1])
		if err != nil {
			return false
		}
		_, ok := targetVersions[major]
		return ok
	}

	// No explicit version suffix: treat as agnostic bundle.
	return true
}

func bridgeOverlaySkills(projectRoot string, manifest OverlayManifest) error {
	agentSkillsDir := filepath.Join(projectRoot, ".agent", "skills")
	if err := os.MkdirAll(agentSkillsDir, 0o755); err != nil {
		return fmt.Errorf("create .agent/skills: %w", err)
	}

	overlayRoot := filepath.Join(projectRoot, ".atl", "overlays", manifest.Name)

	// For Odoo projects, we only bridge skills matching the detected version(s).
	// Read from manifest cache (populated at install time) to avoid O(n) walk on
	// every bridge call. Falls back to live detection if cache is empty.
	var odooVersions map[int]struct{}
	isOdoo := len(manifest.DetectedVersions) > 0
	if isOdoo {
		odooVersions = make(map[int]struct{}, len(manifest.DetectedVersions))
		for _, v := range manifest.DetectedVersions {
			odooVersions[v] = struct{}{}
		}
	} else {
		// Safety net: cache miss → fall back to walk.
		var detErr error
		odooVersions, isOdoo, detErr = detectOdooMajorVersions(projectRoot)
		if detErr != nil {
			fmt.Printf("Warning: version detection fallback failed: %v\n", detErr)
		}
	}

	for _, skillName := range manifest.Skills {
		// Version filtering for Odoo skills
		if isOdoo && len(odooVersions) > 0 && !matchesOverlaySkillVersion(skillName, odooVersions) {
			continue
		}

		// Only bridge if it actually has a SKILL.md in the overlay
		src := filepath.Join(overlayRoot, "skills", skillName)
		if info, err := os.Stat(src); err != nil || !info.IsDir() {
			continue
		}

		// Promote a version-specific skill bundle to the short "odoo" name when it
		// matches one of the detected project versions. This lets agents refer to
		// the Odoo knowledge base as simply "odoo" regardless of target version.
		//
		// Single-overlay model: we used to derive version from manifest.Name (e.g.
		// "odoo-19"). Now manifest.Name is always "odoo-development-skill", so we
		// use the detected versions from __manifest__.py instead.
		bridgeName := skillName
		if isOdoo && len(odooVersions) > 0 && strings.HasPrefix(skillName, "odoo-") {
			for v := range odooVersions {
				vStr := strconv.Itoa(v)
				if skillName == "odoo-"+vStr || skillName == "odoo-"+vStr+".0" {
					bridgeName = "odoo"
					break
				}
			}
		}

		dst := filepath.Join(agentSkillsDir, bridgeName)
		// Remove existing to refresh link/copy
		_ = os.RemoveAll(dst)

		relSrc, err := filepath.Rel(agentSkillsDir, src)
		if err != nil {
			if _, err := copyFSTree(os.DirFS(src), ".", dst, nil); err != nil {
				return fmt.Errorf("copy overlay skill %q: %w", skillName, err)
			}
			continue
		}

		if err := os.Symlink(relSrc, dst); err != nil {
			if _, err := copyFSTree(os.DirFS(src), ".", dst, nil); err != nil {
				return fmt.Errorf("fallback copy overlay skill %q: %w", skillName, err)
			}
		}
	}
	return nil
}

func unbridgeOverlaySkills(projectRoot string, manifest OverlayManifest) error {
	agentSkillsDir := filepath.Join(projectRoot, ".agent", "skills")
	for _, skillName := range manifest.Skills {
		dst := filepath.Join(agentSkillsDir, skillName)

		// Only remove if it's actually an overlay skill (check if it exists)
		if _, err := os.Lstat(dst); err == nil {
			// Optimization: verify it points to our overlay or is a copy of it
			// For now, simple removal is safer for cleanup.
			_ = os.RemoveAll(dst)
		}
	}
	return nil
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func formatVersionSet(versions map[int]struct{}) string {
	items := make([]int, 0, len(versions))
	for v := range versions {
		items = append(items, v)
	}
	sort.Ints(items)

	parts := make([]string, 0, len(items))
	for _, v := range items {
		parts = append(parts, strconv.Itoa(v))
	}
	return strings.Join(parts, ",")
}

// parseVersionIntent is the inverse of formatVersionSet. It accepts
// "19", "18,19", "18, 19" or "18 19" and returns the set of major versions.
// Empty input returns an empty (nil-safe) map — callers treat that as "no filter".
func parseVersionIntent(intent string) map[int]struct{} {
	trimmed := strings.TrimSpace(intent)
	if trimmed == "" {
		return nil
	}
	result := make(map[int]struct{})
	for _, part := range strings.FieldsFunc(trimmed, func(r rune) bool {
		return r == ',' || r == ' '
	}) {
		if v, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
			result[v] = struct{}{}
		}
	}
	return result
}

func resolveEnterpriseRepoPath(explicitPath string) string {
	if explicitPath != "" {
		if info, err := os.Stat(explicitPath); err == nil && info.IsDir() {
			return explicitPath
		}
		return ""
	}

	// Check ODOO_ENTERPRISE_PATH env var before falling back to home-dir heuristics.
	// This enables portable team setups where each developer sets their own path.
	if p := strings.TrimSpace(os.Getenv("ODOO_ENTERPRISE_PATH")); p != "" {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return p
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	candidates := []string{
		filepath.Join(homeDir, "gitproj", "odoo", "enterprise"),
		filepath.Join(homeDir, "gitproj", "odoo-enterprise", "enterprise"),
		filepath.Join(homeDir, "gitproj", "odoo", "addons", "enterprise"),
	}

	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path
		}
	}

	return ""
}

func injectEnterpriseContext(overlayRoot string, enterprisePath string) error {
	targets := make([]string, 0)
	for _, dir := range []string{"skills", "agents", "patterns", "instructions", "prompts"} {
		_ = filepath.WalkDir(filepath.Join(overlayRoot, dir), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() && strings.EqualFold(filepath.Ext(path), ".md") {
				targets = append(targets, path)
			}
			return nil
		})
	}

	for _, path := range targets {
		if err := appendEnterpriseContextBlock(path, enterprisePath); err != nil {
			return err
		}
	}

	return nil
}

func appendEnterpriseContextBlock(path string, enterprisePath string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read markdown for enterprise context %q: %w", path, err)
	}

	content := string(data)
	marker := "## Enterprise Repository Access"
	if strings.Contains(content, marker) {
		return nil
	}

	block := "\n\n## Enterprise Repository Access\n\n" +
		"- Local enterprise repository path: `" + enterprisePath + "`\n" +
		"- Use this repository for enterprise module references, enterprise-only inheritance chains, and parity checks between community and enterprise implementations.\n" +
		"- Keep custom code compatible with the detected Odoo major version in `__manifest__.py`.\n"

	updated := strings.TrimRight(content, "\n") + block
	if err := os.WriteFile(path, []byte(updated+"\n"), 0o644); err != nil {
		return fmt.Errorf("write markdown enterprise context %q: %w", path, err)
	}
	return nil
}

func copyFile(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file %q: %w", src, err)
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create destination parent %q: %w", filepath.Dir(dst), err)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination file %q: %w", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy file %q -> %q: %w", src, dst, err)
	}
	return nil
}

func copyFSFile(sourceFS fs.FS, src string, dst string) error {
	srcFile, err := sourceFS.Open(src)
	if err != nil {
		return fmt.Errorf("open source file %q: %w", src, err)
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create destination parent %q: %w", filepath.Dir(dst), err)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination file %q: %w", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy file %q -> %q: %w", src, dst, err)
	}
	return nil
}

func hasStrongOverlayRepoEvidence(projectRoot string, overlayName string) bool {
	if overlayName != defaultOverlayName {
		return false
	}

	quickHits := []string{"odoo-bin", "odoo.conf"}
	for _, rel := range quickHits {
		if info, err := os.Stat(filepath.Join(projectRoot, rel)); err == nil && !info.IsDir() {
			return true
		}
	}

	found := false
	_ = filepath.WalkDir(projectRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := d.Name()
			if base == ".git" || base == ".atl" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "__manifest__.py" {
			found = true
			return errors.New("found")
		}
		return nil
	})

	return found
}

func RunOverlay(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("overlay command requires a subcommand: install, refresh, or remove")
	}

	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve current working directory: %w", err)
	}

	sub := strings.TrimSpace(args[0])
	switch sub {
	case "install":
		fs := flag.NewFlagSet("overlay install", flag.ContinueOnError)
		fs.SetOutput(ioDiscard{})
		source := fs.String("source", "", "overlay source directory")
		versionIntent := fs.String("ref", "", "overlay version intent")
		enterprisePath := fs.String("enterprise-repo", "", "local Odoo enterprise repository path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return errors.New("usage: architect-ai overlay install <overlay-name> [--source <path>]")
		}

		overlayName := fs.Arg(0)
		manifest, err := InstallOverlay(OverlayInstallOptions{
			OverlayName:     overlayName,
			ProjectRoot:     projectRoot,
			SourcePath:      *source,
			VersionIntent:   *versionIntent,
			EnterprisePath:  *enterprisePath,
			ExplicitRequest: true,
		})
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(stdout, "Installed project-local overlay %q at .atl/overlays/%s\n", manifest.Name, manifest.Name)
		versions, isOdoo, detectErr := detectOdooMajorVersions(projectRoot)
		if detectErr == nil && isOdoo && len(versions) > 0 {
			_, _ = fmt.Fprintf(stdout, "Overlay registry loaded for Odoo versions: %s (agnostic files excluded)\n", formatVersionSet(versions))
		} else {
			_, _ = fmt.Fprintln(stdout, "Overlay files installed. Agnostic skills registered; version-specific skills remain inactive until a __manifest__.py is detected.")
		}
		if manifest.EnterprisePath != "" {
			_, _ = fmt.Fprintf(stdout, "Enterprise repository context enabled: %s\n", manifest.EnterprisePath)
		}
		return nil

	case "refresh":
		fs := flag.NewFlagSet("overlay refresh", flag.ContinueOnError)
		fs.SetOutput(ioDiscard{})
		enterprisePath := fs.String("enterprise-repo", "", "local Odoo enterprise repository path")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 1 {
			return errors.New("usage: architect-ai overlay refresh <overlay-name>")
		}

		overlayName := fs.Arg(0)
		manifest, err := RefreshOverlay(projectRoot, overlayName, *enterprisePath)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(stdout, "Refreshed project-local overlay %q from %s\n", manifest.Name, manifest.SourcePath)
		return nil

	case "remove":
		if len(args) != 2 {
			return errors.New("usage: architect-ai overlay remove <overlay-name>")
		}
		overlayName := strings.TrimSpace(args[1])
		if err := RemoveOverlay(projectRoot, overlayName); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(stdout, "Removed project-local overlay %q\n", overlayName)
		return nil
	case "enable":
		if len(args) != 3 {
			return errors.New("usage: architect-ai overlay enable <overlay-name> <skill-name>")
		}
		overlayName := strings.TrimSpace(args[1])
		skillName := strings.TrimSpace(args[2])
		manifest, err := EnableOverlaySkill(projectRoot, overlayName, skillName)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(stdout, "Enabled optional skill %q in overlay %q\n", skillName, manifest.Name)
		return nil
	default:
		return fmt.Errorf("unknown overlay subcommand %q", sub)
	}
}

func EnableOverlaySkill(projectRoot string, overlayName string, skillName string) (OverlayManifest, error) {
	overlayRoot := filepath.Join(projectRoot, ".atl", "overlays", overlayName)
	manifestPath := filepath.Join(overlayRoot, "manifest.json")
	manifest, err := readOverlayManifest(manifestPath)
	if err != nil {
		return OverlayManifest{}, err
	}

	isOptional := false
	for _, s := range manifest.OptionalSkills {
		if s == skillName {
			isOptional = true
			break
		}
	}
	if !isOptional {
		return OverlayManifest{}, fmt.Errorf("skill %q is not an optional skill for overlay %q", skillName, overlayName)
	}

	for _, s := range manifest.EnabledOptionals {
		if s == skillName {
			// Already enabled, just re-bridge to be sure
			_ = bridgeOverlaySkills(projectRoot, manifest)
			return manifest, nil
		}
	}

	manifest.EnabledOptionals = append(manifest.EnabledOptionals, skillName)
	manifest.Skills = append(manifest.Skills, skillName)
	manifest.Skills = uniqueStrings(manifest.Skills)
	sort.Strings(manifest.Skills)

	if err := writeOverlayManifest(manifestPath, manifest); err != nil {
		return OverlayManifest{}, err
	}

	if err := WriteLocalSkillRegistry(projectRoot); err != nil {
		return OverlayManifest{}, err
	}

	if err := bridgeOverlaySkills(projectRoot, manifest); err != nil {
		return OverlayManifest{}, fmt.Errorf("bridge optional skill: %w", err)
	}

	return manifest, nil
}
func populateRegistryEntries(manifest *OverlayManifest, projectRoot string) {
	overlayRoot := filepath.Join(projectRoot, ".atl", "overlays", manifest.Name)
	
	for _, skillName := range manifest.Skills {
		if skillName == manifest.Name {
			continue
		}
		
		skillPath := filepath.Join(overlayRoot, "skills", skillName, "SKILL.md")
		if _, err := os.Stat(skillPath); err != nil {
			continue
		}
		
		// Use the same logic as collectOverlayContent: prefer .agent/skills/ symlink path
		agentSkillPath := filepath.Join(projectRoot, ".agent", "skills", skillName, "SKILL.md")
		if _, err := os.Stat(agentSkillPath); err == nil {
			skillPath = agentSkillPath
		}
		
		info := parseSkillFile(skillPath)
		if info.Name == "" {
			info.Name = skillName
		}
		
		manifest.RegistryEntries = append(manifest.RegistryEntries, RegistryEntry{
			Skill:   info.Name,
			Trigger: info.Trigger,
			Path:    skillPath,
		})
	}
}
