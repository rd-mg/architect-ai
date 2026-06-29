package antigravitycli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rd-mg/architect-ai/internal/assets"
	"github.com/rd-mg/architect-ai/internal/components/filemerge"
)

// Install writes all plugin files to ~/.gemini/antigravity-cli/plugins/architect-ai/
// and installs global settings.
func Install(homeDir string, engramBin string, dryRun bool) error {
	pluginDir := filepath.Join(homeDir, ".gemini", "antigravity-cli", "plugins", "architect-ai")
	if !dryRun {
		if err := os.MkdirAll(pluginDir, 0o755); err != nil {
			return fmt.Errorf("create plugin dir: %w", err)
		}
		if err := os.MkdirAll(filepath.Join(pluginDir, "skills"), 0o755); err != nil {
			return fmt.Errorf("create skills dir: %w", err)
		}
	}

	// Write plugin.json
	if err := writeAsset(pluginDir, "plugin.json", "antigravity-cli/plugin.json", dryRun); err != nil {
		return err
	}

	// Write hooks.json (CLI named-group format)
	if err := writeAsset(pluginDir, "hooks.json", "antigravity-cli/hooks.json", dryRun); err != nil {
		return err
	}

	// Write mcp_config.json (resolve ENGRAM_BIN placeholder)
	if err := writeMCPConfig(pluginDir, engramBin, dryRun); err != nil {
		return err
	}

	// Write settings.json to global settings path
	settingsPath := filepath.Join(homeDir, ".gemini", "antigravity-cli", "settings.json")
	if err := mergeSettings(settingsPath, dryRun); err != nil {
		return err
	}

	// Install sidecar for archive cleanup
	sidecarDir := filepath.Join(homeDir, ".gemini", "config", "sidecars", "architect-archive")
	if err := writeSidecar(sidecarDir, dryRun); err != nil {
		return err
	}

	return nil
}

func writeMCPConfig(pluginDir, engramBin string, dryRun bool) error {
	// Load template from embedded assets
	template := assets.MustRead("antigravity-cli/mcp_config.json")
	// Replace ENGRAM_BIN placeholder
	content := make([]byte, len(template))
	copy(content, template)
	replaced := replaceEngramBin(content, engramBin)

	if dryRun {
		fmt.Printf("  DRY RUN: would write %s/mcp_config.json\n", pluginDir)
		return nil
	}
	path := filepath.Join(pluginDir, "mcp_config.json")
	_, err := filemerge.WriteFileAtomic(path, replaced, 0o644)
	return err
}

func replaceEngramBin(content []byte, bin string) []byte {
	// Replace "${ENGRAM_BIN}" literal with resolved binary path
	old := []byte(`"${ENGRAM_BIN}"`)
	new_ := []byte(`"` + bin + `"`)
	return bytes.ReplaceAll(content, old, new_)
}

func mergeSettings(settingsPath string, dryRun bool) error {
	overlay := []byte(assets.MustRead("antigravity-cli/settings.json"))
	if dryRun {
		fmt.Printf("  DRY RUN: would merge into %s\n", settingsPath)
		return nil
	}
	existing, _ := os.ReadFile(settingsPath)
	if len(existing) == 0 {
		existing = []byte("{}")
	}
	merged, err := filemerge.MergeJSONObjects(existing, overlay)
	if err != nil {
		return fmt.Errorf("merge settings: %w", err)
	}
	_, err = filemerge.WriteFileAtomic(settingsPath, merged, 0o644)
	return err
}

func writeSidecar(sidecarDir string, dryRun bool) error {
	content := []byte(assets.MustRead("antigravity-cli/sidecars/archive-cleaner.json"))
	if dryRun {
		fmt.Printf("  DRY RUN: would write sidecar to %s\n", sidecarDir)
		return nil
	}
	if err := os.MkdirAll(sidecarDir, 0o755); err != nil {
		return fmt.Errorf("create sidecar dir: %w", err)
	}
	path := filepath.Join(sidecarDir, "sidecar.json")
	_, err := filemerge.WriteFileAtomic(path, content, 0o644)
	return err
}

func writeAsset(dir, filename, assetPath string, dryRun bool) error {
	content := []byte(assets.MustRead(assetPath))
	if dryRun {
		fmt.Printf("  DRY RUN: would write %s/%s\n", dir, filename)
		return nil
	}
	path := filepath.Join(dir, filename)
	_, err := filemerge.WriteFileAtomic(path, content, 0o644)
	return err
}
