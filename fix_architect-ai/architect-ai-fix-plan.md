# architect-ai — Complete Fix Implementation Plan

**Date:** 2026-06-27  
**Based on:** Full codebase analysis (7 batches, all files)  
**Total fixes:** 30  
**Sprints:** 3 (Critical → High → Medium)

> **How to use this document**  
> Each fix has: exact file paths, exact code/content changes, test assertions,
> and acceptance criteria. No placeholders. Implement in the order shown within each sprint.
> All Go changes require `go test -race ./...` before merge.

---

## Sprint 1 — Critical Fixes (Week 1–2)

### FIX-01: Add `InjectCodeGraph()` to `internal/components/mcp/inject.go`

**Problem:** `codegraph` is in `overlay.go` and `generator.go` but no `InjectCodeGraph()` function
exists. Sync/update operations that use the inject path skip codegraph entirely.

**File:** `internal/components/mcp/inject.go`

Add after `InjectSequentialThinking()`:

```go
// InjectCodeGraph adds the codegraph MCP server to the agent's config.
// It follows the same strategy dispatch as InjectSequentialThinking.
func InjectCodeGraph(homeDir string, adapter agents.Adapter) (InjectionResult, error) {
	if !adapter.SupportsMCP() {
		return InjectionResult{}, nil
	}

	switch adapter.MCPStrategy() {
	case model.StrategySeparateMCPFiles:
		path := adapter.MCPConfigPath(homeDir, string(ServerCodeGraph))
		overlay, err := OverlayFor(adapter.Agent(), ServerCodeGraph, Options{})
		if err != nil {
			return InjectionResult{}, err
		}
		writeResult, err := filemerge.WriteFileAtomic(path, overlay, 0o644)
		if err != nil {
			return InjectionResult{}, err
		}
		return InjectionResult{Changed: writeResult.Changed, Files: []string{path}}, nil

	case model.StrategyMergeIntoSettings:
		settingsPath := adapter.SettingsPath(homeDir)
		if settingsPath == "" {
			return InjectionResult{}, nil
		}
		overlay, err := OverlayFor(adapter.Agent(), ServerCodeGraph, Options{})
		if err != nil {
			return InjectionResult{}, err
		}
		settingsWrite, err := mergeJSONFile(settingsPath, overlay)
		if err != nil {
			return InjectionResult{}, err
		}
		return InjectionResult{Changed: settingsWrite.Changed, Files: []string{settingsPath}}, nil

	case model.StrategyMCPConfigFile:
		path := adapter.MCPConfigPath(homeDir, string(ServerCodeGraph))
		if path == "" {
			return InjectionResult{}, nil
		}
		overlay, err := OverlayFor(adapter.Agent(), ServerCodeGraph, Options{})
		if err != nil {
			return InjectionResult{}, err
		}
		settingsWrite, err := mergeJSONFile(path, overlay)
		if err != nil {
			return InjectionResult{}, err
		}
		return InjectionResult{Changed: settingsWrite.Changed, Files: []string{path}}, nil

	case model.StrategyTOMLFile:
		// Codex uses TOML — codegraph not supported via TOML yet.
		// Reference is in system prompt (agents.md) not config file.
		return InjectionResult{}, nil

	default:
		return InjectionResult{}, fmt.Errorf(
			"codegraph injector does not support MCP strategy %d for agent %q",
			adapter.MCPStrategy(), adapter.Agent())
	}
}
```

Also add `InjectContextMode()` immediately after:

```go
// InjectContextMode adds the context-mode MCP server to the agent's config.
func InjectContextMode(homeDir string, adapter agents.Adapter) (InjectionResult, error) {
	if !adapter.SupportsMCP() {
		return InjectionResult{}, nil
	}

	overlay, err := contextModeOverlay(adapter.Agent())
	if err != nil {
		// context-mode overlay not defined for this agent — skip silently.
		return InjectionResult{}, nil
	}

	switch adapter.MCPStrategy() {
	case model.StrategySeparateMCPFiles:
		path := adapter.MCPConfigPath(homeDir, "context-mode")
		writeResult, err := filemerge.WriteFileAtomic(path, overlay, 0o644)
		if err != nil {
			return InjectionResult{}, err
		}
		return InjectionResult{Changed: writeResult.Changed, Files: []string{path}}, nil

	case model.StrategyMergeIntoSettings:
		settingsPath := adapter.SettingsPath(homeDir)
		if settingsPath == "" {
			return InjectionResult{}, nil
		}
		settingsWrite, err := mergeJSONFile(settingsPath, overlay)
		if err != nil {
			return InjectionResult{}, err
		}
		return InjectionResult{Changed: settingsWrite.Changed, Files: []string{settingsPath}}, nil

	case model.StrategyMCPConfigFile:
		path := adapter.MCPConfigPath(homeDir, "context-mode")
		if path == "" {
			return InjectionResult{}, nil
		}
		settingsWrite, err := mergeJSONFile(path, overlay)
		if err != nil {
			return InjectionResult{}, err
		}
		return InjectionResult{Changed: settingsWrite.Changed, Files: []string{path}}, nil

	case model.StrategyTOMLFile:
		return InjectionResult{}, nil

	default:
		return InjectionResult{}, fmt.Errorf(
			"context-mode injector does not support MCP strategy %d for agent %q",
			adapter.MCPStrategy(), adapter.Agent())
	}
}
```

Add `contextModeOverlay()` to `overlay.go`:

```go
func contextModeOverlay(agent model.AgentID) ([]byte, error) {
	const npmPkg = "@mksglu/context-mode"
	const globalBin = "context-mode"

	npxCmd := func(key string) []byte {
		return []byte(fmt.Sprintf(`{
  %s: {
    "context-mode": {
      "command": "npx",
      "args": ["-y", "%s"]
    }
  }
}`, key, npmPkg))
	}

	switch agent {
	case model.AgentGeminiCLI:
		return []byte(fmt.Sprintf(`{
  "mcpServers": {
    "context-mode": {
      "command": "npx",
      "args": ["-y", "%s"],
      "timeout": 15000
    }
  }
}`, npmPkg)), nil
	case model.AgentOpenCode, model.AgentKilocode:
		return []byte(fmt.Sprintf(`{
  "mcp": {
    "context-mode": {
      "type": "local",
      "command": ["npx", "-y", "%s"],
      "enabled": true
    }
  }
}`, npmPkg)), nil
	case model.AgentVSCodeCopilot, model.AgentCursor:
		return []byte(fmt.Sprintf(`{
  "servers": {
    "context-mode": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "%s"]
    }
  }
}`, npmPkg)), nil
	case model.AgentAntigravity, model.AgentWindsurf, model.AgentQwenCode, model.AgentKiroIDE:
		return []byte(fmt.Sprintf(`{
  "mcpServers": {
    "context-mode": {
      "command": "npx",
      "args": ["-y", "%s"]
    }
  }
}`, npmPkg)), nil
	case model.AgentClaudeCode:
		return []byte(fmt.Sprintf(`{
  "command": "npx",
  "args": ["-y", "%s"]
}`, npmPkg)), nil
	default:
		return nil, fmt.Errorf("context-mode overlay not defined for agent %s", agent)
	}
}
```

**Wire into install pipeline** — in `internal/cli/install.go` (wherever `InjectSequentialThinking`
and `InjectNotebookLM` are called), add:

```go
if _, err := mcp.InjectCodeGraph(homeDir, adapter); err != nil {
    return result, fmt.Errorf("inject codegraph: %w", err)
}
if _, err := mcp.InjectContextMode(homeDir, adapter); err != nil {
    return result, fmt.Errorf("inject context-mode: %w", err)
}
```

**Tests to add** in `internal/components/mcp/inject_test.go`:

```go
func TestInjectCodeGraphSeparateFile(t *testing.T) {
    home := t.TempDir()
    adapter := claude.NewAdapter()
    result, err := InjectCodeGraph(home, adapter)
    if err != nil {
        t.Fatalf("InjectCodeGraph: %v", err)
    }
    if !result.Changed {
        t.Fatal("expected Changed=true on first inject")
    }
    content, _ := os.ReadFile(result.Files[0])
    if !bytes.Contains(content, []byte("@colbymchenry/codegraph")) {
        t.Fatal("codegraph package not in output")
    }
}

func TestInjectContextModeAllAgents(t *testing.T) {
    agents := []agents.Adapter{
        opencode.NewAdapter(), gemini.NewAdapter(), cursor.NewAdapter(),
        vscode.NewAdapter(), windsurf.NewAdapter(),
    }
    for _, adapter := range agents {
        home := t.TempDir()
        _, err := InjectContextMode(home, adapter)
        if err != nil {
            t.Errorf("InjectContextMode(%s): %v", adapter.Agent(), err)
        }
    }
}
```

**Acceptance criteria:**
- `go test -race ./internal/components/mcp/...` passes
- Fresh install AND sync both produce codegraph + context-mode in agent config
- All 11 supported agents get context-mode entry

---

### FIX-02: Fix Claude MCP Key Names (Underscores → Hyphens)

**Problem:** `generateClaude()` in `generator.go` uses `sequential_thinking` and `context_mode`
(underscores). Claude Code registers these as the server names exposed to sub-agents.
All phase protocols reference `sequential-thinking` and `context-mode` (hyphens).

**File:** `internal/components/mcp/generator.go`

Replace `generateClaude`:

```go
func generateClaude(engramBin string, opts GenerateOptions) map[string]interface{} {
	servers := map[string]interface{}{
		"engram": map[string]interface{}{
			"command": engramBin,
			"args":    []string{"mcp", "--tools=agent"},
			"type":    "stdio",
		},
		"context7": map[string]interface{}{
			"command": "npx",
			"args":    []string{"-y", "@upstash/context7-mcp"},  // removed @latest pin
			"type":    "stdio",
		},
		"sequential-thinking": map[string]interface{}{  // was: sequential_thinking
			"command": "npx",
			"args":    []string{"-y", "@modelcontextprotocol/server-sequential-thinking"},
			"type":    "stdio",
		},
		"context-mode": map[string]interface{}{  // was: context_mode
			"command": "npx",
			"args":    []string{"-y", "@mksglu/context-mode"},
			"type":    "stdio",
		},
		"codegraph": map[string]interface{}{
			"command": "npx",
			"args":    []string{"-y", "@colbymchenry/codegraph", "serve", "--mcp"},
			"type":    "stdio",
		},
		"notebooklm-mcp": map[string]interface{}{  // was: absent
			"command": "notebooklm-mcp",
			"args":    []string{},
			"type":    "stdio",
		},
	}
	return map[string]interface{}{"mcp_servers": servers}
}
```

**File:** `internal/components/mcp/context7.go`

Change Claude-specific context7 package to remove `@latest` pin (match other agents):

```go
// Old: "@upstash/context7-mcp@latest"
// New: "@upstash/context7-mcp"
```

Update `context7Overlay` for Claude:

```go
case model.AgentClaudeCode:
    return []byte(`{
  "command": "npx",
  "args": [
    "-y",
    "@upstash/context7-mcp"
  ]
}`), nil
```

**Tests:**

```go
func TestGenerateClaudeKeyNames(t *testing.T) {
    cfg, err := GenerateConfig("claude", GenerateOptions{})
    if err != nil {
        t.Fatal(err)
    }
    servers, ok := cfg["mcp_servers"].(map[string]interface{})
    if !ok {
        t.Fatal("expected mcp_servers key")
    }
    for _, key := range []string{"sequential-thinking", "context-mode", "notebooklm-mcp"} {
        if _, exists := servers[key]; !exists {
            t.Errorf("missing server key: %q (check for underscore variant)", key)
        }
    }
    // Ensure old underscore keys are gone
    for _, badKey := range []string{"sequential_thinking", "context_mode"} {
        if _, exists := servers[badKey]; exists {
            t.Errorf("found deprecated underscore key: %q", badKey)
        }
    }
}
```

**Acceptance criteria:**
- `go test ./internal/components/mcp/...` passes
- Claude MCP server names use hyphens consistently
- `notebooklm-mcp` present in Claude config

---

### FIX-03: Add `notebooklm-mcp` to ALL Generators

**Problem:** `notebooklm-mcp` absent from every `generate*` function. NotebookLM
only arrives via `InjectNotebookLM()` which must be called separately.

**File:** `internal/components/mcp/generator.go`

Add to `generateVSCode` servers map:

```go
"notebooklm-mcp": map[string]interface{}{
    "type":    "stdio",
    "command": "notebooklm-mcp",
    "args":    []string{},
},
```

Add to `generateAntigravity` servers map:

```go
"notebooklm-mcp": map[string]interface{}{
    "command": "notebooklm-mcp",
    "args":    []string{},
},
```

Add to `generateGemini` mcpServers map:

```go
"notebooklm-mcp": map[string]interface{}{
    "command": "notebooklm-mcp",
    "args":    []string{},
    "timeout": 30000,
    "trust":   false,
},
```

Add to `generateOpenCode` mcp map:

```go
"notebooklm-mcp": map[string]interface{}{
    "type":    "local",
    "command": []string{"notebooklm-mcp"},
    "enabled": false,  // disabled until user configures notebook
},
```

**Note on `enabled: false` for OpenCode:** NotebookLM requires user-specific notebook
configuration. Setting `enabled: false` registers the server without activating it.
The user enables it manually after configuring their notebook. Add a comment in the
install success output: "NotebookLM configured (disabled). Enable in opencode.json after configuring your notebook."

**Tests:**

```go
func TestAllGeneratorsHaveNotebookLM(t *testing.T) {
    platforms := []string{"claude", "vscode", "antigravity", "gemini", "opencode"}
    for _, p := range platforms {
        cfg, err := GenerateConfig(p, GenerateOptions{})
        if err != nil {
            t.Fatalf("GenerateConfig(%s): %v", p, err)
        }
        jsonBytes, _ := json.Marshal(cfg)
        if !bytes.Contains(jsonBytes, []byte("notebooklm-mcp")) {
            t.Errorf("platform %q missing notebooklm-mcp in generated config", p)
        }
    }
}
```

**Acceptance criteria:**
- Fresh install for any agent includes `notebooklm-mcp`
- OpenCode gets `enabled: false`; all others get enabled entry
- `go test ./internal/components/mcp/...` passes

---

### FIX-04: Create Antigravity CLI Go Adapter

**Problem:** `internal/assets/antigravity-cli/` has all config files but no Go adapter.
The `install` command never touches Antigravity CLI.

**New file:** `internal/agents/antigravity-cli/adapter.go`

```go
package antigravitycli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rd-mg/architect-ai/internal/model"
	"github.com/rd-mg/architect-ai/internal/system"
)

type statResult struct {
	isDir bool
	err   error
}

type Adapter struct {
	lookPath func(string) (string, error)
	statPath func(string) statResult
}

func NewAdapter() *Adapter {
	return &Adapter{
		lookPath: exec.LookPath,
		statPath: defaultStat,
	}
}

// --- Identity ---

func (a *Adapter) Agent() model.AgentID {
	return model.AgentAntigravityCLI
}

func (a *Adapter) Tier() model.SupportTier {
	return model.TierFull
}

// --- Detection ---
// agy is the Antigravity CLI binary.
// Config lives at ~/.gemini/antigravity-cli/plugins/architect-ai/

func (a *Adapter) Detect(_ context.Context, homeDir string) (bool, string, string, bool, error) {
	// Check for agy binary on PATH
	binaryPath, err := a.lookPath("agy")
	installed := err == nil

	// Check plugin dir
	pluginDir := filepath.Join(homeDir, ".gemini", "antigravity-cli", "plugins")
	stat := a.statPath(pluginDir)
	if stat.err != nil {
		if os.IsNotExist(stat.err) {
			return installed, binaryPath, pluginDir, false, nil
		}
		return false, "", "", false, stat.err
	}

	return installed, binaryPath, pluginDir, stat.isDir, nil
}

// --- Installation ---

func (a *Adapter) SupportsAutoInstall() bool {
	// agy is installed via curl script — cannot be automated here.
	return false
}

func (a *Adapter) InstallCommand(_ system.PlatformProfile) ([][]string, error) {
	return nil, AgentNotInstallableError{Agent: model.AgentAntigravityCLI}
}

// --- Config paths ---

func (a *Adapter) pluginRoot(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "antigravity-cli", "plugins", "architect-ai")
}

func (a *Adapter) GlobalConfigDir(homeDir string) string {
	return a.pluginRoot(homeDir)
}

func (a *Adapter) SystemPromptDir(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "antigravity-cli", "skills")
}

// SystemPromptFile — Antigravity CLI uses GEMINI.md in the global gemini dir.
func (a *Adapter) SystemPromptFile(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "GEMINI.md")
}

func (a *Adapter) SkillsDir(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "antigravity-cli", "skills")
}

func (a *Adapter) SettingsPath(homeDir string) string {
	return filepath.Join(homeDir, ".gemini", "antigravity-cli", "settings.json")
}

// --- Config strategies ---

func (a *Adapter) SystemPromptStrategy() model.SystemPromptStrategy {
	return model.StrategyMarkdownSections
}

// MCPStrategy: Antigravity CLI uses plugin-local mcp_config.json (object format).
func (a *Adapter) MCPStrategy() model.MCPStrategy {
	return model.StrategyMCPConfigFile
}

// --- MCP ---

func (a *Adapter) MCPConfigPath(homeDir string, _ string) string {
	return filepath.Join(a.pluginRoot(homeDir), "mcp_config.json")
}

// --- Optional capabilities ---

func (a *Adapter) SupportsOutputStyles() bool  { return false }
func (a *Adapter) OutputStyleDir(_ string) string { return "" }
func (a *Adapter) SupportsSlashCommands() bool { return true }
func (a *Adapter) CommandsDir(_ string) string  { return "" }
func (a *Adapter) SupportsSkills() bool          { return true }
func (a *Adapter) SupportsSystemPrompt() bool    { return true }
func (a *Adapter) SupportsMCP() bool             { return true }

// AgentNotInstallableError is returned when InstallCommand is called on the CLI.
type AgentNotInstallableError struct{ Agent model.AgentID }

func (e AgentNotInstallableError) Error() string {
	return "agent " + string(e.Agent) + " (Antigravity CLI / agy) must be installed manually: curl -fsSL https://antigravity.com/install.sh | bash"
}

func defaultStat(path string) statResult {
	info, err := os.Stat(path)
	if err != nil {
		return statResult{err: err}
	}
	return statResult{isDir: info.IsDir()}
}
```

**New file:** `internal/agents/antigravity-cli/installer.go`

```go
package antigravitycli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rd-mg/architect-ai/internal/assets"
	"github.com/rd-mg/architect-ai/internal/components/filemerge"
	"github.com/rd-mg/architect-ai/internal/components/mcp"
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
	overlay := assets.MustRead("antigravity-cli/settings.json")
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
	content := assets.MustRead("antigravity-cli/sidecars/archive-cleaner.json")
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
	content := assets.MustRead(assetPath)
	if dryRun {
		fmt.Printf("  DRY RUN: would write %s/%s\n", dir, filename)
		return nil
	}
	path := filepath.Join(dir, filename)
	_, err := filemerge.WriteFileAtomic(path, content, 0o644)
	return err
}
```

---

### FIX-05: Register Antigravity CLI in Model, Catalog, Embed, Pipeline

**File:** `internal/model/model.go` (or wherever AgentID constants are)

Add after `AgentKiroIDE`:

```go
AgentAntigravityCLI AgentID = "antigravity-cli"
```

**File:** `internal/catalog/agents.go`

Add to `allAgents` slice:

```go
{
    ID:         model.AgentAntigravityCLI,
    Name:       "Antigravity CLI (agy)",
    Tier:       model.TierFull,
    ConfigPath: "~/.gemini/antigravity-cli/plugins/architect-ai",
},
```

**File:** `internal/assets/assets.go` (or wherever `//go:embed` directive lives)

Add `all:antigravity-cli` to the embed directive:

```go
//go:embed all:claude all:opencode all:generic all:skills all:gga all:gemini all:codex \
//          all:antigravity all:antigravity-cli all:windsurf all:cursor all:qwen all:kiro \
//          all:overlays all:_shared all:vscode all:workflows source-map.json
```

**File:** `internal/app/app.go` (or `internal/cli/install.go`)

Add antigravity-cli adapter to the agent adapter factory:

```go
func antigravityCLIAdapter() agents.Adapter { return antigravitycli.NewAdapter() }
```

Add to the agent dispatch switch:

```go
case model.AgentAntigravityCLI:
    adapter = antigravityCLIAdapter()
```

**File:** `internal/components/mcp/overlay.go`

Add `model.AgentAntigravityCLI` to relevant switch cases:

```go
// In sequentialThinkingOverlay — same format as AgentAntigravity:
case model.AgentAntigravity, model.AgentAntigravityCLI, model.AgentWindsurf, model.AgentQwenCode, model.AgentKiroIDE:
    return []byte(`{"mcpServers": {"sequential-thinking": {...}}}`), nil

// In context7Overlay:
case model.AgentAntigravity, model.AgentAntigravityCLI:
    return AntigravityContext7OverlayJSON(), nil

// In notebookLMOverlay:
case model.AgentAntigravity, model.AgentAntigravityCLI:
    return []byte(`{"mcpServers": {"notebooklm-mcp": {"command": "notebooklm-mcp", "args": []}}}`), nil

// In codegraphOverlay:
case model.AgentAntigravity, model.AgentAntigravityCLI, model.AgentWindsurf, model.AgentQwenCode, model.AgentKiroIDE:
    return []byte(`{"mcpServers": {"codegraph": {"command": "npx", "args": ["-y", "@colbymchenry/codegraph", "serve", "--mcp"]}}}`), nil
```

**File:** `internal/components/mcp/generator.go`

Add case in `GenerateConfig`:

```go
case "antigravity-cli":
    return generateAntigravityCLI(engramBin, opts), nil
```

Add function:

```go
func generateAntigravityCLI(engramBin string, opts GenerateOptions) map[string]interface{} {
	servers := map[string]interface{}{
		"engram":              map[string]interface{}{"command": engramBin, "args": []string{"mcp", "--tools=agent"}},
		"context7":            map[string]interface{}{"serverUrl": "https://mcp.context7.com/mcp"},
		"sequential-thinking": map[string]interface{}{"command": "npx", "args": []string{"-y", "@modelcontextprotocol/server-sequential-thinking"}},
		"context-mode":        map[string]interface{}{"command": "context-mode", "args": []string{"--mcp"}},
		"codegraph":           map[string]interface{}{"command": "npx", "args": []string{"-y", "@colbymchenry/codegraph", "serve", "--mcp"}},
		"notebooklm-mcp":      map[string]interface{}{"command": "notebooklm-mcp", "args": []string{}},
	}
	return map[string]interface{}{"mcpServers": servers}
}
```

**Tests:**

```go
// internal/agents/antigravity-cli/adapter_test.go
func TestAntigravityCLIAdapterIdentity(t *testing.T) {
    a := NewAdapter()
    if a.Agent() != model.AgentAntigravityCLI {
        t.Errorf("wrong agent ID: got %q want %q", a.Agent(), model.AgentAntigravityCLI)
    }
    if a.SupportsAutoInstall() {
        t.Error("antigravity-cli should not support auto-install")
    }
    if !a.SupportsMCP() {
        t.Error("antigravity-cli should support MCP")
    }
}

func TestAntigravityCLIPathsResolveCorrectly(t *testing.T) {
    a := NewAdapter()
    home := "/home/testuser"
    want := "/home/testuser/.gemini/antigravity-cli/plugins/architect-ai/mcp_config.json"
    got := a.MCPConfigPath(home, "any")
    if got != want {
        t.Errorf("MCPConfigPath: got %q want %q", got, want)
    }
}
```

**Acceptance criteria:**
- `architect-ai install --agent antigravity-cli` is recognized
- Detection finds `agy` binary when present
- Plugin dir created at correct path
- All 6 MCPs written to plugin `mcp_config.json`

---

### FIX-06: Add CodeGraph Steps to `sdd-explore.md` — All Agents

**Files to update** (same change in all):
```
internal/assets/claude/sdd-phase-protocols/sdd-explore.md
internal/assets/generic/sdd-phase-protocols/sdd-explore.md
internal/assets/gemini/sdd-phase-protocols/sdd-explore.md
internal/assets/codex/sdd-phase-protocols/sdd-explore.md
internal/assets/cursor/sdd-phase-protocols/sdd-explore.md
internal/assets/kiro/sdd-phase-protocols/sdd-explore.md
internal/assets/opencode/sdd-phase-protocols/sdd-explore.md
internal/assets/gga/sdd-phase-protocols/sdd-explore.md
```

**Change:** Insert after `## Step 0: Deep Code Exploration (Sequential Thinking)` and before
`## ADR Pre-check`:

```markdown
## Step 0b: Semantic Graph Exploration (CodeGraph — Priority over ripgrep)

IF `codegraph_context` is in the verified tool list:

1. **Semantic Context Pack**
   ```
   codegraph_context(
     query: "{change_topic}",
     maxNodes: 25,
     format: "markdown"
   )
   ```
   → Returns related functions, types, files, call chains.
   → Use this output as the primary code map. Proceed to ADR Pre-check.
   → Only run Steps 1-4 of Section B (ripgrep) if codegraph returns < 5 nodes.

2. **Call Chain Trace** (when entrypoint identified from context pack)
   ```
   codegraph_trace(entry: "{identified_entrypoint}")
   ```
   → Full call chain from trigger to leaf.

3. **Blast Radius** (for change impact analysis)
   ```
   codegraph_impact(nodeId: "{primary_node}", depth: 3)
   ```
   → What breaks if this node changes. Always run this.

4. **Inbound References** (LspFindReferences equivalent)
   ```
   codegraph_callers(nodeId: "{primary_node}")
   ```
   → All call sites. Required for impact surface completeness.

IF `codegraph_context` is NOT available:
→ Skip to Section B (ripgrep 5-step protocol). No change to existing flow.

**DO NOT** run both codegraph AND full ripgrep sweep for the same query.
CodeGraph output is authoritative when available.
```

**Also update the Batch Execute Pattern** in Section B to add CodeGraph condition:

```markdown
**Batch Execute Pattern (MANDATORY for 3+ searches AND codegraph unavailable):**
```

---

### FIX-07: Add CodeGraph to Research Routing Policy — All Orchestrators

**Files to update:**
```
internal/assets/claude/general-orchestrator.md
internal/assets/claude/sdd-orchestrator.md
internal/assets/generic/general-orchestrator.md
internal/assets/generic/sdd-orchestrator.md
internal/assets/gemini/general-orchestrator.md
internal/assets/gemini/sdd-orchestrator.md
internal/assets/opencode/general-orchestrator.md
internal/assets/opencode/sdd-orchestrator.md
internal/assets/cursor/general-orchestrator.md
internal/assets/cursor/sdd-orchestrator.md
internal/assets/kiro/general-orchestrator.md
internal/assets/kiro/sdd-orchestrator.md
internal/assets/gga/general-orchestrator.md
internal/assets/gga/sdd-orchestrator.md
internal/assets/generic/context-mode-routing-policy.md
```

**Change:** Replace the 5-step research routing policy with the 6-step version:

```markdown
## RESEARCH-ROUTING POLICY (enforce before any external lookup)

Use sources in strict priority order. Escalate ONLY when current source yields no result.

**STEP 1 — Engram (always first)**
`mem_search(query: "{specific_topic_key}", project: "{project}")`
→ Hit: USE IT. Skip steps 2–6.
→ Miss: proceed to step 2.

**STEP 2 — Local ripgrep (Project Evidence)**
Use when: understanding project's own structure or specific file contents.
`ctx_batch_execute(["rg '{keyword}' -l", "rg '^func|^type' {file}"])`
→ Hit: use it. For semantic relationships → also run Step 2b.
→ Miss: proceed to step 2b.

**STEP 2b — CodeGraph (Semantic Relationships)**
Use when: need call chains, callers, impact radius, or cross-file relationships.
Only available when `codegraph_context` is in verified tool list.
```
codegraph_context(query: "{topic}", maxNodes: 25, format: "markdown")
codegraph_callers(nodeId: "{node}")    // who calls this?
codegraph_impact(nodeId: "{node}")     // what breaks?
```
→ Result supplements or replaces multi-file ripgrep chaining.
→ Miss or unavailable: proceed to step 3.

**STEP 3 — Context7 (Framework/Library Docs)**
Use when: documentation for third-party library or API needed.
```
mcp__context7__resolve-library-id(libraryName: "{detected_library}")
mcp__context7__get-library-docs(context7CompatibleLibraryID: "{id}", topic: "{specific_aspect}", tokens: 5000)
```
ALWAYS specify `topic`. Never fetch full docs without topic filter.
→ Hit: use it. → Miss: proceed to step 4.

**STEP 4 — NotebookLM (Optional Synthesis)**
Use when: version-specific changes, migration guides, domain synthesis.
ONLY available in Mode 1 or Mode 2. NOT in Mode 3.
→ Result persists to Engram via after_model hook.

**STEP 5 — Web search (last resort)**
Use when: steps 1–4 all yield no result.
Include `site:` filter when possible.
NOT available in Mode 3.
```

---

### FIX-08: Separate Judge Agent Instantiation in `sdd-verify.md`

**Files to update:**
```
internal/assets/claude/sdd-phase-protocols/sdd-verify.md
internal/assets/generic/sdd-phase-protocols/sdd-verify.md
internal/assets/gemini/sdd-phase-protocols/sdd-verify.md
internal/assets/codex/sdd-phase-protocols/sdd-verify.md
internal/assets/cursor/sdd-phase-protocols/sdd-verify.md
internal/assets/kiro/sdd-phase-protocols/sdd-verify.md
internal/assets/opencode/sdd-phase-protocols/sdd-verify.md
internal/assets/gga/sdd-phase-protocols/sdd-verify.md
```

**Change:** Replace the single sub-agent template with a two-judge parallel dispatch.

In the SDD Orchestrator's `sdd-verify` delegation, replace:

```markdown
## Judge Instantiation (MANDATORY — spawn BOTH in same orchestrator response)

The orchestrator MUST emit TWO parallel Task tool calls for verification.
Never use a single agent for adversarial review.

```
[Response must contain BOTH of these in the same turn:]

Task(
  agent="judge-primary",
  model="sonnet",
  description="PRIMARY JUDGE for {change-name}",
  prompt="""
+++Adversarial
Find defects in the implementation. Assume NOTHING is correct until proven.

## Your role: Primary Judge
You review the implementation for CORRECTNESS.
Focus: Does the implementation match the spec? Are test assertions meaningful?
Construct: counterexamples, hostile inputs, data edge cases.

## Evidence Required
Load: sdd/{change-name}/spec, sdd/{change-name}/design, sdd/{change-name}/tasks
Load: apply-progress for all batches
Run: test suite

## Validation Checklist
- [ ] All spec capabilities implemented
- [ ] All success criteria measurable and met
- [ ] No TODO/FIXME/XXX in changed code
- [ ] Tests pass (or match documented baseline failures)
- [ ] No test weakening

## Verdict
APPROVED | CONDITIONALLY_APPROVED | NEEDS_CHANGES | UNRESOLVED

## Return Envelope (JSON)
{"judge": "primary", "verdict": "...", "findings": [...], "critical_count": N}
"""
)

Task(
  agent="judge-secondary",
  model="sonnet",
  description="SECONDARY JUDGE for {change-name}",
  prompt="""
+++Forensic
Trace failure modes and edge cases. Evidence chains required for every finding.

## Your role: Secondary Judge
You review the implementation for ROBUSTNESS and FAILURE HANDLING.
Focus: What breaks in production? Race conditions, error paths, resource leaks.
Construct: failure sequences, concurrent access patterns, upgrade hazards.

## Evidence Required
Load: sdd/{change-name}/spec (FMEA table), sdd/{change-name}/design (error propagation section)
Load: apply-progress — check error handling blocks

## Validation Checklist
- [ ] FMEA failure modes handled in implementation
- [ ] Sad-path BDD scenarios pass (if present)
- [ ] Error propagation matches design
- [ ] No silent failures (unchecked errors)
- [ ] Rollback plan still valid after implementation

## Verdict
APPROVED | CONDITIONALLY_APPROVED | NEEDS_CHANGES | UNRESOLVED

## Return Envelope (JSON)
{"judge": "secondary", "verdict": "...", "findings": [...], "critical_count": N}
"""
)
```

## Verdict Synthesis (orchestrator — after BOTH judges complete)

```
IF primary.verdict == APPROVED AND secondary.verdict == APPROVED:
  → Final verdict: APPROVED
IF either judge == NEEDS_CHANGES:
  → Final verdict: NEEDS_CHANGES (route to sdd-apply with combined findings)
IF both == CONDITIONALLY_APPROVED:
  → Final verdict: CONDITIONALLY_APPROVED (present to user)
IF judges disagree (one APPROVED, one CONDITIONALLY_APPROVED):
  → Final verdict: CONDITIONALLY_APPROVED (conservative)
IF either judge == UNRESOLVED:
  → Final verdict: UNRESOLVED (escalate to user)

Save combined findings:
mem_save(topic_key: "sdd/{change-name}/verify-report", content: {combined_report})
```
```

---

### FIX-09: Pre-Apply Completeness Gate in `sdd-apply.md`

**Files to update:** Same 8 agents as FIX-08 but for `sdd-apply.md`.

**Change:** Replace the existing TDD Prerequisite Lock with a comprehensive completeness gate:

```markdown
## Pre-Apply Completeness Gate (MANDATORY — HALT if ANY check fails)

Run BEFORE any file modification. Load and validate all prior artifacts.

### Load artifacts
```
mem_search("sdd/{change-name}/spec")    → mem_get_observation(id) → SPEC
mem_search("sdd/{change-name}/design")  → mem_get_observation(id) → DESIGN
mem_search("sdd/{change-name}/tasks")   → mem_get_observation(id) → TASKS
```

### Spec Completeness (fail if ANY is true)
- [ ] Any capability section contains: "TODO", "TBD", "PLACEHOLDER", "N/A" without justification
- [ ] Any capability missing: Purpose, Preconditions, Behavior, Postconditions, Error Handling, Invariants, Test Hooks
- [ ] External I/O capability exists with no FMEA table
- [ ] FMEA severity ≥ 3 exists with no Sad-path BDD scenario
- [ ] Any success criterion is unmeasurable ("should work", "as expected")

### Design Completeness (fail if ANY is true)
- [ ] Architecture diagram absent
- [ ] Any module boundary section says "to be designed" or is empty
- [ ] Interface contracts section absent or has stubs
- [ ] ADR table is empty (0 entries)
- [ ] Open Questions section is non-empty (must be resolved before apply)
- [ ] YAGNI Gate table absent

### Tasks Completeness (fail if ANY is true)
- [ ] Any task lacks an acceptance criterion
- [ ] Any task describes a whole feature (not atomic: must be < 30 min work)
- [ ] Any HIGH-risk task has no Risk-reason
- [ ] Task count ≥ 5 but no Execution Graph (Mermaid) present
- [ ] Cross-phase reference check: any task references a capability not in SPEC

### Cross-Phase Reference Check
```
FOR EACH task in TASKS:
  scan task acceptance criterion for spec capability name
  IF no matching capability found in SPEC → FAIL
  emit: "Task {N.N} references capability not in spec: {criterion}"
```

### On ANY failure:
- Set status: blocked
- List ALL failed checks (not just first)
- Route: spec failures → sdd-spec; design failures → sdd-design; task failures → sdd-tasks
- DO NOT proceed with any file modification

### On ALL checks pass:
- Emit: "Pre-apply gate: PASSED. {N} tasks ready for implementation."
- Proceed with batch execution
```

---

### FIX-10: Fix Duplicate Step 4 in `sdd-archive.md`

**Files to update:** Both Claude and generic versions:
```
internal/assets/claude/sdd-phase-protocols/sdd-archive.md
internal/assets/generic/sdd-phase-protocols/sdd-archive.md
```

**Change:** Fix procedure numbering in the `## Procedure` section:

Current (broken):
```
3. Generate archive summary...
3b. Eval Gate Check...
4. Persistence (Learned Patterns)...
4. If OpenSpec mode: move change directory...  ← DUPLICATE
5. Update DAG state to "archived"
```

Corrected:
```
3. Generate archive summary (including Entity Tag Extraction)
3b. Eval Gate Check: verify NO HIGH-risk tasks lack eval evidence
4. Project File Updates (see FIX-13 section below)
5. Persistence (Learned Patterns):
   Search knowledge/_global/skill/{skill-name}/learned-patterns
   → found: mem_update (append + increment version)
   → not found: mem_save new entry
6. If OpenSpec mode: move change directory to archive/YYYY-MM-DD-{change-name}/
7. Update DAG state to "archived"
```

---

### FIX-11: Fix L1b Label in Generic `general-orchestrator.md`

**Files to update:**
```
internal/assets/generic/general-orchestrator.md
internal/assets/opencode/general-orchestrator.md
internal/assets/cursor/general-orchestrator.md
internal/assets/gemini/general-orchestrator.md
internal/assets/kiro/general-orchestrator.md
internal/assets/gga/general-orchestrator.md
```

For each file, find and replace the header:

```markdown
# L1b General Orchestrator Core (...)
```

Replace with:

```markdown
# L1a General Orchestrator Core (...)
```

Also update any description frontmatter that says "L1b":

```yaml
# Before:
description: >
  L1b General Orchestrator. Handles non-SDD workflows...

# After:
description: >
  L1a General Orchestrator. Handles non-SDD workflows...
```

**Note:** L1a = General Orchestrator, L1b = SDD Orchestrator. This is the canonical mapping.

---

### FIX-12: Add `manifest.json` Creation to `sdd-init.md`

**Files to update:** All `sdd-init.md` phase protocols.

**Change:** Add Step 6b to the Detection Procedure section:

```markdown
### Step 6b: Overlay Registration

IF Odoo project detected in Steps 1-6 (pyproject.toml contains "odoo" OR
any `__manifest__.py` found in addons/ directory):

1. Determine Odoo version from:
   - `pyproject.toml` `[tool.odoo]` section, OR
   - `__manifest__.py` `"version"` field prefix (e.g., "17.0.x.y.z" → v17), OR
   - ODOO_VERSION environment variable

2. Create overlay manifest:
   ```
   mkdir -p .atl/overlays/odoo-{version}
   write .atl/overlays/odoo-{version}/manifest.json:
   {
     "overlay": "odoo-development-skill",
     "version": "{detected_version}",
     "active": true,
     "addons_path": "{detected_addons_path}",
     "detected_at": "{ISO_8601_timestamp}"
   }
   ```

3. Record IS_ODOO = true in sdd-init artifact (under `active_overlays` key)

4. Notify user:
   "Odoo {version} detected. Overlay activated. Phase protocols will include Odoo-specific rules."

IF NOT Odoo project:
   Skip this step entirely.
```

---

## Sprint 2 — High Priority (Week 3–4)

### FIX-13: Add README.md and `__manifest__.py` to `sdd-archive.md`

**Files to update:** All `sdd-archive.md` phase protocols.

**Change:** Add Step 4 (Project File Updates) to the Procedure section (after FIX-10 renumbering):

```markdown
## Step 4: Project File Updates (MANDATORY — run BEFORE moving to archive/)

### 4a. CHANGELOG.md (all projects)

IF `CHANGELOG.md` exists in project root:
  Read current content
  Prepend new entry:
  ```
  ## [{computed_version}] — {date_today_ISO}
  ### {category: Added|Changed|Fixed|Removed}
  - {one-line summary per implemented capability from proposal}
  
  Full change: openspec/changes/{change-name}/archive-report
  ```

IF `CHANGELOG.md` does not exist:
  Create it with the above entry and a standard Keep a Changelog header.

### 4b. README.md (all projects)

Read `README.md`. Check if any of the following sections need updating:

- **Installation**: New dependency added in sdd-apply? → update install steps
- **Usage**: New command, endpoint, or behavior? → update usage examples
- **Configuration**: New env var or config key? → document it
- **Known Issues**: Any CONDITIONALLY_APPROVED items from verify-report? → document them

IF any update needed: make targeted edit (do NOT rewrite whole README).
IF README.md does not exist: create minimal one with project name + change summary.

### 4c. `__manifest__.py` (Odoo projects only — when IS_ODOO = true)

1. Read current `__manifest__.py`
2. Bump `version` field:
   ```python
   # Current: "version": "17.0.1.2.3"
   # Bump last segment:   "version": "17.0.1.2.4"
   # If minor feature:    "version": "17.0.1.3.0"
   # If breaking:         "version": "17.0.2.0.0"
   ```
   Ask user for bump type if unclear: `patch | minor | major (within Odoo version)`
3. Update `depends` if sdd-apply added new module dependencies:
   ```python
   "depends": ["base", "mail", "new_dependency_added_in_apply"]
   ```
4. Update `data` list if new XML view files were created
5. Update `assets` dict if new JS/CSS/SCSS files were added
6. Verify `__manifest__.py` is valid Python syntax:
   ```bash
   python3 -c "import ast; ast.parse(open('__manifest__.py').read())"
   ```
   If syntax error → STOP, report error, do NOT archive.

### 4d. `package.json` / `pyproject.toml` / `go.mod` (if applicable)

IF a versioned manifest file exists at project root:
  Check if `version` field should be bumped per the same logic as 4c.
  Only bump if the change introduced a user-visible API change.
```

---

### FIX-14: Add `context-mode` Overlay for Cursor, Kiro, Windsurf, Qwen

**Problem:** `contextModeOverlay()` in `overlay.go` (added in FIX-01) handles these agents.
But the install pipeline must call `InjectContextMode()` for them.

**File:** `internal/cli/install.go` (or wherever component injection is orchestrated)

Ensure `InjectContextMode()` is called for ALL agents that support MCP, not just the
subset that previously had it. The `InjectContextMode()` function in FIX-01 already
handles per-agent format selection, so just ensuring it's called is sufficient.

```go
// In the MCP injection loop (after existing Inject/InjectSequentialThinking calls):
for _, adapter := range selectedAdapters {
    if _, err := mcp.Inject(homeDir, adapter); err != nil {
        return fmt.Errorf("inject context7 for %s: %w", adapter.Agent(), err)
    }
    if _, err := mcp.InjectSequentialThinking(homeDir, adapter); err != nil {
        return fmt.Errorf("inject sequential-thinking for %s: %w", adapter.Agent(), err)
    }
    if _, err := mcp.InjectCodeGraph(homeDir, adapter); err != nil {       // FIX-01
        return fmt.Errorf("inject codegraph for %s: %w", adapter.Agent(), err)
    }
    if _, err := mcp.InjectContextMode(homeDir, adapter); err != nil {     // FIX-01
        return fmt.Errorf("inject context-mode for %s: %w", adapter.Agent(), err)
    }
    if _, err := mcp.InjectNotebookLM(homeDir, adapter); err != nil {
        return fmt.Errorf("inject notebooklm for %s: %w", adapter.Agent(), err)
    }
}
```

---

### FIX-15: Fix Max Postures Violation in Routing Tables (3 → 2)

**Problem:** Routing tables assign 3 postures to Analyst/Solver and Ideator,
violating the invariant "Max Postures per prompt: 2".

**Files to update:** ALL `general-orchestrator.md` files across all agents.

**Change:** Replace the Intent Resolution routing table with the corrected 2-posture version:

```markdown
## Intent Resolution & Task Router

Scan for intent in free-text BEFORE responding. Route to correct specialist.
Enforce Max 2 Postures invariant: NO agent receives more than 2 postures.

### Routing Table

| User phrase | Workflow | Target Agent | Postures (max 2) |
|------------|----------|--------------|-----------------|
| "fix this", "why is X crashing", "solve", "repair", "broken" | `/solve` | **Solver** | +++Forensic + +++Systemic |
| "debug", "trace", "what's causing", "stack trace" | `/debug` | **Solver** | +++Forensic + +++Adversarial |
| "research", "how does X work", "investigate", "look into" | `/investigate` | **Researcher** | +++Socratic + +++Empirical |
| "give me ideas", "brainstorm", "ideate", "options for" | `/brainstorm` | **Ideator** | +++Divergent + +++Lateral |
| "build a quick", "prototype", "draft", "scaffold" | `/prototype` | **Generalist** | +++Pragmatic |
| Other / general tasks | (auto) | **Generalist** | D1-D4 → see Posture Map below |
| "use sdd", "start sdd", "spec-driven" | `/sdd-new` | **SDD Orchestrator** | (phase-assigned) |

### D1-D4 Auto-Posture Map (for Generalist and unlisted intents)

| Mode | Posture |
|------|---------|
| Mode 1 (D1+D2 ≤ 2, D3+D4 ≤ 2) | +++Pragmatic |
| Mode 2 (D1+D2 ≥ 3 OR D3 ≥ 1) | +++Critical |
| Mode 2-ERR (D3 = 1) | +++Forensic |
| Mode 3 (D3 ≥ 2 OR D4 ≥ 3) | +++Adversarial |
| Mode 3-CTX (D4 ≥ 3) | +++Pragmatic |
```

---

### FIX-16: Unify Agent Routing Tables (Claude ↔ Generic)

**Problem:** Claude uses `/analyze → Analyst` while Generic uses `/solve → Solver` and
`/investigate → Researcher`. These are the same concepts with different names.

**Decision:** Adopt the Generic version (more precise, correct split) as the canonical
routing table. Apply to Claude orchestrator too.

**File:** `internal/assets/claude/general-orchestrator.md`

Replace the Claude-specific routing table with the unified version from FIX-15.

Also rename sub-agent labels in Claude to match Generic:
- `Analyst` → `Solver` (for `/solve`, `/debug`)
- Keep `Researcher` for `/investigate`
- Keep `Ideator` for `/brainstorm`
- Keep `Generalist` for `/prototype` and general

Update model assignments in Claude orchestrator to match:

```markdown
## Model Assignments

| Agent Type | Model | Reason |
|-----------|-------|--------|
| orchestrator | opus | Routing judgment + complex decisions |
| solver | opus | Deep debugging, complex root cause analysis |
| researcher | sonnet | Research, investigation, documentation lookup |
| ideator | sonnet | Creative generation, lateral connections |
| generalist | sonnet | Execution, mechanical tasks, prototypes |
```

---

### FIX-17: Remove Spanish Regionalisms from All Routing Patterns

**Files to update:**
```
internal/assets/_shared/architect-identity.md
internal/assets/claude/sdd-orchestrator.md
internal/assets/generic/sdd-orchestrator.md
internal/assets/gemini/sdd-orchestrator.md
internal/assets/cursor/sdd-orchestrator.md
internal/assets/opencode/sdd-orchestrator.md
internal/assets/kiro/sdd-orchestrator.md
internal/assets/gga/sdd-orchestrator.md
internal/assets/claude/general-orchestrator.md
internal/assets/generic/general-orchestrator.md
```

**Change in `architect-identity.md`:**

```markdown
# Before:
Regex: \b(use sdd|start sdd|sdd mode|spec-driven|iniciar sdd|haceme un sdd)\b

# After:
Regex: \b(use sdd|start sdd|begin sdd|sdd mode|spec-driven)\b
```

**Change in all `sdd-orchestrator.md` files:**

Remove ALL `(ES: ...)` annotations from the Intent Resolution table. Remove all
Spanish phrases from pattern columns. The tables become English-only:

```markdown
# Before:
| "use sdd", "let's do sdd", "start sdd" (ES: "usa sdd", "vamos con sdd") | /sdd-new |

# After:
| "use sdd", "start sdd", "begin sdd", "let's do sdd", "apply spec-driven" | /sdd-new |
```

```markdown
# Before:
| "continue", "next phase", "keep going" (ES: "sigue", "continua") | /sdd-continue |

# After:
| "continue", "next phase", "keep going", "what's next" | /sdd-continue |
```

Apply this pattern to ALL rows that have `(ES: ...)` annotations.

**Note:** Keep `Terra Incógnita` in D2 dimension description — it is an internationally
recognized technical term, not a regionalism.

**eintegrate check to add** (`cmd/eintegrate/main.go`):

```go
// E-24: No Spanish phrases in routing patterns
regionalisms := []string{"haceme", "vamos con", "usa sdd", "iniciar sdd",
    "ES: ", "(ES:", "sigue", "continua", "guíame", "valida", "cierra el cambio"}
for _, file := range orchestratorFiles {
    for _, term := range regionalisms {
        if checkFile(file, term) {
            errs = append(errs, fmt.Sprintf("E-24: Spanish regionalism %q found in %s", term, file))
        }
    }
}
```

---

### FIX-18: Add Guardrails and MCP Listing to All Persona Files

**Files to update:**
```
internal/assets/claude/persona-architect.md
internal/assets/generic/persona-architect.md
internal/assets/generic/persona-neutral.md
internal/assets/opencode/persona-architect.md
internal/assets/kiro/persona-architect.md
```

**Change:** Add two new sections at the end of each persona file (after `## Tools`):

```markdown
## Architecture Constitution (MANDATORY — governs all behavior)

Five inviolable rules for all actions:
1. **Source of Truth**: State lives in ONE place. No replication without sync.
2. **Thin Adapters**: Business logic in domain/core. Integrations are thin wrappers.
3. **Explicit Boundaries**: No hidden cross-system coupling in helpers/utilities.
4. **Mental Model First**: Fit new features into logical model BEFORE designing implementation.
5. **Sandbox Security**: L2 agents CANNOT perform destructive mutations without L0/L1 authorization.
   Report RISK. Defer to human if escalation required.

Full reference: `_shared/architecture-guardrails.md` (load when D1 ≥ 2)

## Active MCP Servers

Available tools — probe at session start, pass availability to sub-agents:

| Server | Primary tools | When to use |
|--------|--------------|------------|
| **engram** | mem_search, mem_save, mem_get_observation, mem_context, mem_session_summary | Always — session memory, SDD artifacts, decision records |
| **context7** | resolve-library-id, get-library-docs(topic, tokens) | External library/framework docs. ALWAYS specify topic. Cap tokens at 5000. |
| **sequential-thinking** | sequential_thinking | Before any complex design or multi-path analysis |
| **context-mode** | ctx_execute, ctx_batch_execute, ctx_fetch_and_index, ctx_search | Protecting context window from raw output flooding |
| **codegraph** | codegraph_context, codegraph_trace, codegraph_callers, codegraph_impact | Semantic code exploration, impact analysis, LspFindReferences |
| **notebooklm-mcp** | notebooklm_* | Research synthesis, migration guides (Mode 1/2 only) |

**MCP Usage Rules:**
- Run tool probe at session start. Cache result. Do not re-probe per sub-agent.
- context-mode BLOCKED list: raw `curl`, `cat` on large files, direct web fetch
- CodeGraph priority over ripgrep for relationship queries
- context7 ALWAYS with topic parameter; never fetch full docs
- Engram FIRST in all research lookups before any external source

## Context-Mode Routing (MANDATORY)

{content of generic/context-mode-routing-policy.md — inline at install time}
```

**For persona-neutral.md:** Remove any mention of regional language or cultural style.
Replace "style preferences" with neutral professional tone description:

```markdown
## Communication Style

Professional, concise, and direct. No regional expressions, colloquialisms, or
cultural idioms. All output in the user's language (detected from first message)
using standard grammar. English for all code, comments, and technical artifacts.
```

---

### FIX-19: Add Unified Posture Assignment Specification to `adaptive-reasoning-gate-v2.md`

**File:** `internal/assets/_shared/adaptive-reasoning-gate-v2.md`

Add new section between `### Mode Reference Table` and `### Response Header`:

```markdown
### Posture Assignment — Two Sources (Explicit Precedence)

Postures come from two orthogonal sources. Both apply; Source 2 overrides Source 1 in
error/saturation conditions.

#### Source 1: Task Router (workflow-triggered postures)
Applied first. Set by user intent and SDD phase:

| Context | Postures |
|---------|---------|
| SDD: sdd-explore, sdd-onboard | +++Socratic |
| SDD: sdd-propose | +++Critical |
| SDD: sdd-spec | +++Systemic |
| SDD: sdd-design | +++Critical + +++Systemic |
| SDD: sdd-tasks | +++Pragmatic + +++Economic |
| SDD: sdd-apply | +++Pragmatic |
| SDD: sdd-verify (judge-primary) | +++Adversarial |
| SDD: sdd-verify (judge-secondary) | +++Forensic |
| SDD: sdd-archive | (none) |
| General: /investigate | +++Socratic + +++Empirical |
| General: /brainstorm | +++Divergent + +++Lateral |
| General: /debug | +++Forensic + +++Adversarial |
| General: /solve | +++Forensic + +++Systemic |
| General: /prototype | +++Pragmatic |
| General: /analyze (complex) | +++Critical + +++Systemic |
| Numeric SLA in spec | Add +++Empirical (replaces one posture; keep max 2) |

#### Source 2: Routing Matrix (complexity/error override)
Applied second. Overrides Source 1 posture ONLY for error/saturation modes:

| Mode | Override | Postures |
|------|---------|---------|
| Mode 1, Mode 2 | No override | Retain Source 1 postures |
| Mode 2-ERR (D3=1) | Yes | +++Forensic (replaces ALL Source 1 postures) |
| Mode 3 (D3≥2 OR D4≥3) | Yes | +++Adversarial + +++Systemic (replaces ALL) |
| Mode 3-CTX (D4≥3) | Yes | +++Pragmatic (replaces ALL — compressed mode) |
| Circuit Breaker (attempt_count≥2) | Yes | +++Forensic + +++Adversarial (replaces ALL) |

#### Invariant: Max 2 Postures Per Prompt

After applying Source 1 + Source 2:
- Count active postures
- IF count > 2: REJECT. Split task into two delegations OR select top 2 by relevance.
- NEVER inject 3 postures into one sub-agent prompt.
```

---

### FIX-20: Remove Duplicate Adaptive-Reasoning Blocks from Orchestrators

**Problem:** `sdd-orchestrator.md` and `general-orchestrator.md` embed the full
`<!-- adaptive-reasoning-gate:START/END -->` block. Sub-agents also receive `adaptive-reasoning-gate-v2.md`.
This creates ~200-token waste per sub-agent dispatch.

**Files to update:**
```
internal/assets/claude/sdd-orchestrator.md
internal/assets/claude/general-orchestrator.md
internal/assets/generic/sdd-orchestrator.md
internal/assets/generic/general-orchestrator.md
(and all other agent variants)
```

**Change:** Remove the inline `<!-- adaptive-reasoning-gate:START -->` through
`<!-- adaptive-reasoning-gate:END -->` blocks from both orchestrators.

Replace with a compact reference:

```markdown
## Adaptive Reasoning Mode (MANDATORY)

Self-classify before delegating. Emit as first line:
`[MODE N | D1=X, D2=X, D3=X, D4=X] {one-line rationale}`

Full gate: sub-agents receive `_shared/adaptive-reasoning-gate-v2.md` which contains
the complete routing matrix, posture assignment specification, and circuit breaker rules.
```

**Note:** The orchestrator still self-classifies (it needs to compute D1-D4 and inject
the MODE into sub-agents). It just doesn't need the full gate definition in its own prompt.

---

### FIX-21: Fix Gemini Generator — Only Write mcpServers Section

**Problem:** `generateGemini` returns a full settings object (general, ide, model, security,
ui, mcpServers). When used for MCP updates, this overwrites user settings.

**File:** `internal/components/mcp/generator.go`

Split `generateGemini` into two functions:

```go
// generateGemini returns a full settings.json for FRESH installs.
// Only used when the file does not exist yet.
func generateGemini(engramBin string, opts GenerateOptions) map[string]interface{} {
	return map[string]interface{}{
		"general":    map[string]interface{}{"defaultApprovalMode": "auto_edit"},
		"ide":        map[string]interface{}{"enabled": true},
		"mcpServers": generateGeminiMCPServers(engramBin, opts),
		"model":      map[string]interface{}{"name": ""},
		"security":   map[string]interface{}{"auth": map[string]interface{}{"selectedType": "oauth-personal"}},
		"ui":         map[string]interface{}{"hideFooter": true, "showCitations": true, "showMemoryUsage": true, "showModelInfoInChat": true},
	}
}

// generateGeminiMCPOnly returns ONLY the mcpServers section for MERGE operations.
// Used by InjectContextMode, InjectCodeGraph, etc. when settings.json already exists.
func generateGeminiMCPOnly(engramBin string, opts GenerateOptions) map[string]interface{} {
	return map[string]interface{}{
		"mcpServers": generateGeminiMCPServers(engramBin, opts),
	}
}

func generateGeminiMCPServers(engramBin string, opts GenerateOptions) map[string]interface{} {
	return map[string]interface{}{
		"context7":            map[string]interface{}{"httpUrl": "https://mcp.context7.com/mcp", "timeout": 30000, "trust": false},
		"context-mode":        map[string]interface{}{"command": "npx", "args": []string{"-y", "@mksglu/context-mode"}, "timeout": 15000},
		"engram":              map[string]interface{}{"command": engramBin, "args": []string{"mcp", "--tools=agent"}},
		"sequential-thinking": map[string]interface{}{"command": "npx", "args": []string{"-y", "@modelcontextprotocol/server-sequential-thinking"}, "timeout": 30000, "trust": true},
		"codegraph":           map[string]interface{}{"command": "npx", "args": []string{"-y", "@colbymchenry/codegraph", "serve", "--mcp"}, "timeout": 30000, "trust": true},
		"notebooklm-mcp":      map[string]interface{}{"command": "notebooklm-mcp", "args": []string{}, "timeout": 30000, "trust": false},
	}
}
```

**Update `injectMergeIntoSettings` in `inject.go`** to use `generateGeminiMCPOnly` when
the file already exists:

```go
// In injectMergeIntoSettings, before calling OverlayFor:
// Check if this is Gemini and settings already exist — if so, only merge MCP section
if adapter.Agent() == model.AgentGeminiCLI {
    if _, err := os.Stat(settingsPath); err == nil {
        // File exists — only inject MCP servers, preserve other settings
        overlay, err = json.Marshal(generateGeminiMCPOnly(engramBin, GenerateOptions{}))
        // ... merge
    }
}
```

---

### FIX-22: Make WCAG Check Conditional in `sdd-verify.md`

**Files:** All `sdd-verify.md` variants.

**Change:** Wrap the WCAG check:

```markdown
- [ ] **WCAG Compliance Check** (ONLY if change includes UI components):
      Condition: verify-report scope includes frontend files (.tsx, .vue, .html, .css, .scss)
      IF condition true: verify aria-labels, contrast ratios, keyboard navigation, focus management.
      IF condition false (backend-only, API-only, CLI, migration, data-only): SKIP — mark N/A.
```

---

## Sprint 3 — Medium Priority (Week 5–6)

### FIX-23: Add Context-Mode Hook Writer

**Problem:** `context-mode` routing is enforced via prompt text but no actual hook file
intercepts tool calls at the agent level.

**New file:** `internal/components/hooks/context_mode.go`

```go
package hooks

import (
	"os"
	"path/filepath"

	"github.com/rd-mg/architect-ai/internal/agents"
	"github.com/rd-mg/architect-ai/internal/components/filemerge"
	"github.com/rd-mg/architect-ai/internal/model"
)

// InjectContextModeHook writes a pre-tool-use hook file that routes
// shell/read/grep tool calls through context-mode.
func InjectContextModeHook(homeDir string, adapter agents.Adapter) error {
	hookDir, hookFile, hookContent := hookConfigFor(adapter.Agent(), homeDir)
	if hookDir == "" {
		return nil // agent does not support hook files
	}

	if err := os.MkdirAll(hookDir, 0o755); err != nil {
		return err
	}

	_, err := filemerge.WriteFileAtomic(filepath.Join(hookDir, hookFile), hookContent, 0o644)
	return err
}

func hookConfigFor(agent model.AgentID, homeDir string) (dir, file string, content []byte) {
	switch agent {
	case model.AgentClaudeCode:
		return filepath.Join(homeDir, ".claude", "hooks"),
			"context-mode.json",
			claudeContextModeHookJSON

	case model.AgentCursor:
		return filepath.Join(homeDir, ".cursor", "hooks"),
			"context-mode.json",
			cursorContextModeHookJSON

	case model.AgentGeminiCLI:
		return filepath.Join(homeDir, ".gemini", "hooks"),
			"context-mode.json",
			geminiContextModeHookJSON

	case model.AgentAntigravityCLI:
		// CLI uses plugin-level hooks.json (written by installer.go)
		return "", "", nil

	default:
		return "", "", nil
	}
}

var claudeContextModeHookJSON = []byte(`{
  "preToolUse": [
    {
      "type": "command",
      "command": "context-mode hook claude pretooluse",
      "matcher": "Bash|Shell|Read|Grep|WebFetch|Task",
      "timeout": 5
    }
  ]
}
`)

var cursorContextModeHookJSON = []byte(`{
  "preToolUse": [
    {
      "type": "command",
      "command": "context-mode hook cursor pretooluse",
      "matcher": "shell|read|grep|fetch",
      "timeout": 5
    }
  ]
}
`)

var geminiContextModeHookJSON = []byte(`{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "run_command",
        "hooks": [
          {
            "type": "command",
            "command": "context-mode hook gemini pretooluse",
            "timeout": 5
          }
        ]
      }
    ]
  }
}
`)
```

**Wire into install pipeline:**

```go
// In cli/install.go, after MCP injection loop:
for _, adapter := range selectedAdapters {
    if err := hooks.InjectContextModeHook(homeDir, adapter); err != nil {
        // Non-fatal: log warning, continue
        fmt.Fprintf(stderr, "Warning: context-mode hook not installed for %s: %v\n",
            adapter.Agent(), err)
    }
}
```

---

### FIX-24: Process `{{ include "..." }}` Template Directives

**Problem:** Phase protocol files contain `{{ include "_shared/caveman-identity-block.md" }}`
which is Go template syntax but is written to disk as literal text. Sub-agents see
raw `{{ include ... }}` which is non-functional.

**New file:** `internal/assets/renderer.go`

```go
package assets

import (
	"bytes"
	"fmt"
	"regexp"
)

var includePattern = regexp.MustCompile(`\{\{\s*include\s+"([^"]+)"\s*\}\}`)

// RenderIncludes replaces {{ include "path" }} directives with the
// content of the referenced embedded asset file.
// Panics if any referenced asset cannot be found (programming error).
func RenderIncludes(content []byte) []byte {
	return includePattern.ReplaceAllFunc(content, func(match []byte) []byte {
		submatches := includePattern.FindSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		assetPath := string(submatches[1])
		included := MustRead(assetPath)
		return included
	})
}
```

**Update all callers** that write asset content to disk to call `RenderIncludes` first:

```go
// Example in internal/components/sdd/inject.go or equivalent:
content := assets.MustRead("claude/general-orchestrator.md")
rendered := assets.RenderIncludes(content)  // ← add this
_, err = filemerge.WriteFileAtomic(path, rendered, 0o644)
```

Apply to:
- SDD inject (writes orchestrators + phase protocols)
- Engram inject (writes system prompts)
- Persona inject (writes persona files)
- Skills inject (writes skill files)

**Tests:**

```go
func TestRenderIncludes(t *testing.T) {
    // Ensure caveman-identity-block is included correctly
    content := []byte(`preamble\n{{ include "_shared/caveman-identity-block.md" }}\npostamble`)
    result := RenderIncludes(content)
    if bytes.Contains(result, []byte("{{ include")) {
        t.Error("RenderIncludes left unreplaced {{ include }} directive")
    }
    if !bytes.Contains(result, []byte("CAVEMAN")) {
        t.Error("expected caveman-identity-block content in rendered output")
    }
}
```

---

### FIX-25: Add Keyword Index to Skill Registry Generator

**Problem:** `.atl/skill-registry.md` is flat markdown. No machine-readable keyword index
for Dynamic Context Assembler to filter skills by phase/task.

**File:** `internal/app/skills_cmd.go`

Add a keyword index section at the top of the generated `skill-registry.md`:

```go
// buildSkillRegistry generates the .atl/skill-registry.md with keyword index header.
func buildSkillRegistryContent(skills []Skill) string {
    var buf strings.Builder

    // Machine-readable keyword index (JSON block in HTML comment)
    buf.WriteString("<!-- SKILL-REGISTRY-INDEX-V4\n")
    buf.WriteString("Format: skill_id: [keyword1, keyword2, ...]\n")
    buf.WriteString("Used by: Dynamic Context Assembler for keyword-filtered injection\n")
    
    index := buildKeywordIndex(skills)
    indexJSON, _ := json.MarshalIndent(index, "", "  ")
    buf.Write(indexJSON)
    buf.WriteString("\nSKILL-REGISTRY-INDEX-END -->\n\n")

    // Rest of registry content (existing)
    buf.WriteString(buildPhaseContent(skills))
    return buf.String()
}

func buildKeywordIndex(skills []Skill) map[string][]string {
    index := map[string][]string{}
    for _, s := range skills {
        index[s.ID] = s.TriggerKeywords
    }
    return index
}
```

Default keyword mappings for built-in skills:

```go
var defaultSkillKeywords = map[string][]string{
    "ripgrep":            {"search", "grep", "find", "rg", "sdd-explore", "explore"},
    "bash-expert":        {"shell", "bash", "script", "sdd-apply", "run", "execute"},
    "context-guardian":   {"context", "token", "window", "saturation", "budget"},
    "codegraph":          {"explore", "trace", "callers", "impact", "ast", "graph", "LspFindReferences"},
    "adaptive-reasoning": {"mode", "circuit-breaker", "posture", "D1", "D2", "D3", "D4"},
    "mcp-notebooklm":    {"research", "notebooklm", "academic", "paper", "migration"},
    "ctx_fetch_and_index":{"web", "fetch", "url", "http", "context-mode"},
    "sequential-thinking":{"plan", "analyze", "think", "decompose", "branch"},
    "engram":             {"memory", "persist", "recall", "session", "history"},
    "sdd-apply":          {"apply", "implement", "tdd", "strict-tdd", "batch"},
    "sdd-verify":         {"verify", "test", "adversarial", "validate", "judge"},
    "sdd-archive":        {"archive", "close", "manifest", "changelog"},
}
```

---

### FIX-26: Add `uv`/`uvx` Detection for Odoo MCP Install

**Problem:** The Odoo MCP (`mcp-server-odoo`) requires `uvx` from the `uv` Python package
manager. No detection or warning exists if `uv` is absent.

**File:** `internal/system/deps.go` (or equivalent dependency checker)

```go
// CheckUV verifies that uv/uvx is available for Odoo MCP.
func CheckUV() (bool, string) {
    path, err := exec.LookPath("uvx")
    if err == nil {
        out, _ := exec.Command("uvx", "--version").Output()
        return true, strings.TrimSpace(string(out))
    }
    // Try uv (which includes uvx)
    path, err = exec.LookPath("uv")
    if err == nil {
        out, _ := exec.Command("uv", "--version").Output()
        return true, "uv " + strings.TrimSpace(string(out))
    }
    _ = path
    return false, ""
}
```

**Wire into Odoo MCP generation:**

```go
// In generator.go, before adding "odoo" server to the map:
if opts.IsOdooProject {
    uvAvailable, uvVersion := system.CheckUV()
    if !uvAvailable {
        // Add odoo server to config but emit warning
        fmt.Fprintln(os.Stderr,
            "WARNING: Odoo MCP requires 'uvx' (uv package manager). "+
            "Install: curl -LsSf https://astral.sh/uv/install.sh | sh")
    }
    servers["odoo"] = map[string]interface{}{...} // always write the config
    _ = uvVersion
}
```

**Verify check** (`internal/verify/`):

```go
// verifyOdooMCP checks that uvx is available when Odoo MCP is configured.
func verifyOdooMCP(homeDir string) verify.Check {
    return verify.Check{
        ID: "verify:odoo-mcp:uvx",
        Run: func() verify.Result {
            available, version := system.CheckUV()
            if !available {
                return verify.Result{
                    OK:      false,
                    Message: "uvx not found — Odoo MCP (mcp-server-odoo) will fail to start",
                    Fix:     "curl -LsSf https://astral.sh/uv/install.sh | sh",
                }
            }
            return verify.Result{OK: true, Message: "uvx available: " + version}
        },
    }
}
```

---

### FIX-27: Normalize Context7 Package Name

**Problem:** Claude uses `@upstash/context7-mcp` (no version pin, already correct in FIX-02).
Some overlay entries use `@upstash/context7-mcp` while the remote URL `https://mcp.context7.com/mcp`
is used for non-Claude agents. These are two different access methods (npm vs HTTP remote), both valid.

**Normalize the naming principle:**

- **Claude Code**: npm package `@upstash/context7-mcp` via stdio (correct, no pin needed)
- **All other agents**: Remote HTTP URL `https://mcp.context7.com/mcp` (correct, no package)

Document this in `internal/components/mcp/context7.go` with a clear comment:

```go
// Context7 is available via two methods:
//
// Method A — Claude Code (stdio):
//   Uses npm package "@upstash/context7-mcp" run via npx.
//   Claude Code MCP strategy is StrategySeparateMCPFiles (stdio).
//
// Method B — All other agents (remote HTTP):
//   Uses the remote MCP endpoint https://mcp.context7.com/mcp.
//   No npm package needed; the service is cloud-hosted.
//
// Do NOT mix methods. Do NOT use npm package for HTTP-capable agents.
// Do NOT use remote URL for Claude Code (it requires stdio).
```

No code change needed beyond what FIX-02 already applied (removing `@latest` pin).

---

### FIX-28: Add D1-D4 → Posture Mapping for Generalist

**Files:** All `general-orchestrator.md` variants.

This is already covered by the table in FIX-15 (`D1-D4 Auto-Posture Map`).

**Additional change:** In the Sub-Agent Launch Template, when delegating to Generalist
for auto-detected (non-workflow) tasks, the orchestrator MUST compute and inject posture:

```markdown
### Generalist Delegation (auto-detected tasks)

Before building prompt:
1. Compute D1, D2, D3, D4 from task description
2. Apply Routing Matrix → determine MODE
3. Apply D1-D4 Auto-Posture Map → determine posture
4. Inject posture as FIRST LINE of sub-agent prompt

Example for MODE 2 task:
```
+++Critical
Evaluate objectively based on evidence. For each claim or proposed change:
(1) What evidence supports it?
(2) What evidence contradicts it?
(3) What alternative approach exists?
```
5. Inject [MODE 2 | D1=2, D2=1, D3=0, D4=1] header
6. Continue with Tier 0/1/2/3 injection
```

---

### FIX-29 & FIX-30 — Combined: Engram Missing from Install Verification

**Problem:** No test verifies that EVERY selected agent gets ALL required MCPs after install.

**New test:** `internal/components/mcp/mcp_completeness_test.go`

```go
package mcp_test

import (
    "encoding/json"
    "os"
    "testing"
    
    "github.com/rd-mg/architect-ai/internal/agents/claude"
    "github.com/rd-mg/architect-ai/internal/agents/opencode"
    "github.com/rd-mg/architect-ai/internal/agents/gemini"
    "github.com/rd-mg/architect-ai/internal/agents/cursor"
    "github.com/rd-mg/architect-ai/internal/agents/vscode"
    "github.com/rd-mg/architect-ai/internal/agents/windsurf"
    "github.com/rd-mg/architect-ai/internal/agents/antigravity"
    antigravitycli "github.com/rd-mg/architect-ai/internal/agents/antigravity-cli"
    "github.com/rd-mg/architect-ai/internal/components/mcp"
)

// requiredMCPs lists the server names that EVERY agent must have after install.
var requiredMCPs = []string{
    "engram",
    "context7",
    "sequential-thinking",
    "context-mode",
    "codegraph",
    "notebooklm-mcp",
}

var allAdapters = []struct {
    name    string
    adapter interface{ SupportsMCP() bool; MCPConfigPath(string, string) string }
}{
    {"claude", claude.NewAdapter()},
    {"opencode", opencode.NewAdapter()},
    {"gemini", gemini.NewAdapter()},
    {"cursor", cursor.NewAdapter()},
    {"vscode", vscode.NewAdapter()},
    {"windsurf", windsurf.NewAdapter()},
    {"antigravity", antigravity.NewAdapter()},
    {"antigravity-cli", antigravitycli.NewAdapter()},
}

func TestAllAgentsHaveRequiredMCPs(t *testing.T) {
    engramBin := "engram"
    
    for _, tc := range allAdapters {
        t.Run(tc.name, func(t *testing.T) {
            home := t.TempDir()
            
            // Generate fresh config
            cfg, err := mcp.GenerateConfig(tc.name, mcp.GenerateOptions{})
            if err != nil {
                t.Fatalf("GenerateConfig(%s): %v", tc.name, err)
            }
            
            cfgJSON, _ := json.Marshal(cfg)
            
            for _, required := range requiredMCPs {
                if !containsString(string(cfgJSON), required) {
                    t.Errorf("agent %q missing MCP server %q in generated config", tc.name, required)
                }
            }
        })
    }
}

func containsString(s, substr string) bool {
    return strings.Contains(s, `"`+substr+`"`) || strings.Contains(s, `'`+substr+`'`)
}
```

---

## Acceptance Criteria — Full Suite

All 30 fixes are complete when:

1. `go test -race ./...` passes with no failures
2. `go vet ./...` reports no issues
3. `go test ./internal/components/mcp/...` — all MCP tests pass including new completeness test
4. `go test ./internal/agents/antigravity-cli/...` — new adapter tests pass
5. `architect-ai install --agent antigravity-cli --dry-run` — shows correct plan
6. `architect-ai install --dry-run` for ANY agent — generated config includes all 6 MCPs
7. E2E test: Fresh install for Claude → `~/.claude/mcp/*.json` — no underscore key names
8. E2E test: Fresh install for any agent → config includes `notebooklm-mcp`
9. E2E test: Gemini sync (not fresh install) → does NOT overwrite `general`/`ui` settings keys
10. `cmd/eintegrate/main.go` — all existing checks pass + new E-24 (no Spanish regionalisms)
11. Manual review: sdd-archive.md — has Project File Updates section with README + __manifest__.py
12. Manual review: sdd-apply.md — has Pre-Apply Completeness Gate section
13. Manual review: sdd-verify.md — spawns two parallel Task tool calls
14. Manual review: persona files — have Architecture Constitution + MCP table sections
15. Manual review: all general-orchestrator.md — routing tables have max 2 postures per agent
16. Manual review: no `(ES: ...)` annotations in any orchestrator routing table

---

## File Change Summary

| File | Type | Fixes |
|------|------|-------|
| `internal/components/mcp/inject.go` | Modify | FIX-01, FIX-14 |
| `internal/components/mcp/overlay.go` | Modify | FIX-01, FIX-05 |
| `internal/components/mcp/generator.go` | Modify | FIX-02, FIX-03, FIX-05, FIX-21 |
| `internal/components/mcp/context7.go` | Modify | FIX-27 |
| `internal/components/mcp/mcp_completeness_test.go` | New | FIX-29 |
| `internal/agents/antigravity-cli/adapter.go` | New | FIX-04 |
| `internal/agents/antigravity-cli/installer.go` | New | FIX-04 |
| `internal/agents/antigravity-cli/adapter_test.go` | New | FIX-05 |
| `internal/model/model.go` | Modify | FIX-05 |
| `internal/catalog/agents.go` | Modify | FIX-05 |
| `internal/assets/assets.go` | Modify | FIX-05 |
| `internal/assets/renderer.go` | New | FIX-24 |
| `internal/components/hooks/context_mode.go` | New | FIX-23 |
| `internal/app/skills_cmd.go` | Modify | FIX-25 |
| `internal/system/deps.go` | Modify | FIX-26 |
| `cmd/eintegrate/main.go` | Modify | FIX-17 (E-24 check) |
| `_shared/adaptive-reasoning-gate-v2.md` | Modify | FIX-19, FIX-20 |
| `_shared/architect-identity.md` | Modify | FIX-17 |
| `{all}/sdd-explore.md` (8 agents) | Modify | FIX-06 |
| `{all}/sdd-verify.md` (8 agents) | Modify | FIX-08, FIX-22 |
| `{all}/sdd-apply.md` (8 agents) | Modify | FIX-09 |
| `{all}/sdd-archive.md` (8 agents) | Modify | FIX-10, FIX-13 |
| `{all}/sdd-init.md` (8 agents) | Modify | FIX-12 |
| `{all}/general-orchestrator.md` (all agents) | Modify | FIX-11, FIX-15, FIX-16, FIX-17, FIX-20 |
| `{all}/sdd-orchestrator.md` (all agents) | Modify | FIX-07, FIX-17, FIX-20 |
| `{all}/persona-architect.md` (all agents) | Modify | FIX-18 |
| `generic/persona-neutral.md` | Modify | FIX-18 |
| `generic/context-mode-routing-policy.md` | Modify | FIX-07 |
| All orchestrators (routing tables) | Modify | FIX-15, FIX-16 |

---

## Appendix A — Additional Critical Bugs Found During Deep Read

### BUG-A1: L1a/L1b Labels SWAPPED in `_shared/` Identity Files 🔴

**Discovered in:** `internal/assets/_shared/general-orchestrator-identity.md` line 2 and
`internal/assets/_shared/sdd-orchestrator-identity.md` line 1.

```
general-orchestrator-identity.md says: "L1b General Orchestrator"  ← WRONG, should be L1a
sdd-orchestrator-identity.md says:     "L1a Tactical Orchestrator" ← WRONG, should be L1b
```

But `adaptive-reasoning-gate-v2.md` correctly states:
```
"The orchestrator (L1a for general tasks; L1b for SDD tasks) has computed your D1-D4"
```

The L1 labels are **completely swapped** in the identity files. Every sub-agent that reads
these identity files will see the wrong layer designation.

**Fix — File: `internal/assets/_shared/general-orchestrator-identity.md`:**

```markdown
# general-orchestrator — L1a General Orchestrator

You are **general-orchestrator**, L1a General Orchestrator of architect-ai.
```

**Fix — File: `internal/assets/_shared/sdd-orchestrator-identity.md`:**

```markdown
# sdd-orchestrator — L1b SDD Orchestrator

You are **sdd-orchestrator**, L1b SDD Orchestrator of architect-ai.

## Role

Drive Spec-Driven Development (SDD) pipeline. Coordinate change lifecycle across phases:
`init` → `onboard` → `explore` → `propose` → `spec` → `design` → `tasks` → `apply` → `verify` → `archive`

## Authority Scope

- READ: Change state, spec files, tasks, implementation code, Engram memory
- WRITE: state.yaml, apply-progress.md, verify-report.md
- DELEGATE: Phase execution to L2 specialized sub-agents (sdd-explore, sdd-apply, sdd-verify, etc.)
```

**eintegrate check:**

```go
// E-25: L1 labels correct in identity files
if checkFile("_shared/general-orchestrator-identity.md", "L1b") {
    errs = append(errs, "E-25a: general-orchestrator-identity.md has wrong L1b label (should be L1a)")
}
if checkFile("_shared/sdd-orchestrator-identity.md", "L1a") {
    errs = append(errs, "E-25b: sdd-orchestrator-identity.md has wrong L1a label (should be L1b)")
}
```

---

### BUG-A2: GATE_ERROR on Missing Placeholder — Gate Blocks Itself 🔴

**Discovered in:** `_shared/adaptive-reasoning-gate-v2.md`

The Gate Error Protocol says:
```
IF any of the above fields contain unfilled placeholders ({INJECTED_MODE}, {D1}, etc.):
  → DO NOT self-score
  → DO NOT proceed with work
  → Emit: [GATE_ERROR: orchestrator did not inject mode — required fields missing]
  → Set status: blocked in return envelope
  → Stop
```

This is correct behavior. However, the orchestrators do NOT consistently verify they have
filled all placeholders before injecting the gate. If the orchestrator injects the gate
text literally without resolving `{INJECTED_MODE}`, `{D1}`, etc., the sub-agent silently
blocks with GATE_ERROR and the orchestrator may not detect it.

**Fix:** Add pre-dispatch validation to ALL orchestrators before sub-agent Task call:

```markdown
## Sub-Agent Dispatch Validation (MANDATORY before Task call)

Before calling Task(agent, prompt):
1. Verify prompt contains `[MODE N | D1=X, D2=X, D3=X, D4=X]` with ACTUAL values.
   - N must be: 1, 2, 2-ERR, 3, or 3-CTX
   - X values must be integers 0-3
2. Verify prompt contains `+++{POSTURE}` with ACTUAL posture name.
3. Verify `{ATTEMPT_COUNT}` is replaced with actual integer.

IF any placeholder remains unfilled:
  → STOP. Do not call Task.
  → Log: "[ORCHESTRATOR_ERROR] Unfilled placeholders in gate: {list}"
  → Re-compute D1-D4 and retry template render (max 1 retry)
  → If still unfilled after retry: emit BLOCKED status to user
```

---

### BUG-A3: `sdd-orchestrator-identity.md` Calls SDD "L1a Tactical" — Naming Inconsistency 🟠

Multiple files use inconsistent naming for the SDD orchestrator:
- Some files: "L1a Tactical Orchestrator"
- Some files: "L1b SDD Orchestrator"  
- Blueprint: "L1b SDD Orchestrator"

The canonical naming (to be enforced everywhere):
- L0 = Thinking Agent (Proxy Router)
- L1a = General Orchestrator
- L1b = SDD Orchestrator (NOT "Tactical")

The word "Tactical" came from an earlier design where L1a=Strategic, L1b=Tactical. This was
superseded. The word "Tactical" should be removed from all files.

**Search and replace (all agents, all orchestrator files):**
```bash
# Find all occurrences
rg "L1a Tactical\|Tactical Orchestrator" internal/assets/ -l
# Replace
sed -i 's/L1a Tactical Orchestrator/L1b SDD Orchestrator/g' {files}
sed -i 's/Tactical Orchestrator/SDD Orchestrator/g' {files}
```

---

### BUG-A4: `_shared/super-orchestrator-gate.md` References L0 as "Super-Orchestrator" 🟡

**File:** `internal/assets/_shared/super-orchestrator-gate.md`

This file predates the Thin-Proxy L0 redesign. It describes L0 as a "Super-Orchestrator"
with sequential_thinking Intention Gate. Per the V4 redesign (orchestrator-redesign plan),
L0 should be a thin proxy router — NOT a super-orchestrator.

**Fix:** Update `super-orchestrator-gate.md` to reflect the new L0 design:

```markdown
<!-- architect-ai:super-orchestrator-gate:v2 -->
## Intent Router Gate (L0 — Stateless Proxy)

L0 is a ROUTING-ONLY layer. It does NOT plan, compute, or execute.

### Step 1: Session Cache (optional)
mem_search(query: "session-state/{project}/tools", project: "{project}")
IF hit AND age < 30min → extract {session_state} for forwarding
IF miss → {session_state} = {} (L1 runs its own probe and caches)

### Step 2: String Match (O(1), no LLM call, no tool call)

SDD Pattern — any match → SDD_INTENT:
- "use sdd", "start sdd", "begin sdd", "sdd-new", "sdd-continue", "sdd-ff",
  "sdd-explore", "sdd-init", "sdd-verify", "sdd-archive", "spec-driven", "/sdd"
Regex: \b(use sdd|start sdd|begin sdd|sdd[-\s]|spec-driven)\b

SDD_INTENT → forward to L1b with session_state. L1b owns conversation.
NON_SDD   → forward to L1a with session_state. L1a owns conversation.

### L0 Does NOT:
- Call sequential_thinking
- Run tool availability probe
- Compute D1-D4
- Synthesize or post-process L1 responses
- Maintain conversation state (all state via Engram)

### Architecture Constitution (always active, ~150 tokens)
{content of _shared/architecture-guardrails.md compact form}
```

---

## Appendix B — New Shared Asset Files Required

The following `_shared/` files need to be created to support the fixes above:

### B1: `internal/assets/_shared/codegraph-tools-reference.md` (NEW)

This file is injected into `sdd-explore.md`, `sdd-verify.md`, and the
research routing policy when codegraph is available.

```markdown
<!-- architect-ai:codegraph-tools-reference:v1 -->
## CodeGraph MCP Tools Reference

Requires: `codegraph` server active in MCP config.
Priority: Use BEFORE ripgrep for all relationship/impact queries.

### Tool Signatures

```
codegraph_context(query: string, maxNodes?: int = 25, format?: "markdown" | "json")
  → Semantic context pack: nodes related to query (functions, types, files, call chains)
  → Use as PRIMARY exploration tool (replaces multi-file ripgrep)

codegraph_trace(entry: string)
  → Full call chain from entry point to all leaf calls
  → Use for: "what does this function ultimately call?"

codegraph_callers(nodeId: string)
  → All nodes that call this node (inbound references)
  → Equivalent to: LspFindReferences, "who calls this?"

codegraph_callees(nodeId: string)
  → All nodes this node calls (outbound references)
  → Equivalent to: "what does this call?"

codegraph_impact(nodeId: string, depth?: int = 3)
  → Blast radius: what breaks if nodeId changes
  → Use in sdd-explore (impact surface) and sdd-verify (change validation)

codegraph_search(query: string, kind?: string, limit?: int = 10)
  → Symbol search: find function/type/class by name
  → Use instead of: rg "^func {name}"
```

### Decision Matrix: CodeGraph vs ripgrep

| Query type | Use CodeGraph | Use ripgrep |
|-----------|--------------|-------------|
| Find all callers of function X | ✅ codegraph_callers | ❌ slow, error-prone |
| Find impact radius of change X | ✅ codegraph_impact | ❌ impossible |
| Find string literal "config_key" | ❌ | ✅ rg -l "config_key" |
| Find all YAML env var references | ❌ | ✅ rg "MY_VAR" --type yaml |
| Find related functions/types | ✅ codegraph_context | ❌ pattern-only |
| Find file containing "func AuthHandler" | ✅ codegraph_search | ✅ either |

### Initialization
First use in a session: verify codegraph index is current.
Run in project root:
```bash
codegraph init --quiet  # idempotent — only re-indexes changed files
```
If codegraph command not found: skip to ripgrep fallback silently.

### Fallback (if codegraph unavailable)
All codegraph tool calls → replace with:
- codegraph_callers → rg "{function_name}" -l
- codegraph_context → sequential_thinking + rg "^func|^type|^class" sweep
- codegraph_impact → manual multi-file grep + dependency tracing
```

### B2: `internal/assets/_shared/judge-protocol.md` (NEW)

Referenced from `sdd-verify.md` for the two-judge instantiation pattern.

```markdown
<!-- architect-ai:judge-protocol:v1 -->
## Judge Protocol (Adversarial Verification)

MANDATORY in sdd-verify. TWO SEPARATE judge agents. Never combine into one.

### Judge Primary — +++Adversarial
Role: Correctness. Happy-path failures. Implementation vs spec.
Model: sonnet (or agent default)
Focus areas:
- Does implementation match every spec capability?
- Are test assertions actually meaningful (not trivially true)?
- Hostile inputs: null, empty, max, concurrent, unicode edge cases
- No TODO/FIXME/XXX in changed code

### Judge Secondary — +++Forensic
Role: Robustness. Failure modes. Production survivability.
Model: sonnet (or agent default)
Focus areas:
- FMEA failure modes handled in implementation?
- Error paths explicitly tested?
- Race conditions under concurrent load?
- Resource leak paths (file handles, goroutines, connections)?
- Upgrade/downgrade hazards?

### Verdict Scale (both judges)
APPROVED                — meets all criteria
CONDITIONALLY_APPROVED  — minor issues, can ship with follow-up task
NEEDS_CHANGES           — blocking issues, must fix before archive
UNRESOLVED              — cannot determine without more information

### Synthesis Rules
| Primary      | Secondary    | Final verdict |
|-------------|-------------|---------------|
| APPROVED    | APPROVED    | APPROVED      |
| APPROVED    | COND_APPR   | COND_APPR     |
| COND_APPR   | APPROVED    | COND_APPR     |
| COND_APPR   | COND_APPR   | COND_APPR     |
| Any NEEDS_CHANGES | Any | NEEDS_CHANGES |
| Any UNRESOLVED | Any   | UNRESOLVED (escalate) |

Combined report: mem_save("sdd/{change-name}/verify-report", combined)
```

---

## Appendix C — Complete eintegrate Checks Reference

All existing and new checks. File: `cmd/eintegrate/main.go`.

```go
// Existing checks (preserved):
// E-03: sdd-propose in skill-registry
// E-04: PHASE 1 — Gather in skill-registry
// E-05: THEY ARE PROBABLY LYING in sdd-verify
// E-06: ctx_fetch_and_index in skill-registry
// E-07: LspFindReferences / codegraph_callers in skill-registry
// E-08: KEYWORD ROUTING in GEMINI.md + antigravity-agent templates
// E-09: MATERIALIZE_COMPLETE sentinel in arch-materialize.md
// E-10: HARDENING_COMPLETE sentinel in arch-hardening.md

// Sprint 1 new checks:
func runEintegrateChecks(projectRoot string) []string {
	var errs []string

	// E-11: Token budget config present
	if !checkFile(filepath.Join(projectRoot, ".atl/config.yaml"), "token_budget") {
		errs = append(errs, "E-11: token_budget section missing from .atl/config.yaml")
	}

	// E-12: CodeGraph in at least one agent config after install
	if !checkAnyAgentConfigContains(projectRoot, "codegraph") {
		errs = append(errs, "E-12: codegraph MCP not configured for any agent")
	}

	// E-13: sdd-explore has codegraph_context call
	if !checkAnyPhaseFile(projectRoot, "sdd-explore.md", "codegraph_context") {
		errs = append(errs, "E-13: sdd-explore.md missing codegraph_context step")
	}

	// E-14: sequential-thinking fallback block present in thinking-agent
	if !checkAnyFile(projectRoot, "thinking-agent.md", "sequential_thinking tool unavailable") {
		errs = append(errs, "E-14: thinking-agent.md missing sequential-thinking fallback block")
	}

	// E-15: ctx_batch_execute in context-mode-routing-policy
	if !checkAnyFile(projectRoot, "context-mode-routing-policy.md", "ctx_batch_execute") {
		errs = append(errs, "E-15: context-mode-routing-policy.md missing ctx_batch_execute pattern")
	}

	// E-16: Protocol Shell header in at least one phase protocol
	if !checkAnyPhaseFile(projectRoot, "sdd-explore.md", "/sdd.explore{") {
		errs = append(errs, "E-16: sdd-explore.md missing Protocol Shell header")
	}

	// E-17: L0 thinking-agent must NOT have sequential_thinking Intention Gate
	if checkAnyFile(projectRoot, "thinking-agent.md", "## Intention Gate") {
		errs = append(errs, "E-17: thinking-agent.md still has sequential_thinking Intention Gate (L0 should be thin proxy)")
	}

	// E-18: L1a general-orchestrator must NOT have ROUTER GATE section
	if checkAnyFile(projectRoot, "general-orchestrator.md", "## ROUTER GATE") {
		errs = append(errs, "E-18: general-orchestrator.md still has ROUTER GATE (now owned by L0)")
	}

	// E-19: L1b sdd-orchestrator must have Step 0 D1-D4 SDD Classification
	if !checkAnyFile(projectRoot, "sdd-orchestrator.md", "D1-D4 SDD") {
		errs = append(errs, "E-19: sdd-orchestrator.md missing Step 0 D1-D4 SDD Classification")
	}

	// E-20: L0 must have Intent Router section
	if !checkAnyFile(projectRoot, "thinking-agent.md", "## Intent Router") {
		errs = append(errs, "E-20: thinking-agent.md missing Intent Router section (L0 thin proxy)")
	}

	// E-21: L0 Architecture Constitution compact form ≤ 200 tokens (~800 chars)
	constitutionText := extractSection(projectRoot, "thinking-agent.md", "## Architecture Constitution")
	if len(constitutionText) > 800 {
		errs = append(errs, fmt.Sprintf(
			"E-21: L0 Architecture Constitution exceeds 200 tokens (%d chars) — compress it", len(constitutionText)))
	}

	// E-22: L1b sdd-orchestrator has forwarded session_state check
	if !checkAnyFile(projectRoot, "sdd-orchestrator.md", "session_state") {
		errs = append(errs, "E-22: sdd-orchestrator.md missing session_state forwarding check")
	}

	// E-23: L1a general-orchestrator has Session State Reader section
	if !checkAnyFile(projectRoot, "general-orchestrator.md", "Session State Reader") {
		errs = append(errs, "E-23: general-orchestrator.md missing Session State Reader section")
	}

	// E-24: No Spanish regionalisms in routing patterns
	regionalisms := []string{
		"haceme", "vamos con", "usa sdd", "iniciar sdd", "(ES:", "ES: ",
		"sigue", "continua sdd", "guíame", "cierra el cambio",
	}
	orchestratorFiles := findFiles(projectRoot, "sdd-orchestrator.md", "general-orchestrator.md",
		"architect-identity.md")
	for _, file := range orchestratorFiles {
		for _, term := range regionalisms {
			if fileContains(file, term) {
				errs = append(errs, fmt.Sprintf(
					"E-24: Spanish regionalism %q found in %s", term, filepath.Base(file)))
			}
		}
	}

	// E-25: L1 labels correct in _shared identity files
	if fileContains(filepath.Join(projectRoot, "internal/assets/_shared/general-orchestrator-identity.md"), "L1b") {
		errs = append(errs, "E-25a: general-orchestrator-identity.md has wrong L1b label (should be L1a)")
	}
	if fileContains(filepath.Join(projectRoot, "internal/assets/_shared/sdd-orchestrator-identity.md"), "L1a") {
		errs = append(errs, "E-25b: sdd-orchestrator-identity.md has wrong L1a label (should be L1b)")
	}

	// E-26: sdd-verify.md must contain both judge-primary and judge-secondary Task calls
	if !checkAnyPhaseFile(projectRoot, "sdd-verify.md", "judge-primary") ||
		!checkAnyPhaseFile(projectRoot, "sdd-verify.md", "judge-secondary") {
		errs = append(errs, "E-26: sdd-verify.md missing dual judge instantiation (judge-primary + judge-secondary)")
	}

	// E-27: sdd-apply.md must have Pre-Apply Completeness Gate
	if !checkAnyPhaseFile(projectRoot, "sdd-apply.md", "Pre-Apply Completeness Gate") {
		errs = append(errs, "E-27: sdd-apply.md missing Pre-Apply Completeness Gate section")
	}

	// E-28: sdd-archive.md must have README.md update step
	if !checkAnyPhaseFile(projectRoot, "sdd-archive.md", "README.md") {
		errs = append(errs, "E-28: sdd-archive.md missing README.md update step")
	}

	// E-29: sdd-archive.md must have __manifest__.py update step
	if !checkAnyPhaseFile(projectRoot, "sdd-archive.md", "__manifest__.py") {
		errs = append(errs, "E-29: sdd-archive.md missing __manifest__.py update step")
	}

	// E-30: sdd-init.md must have overlay manifest.json creation
	if !checkAnyPhaseFile(projectRoot, "sdd-init.md", "manifest.json") {
		errs = append(errs, "E-30: sdd-init.md missing overlay manifest.json creation step")
	}

	// E-31: persona files must include Architecture Constitution section
	personaFiles := findFiles(projectRoot, "persona-architect.md", "persona-neutral.md")
	for _, f := range personaFiles {
		if !fileContains(f, "Architecture Constitution") {
			errs = append(errs, fmt.Sprintf("E-31: %s missing Architecture Constitution section", filepath.Base(f)))
		}
	}

	// E-32: persona files must include MCP server table
	for _, f := range personaFiles {
		if !fileContains(f, "Active MCP Servers") {
			errs = append(errs, fmt.Sprintf("E-32: %s missing Active MCP Servers section", filepath.Base(f)))
		}
	}

	// E-33: notebooklm-mcp in generated config for primary agents
	for _, agent := range []string{"claude", "vscode", "antigravity", "gemini"} {
		cfgPath := findAgentConfig(projectRoot, agent)
		if cfgPath != "" && !fileContains(cfgPath, "notebooklm-mcp") {
			errs = append(errs, fmt.Sprintf("E-33: %s agent config missing notebooklm-mcp", agent))
		}
	}

	// E-34: No underscore MCP key names for Claude
	claudeMCPDir := filepath.Join(projectRoot, ".claude", "mcp")
	for _, badKey := range []string{"sequential_thinking", "context_mode"} {
		if dirContains(claudeMCPDir, badKey) {
			errs = append(errs, fmt.Sprintf("E-34: Claude MCP config has deprecated underscore key: %s", badKey))
		}
	}

	return errs
}
```

---

## Appendix D — Odoo Overlay Go Implementation (Complete)

### D1: Overlay Installer Component

**New file:** `internal/components/overlay/odoo.go`

```go
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

// OdooManifest is written to .atl/overlays/odoo-{version}/manifest.json
// It activates the Odoo overlay in the sdd-orchestrator.
type OdooManifest struct {
	Overlay     string    `json:"overlay"`
	Version     string    `json:"version"`
	Active      bool      `json:"active"`
	AddonsPath  string    `json:"addons_path"`
	DetectedAt  time.Time `json:"detected_at"`
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
```

**Wire into install pipeline** (`internal/cli/install.go`):

```go
// After basic agent install, detect and configure Odoo overlay:
if projectDir, err := findProjectRoot(); err == nil {
    if detected, version, addonsPath := overlay.DetectOdoo(projectDir); detected {
        fmt.Printf("  Odoo %s project detected\n", version)
        if err := overlay.InstallOverlayManifest(projectDir, version, addonsPath, installOpts.DryRun); err != nil {
            fmt.Fprintf(os.Stderr, "Warning: Odoo overlay activation failed: %v\n", err)
        }
        if err := overlay.UpdateATLConfig(projectDir, version, addonsPath, installOpts.DryRun); err != nil {
            fmt.Fprintf(os.Stderr, "Warning: .atl/config.yaml update failed: %v\n", err)
        }
    }
}
```

---

## Appendix E — Sprint Execution Order & Dependencies

```
WEEK 1:
  Day 1-2: BUG-A1 (L1 label swap) — foundational, blocks all orchestrator work
  Day 2-3: FIX-02 (Claude key names) — must fix before any Claude testing
  Day 3-4: FIX-03 (notebooklm generators) — small, independent
  Day 4-5: FIX-10 (archive step numbering) — small fix, needed for FIX-13

WEEK 2:
  Day 1-3: FIX-04 + FIX-05 (Antigravity CLI adapter + registration) — large, independent
  Day 3-4: FIX-01 (InjectCodeGraph + InjectContextMode) — depends on overlay.go pattern
  Day 4-5: FIX-14 (wire InjectContextMode into pipeline) — depends on FIX-01

WEEK 3:
  Day 1-2: FIX-06 (codegraph in sdd-explore — all agents) — prompt only
  Day 2-3: FIX-07 (codegraph in research routing — all orchestrators) — prompt only
  Day 3-4: FIX-08 (dual judge instantiation in sdd-verify) — prompt only
  Day 4-5: FIX-09 (pre-apply completeness gate) — prompt only

WEEK 4:
  Day 1-2: FIX-11 (L1a/L1b label unification — all orchestrators) — large prompt sweep
  Day 2-3: FIX-15 + FIX-16 (posture max-2 fix + routing table unification)
  Day 3-4: FIX-17 (remove Spanish regionalisms — all files)
  Day 4-5: FIX-19 (posture assignment spec in gate-v2) + FIX-20 (remove duplicate gate)

WEEK 5:
  Day 1-2: FIX-12 + FIX-13 (sdd-init manifest.json + sdd-archive file updates)
  Day 2-3: FIX-18 (persona guardrails + MCP sections)
  Day 3-4: FIX-21 (Gemini generator split) + FIX-22 (WCAG conditional)
  Day 4-5: FIX-23 (context-mode hook writer)

WEEK 6:
  Day 1-2: FIX-24 ({{ include }} template rendering)
  Day 2-3: FIX-25 (skill registry keyword index)
  Day 3-4: FIX-26 (uv/uvx detection) + Appendix D (overlay installer)
  Day 4-5: Appendix C (all eintegrate checks E-11 through E-34)
           BUG-A2 (gate error dispatch validation)
           BUG-A3 (Tactical → SDD rename sweep)
           BUG-A4 (super-orchestrator-gate.md update)
           New _shared files: codegraph-tools-reference.md, judge-protocol.md

WEEK 7:
  Full regression: go test -race ./... + e2e/docker-test.sh
  Manual verification of all 16 acceptance criteria
  eintegrate run: expect all E-03 through E-34 to pass (0 failures)
```

---

## Appendix F — Rollback Strategy

Each sprint has an atomic rollback point:

```bash
# Before starting each sprint — create backup:
architect-ai install --dry-run  # verify current state clean
git stash                       # save WIP
git tag sprint-N-start          # tag before changes

# If sprint goes wrong:
git checkout sprint-N-start
git stash pop

# Config files (not in git):
architect-ai restore --latest   # restore pre-sprint backup
```

**Per-fix rollback:**

| Fix | Rollback action |
|-----|----------------|
| FIX-01,02,03 (generator) | Revert `generator.go`, re-run `go generate` |
| FIX-04,05 (CLI adapter) | Delete `internal/agents/antigravity-cli/`, revert model + catalog |
| FIX-06-09 (prompt files) | `git checkout internal/assets/` for affected phase files |
| FIX-10-13 (archive/init) | `git checkout` affected phase files |
| FIX-14-20 (orchestrators) | `git checkout internal/assets/` for all orchestrator files |
| FIX-21 (Gemini generator) | Revert `generator.go:generateGemini()` |
| FIX-23 (hook writer) | Delete `internal/components/hooks/context_mode.go` |
| FIX-24 (template renderer) | Delete `internal/assets/renderer.go`, revert callers |
| FIX-25 (skill keyword index) | Revert `skills_cmd.go`, regenerate skill-registry |
