package odoo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePaths(t *testing.T) {
	tmp, err := os.MkdirTemp("", "atl-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmp)

	atlDir := filepath.Join(tmp, ".atl")
	if err := os.Mkdir(atlDir, 0755); err != nil {
		t.Fatalf("Failed to create .atl dir: %v", err)
	}

	configYAML := `
odoo_community_path: "/mock/community"
odoo_version: "18"
`
	if err := os.WriteFile(filepath.Join(atlDir, "config.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatalf("Failed to write config.yaml: %v", err)
	}

	cfg := ResolvePaths(tmp)
	if cfg.CommunityPath != "/mock/community" {
		t.Errorf("Expected CommunityPath '/mock/community', got '%s'", cfg.CommunityPath)
	}
	if cfg.OdooVersion != "18" {
		t.Errorf("Expected OdooVersion '18', got '%s'", cfg.OdooVersion)
	}

	env := cfg.ToEnvMap()
	if env["ODOO_COMMUNITY_PATH"] != "/mock/community" {
		t.Errorf("Expected ODOO_COMMUNITY_PATH environment mapping, got '%s'", env["ODOO_COMMUNITY_PATH"])
	}
}
