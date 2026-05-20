package gga

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetect_Generic(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.IsOdoo {
		t.Error("empty dir should not be Odoo")
	}
	if cfg.HasCudioGit {
		t.Error("empty dir should not have cudio-git")
	}
	if cfg.Endpoint == "" {
		t.Error("endpoint should have default value")
	}
}

func TestDetect_OdooManifest(t *testing.T) {
	dir := t.TempDir()
	modDir := filepath.Join(dir, "my_module")
	err := os.MkdirAll(modDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(modDir, "__manifest__.py"),
		[]byte(`{'name':'Test','version':'18.0.1.0.0','depends':['base']}`), 0644)
	if err != nil {
		t.Fatal(err)
	}
	cfg, _ := Detect(dir)
	if !cfg.IsOdoo {
		t.Error("should detect Odoo")
	}
	if cfg.OdooVersion != "18" && cfg.OdooVersion != "unknown" {
		t.Errorf("OdooVersion was %s, expected 18 or unknown", cfg.OdooVersion)
	}
}

func TestDetect_CudioGit(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "cudio-git.md"), []byte("# cudio-git rules"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	cfg, _ := Detect(dir)
	if !cfg.HasCudioGit {
		t.Error("should detect cudio-git")
	}
}

func TestInstall_CreatesHook(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git", "hooks")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &Config{RepoDir: dir, IsOdoo: false, Endpoint: "http://localhost:8765/audit"}
	if err := Install(cfg); err != nil {
		t.Fatalf("Install: %v", err)
	}

	// Verify hook exists
	hookName := "pre-commit"
	if strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
		hookName = "pre-commit.ps1"
	}
	hookPath := filepath.Join(gitDir, hookName)
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		t.Errorf("hook not created: %s", hookPath)
	}
	// Verify .gga/config exists
	if _, err := os.Stat(filepath.Join(dir, ".gga", "config")); os.IsNotExist(err) {
		t.Error(".gga/config not created")
	}
}

func TestRenderBash_ContainsSecretPattern(t *testing.T) {
	cfg := &Config{IsOdoo: false, OdooVersion: "unknown", HasCudioGit: false, Endpoint: "http://localhost"}
	script := renderBash(cfg)
	if !strings.Contains(script, "api_key|secret|password|token") {
		t.Error("bash hook missing secret detection patterns")
	}
	if !strings.Contains(script, "GGA_SKIP") {
		t.Error("bash hook missing skip-ai support")
	}
	if !strings.Contains(script, "CI:-false") {
		t.Error("bash hook missing CI detection")
	}
}

func TestRenderPowerShell_NojqRequired(t *testing.T) {
	cfg := &Config{IsOdoo: false, OdooVersion: "unknown", Endpoint: "http://localhost"}
	script := renderPowerShell(cfg)
	if strings.Contains(script, " jq ") || strings.Contains(script, "| jq") {
		t.Error("PowerShell hook must not use jq — use ConvertFrom-Json instead")
	}
	if !strings.Contains(script, "ConvertTo-Json") && !strings.Contains(script, "Invoke-RestMethod") {
		t.Error("PowerShell hook should use native PS cmdlets")
	}
}
