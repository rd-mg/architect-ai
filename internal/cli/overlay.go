package cli

// Package cli — overlay.go — PATCHED VERSION
//
// This is the v3 patched version. The critical fix is the REMOVAL of the
// copyFSDirFiles("skills/", *.md) call that previously copied all loose .md
// files without version filtering. With the v3 restructure, all patterns are
// inside versioned bundle directories (patterns-14/, patterns-18/, etc.)
// which are already handled by copyOverlaySkillBundles() with proper
// version filtering.
//
// NOTE: This file is a TEMPLATE showing the critical change. You MUST
// integrate it with the full overlay.go in your repository. The full file
// includes many helpers, imports, and utility functions that are NOT shown
// here (they are unchanged). Apply ONLY the patch described in the "PATCH"
// section below to your existing overlay.go.

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// NOTE: This file shows ONLY the functions that change.
// Functions not shown (copyFSFile, copyFSTree, copyFSDirFiles,
// bridgeOverlaySkills, detectOdooMajorVersions, matchesOverlaySkillVersion,
// etc.) are UNCHANGED — keep them as they are in your current overlay.go.

// ============================================================================
// PATCH — The `InstallOverlay` function (or equivalent) changes as follows.
// ============================================================================
//
// BEFORE (v2.x — the buggy version):
//
//   func InstallOverlay(sourceFS fs.FS, overlayName string, projectRoot string) error {
//       ...
//
//       // Copy the root SKILL.md
//       if err := copyFSFile(sourceFS, "SKILL.md", filepath.Join(overlayRoot, "SKILL.md")); err != nil {
//           return fmt.Errorf("copy overlay SKILL.md: %w", err)
//       }
//
//       // Copy skill bundles (directories under skills/)
//       bundles, err := copyOverlaySkillBundles(sourceFS, "skills", overlayRoot)
//       if err != nil {
//           return fmt.Errorf("copy skill bundles: %w", err)
//       }
//
//       // ❌ BUG: This copies ALL loose .md files from skills/ as "patterns"
//       //        WITHOUT version filtering. All v14-v19 patterns get installed
//       //        regardless of the project's Odoo version.
//       patternFiles, err := copyFSDirFiles(sourceFS, "skills",
//           filepath.Join(overlayRoot, "patterns"),
//           func(assetPath string, _ fs.DirEntry) bool {
//               return strings.EqualFold(filepath.Ext(assetPath), ".md")
//           })
//       if err != nil {
//           return fmt.Errorf("copy pattern files: %w", err)
//       }
//
//       // Copy agents, instructions, prompts, rules
//       ...
//
//       // Bridge bundles according to detected Odoo version
//       if err := bridgeOverlaySkills(projectRoot, overlayRoot, bundles); err != nil {
//           return fmt.Errorf("bridge overlay skills: %w", err)
//       }
//
//       return nil
//   }
//
// AFTER (v3.0 — the fixed version):

// InstallOverlay installs the overlay into the project.
// v3 change: removed the copyFSDirFiles for loose patterns (see PATCH above).
// All patterns are now inside versioned bundle directories which are already
// handled by copyOverlaySkillBundles with proper version filtering via
// bridgeOverlaySkills.
func InstallOverlay(sourceFS fs.FS, overlayName string, projectRoot string) error {
	overlayRoot, err := resolveOverlayRoot(projectRoot, overlayName)
	if err != nil {
		return fmt.Errorf("resolve overlay root: %w", err)
	}

	// Copy the root SKILL.md
	if err := copyFSFile(sourceFS, "SKILL.md", filepath.Join(overlayRoot, "SKILL.md")); err != nil {
		return fmt.Errorf("copy overlay SKILL.md: %w", err)
	}

	// Copy skill bundles (directories under skills/)
	// This includes the new patterns-14/, patterns-15/, ..., patterns-19/,
	// patterns-agnostic/, and migration-*/ bundles.
	// bridgeOverlaySkills will version-filter them later.
	bundles, err := copyOverlaySkillBundles(sourceFS, "skills", overlayRoot)
	if err != nil {
		return fmt.Errorf("copy skill bundles: %w", err)
	}

	// REMOVED in v3: copyFSDirFiles for loose patterns.
	// All patterns are now inside versioned bundle directories above.
	// No loose .md files should exist under skills/ in the source tree.

	// Copy sdd-supplements (NEW in v3)
	// These are injected into sub-agent prompts by the orchestrator based on
	// the active SDD phase.
	if err := copyFSTree(sourceFS, "sdd-supplements", filepath.Join(overlayRoot, "sdd-supplements")); err != nil {
		// Missing sdd-supplements is not fatal — it's new in v3 and some
		// older overlays may not have it. Log a debug message if desired.
		if !isNotExistError(err) {
			return fmt.Errorf("copy sdd-supplements: %w", err)
		}
	}

	// Copy agents (only the non-SDD independent agents remain)
	if err := copyFSTree(sourceFS, "agents", filepath.Join(overlayRoot, "agents")); err != nil {
		if !isNotExistError(err) {
			return fmt.Errorf("copy agents: %w", err)
		}
	}

	// Copy instructions (Python, XML, manifest)
	if err := copyFSTree(sourceFS, "instructions", filepath.Join(overlayRoot, "instructions")); err != nil {
		if !isNotExistError(err) {
			return fmt.Errorf("copy instructions: %w", err)
		}
	}

	// Copy prompts
	if err := copyFSTree(sourceFS, "prompts", filepath.Join(overlayRoot, "prompts")); err != nil {
		if !isNotExistError(err) {
			return fmt.Errorf("copy prompts: %w", err)
		}
	}

	// Copy rules (now includes cudio-naming, cudio-git if Cudio overlay active)
	if err := copyFSTree(sourceFS, "rules", filepath.Join(overlayRoot, "rules")); err != nil {
		if !isNotExistError(err) {
			return fmt.Errorf("copy rules: %w", err)
		}
	}

	// Detect Odoo version(s) from project manifest(s)
	versions, err := detectOdooMajorVersions(projectRoot)
	if err != nil {
		return fmt.Errorf("detect Odoo versions: %w", err)
	}

	// Bridge bundles according to detected versions.
	// This is where version filtering happens:
	//   - patterns-agnostic/ is always bridged
	//   - patterns-{V}/ is bridged ONLY if V is in the detected versions
	//   - migration-{F}-{T}/ is bridged ONLY if BOTH F and T are detected
	if err := bridgeOverlaySkills(projectRoot, overlayRoot, bundles, versions); err != nil {
		return fmt.Errorf("bridge overlay skills: %w", err)
	}

	return nil
}

// matchesOverlaySkillVersion reports whether a skill bundle's name matches
// any of the detected Odoo versions.
//
// Bundle naming conventions recognized:
//   - "patterns-agnostic" — always matches (version-independent)
//   - "patterns-{V}" — matches only if V is in detectedVersions
//   - "migration-{F}-{T}" — matches only if BOTH F AND T are in detectedVersions
//   - "odoo-{V}.0" — matches only if V is in detectedVersions (legacy bundle style)
//
// This function is called by bridgeOverlaySkills for each bundle.
func matchesOverlaySkillVersion(bundleName string, detectedVersions []int) bool {
	// patterns-agnostic always matches
	if bundleName == "patterns-agnostic" {
		return true
	}

	// Code-review and other non-versioned bundles — match by not having a version suffix
	// (extend this list as needed)
	nonVersionedBundles := map[string]bool{
		"code-review":                      true,
		"odoo-migration":                   true,
		"odoo-module-builder":              true,
		"odoo-minimax-xlsx-o-spreadsheets": true,
		"odoo-quote-calculator":            true,
	}
	if nonVersionedBundles[bundleName] {
		return true
	}

	// patterns-{V} bundle
	if strings.HasPrefix(bundleName, "patterns-") {
		versionPart := strings.TrimPrefix(bundleName, "patterns-")
		v, err := parseVersion(versionPart)
		if err != nil {
			return false
		}
		return containsInt(detectedVersions, v)
	}

	// migration-{F}-{T} bundle
	if strings.HasPrefix(bundleName, "migration-") {
		versionPart := strings.TrimPrefix(bundleName, "migration-")
		parts := strings.Split(versionPart, "-")
		if len(parts) != 2 {
			return false
		}
		from, err1 := parseVersion(parts[0])
		to, err2 := parseVersion(parts[1])
		if err1 != nil || err2 != nil {
			return false
		}
		return containsInt(detectedVersions, from) && containsInt(detectedVersions, to)
	}

	// Legacy odoo-{V}.0 bundle
	if strings.HasPrefix(bundleName, "odoo-") && strings.HasSuffix(bundleName, ".0") {
		versionPart := strings.TrimSuffix(strings.TrimPrefix(bundleName, "odoo-"), ".0")
		v, err := parseVersion(versionPart)
		if err != nil {
			return false
		}
		return containsInt(detectedVersions, v)
	}

	// Unknown bundle naming — bridge by default (conservative)
	return true
}

// Helper: parses an integer version string.
func parseVersion(s string) (int, error) {
	var v int
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

// Helper: checks if an int is in a slice.
func containsInt(slice []int, n int) bool {
	for _, v := range slice {
		if v == n {
			return true
		}
	}
	return false
}

// Helper: checks if an error indicates a file/directory does not exist.
// Use the existing utility from your current overlay.go — this is a stub.
func isNotExistError(err error) bool {
	// In the real overlay.go, this uses errors.Is(err, fs.ErrNotExist) or similar.
	// Keep your existing implementation.
	return false
}

// Stub signatures for functions that already exist in your overlay.go
// DO NOT copy these stubs — use the real implementations from your file.
func resolveOverlayRoot(projectRoot, overlayName string) (string, error) { return "", nil }
func copyFSFile(src fs.FS, srcPath, destPath string) error               { return nil }
func copyFSTree(src fs.FS, srcPath, destPath string) error               { return nil }
func copyOverlaySkillBundles(src fs.FS, skillsDir, overlayRoot string) ([]string, error) {
	return nil, nil
}
func bridgeOverlaySkills(projectRoot, overlayRoot string, bundles []string, versions []int) error {
	return nil
}
func detectOdooMajorVersions(projectRoot string) ([]int, error) { return nil, nil }

// ============================================================================
// DIFF FORMAT (alternative — apply as a `git apply` patch if preferred)
// ============================================================================
//
// --- a/internal/cli/overlay.go
// +++ b/internal/cli/overlay.go
// @@ -{line},{N} +{line},{M} @@ func InstallOverlay(
// -    // Copy loose pattern .md files
// -    patternFiles, err := copyFSDirFiles(sourceFS, "skills",
// -        filepath.Join(overlayRoot, "patterns"),
// -        func(assetPath string, _ fs.DirEntry) bool {
// -            return strings.EqualFold(filepath.Ext(assetPath), ".md")
// -        })
// -    if err != nil {
// -        return fmt.Errorf("copy pattern files: %w", err)
// -    }
// +    // REMOVED in v3: copyFSDirFiles for loose patterns.
// +    // All patterns are now inside versioned bundle directories
// +    // (patterns-14/, patterns-18/, etc.) which copyOverlaySkillBundles
// +    // already handles with version filtering via bridgeOverlaySkills.
//
// +    // Copy sdd-supplements (NEW in v3)
// +    if err := copyFSTree(sourceFS, "sdd-supplements",
// +        filepath.Join(overlayRoot, "sdd-supplements")); err != nil {
// +        if !isNotExistError(err) {
// +            return fmt.Errorf("copy sdd-supplements: %w", err)
// +        }
// +    }
