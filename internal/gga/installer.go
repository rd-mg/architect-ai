package gga

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rd-mg/architect-ai/internal/assets"
)

// Config holds detected project configuration for GGA hook generation
type Config struct {
	RepoDir     string
	IsOdoo      bool
	OdooVersion string
	HasCudioGit bool
	Endpoint    string
	Platform    string
}

// Detect inspects the repo to build GGA configuration
func Detect(repoDir string) (*Config, error) {
	cfg := &Config{
		RepoDir:  repoDir,
		Endpoint: envOr("GGA_ENDPOINT", "http://localhost:8765/audit"),
		Platform: detectIDE(repoDir),
	}
	cfg.IsOdoo, cfg.OdooVersion = detectOdoo(repoDir)
	cfg.HasCudioGit = detectCudioGit(repoDir)
	return cfg, nil
}

// Install writes the pre-commit hook appropriate for the OS
func Install(cfg *Config) error {
	hookDir := filepath.Join(cfg.RepoDir, ".git", "hooks")
	if err := os.MkdirAll(hookDir, 0755); err != nil {
		return fmt.Errorf("create hooks dir: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(cfg.RepoDir, ".gga"), 0755); err != nil {
		return fmt.Errorf("create .gga dir: %w", err)
	}

	var hookPath, hookContent string
	if runtime.GOOS == "windows" {
		hookPath = filepath.Join(hookDir, "pre-commit.ps1")
		hookContent = renderPowerShell(cfg)
	} else {
		hookPath = filepath.Join(hookDir, "pre-commit")
		hookContent = renderBash(cfg)
	}

	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		return fmt.Errorf("write hook: %w", err)
	}

	return writeGGAConfig(cfg)
}

func writeGGAConfig(cfg *Config) error {
	content := fmt.Sprintf(`# .gga/config
GGA_VERSION=2.0
IS_ODOO=%v
ODOO_VERSION=%s
HAS_CUDIO_GIT=%v
ENDPOINT=%s
`,
		cfg.IsOdoo, cfg.OdooVersion, cfg.HasCudioGit, cfg.Endpoint)
	return os.WriteFile(filepath.Join(cfg.RepoDir, ".gga", "config"), []byte(content), 0644)
}

func detectOdoo(dir string) (bool, string) {
	version := "unknown"
	var manifestFound string

	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if !d.IsDir() && d.Name() == "__manifest__.py" {
			manifestFound = p
			return filepath.SkipAll
		}
		return nil
	})

	if manifestFound == "" {
		// Check requirements.txt
		req, _ := os.ReadFile(filepath.Join(dir, "requirements.txt"))
		if strings.Contains(string(req), "odoo") {
			return true, "unknown"
		}
		return false, ""
	}

	content, _ := os.ReadFile(manifestFound)
	// Try single quote and double quote matching for version
	for _, line := range strings.Split(string(content), "\n") {
		if strings.Contains(line, "version") {
			// Find version number like 'version': '18.0.1.0.0' or "version": "18.0.1.0"
			parts := strings.FieldsFunc(line, func(r rune) bool {
				return r == '\'' || r == '"'
			})
			for _, p := range parts {
				if len(p) > 2 && strings.Contains(p, ".") {
					// Extract main version e.g. "18.0" -> "18"
					versionParts := strings.Split(p, ".")
					if len(versionParts) > 0 {
						version = versionParts[0]
						return true, version
					}
				}
			}
		}
	}
	return true, version
}

func detectCudioGit(dir string) bool {
	paths := []string{
		filepath.Join(dir, "cudio-git.md"),
		filepath.Join(dir, ".atl", "overlays", "odoo-development-skill", "rules", "cudio-git.md"),
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

func detectIDE(dir string) string {
	signals := map[string]string{
		"opencode.json":                   "opencode",
		"CLAUDE.md":                       "claude",
		".github/copilot-instructions.md": "cursor",
		"GEMINI.md":                       "gemini",
	}
	for file, ide := range signals {
		if _, err := os.Stat(filepath.Join(dir, file)); err == nil {
			return ide
		}
	}
	return "generic"
}

func renderBash(cfg *Config) string {
	content := assets.MustRead("gga/pre-commit.bash.tpl")
	// Replace default endpoint with the configured one
	content = strings.Replace(content, `GGA_ENDPOINT="${GGA_ENDPOINT:-http://localhost:8765/audit}"`, fmt.Sprintf(`GGA_ENDPOINT="${GGA_ENDPOINT:-%s}"`, cfg.Endpoint), 1)
	return content
}

func renderPowerShell(cfg *Config) string {
	content := assets.MustRead("gga/pre-commit.ps1.tpl")
	// Replace default endpoint with the configured one
	content = strings.Replace(content, `$GgaEndpoint = $env:GGA_ENDPOINT ?? "http://localhost:8765/audit"`, fmt.Sprintf(`$GgaEndpoint = $env:GGA_ENDPOINT ?? "%s"`, cfg.Endpoint), 1)
	return content
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
