package overlay

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rd-mg/architect-ai/internal/components/filemerge"
)

// OdooManifest is written to .atl/overlays/odoo-{version}/manifest.json.
// It activates the Odoo overlay in the sdd-orchestrator.
type OdooManifest struct {
	Overlay    string    `json:"overlay"`
	Version    string    `json:"version"`
	Active     bool      `json:"active"`
	AddonsPath string    `json:"addons_path"`
	DetectedAt time.Time `json:"detected_at"`
}

// DetectOdoo checks for Odoo markers in the given project directory.
// Returns (detected bool, version string, addonsPath string).
func DetectOdoo(projectDir string) (bool, string, string) {
	// Signal 1: __manifest__.py with version prefix
	manifestFiles, _ := findManifestFiles(projectDir)
	for _, mf := range manifestFiles {
		if version := extractOdooVersion(mf); version != "" {
			addonsPath := filepath.Dir(filepath.Dir(mf)) // parent of module dir
			return true, version, addonsPath
		}
	}

	// Signal 2: pyproject.toml or requirements.txt contains odoo
	for _, f := range []string{"pyproject.toml", "requirements.txt"} {
		content, err := os.ReadFile(filepath.Join(projectDir, f))
		if err == nil && (strings.Contains(string(content), "odoo") ||
			strings.Contains(string(content), "Odoo")) {
			return true, "unknown", projectDir
		}
	}

	// Signal 3: .atl/config.yaml has odoo_version
	cfg, err := os.ReadFile(filepath.Join(projectDir, ".atl", "config.yaml"))
	if err == nil && strings.Contains(string(cfg), "odoo_version") {
		version := extractVersionFromConfig(string(cfg))
		return true, version, projectDir
	}

	return false, "", ""
}

// InstallOverlayManifest creates .atl/overlays/odoo-{version}/manifest.json.
// This file is the activation marker that sdd-orchestrator looks for.
func InstallOverlayManifest(projectDir, version, addonsPath string, dryRun bool) error {
	overlayDir := filepath.Join(projectDir, ".atl", "overlays",
		fmt.Sprintf("odoo-%s", version))

	if dryRun {
		fmt.Printf("  DRY RUN: would create %s/manifest.json\n", overlayDir)
		return nil
	}

	if err := os.MkdirAll(overlayDir, 0o755); err != nil {
		return fmt.Errorf("create overlay dir: %w", err)
	}

	manifest := OdooManifest{
		Overlay:    "odoo-development-skill",
		Version:    version,
		Active:     true,
		AddonsPath: addonsPath,
		DetectedAt: time.Now().UTC(),
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	manifestPath := filepath.Join(overlayDir, "manifest.json")
	_, err = filemerge.WriteFileAtomic(manifestPath, append(data, '\n'), 0o644)
	if err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	fmt.Printf("  Odoo %s overlay activated: %s\n", version, manifestPath)
	return nil
}

// UpdateATLConfig adds Odoo config keys to .atl/config.yaml.
func UpdateATLConfig(projectDir, version, addonsPath string, dryRun bool) error {
	configPath := filepath.Join(projectDir, ".atl", "config.yaml")
	if dryRun {
		fmt.Printf("  DRY RUN: would update %s with Odoo config\n", configPath)
		return nil
	}

	odooConfig := fmt.Sprintf(`
# Odoo configuration (auto-detected by architect-ai)
odoo_version: "%s"
odoo_addons_path: "%s"
`, version, addonsPath)

	existing, _ := os.ReadFile(configPath)
	if strings.Contains(string(existing), "odoo_version") {
		return nil // already configured
	}

	updated := string(existing) + odooConfig
	_, err := filemerge.WriteFileAtomic(configPath, []byte(updated), 0o644)
	return err
}

func findManifestFiles(projectDir string) ([]string, error) {
	var results []string
	out, err := exec.Command("rg", "--files", "--glob", "*/__manifest__.py",
		"--max-depth", "6", projectDir).Output()
	if err != nil {
		// rg not available or no results
		return results, nil
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			results = append(results, line)
		}
	}
	return results, nil
}

func extractOdooVersion(manifestPath string) string {
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return ""
	}
	text := string(content)
	// Look for "version": "17.0.x.y.z" or "version": "18.0.x.y.z"
	for _, prefix := range []string{"17.0", "18.0", "16.0", "15.0", "14.0"} {
		if strings.Contains(text, prefix) {
			return strings.Split(prefix, ".")[0] // return major version: "17", "18", etc.
		}
	}
	return ""
}

func extractVersionFromConfig(config string) string {
	for _, line := range strings.Split(config, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "odoo_version:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			}
		}
	}
	return "unknown"
}
