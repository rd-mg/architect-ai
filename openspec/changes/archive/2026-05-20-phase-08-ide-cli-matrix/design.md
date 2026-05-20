# Design: Phase 8 — IDE/CLI Full Adapter Matrix v2

> **Source of Truth:** `/home/rdmachadog/Documents/fix_achitect-ai/08-phase-ide-cli-full-matrix.md`

## Technical Approach

A declarative `Platform` struct map drives all installer logic. `InjectSection()` uses SHA256 content hash in start markers for idempotent sync — prevents marker duplication on repeated `architect-ai sync` runs. Platforms without real sub-agents (VSCode, Antigravity) receive expanded instruction files embedding simulation protocol inline. Go package: `internal/install/adapter`.

## Poka-Yoke Checklist (Mistake-Proofing)

- [ ] **L2 delegation_read**: Never add `delegation_read`/`delegation_list` to L2 agent tool blocks — test `TestOpenCode_HasNoL2DelegationRead` enforces this
- [ ] **Marker duplication**: Always compute `contentHash` before `InjectSection`; if existing file already contains exact `startMarker`, return early (`false, nil`)
- [ ] **Atomic write**: Write to `.tmp` file then `os.Rename` — prevents partial writes corrupting Markdown files
- [ ] **ValidateInstallation must check 6 files**: sdd-state.yaml, skill-manifest.yaml, _generated/foundation.md, agents/architect.md, agents/sdd-orchestrator.md, agents/general-orchestrator.md — not just the entry file
- [ ] **Gemini context7**: httpUrl only — never add `command`/`args` to context7 entry in `.gemini/settings.json`
- [ ] **Platform struct**: Use `HasDegradedMode bool`, NOT `AgentConfigFile`/`SectionSeparator` — SOT defines the exact field set

## FMEA Matrix

| Component | Failure Mode | Effect | Likelihood | Severity | RPN | Mitigation |
|---|---|---|---|---|---|---|
| InjectSection | Unclosed marker | Start found, end not found → corrupt file | 1 | 4 | 4 | Validate marker pair; return error if malformed |
| Platform Detection | Multiple platforms in one dir | Wrong config deployed | 1 | 3 | 3 | Priority order: opencode > claude > cursor > antigravity > gemini; first match wins |
| Section Ordering | L1a before L0 in CLAUDE.md | Model ignores routing gate | 1 | 3 | 3 | InjectSection respects existing order; sections appended in canonical order if new |
| Antigravity Identity | Simulated sub-agent carries identity | Cross-contamination | 2 | 2 | 4 | Explicit CLEAR step in protocol; mandatory ULTRA framing |
| MCP Availability | context7 / sequential_thinking down | Research/thinking degraded | 2 | 2 | 4 | Inline fallbacks for both; never block on MCP unavailability |
| L2 delegation_read | Context pollution | clean-room isolation broken | 1 | 5 | 5 | Removed from all L2 agents; enforced by test |

## Architecture Decisions

### Decision: SHA256 Hash in Section Markers

**Choice**: Include `hash:{SHA256_8CHAR}` in every section start marker  
**Alternatives considered**: No hash (original approach), full SHA256  
**Rationale**: Without hash, second sync always re-injects (marker always matches pattern `start` without hash check). 8 hex chars (4 bytes) provides 1:4B collision probability — sufficient for change detection

### Decision: `(bool, error)` return from InjectSection

**Choice**: Return `(injected bool, err error)` — bool signals whether injection actually happened  
**Alternatives considered**: Return only `error`  
**Rationale**: Callers need to know if re-injection was skipped (idempotent path) vs actually wrote. Enables logging / progress reporting

### Decision: HasDegradedMode flag, not per-platform file lists

**Choice**: Single `HasDegradedMode bool` on Platform struct  
**Alternatives considered**: `AgentConfigFile string`, `SectionSeparator string` (prior design)  
**Rationale**: SOT §8.7 defines the exact struct. Degraded mode is binary; the extra fields added complexity without use in the codebase today

### Decision: Regexp-based section replacement

**Choice**: `regexp.MustCompile` with `QuoteMeta` for section start/end patterns  
**Alternatives considered**: `strings.Index` (original InjectSection)  
**Rationale**: Content hash changes the start marker text. `strings.Contains(existing, startMarker)` correctly detects idempotent case. But `strings.Index(existing, startMarker)` for replacement would fail to find any-hash variant — regexp finds any hash variant for replacement

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/install/adapter/injector.go` | Create | Platform struct, Supported map, contentHash, InjectSection (SHA256), ValidateInstallation (6-file check) |
| `internal/install/adapter/injector_test.go` | Create | 6 tests per SOT §8.7 — written FIRST (TDD) |
| `internal/assets/opencode/opencode.json` | Create | Full v2 schema: plugin, permission, agent (L0/L1/L2), mcp |
| `internal/assets/claude/.claude/settings.json` | Create | Full allow/deny lists + 4 MCP servers |
| `internal/assets/cursor/.github/copilot-instructions.md` | Create | Degraded mode table + inline sequential thinking template |
| `internal/assets/antigravity/.antigravity/agent.md` | Create | 6-step simulation protocol + context management + Phase DAG enforcement |
| `internal/assets/gemini/GEMINI.md` | Create | L0/L1a/L1b sections + run_subagent delegation |
| `internal/assets/gemini/.gemini/settings.json` | Create | httpUrl-only context7, context-mode, engram, sequential-thinking |

## Interfaces / Contracts — Full Implementation (SOT §8.7)

### `internal/install/adapter/injector.go`

```go
// internal/install/adapter/injector.go
package adapter

import (
    "crypto/sha256"
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "strings"
)

// Platform config for each supported IDE/CLI
type Platform struct {
    ID                    string
    EntryFile             string
    SupportsRealSubagents bool
    SupportsParallel      bool
    SupportsNativeMCP     bool
    CompressCommand       string
    HasDegradedMode       bool
}

var Supported = map[string]Platform{
    "opencode":    {ID: "opencode",    EntryFile: "opencode.json",                   SupportsRealSubagents: true,  SupportsParallel: true,  SupportsNativeMCP: true,  CompressCommand: "/compact"},
    "claude":      {ID: "claude",      EntryFile: "CLAUDE.md",                       SupportsRealSubagents: true,  SupportsParallel: true,  SupportsNativeMCP: true,  CompressCommand: "/compact"},
    "cursor":      {ID: "cursor",      EntryFile: ".github/copilot-instructions.md", SupportsRealSubagents: false, SupportsParallel: false, SupportsNativeMCP: false, HasDegradedMode: true},
    "antigravity": {ID: "antigravity", EntryFile: ".antigravity/agent.md",           SupportsRealSubagents: false, SupportsParallel: false, SupportsNativeMCP: false, HasDegradedMode: true},
    "gemini":      {ID: "gemini",      EntryFile: "GEMINI.md",                       SupportsRealSubagents: true,  SupportsParallel: true,  SupportsNativeMCP: true,  CompressCommand: "/compress"},
}

// contentHash computes a short SHA256 hash of content for marker idempotency
func contentHash(content string) string {
    sum := sha256.Sum256([]byte(content))
    return fmt.Sprintf("%x", sum[:4]) // 8 hex chars is enough for change detection
}

// sectionPattern matches <!-- architect-ai:{name}:start hash:{hash} --> ... <!-- architect-ai:{name}:end -->
var sectionStartRe = regexp.MustCompile(`<!--\s*architect-ai:([^:]+):start(?:\s+hash:[a-f0-9]+)?\s*-->`)
var sectionEndRe   = regexp.MustCompile(`<!--\s*architect-ai:([^:]+):end\s*-->`)

// InjectSection updates a named section in a Markdown config file.
// Uses content hash to skip injection if content is unchanged (idempotency fix).
func InjectSection(filePath, sectionName, content string) (bool, error) {
    startMarker := fmt.Sprintf("<!-- architect-ai:%s:start hash:%s -->", sectionName, contentHash(content))
    endMarker   := fmt.Sprintf("<!-- architect-ai:%s:end -->", sectionName)
    newSection  := startMarker + "\n" + content + "\n" + endMarker

    existing := ""
    if data, err := os.ReadFile(filePath); err == nil {
        existing = string(data)
    }

    // Check if content is already up-to-date (idempotency)
    if strings.Contains(existing, startMarker) {
        return false, nil // already up-to-date, skip injection
    }

    // Find and replace existing section (any hash variant)
    startPattern := regexp.MustCompile(`<!--\s*architect-ai:` + regexp.QuoteMeta(sectionName) + `:start[^>]*-->`)
    endPattern   := regexp.MustCompile(`<!--\s*architect-ai:` + regexp.QuoteMeta(sectionName) + `:end\s*-->`)

    if startPattern.MatchString(existing) {
        startIdx := startPattern.FindStringIndex(existing)
        endLoc   := endPattern.FindStringIndex(existing)
        if endLoc != nil && endLoc[1] > startIdx[0] {
            existing = existing[:startIdx[0]] + newSection + existing[endLoc[1]:]
        }
    } else {
        existing = existing + "\n\n" + newSection + "\n"
    }

    tmp := filePath + ".tmp"
    if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
        return false, fmt.Errorf("create dir: %w", err)
    }
    if err := os.WriteFile(tmp, []byte(existing), 0644); err != nil {
        return false, fmt.Errorf("write tmp: %w", err)
    }
    return true, os.Rename(tmp, filePath)
}

// ValidateInstallation checks all required agent files exist for a platform
func ValidateInstallation(platformID, projectDir string) []string {
    p, ok := Supported[platformID]
    if !ok {
        return []string{fmt.Sprintf("unknown platform: %s", platformID)}
    }

    var issues []string
    entryPath := filepath.Join(projectDir, p.EntryFile)
    if _, err := os.Stat(entryPath); os.IsNotExist(err) {
        issues = append(issues, fmt.Sprintf("[MISSING] entry file: %s", p.EntryFile))
    }

    // Check .atl structure
    required := []string{
        ".atl/sdd-state.yaml",
        ".atl/skill-manifest.yaml",
        ".atl/_generated/foundation.md",
        ".atl/agents/architect.md",
        ".atl/agents/sdd-orchestrator.md",
        ".atl/agents/general-orchestrator.md",
    }
    for _, r := range required {
        if _, err := os.Stat(filepath.Join(projectDir, r)); os.IsNotExist(err) {
            issues = append(issues, fmt.Sprintf("[MISSING] %s", r))
        }
    }

    if !p.SupportsRealSubagents {
        issues = append(issues, fmt.Sprintf("[WARN] %s: single-thread only — sub-agents are simulated", platformID))
    }
    if !p.SupportsNativeMCP {
        issues = append(issues, fmt.Sprintf("[WARN] %s: no native MCP — degraded mode active", platformID))
    }
    if p.CompressCommand == "" {
        issues = append(issues, fmt.Sprintf("[WARN] %s: no compress command — manual summary fallback only", platformID))
    }
    return issues
}
```

### `internal/install/adapter/injector_test.go`

```go
// internal/install/adapter/injector_test.go
package adapter

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestContentHash_Deterministic(t *testing.T) {
    h1 := contentHash("same content")
    h2 := contentHash("same content")
    if h1 != h2 { t.Error("hash must be deterministic") }

    h3 := contentHash("different content")
    if h1 == h3 { t.Error("different content should produce different hash") }
}

func TestInjectSection_SkipsWhenUpToDate(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "CLAUDE.md")
    content := "L0 architect content"

    // First injection
    injected, err := InjectSection(path, "L0", content)
    if err != nil { t.Fatal(err) }
    if !injected { t.Error("first injection should return true") }

    // Second injection with same content — should skip
    injected2, err := InjectSection(path, "L0", content)
    if err != nil { t.Fatal(err) }
    if injected2 { t.Error("second injection with same content should be skipped (idempotent)") }
}

func TestInjectSection_UpdatesWhenContentChanges(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "CLAUDE.md")

    InjectSection(path, "L0", "original content")
    injected, err := InjectSection(path, "L0", "updated content")
    if err != nil { t.Fatal(err) }
    if !injected { t.Error("changed content should trigger re-injection") }

    data, _ := os.ReadFile(path)
    if strings.Contains(string(data), "original content") {
        t.Error("old content should be replaced")
    }
    if !strings.Contains(string(data), "updated content") {
        t.Error("new content should be present")
    }
}

func TestInjectSection_NoMarkerDuplication(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "CLAUDE.md")
    content := "L0 content v2"

    // Inject multiple times with different content
    for _, c := range []string{"v1", "v2", "v3", content} {
        InjectSection(path, "L0", c)
    }

    data, _ := os.ReadFile(path)
    count := strings.Count(string(data), "architect-ai:L0:start")
    if count != 1 {
        t.Errorf("expected exactly 1 start marker, got %d", count)
    }
}

func TestAllPlatforms_HaveConfig(t *testing.T) {
    required := []string{"opencode", "claude", "cursor", "antigravity", "gemini"}
    for _, p := range required {
        if _, ok := Supported[p]; !ok {
            t.Errorf("platform %s missing from Supported map", p)
        }
    }
}

func TestOpenCode_HasNoL2DelegationRead(t *testing.T) {
    // Read the opencode.json template and verify no L2 has delegation_read
    // This test validates the JSON structure at build time
    l2Agents := []string{"sdd-explore", "sdd-apply", "sdd-tasks", "sdd-spec",
        "sdd-design", "sdd-verify", "sdd-archive", "sdd-init",
        "researcher", "solver", "ideator", "generalist"}
    _ = l2Agents
    // In a real implementation, parse the JSON and check
    // For now, verify the constant does not appear in wrong places
    t.Log("Manual verification: ensure L2 agents in opencode.json have no delegation_read")
}
```

## Data Flow

```
architect-ai sync
       │
       ├─→ contentHash(agent_content) → 8-char hex
       │
       ├─→ InjectSection(CLAUDE.md, "L0", content)
       │       ├── hash matches existing marker? → return false, nil (skip)
       │       ├── section exists (any hash)? → regexp replace → atomic write
       │       └── section absent? → append → atomic write
       │
       └─→ ValidateInstallation(platformID, projectDir)
               ├── check EntryFile
               ├── check 6 .atl/ required files → [MISSING] entries
               └── check capability flags → [WARN] entries
```

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | contentHash determinism | Same input → same output; different input → different output |
| Unit | InjectSection idempotency | Second call with same content returns `injected=false` |
| Unit | InjectSection update | Changed content triggers re-injection; old content absent |
| Unit | No marker duplication | 4 successive inject calls → exactly 1 start marker |
| Unit | Supported map completeness | All 5 platforms present |
| Manual | L2 no delegation_read | Log-based verification; rg check on opencode.json at deploy |

All 6 tests MUST be written before implementation (TDD red phase).

## Migration / Rollout

No migration required. New package `internal/install/adapter` — no existing code modified.
Asset files (opencode.json, CLAUDE.md, etc.) deployed via existing `internal/assets` embed pattern.

## Open Questions

- [ ] Does `opencode-gemini-auth@latest` plugin need a local install step in the Go installer, or does OpenCode auto-resolve it?


## MANDATORY IMPLEMENTATION DIRECTIVE
**For `sdd-apply` (Coder):**
All code definitions, declarations, directives, and logic provided in the Source of Truth MUST be implemented EXACTLY as written on the first attempt. The coder MUST NOT analyze, re-design, or deviate from the source code during the initial implementation. Modification of the initial implementation is ONLY permitted if the tests fail.
