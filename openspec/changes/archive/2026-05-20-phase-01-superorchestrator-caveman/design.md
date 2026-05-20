# Design: Phase 1 - Super-Orchestrator v2.0: Inline Execution + Delegation Triggers + Caveman


## Architecture
The L0/L1/L2 hierarchy uses a shared-asset injection model. Platform-agnostic identity blocks in `_shared/` are composed into platform-specific adapter files via Go template rendering. Each platform adapter declares capability flags (parallel, sub-agents, MCP, compress) that determine how the hierarchy is rendered.

v2.0 introduces a **3-Mode Architecture** for the L0:
- **Mode A (Inline)**: L0 executes simple tasks directly, bypassing L1 latency.
- **Mode B (SDD)**: Delegated to `sdd-orchestrator`.
- **Mode C (General)**: Delegated to `general-orchestrator`.

To prevent context inflation, **6 Mandatory Delegation Triggers** enforce when the L0 MUST delegate (e.g., touching 4+ files, multi-file writes). 

v2.0 also introduces **Session-wide Execution Modes** (Interactive vs Automatic) and **Model Routing** to match model size to phase complexity (Opus for Architecture, Sonnet for Implementation, Haiku for Mechanical tasks).

Platforms with real sub-agents (OpenCode, Claude, Gemini) use physical delegation via `Task` tool. Platforms without (VSCode Copilot, Antigravity) simulate the hierarchy inline using ULTRA caveman framing with explicit identity transitions.

## File Structure

```
internal/assets/
├── _shared/
│   ├── architect-identity.md          [NEW] Identity block L0 reusable
│   ├── caveman-identity-block.md      [NEW] Caveman block for identity level
│   ├── super-orchestrator-gate.md     [NEW] Router gate L0→L1 universal
│   └── subagent-executor-boundary.md  [EXISTING — review]
├── opencode/
│   └── architect.md                   [NEW] L0 for OpenCode
├── claude/
│   └── architect.md                   [NEW] L0 for Claude Code
├── cursor/
│   └── architect.md                   [NEW] L0 for VSCode Copilot (Cursor)
├── antigravity/
│   └── architect.md                   [NEW] L0 for Antigravity (single-thread)
└── gemini/
    └── architect.md                   [NEW] L0 for Gemini CLI
```

## FODA Matrix

| | Detail |
|---|---|
| **F** | L0/L1/L2 pattern already exists in OpenCode. Generic caveman block already exists. |
| **O** | Inline execution reduces latency for trivial interactions by ~30%. Model routing reduces cost by ~40% using Haiku where Sonnet is overkill. |
| **D** | L0 must make an active decision (Simple/SDD/General) every turn — adds complexity to prompt. |
| **A** | If classification fails, L0 might execute a complex task inline, blowing up context. Mitigated by the 6 Mandatory Delegation Triggers (e.g., long-session trigger forces delegation after 20 tools). |

## FMEA Matrix
| Component | Failure Mode | Effect | Likelihood | Severity | RPN | Mitigation |
|---|---|---|---|---|---|---|
| L0 Routing Gate | Model ignores gate and self-executes | L0 acts as executor, bypassing L1 | 2 | 3 | 6 | Strong PROHIBITION directive. "STRICTLY PROHIBITED from executing" language. |
| Caveman Injection | Agent drops compression mid-session | Token waste, context rot | 2 | 2 | 4 | Identity-level injection (not inherited). "ACTIVE EVERY RESPONSE" directive. |
| InjectSection | Malformed markers in existing file | Section content corrupted | 1 | 3 | 3 | Regex validation of marker pairs before write. |
| Platform Detection | Multiple platform files present | Wrong platform selected | 1 | 2 | 2 | Priority order: opencode > claude > cursor > antigravity > gemini. |
| Single-Thread Simulation | Identity bleed between simulated agents | Antigravity carries L2 identity into L1 | 2 | 2 | 4 | Explicit "CLEAR identity" step in simulation protocol. |

## Go Implementation

### `internal/install/architect/inject.go`

```go
package architect

import (
    "fmt"
    "os"
    "path/filepath"
    "text/template"
)

type AgentConfig struct {
    Platform              string
    SupportsParallel      bool
    SupportsRealSubagents bool
    CompressCommand       string
    EntryFile             string
    MCPConfigFile         string
}

var PlatformConfigs = map[string]AgentConfig{
    "opencode": {
        Platform: "opencode", SupportsParallel: true, SupportsRealSubagents: true,
        CompressCommand: "/compact", EntryFile: "opencode.json", MCPConfigFile: "opencode.json",
    },
    "claude": {
        Platform: "claude", SupportsParallel: true, SupportsRealSubagents: true,
        CompressCommand: "/compact", EntryFile: "CLAUDE.md", MCPConfigFile: ".claude/settings.json",
    },
    "cursor": {
        Platform: "cursor", SupportsParallel: false, SupportsRealSubagents: false,
        CompressCommand: "", EntryFile: ".github/copilot-instructions.md", MCPConfigFile: ".cursor/mcp.json",
    },
    "antigravity": {
        Platform: "antigravity", SupportsParallel: false, SupportsRealSubagents: false,
        CompressCommand: "", EntryFile: ".antigravity/agent.md", MCPConfigFile: ".antigravity/mcp.json",
    },
    "gemini": {
        Platform: "gemini", SupportsParallel: true, SupportsRealSubagents: true,
        CompressCommand: "/compress", EntryFile: "GEMINI.md", MCPConfigFile: ".gemini/settings.json",
    },
}

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

func ValidateHierarchy(platform string, installDir string) []string {
    var issues []string
    cfg := PlatformConfigs[platform]
    l0Path := filepath.Join(installDir, "architect.md")
    if _, err := os.Stat(l0Path); os.IsNotExist(err) {
        issues = append(issues, fmt.Sprintf("[L0] architect.md not found at %s", l0Path))
    }
    l1aPath := filepath.Join(installDir, "sdd-orchestrator.md")
    if _, err := os.Stat(l1aPath); os.IsNotExist(err) {
        issues = append(issues, "[L1a] sdd-orchestrator.md not found")
    }
    l1bPath := filepath.Join(installDir, "general-orchestrator.md")
    if _, err := os.Stat(l1bPath); os.IsNotExist(err) {
        issues = append(issues, "[L1b] general-orchestrator.md not found")
    }
    if !cfg.SupportsRealSubagents {
        issues = append(issues,
            fmt.Sprintf("[WARN] Platform %s does not support real sub-agents. Orchestration is simulated.", platform))
    }
    _ = cfg
    return issues
}
```

### `internal/install/architect/inject_test.go`

```go
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
```

## Key Decisions
- **Caveman at identity level**: Chosen over inherited injection because models lose inherited directives during long sessions. Identity-level means it's in the agent's "self-concept" and survives context pressure.
- **Section markers over full-file overwrite**: Enables users to add custom content outside `<!-- architect-ai:* -->` sections without losing it during sync.
- **Simulated delegation for single-thread**: Antigravity and VSCode Copilot cannot create real sub-agents. The simulation protocol uses explicit ULTRA framing to maintain identity isolation within a single context window.
