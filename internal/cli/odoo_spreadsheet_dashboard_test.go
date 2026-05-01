package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// TestParseSkillFileTriggerFrontmatter
//
// parseSkillFile must read the explicit `trigger:` YAML frontmatter key and
// return it verbatim. Previously the skill used only `description:` and the
// trigger was extracted via heuristic fallback.
// ---------------------------------------------------------------------------

func TestParseSkillFileTriggerFrontmatter(t *testing.T) {
	dir := t.TempDir()
	skillPath := filepath.Join(dir, "SKILL.md")

	content := `---
name: odoo-spreadsheet-dashboard
trigger: When the user asks to create or edit an Odoo 19 osheet dashboard. Triggers on 'odoo dashboard', 'osheet'.
description: Design and validate Odoo 19 native dashboard spreadsheet files.
license: MIT
---

# Odoo 19 Spreadsheet Dashboard Skill

## Key rules

1. Resolve PIVOT/LIST formulas by both UUID dict key AND numeric formulaId.
2. Always wrap: IFERROR(PIVOT.VALUE(...), 0).
`

	if err := os.WriteFile(skillPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	entry := parseSkillFile(skillPath)

	if entry.Name != "odoo-spreadsheet-dashboard" {
		t.Errorf("Name = %q, want %q", entry.Name, "odoo-spreadsheet-dashboard")
	}

	// Trigger must come from the explicit `trigger:` key, not a description heuristic.
	if !strings.Contains(entry.Trigger, "odoo dashboard") {
		t.Errorf("Trigger = %q — expected it to contain 'odoo dashboard'", entry.Trigger)
	}
	if !strings.Contains(entry.Trigger, "osheet") {
		t.Errorf("Trigger = %q — expected it to contain 'osheet'", entry.Trigger)
	}

	// CompactRules must capture the Key rules section.
	if !strings.Contains(entry.CompactRules, "PIVOT") {
		t.Errorf("CompactRules = %q — expected it to capture the Key rules section", entry.CompactRules)
	}
}

// ---------------------------------------------------------------------------
// TestParseSkillFileTriggerFallbackToDescription
//
// When `trigger:` is absent, parseSkillFile should still populate the trigger
// field from the `description:` value (existing behaviour, regression guard).
// ---------------------------------------------------------------------------

func TestParseSkillFileTriggerFallbackToDescription(t *testing.T) {
	dir := t.TempDir()
	skillPath := filepath.Join(dir, "SKILL.md")

	content := `---
name: some-skill
description: Use this skill for building Odoo modules.
---

# Some Skill
`
	if err := os.WriteFile(skillPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	entry := parseSkillFile(skillPath)

	if entry.Name != "some-skill" {
		t.Errorf("Name = %q, want %q", entry.Name, "some-skill")
	}
	if entry.Trigger == "" {
		t.Errorf("Trigger is empty — expected fallback from description")
	}
	if !strings.Contains(strings.ToLower(entry.Trigger), "odoo") {
		t.Errorf("Trigger = %q — expected description content as fallback", entry.Trigger)
	}
}

// ---------------------------------------------------------------------------
// TestInstallOverlayClassifiesDashboardSkillAsOptional
//
// `odoo-spreadsheet-dashboard` must appear in OptionalSkills (not Skills) when
// the embedded overlay is installed. This guards the optionalSkillSet entry
// added in overlay.go.
// ---------------------------------------------------------------------------

func TestInstallOverlayClassifiesDashboardSkillAsOptional(t *testing.T) {
	projectRoot := t.TempDir()

	manifest, err := InstallOverlay(OverlayInstallOptions{
		ProjectRoot:     projectRoot,
		OverlayName:     defaultOverlayName,
		ExplicitRequest: true,
	})
	if err != nil {
		t.Fatalf("InstallOverlay() error = %v", err)
	}

	// odoo-spreadsheet-dashboard must be in OptionalSkills.
	foundOptional := false
	for _, s := range manifest.OptionalSkills {
		if s == "odoo-spreadsheet-dashboard" {
			foundOptional = true
			break
		}
	}
	if !foundOptional {
		t.Errorf("odoo-spreadsheet-dashboard not in OptionalSkills; OptionalSkills = %v", manifest.OptionalSkills)
	}

	// Must NOT appear in core Skills (would mean it is bridged by default).
	for _, s := range manifest.Skills {
		if s == "odoo-spreadsheet-dashboard" {
			t.Errorf("odoo-spreadsheet-dashboard is in core Skills — it should be optional-only; Skills = %v", manifest.Skills)
		}
	}
}

// ---------------------------------------------------------------------------
// TestInstallOverlayOtherOptionalSkillsUnchanged
//
// The three pre-existing optional skills must still be optional after the
// addition of odoo-spreadsheet-dashboard.
// ---------------------------------------------------------------------------

func TestInstallOverlayOtherOptionalSkillsUnchanged(t *testing.T) {
	projectRoot := t.TempDir()

	manifest, err := InstallOverlay(OverlayInstallOptions{
		ProjectRoot:     projectRoot,
		OverlayName:     defaultOverlayName,
		ExplicitRequest: true,
	})
	if err != nil {
		t.Fatalf("InstallOverlay() error = %v", err)
	}

	wantOptional := []string{
		"odoo-minimax-xlsx-o-spreadsheets",
		"odoo-module-builder",
		"odoo-quote-calculator",
	}

	optionalSet := make(map[string]bool, len(manifest.OptionalSkills))
	for _, s := range manifest.OptionalSkills {
		optionalSet[s] = true
	}

	for _, want := range wantOptional {
		if !optionalSet[want] {
			t.Errorf("expected %q in OptionalSkills; got %v", want, manifest.OptionalSkills)
		}
	}
}

// ---------------------------------------------------------------------------
// TestEnableOverlaySkillDashboard
//
// EnableOverlaySkill must move odoo-spreadsheet-dashboard from OptionalSkills
// to Skills, then call bridgeOverlaySkills — which copies the skill dir to
// .agent/skills/. We only verify the manifest state and that the SKILL.md
// appears in the agent skills dir afterward.
// ---------------------------------------------------------------------------

func TestEnableOverlaySkillDashboard(t *testing.T) {
	projectRoot := t.TempDir()

	_, err := InstallOverlay(OverlayInstallOptions{
		ProjectRoot:     projectRoot,
		OverlayName:     defaultOverlayName,
		ExplicitRequest: true,
	})
	if err != nil {
		t.Fatalf("InstallOverlay() error = %v", err)
	}

	manifest, err := EnableOverlaySkill(projectRoot, defaultOverlayName, "odoo-spreadsheet-dashboard")
	if err != nil {
		t.Fatalf("EnableOverlaySkill() error = %v", err)
	}

	// After enable, it must be in manifest.Skills.
	foundCore := false
	for _, s := range manifest.Skills {
		if s == "odoo-spreadsheet-dashboard" {
			foundCore = true
			break
		}
	}
	if !foundCore {
		t.Errorf("odoo-spreadsheet-dashboard not in Skills after enable; Skills = %v", manifest.Skills)
	}

	// The SKILL.md must be accessible at the bridged path.
	bridged := filepath.Join(projectRoot, ".agent", "skills", "odoo-spreadsheet-dashboard", "SKILL.md")
	if _, statErr := os.Stat(bridged); statErr != nil {
		t.Errorf("SKILL.md not found at bridged path %q after enable: %v", bridged, statErr)
	}
}

// ---------------------------------------------------------------------------
// TestWriteLocalSkillRegistryIncludesDashboardTrigger
//
// After installing the overlay and enabling the dashboard skill,
// WriteLocalSkillRegistry must produce a registry that contains the
// odoo-spreadsheet-dashboard trigger phrase.
// ---------------------------------------------------------------------------

func TestWriteLocalSkillRegistryIncludesDashboardTrigger(t *testing.T) {
	projectRoot := t.TempDir()

	// Save and restore osUserHomeDir to isolate user skill scanning.
	oldHomeDir := osUserHomeDir
	osUserHomeDir = func() (string, error) { return t.TempDir(), nil }
	defer func() { osUserHomeDir = oldHomeDir }()

	_, err := InstallOverlay(OverlayInstallOptions{
		ProjectRoot:     projectRoot,
		OverlayName:     defaultOverlayName,
		ExplicitRequest: true,
	})
	if err != nil {
		t.Fatalf("InstallOverlay() error = %v", err)
	}

	// Enable the dashboard skill so it is included in the registry.
	if _, err := EnableOverlaySkill(projectRoot, defaultOverlayName, "odoo-spreadsheet-dashboard"); err != nil {
		t.Fatalf("EnableOverlaySkill() error = %v", err)
	}

	if err := WriteLocalSkillRegistry(projectRoot); err != nil {
		t.Fatalf("WriteLocalSkillRegistry() error = %v", err)
	}

	registryPath := filepath.Join(projectRoot, ".atl", "skill-registry.md")
	content, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatalf("ReadFile(skill-registry.md) error = %v", err)
	}

	md := string(content)

	// The skill name must appear.
	if !strings.Contains(md, "odoo-spreadsheet-dashboard") {
		t.Errorf("registry does not contain 'odoo-spreadsheet-dashboard';\nregistry (first 2000 chars):\n%s", truncateStr(md, 2000))
	}

	// At least one osheet-related keyword must be present — it appears either in the
	// trigger column or the compact rules section.
	if !strings.Contains(md, "osheet") && !strings.Contains(md, "odoo dashboard") &&
		!strings.Contains(md, "PIVOT") && !strings.Contains(md, "scorecard") {
		t.Errorf("registry does not contain any osheet/dashboard keyword;\nregistry (first 3000 chars):\n%s", truncateStr(md, 3000))
	}
}

// truncateStr returns the first n bytes of s for readable test output.
func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
