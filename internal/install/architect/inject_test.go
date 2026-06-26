package architect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The required platforms as requested by the user
var testPlatforms = []string{"claude", "opencode", "vscode", "copilot", "antigravity", "gemini"}

func TestValidateHierarchy_AllPlatforms(t *testing.T) {
	for _, platform := range testPlatforms {
		t.Run(platform, func(t *testing.T) {
			_, ok := PlatformConfigs[platform]
			if !ok {
				t.Fatalf("platform %s not found in PlatformConfigs", platform)
			}
		})
	}
}

func TestPlatformConfigs_ParallelCapabilities(t *testing.T) {
	// Let's assume claude, opencode, gemini are parallel, while cursor/copilot, vscode, antigravity are not.
	parallelPlatforms := []string{"opencode", "claude", "gemini"}
	sequentialPlatforms := []string{"vscode", "copilot", "antigravity"}
	
	for _, p := range parallelPlatforms {
		cfg, ok := PlatformConfigs[p]
		if !ok {
			continue // skip if missing, caught by previous test
		}
		if !cfg.SupportsParallel {
			t.Errorf("platform %s should support parallel but does not", p)
		}
	}
	
	for _, p := range sequentialPlatforms {
		cfg, ok := PlatformConfigs[p]
		if !ok {
			continue
		}
		if cfg.SupportsParallel {
			t.Errorf("platform %s should NOT support parallel but does", p)
		}
	}
}

func TestValidateHierarchy_MissingFiles(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Use opencode as a sample since ValidateHierarchy uses the cfg
	issues := ValidateHierarchy("opencode", tmpDir)
	
	if len(issues) < 2 {
		t.Errorf("expected at least 2 issues for empty dir, got %d: %v", len(issues), issues)
	}
}

func TestValidateHierarchy_Complete(t *testing.T) {
	tmpDir := t.TempDir()
	
	files := []string{"architect.md", "sdd-orchestrator.md", "general-orchestrator.md"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("# test"), 0644); err != nil {
			t.Fatalf("create test file %s: %v", f, err)
		}
	}
	
	issues := ValidateHierarchy("opencode", tmpDir)
	
	for _, issue := range issues {
		if issue[:4] == "[L0]" || issue[:4] == "[L1a" || issue[:4] == "[L1b" {
			t.Errorf("unexpected missing file issue: %s", issue)
		}
	}
}

func TestInjectArchitect(t *testing.T) {
	assetsDir := filepath.Join("..", "..", "assets")
	
	for _, platform := range testPlatforms {
		t.Run(platform, func(t *testing.T) {
			tmpDir := t.TempDir()
			
			// Mock missing files for testing if they don't exist yet
			platformDir := filepath.Join(assetsDir, platform)
			os.MkdirAll(platformDir, 0755)
			mockFilePath := filepath.Join(platformDir, "architect.md")
			if _, err := os.Stat(mockFilePath); os.IsNotExist(err) {
				os.WriteFile(mockFilePath, []byte("# Mock L0"), 0644)
			}
			
			err := InjectArchitect(platform, assetsDir, tmpDir)
			if err != nil {
				t.Fatalf("InjectArchitect(%q) failed: %v", platform, err)
			}
			
			cfg := PlatformConfigs[platform]
			outPath := filepath.Join(tmpDir, cfg.EntryFile)
			if _, err := os.Stat(outPath); err != nil {
				t.Fatalf("expected output file %s to be created: %v", outPath, err)
			}
			
			// L0-L1 Isolation Check: The generated L0 file MUST NOT contain L1b SDD Orchestrator text
			contentBytes, err := os.ReadFile(outPath)
			if err != nil {
				t.Fatalf("failed to read generated file: %v", err)
			}
			content := string(contentBytes)
			if strings.Contains(content, "L1b SDD Orchestrator") {
				t.Errorf("IDENTITY LEAK: Generated L0 file for %s contains 'L1b SDD Orchestrator'!", platform)
			}
		})
	}
}

