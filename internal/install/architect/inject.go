package architect

import (
    "fmt"
    "os"
    "path/filepath"
    "text/template"
)

// AgentConfig defines the L0 architect agent configuration per platform
type AgentConfig struct {
    Platform        string
    SupportsParallel bool
    SupportsRealSubagents bool
    CompressCommand  string
    EntryFile        string
    MCPConfigFile    string
}

var PlatformConfigs = map[string]AgentConfig{
    "opencode": {
        Platform:              "opencode",
        SupportsParallel:      true,
        SupportsRealSubagents: true,
        CompressCommand:       "/compact",
        EntryFile:             "opencode.json",
        MCPConfigFile:         "opencode.json",
    },
    "claude": {
        Platform:              "claude",
        SupportsParallel:      true,
        SupportsRealSubagents: true,
        CompressCommand:       "/compact",
        EntryFile:             "CLAUDE.md",
        MCPConfigFile:         ".claude/settings.json",
    },
    "cursor": {
        Platform:              "cursor",
        SupportsParallel:      false,
        SupportsRealSubagents: false,
        CompressCommand:       "",
        EntryFile:             ".github/copilot-instructions.md",
        MCPConfigFile:         ".cursor/mcp.json",
    },
    "antigravity": {
        Platform:              "antigravity",
        SupportsParallel:      false,
        SupportsRealSubagents: false,
        CompressCommand:       "",
        EntryFile:             ".antigravity/agent.md",
        MCPConfigFile:         ".antigravity/mcp.json",
    },
    "gemini": {
        Platform:              "gemini",
        SupportsParallel:      true,
        SupportsRealSubagents: true,
        CompressCommand:       "/compress",
        EntryFile:             "GEMINI.md",
        MCPConfigFile:         ".gemini/settings.json",
    },
}

// InjectArchitect generates the L0 architect agent file for the given platform
func InjectArchitect(platform string, assetsDir string, outputDir string) error {
    cfg, ok := PlatformConfigs[platform]
    if !ok {
        return fmt.Errorf("unknown platform: %s", platform)
    }

    templatePath := filepath.Join(assetsDir, platform, "architect.md")
    tmpl, err := template.ParseFiles(templatePath)
    if err != nil {
        return fmt.Errorf("parse architect template for %s: %w", platform, err)
    }

    outPath := filepath.Join(outputDir, cfg.EntryFile)
    if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
        return fmt.Errorf("create output dir: %w", err)
    }

    f, err := os.Create(outPath)
    if err != nil {
        return fmt.Errorf("create output file: %w", err)
    }
    defer f.Close()

    return tmpl.Execute(f, cfg)
}

// ValidateHierarchy checks that L0, L1a, L1b are all present and properly configured
func ValidateHierarchy(platform string, installDir string) []string {
    var issues []string
    
    cfg := PlatformConfigs[platform]
    
    // Check L0 exists
    l0Path := filepath.Join(installDir, "architect.md")
    if _, err := os.Stat(l0Path); os.IsNotExist(err) {
        issues = append(issues, fmt.Sprintf("[L0] architect.md not found at %s", l0Path))
    }
    
    // Check L1a exists
    l1aPath := filepath.Join(installDir, "sdd-orchestrator.md")
    if _, err := os.Stat(l1aPath); os.IsNotExist(err) {
        issues = append(issues, fmt.Sprintf("[L1a] sdd-orchestrator.md not found"))
    }
    
    // Check L1b exists
    l1bPath := filepath.Join(installDir, "general-orchestrator.md")
    if _, err := os.Stat(l1bPath); os.IsNotExist(err) {
        issues = append(issues, fmt.Sprintf("[L1b] general-orchestrator.md not found"))
    }
    
    // Platform-specific validation
    if !cfg.SupportsRealSubagents {
        issues = append(issues, 
            fmt.Sprintf("[WARN] Platform %s does not support real sub-agents. Orchestration is simulated.", platform))
    }
    
    _ = cfg // suppress unused warning
    return issues
}
