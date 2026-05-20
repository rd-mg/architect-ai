package architect

import (
    "os"
    "path/filepath"
    "testing"
)

func TestValidateHierarchy_AllPlatforms(t *testing.T) {
    platforms := []string{"opencode", "claude", "cursor", "antigravity", "gemini"}
    
    for _, platform := range platforms {
        t.Run(platform, func(t *testing.T) {
            _, ok := PlatformConfigs[platform]
            if !ok {
                t.Fatalf("platform %s not found in PlatformConfigs", platform)
            }
        })
    }
}

func TestPlatformConfigs_ParallelCapabilities(t *testing.T) {
    parallelPlatforms := []string{"opencode", "claude", "gemini"}
    sequentialPlatforms := []string{"cursor", "antigravity"}
    
    for _, p := range parallelPlatforms {
        cfg := PlatformConfigs[p]
        if !cfg.SupportsParallel {
            t.Errorf("platform %s should support parallel but does not", p)
        }
    }
    
    for _, p := range sequentialPlatforms {
        cfg := PlatformConfigs[p]
        if cfg.SupportsParallel {
            t.Errorf("platform %s should NOT support parallel but does", p)
        }
    }
}

func TestValidateHierarchy_MissingFiles(t *testing.T) {
    tmpDir := t.TempDir()
    
    issues := ValidateHierarchy("opencode", tmpDir)
    
    // Expect issues for missing L0, L1a, L1b
    if len(issues) < 2 {
        t.Errorf("expected at least 2 issues for empty dir, got %d: %v", len(issues), issues)
    }
}

func TestValidateHierarchy_Complete(t *testing.T) {
    tmpDir := t.TempDir()
    
    // Create required files
    files := []string{"architect.md", "sdd-orchestrator.md", "general-orchestrator.md"}
    for _, f := range files {
        path := filepath.Join(tmpDir, f)
        if err := os.WriteFile(path, []byte("# test"), 0644); err != nil {
            t.Fatalf("create test file %s: %v", f, err)
        }
    }
    
    issues := ValidateHierarchy("opencode", tmpDir)
    
    // Only expect the parallel/subagent warnings, not missing file errors
    for _, issue := range issues {
        if issue[:4] == "[L0]" || issue[:4] == "[L1a" || issue[:4] == "[L1b" {
            t.Errorf("unexpected missing file issue: %s", issue)
        }
    }
}

func TestInjectArchitect(t *testing.T) {
    platforms := []string{"opencode", "claude", "cursor", "antigravity", "gemini"}
    assetsDir := filepath.Join("..", "..", "assets")
    
    for _, platform := range platforms {
        t.Run(platform, func(t *testing.T) {
            tmpDir := t.TempDir()
            err := InjectArchitect(platform, assetsDir, tmpDir)
            if err != nil {
                t.Fatalf("InjectArchitect(%q) failed: %v", platform, err)
            }
            
            cfg := PlatformConfigs[platform]
            outPath := filepath.Join(tmpDir, cfg.EntryFile)
            if _, err := os.Stat(outPath); err != nil {
                t.Fatalf("expected output file %s to be created: %v", outPath, err)
            }
        })
    }
}

