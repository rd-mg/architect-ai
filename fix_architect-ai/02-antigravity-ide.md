# Fix & Improvement Plan — `antigravity-ide` Agent

**Agent ID:** `antigravity` (internal) → rename to `antigravity-ide` in V4
**Runtime:** Antigravity IDE — Google's standalone IDE coding agent
**Product string:** `antigravity` (distinct from `antigravity-cli`)
**NOT a VS Code extension** — it is a separate, standalone IDE with its own editor,
  workspace management, transcript system, and Python SDK
**SDK:** `pip install google-antigravity`
**MCP config:** `mcp_config.json` (array format, `authProviderType: google_credentials`)
**Skills:** `.agents/skills/<name>/SKILL.md` with YAML frontmatter
**Assets:** `internal/assets/antigravity/` → rename to `internal/assets/antigravity-ide/`
**Go Adapter:** None (config files written to project directory)
**Priority:** 🟠 High — IDE hooks use named-group format; MCP uses array format; both differ from CLI

> ⚠️ **CODE VERIFICATION REQUIRED**: All paths, commands, binaries, JSON structures, and
> field names in this plan must be verified against the actual Antigravity IDE documentation
> at `https://antigravity.google/docs/ide-*` before implementation.

---

## 1. What Antigravity IDE Actually Is

Antigravity IDE is Google's **own standalone coding agent IDE** — not a VS Code extension,
not a plugin for another editor. It has:

- Its **own editor environment** with workspace management
- A **Python SDK** (`google-antigravity`) for building custom agents on top of the IDE
- An **`mcp_config.json`** for MCP servers using an **array format** (different from CLI's object format)
- A **hooks system** using **named hook groups** (structurally different from CLI)
- A **skills system** at `.agents/skills/<skill-name>/SKILL.md`
- **Declarative safety policies** via `from google.antigravity.hooks.policy import deny, allow, ask_user`
- **Google Application Default Credentials** via `gcloud auth application-default login`
- **OAuth support** in MCP config
- Workspace fields: `workspacePaths`, `transcriptPath`, `artifactDirectoryPath`
- Files stored in `.gemini/antigravity/` within the workspace
- PreInvocation hook can inject `ephemeralMessage` (zero context cost)
- Stop hook returns `{"decision": "continue|terminate", "reason": "..."}`

---

## 2. Hooks Format Differences (IDE vs CLI)

### IDE hooks.json — Named Group Format

```json
{
  "my-linter-hook": {
    "PostToolUse": [
      {
        "matcher": "run_command",
        "hooks": [
          { "type": "command", "command": "./scripts/lint.sh", "timeout": 10 }
        ]
      }
    ]
  },
  "safety-gate": {
    "enabled": false,
    "PreToolUse": [
      {
        "matcher": "run_command",
        "hooks": [{ "command": "./scripts/safety-check.sh" }]
      }
    ]
  },
  "reminder": {
    "PreInvocation": [
      { "type": "command", "command": "./scripts/reminder.sh" }
    ]
  }
}
```

### IDE PreToolUse output — SAME as CLI (NOT empty `{}`)

```json
{
  "decision": "ask",
  "reason": "Requires confirmation for test execution.",
  "permissionOverrides": ["command(npm test)"]
}
```

### IDE PostToolUse output — Empty object

```json
{}
```

### IDE PreInvocation output — Can inject ephemeralMessage (IDE-specific)

```json
{
  "injectSteps": [
    { "ephemeralMessage": "Remember to lint before committing." }
  ]
}
```

### IDE Stop hook output — decision field (IDE-specific)

```json
{ "decision": "continue", "reason": "Not done yet" }
```

Input to Stop hook:

```json
{
  "executionNum": 1,
  "terminationReason": "model_stop",
  "error": "",
  "fullyIdle": true,
  "conversationId": "ec33ebf9-...",
  "workspacePaths": ["/workspace/project"],
  "transcriptPath": "/workspace/project/.gemini/antigravity/transcript.jsonl",
  "artifactDirectoryPath": "/workspace/project/.gemini/antigravity/artifacts"
}
```

---

## 3. MCP Configuration Format (IDE)

IDE uses **array format** for `mcpServers` (different from CLI's object/plugin format):

```json
{
  "mcpServers": [
    {
      "name": "MyLocalServer",
      "command": "/path/to/mcp/server",
      "args": ["--port", "8080"],
      "env": { "API_KEY": "YOUR_KEY" },
      "cwd": "/path/to/working/dir",
      "authProviderType": "google_credentials",
      "oauth": {
        "clientId": "YOUR_CLIENT_ID",
        "clientSecret": "YOUR_CLIENT_SECRET"
      },
      "disabled": false,
      "disabledTools": []
    },
    {
      "name": "RemoteService",
      "serverUrl": "https://api.example.com/mcp",
      "headers": { "Authorization": "Bearer YOUR_TOKEN" }
    }
  ]
}
```

> ⚠️ **VERIFY**: The actual path where `mcp_config.json` is written for the IDE.
> Based on docs the transcript is at `.gemini/antigravity/` — confirm `mcp_config.json` location.

---

## 4. Current State in architect-ai

### Assets Present ✅

All 9 SDD phase protocols plus `architect.md`, `sdd-orchestrator.md`, `thinking-agent.md`.

### Critical Bugs

1. **Hook format is WRONG**: Current code writes CLI-style `{"hooks": {...}}` wrapper.
   IDE format uses named groups directly at root level.

2. **MCP format is WRONG**: Current code may write object-format mcpServers.
   IDE requires array-format mcpServers.

3. **Asset directory not split**: `antigravity/` serves both IDE and CLI (there is no CLI-specific dir).

4. **sequential-thinking absent from MCP config** ❌

5. **context-mode absent from MCP config** ❌

6. **codegraph absent from MCP config** ❌

---

## 5. Gap Analysis

| ID | Gap | Severity |
|----|-----|---------|
| AI-01 | IDE hooks.json uses CLI wrapper format — WRONG | 🔴 Critical |
| AI-02 | MCP array format not used (may be object format) | 🔴 Critical |
| AI-03 | sequential-thinking absent from IDE mcp_config.json | 🔴 Critical |
| AI-04 | context-mode absent from IDE mcp_config.json | 🟠 High |
| AI-05 | codegraph absent from IDE mcp_config.json | 🟠 High |
| AI-06 | PreInvocation ephemeralMessage not used for routing reminder | 🟡 Medium |
| AI-07 | Stop hook not implemented (no completion sentinel) | 🟡 Medium |
| AI-08 | Asset dir not split from CLI assets | 🔴 Critical |
| AI-09 | Python SDK (`google-antigravity`) not referenced in install docs | Low |
| AI-10 | Google ADC setup not triggered at install time | 🟡 Medium |
| AI-11 | Protocol Shell format (David Kim CE) absent from phase protocols | 🟠 High |
| AI-12 | Token budget tracking (David Kim CE) absent from orchestrator | 🟠 High |
| AI-13 | Self-refinement quality loop (David Kim CE) absent | 🟡 Medium |
| AI-14 | Context7 targeted topic fetch not in sdd-explore | 🟡 Medium |

---

## 6. Fix Plan

### Fix AI-01 — IDE hooks.json Format

**File to create/update:** `internal/assets/antigravity-ide/hooks.json` template

```go
// internal/components/mcp/antigravity_ide.go
// ⚠️ VERIFY: Confirm named-group root format against live IDE docs before shipping

var antigravityIDEHooksJSON = []byte(`{
  "architect-ai-safety": {
    "enabled": true,
    "PreToolUse": [
      {
        "matcher": "run_command",
        "hooks": [
          {
            "type": "command",
            "command": "architect-ai hook ide-pretooluse",
            "timeout": 5
          }
        ]
      }
    ]
  },
  "architect-ai-routing": {
    "PreInvocation": [
      {
        "type": "command",
        "command": "architect-ai hook ide-preinvocation",
        "timeout": 2
      }
    ]
  },
  "architect-ai-stop": {
    "Stop": [
      {
        "type": "command",
        "command": "architect-ai hook ide-stop",
        "timeout": 3
      }
    ]
  }
}`)
// PreToolUse output contract (VERIFY against docs):
// { "decision": "allow|deny|ask|force_ask", "reason": "...", "permissionOverrides": [...] }
// PostToolUse output: {}
// Stop output: { "decision": "continue|terminate", "reason": "..." }
```

### Fix AI-02 + AI-03 + AI-04 + AI-05 — MCP Array Config

```go
// internal/components/mcp/antigravity_ide.go
// ⚠️ VERIFY: mcp_config.json actual path in IDE workspace
// ⚠️ VERIFY: serverUrl vs url field name for remote servers
// ⚠️ VERIFY: whether npx is available in IDE's PATH

func AntigravityIDEMCPConfig(cfg AgentMCPConfig) []byte {
    servers := []map[string]any{
        {
            "name":    "engram",
            "command": cfg.EngramBin,   // ⚠️ VERIFY: resolve ${ENGRAM_BIN} before writing
            "args":    []string{"mcp", "--tools=agent"},
        },
        {
            "name":      "context7",
            "serverUrl": "https://mcp.context7.com/mcp",
            // ⚠️ VERIFY: IDE supports serverUrl key (vs url or httpUrl)
        },
    }
    if cfg.NPXAvailable {
        servers = append(servers, map[string]any{
            "name":    "sequential-thinking",
            "command": "npx",
            "args":    []string{"-y", "@modelcontextprotocol/server-sequential-thinking"},
            // ⚠️ VERIFY: npx in IDE PATH; may need full path /usr/local/bin/npx
        })
    }
    if cfg.ContextModeAvailable {
        servers = append(servers, map[string]any{
            "name":    "context-mode",
            "command": "context-mode",  // ⚠️ VERIFY: binary name in IDE environment
            "args":    []string{"--mcp"},
        })
    }
    if cfg.CodeGraphAvailable {
        servers = append(servers, map[string]any{
            "name":    "codegraph",
            "command": "npx",
            "args":    []string{"-y", "@colbymchenry/codegraph", "serve", "--mcp"},
            // ⚠️ VERIFY: correct package name @colbymchenry/codegraph
        })
    }
    out, _ := json.MarshalIndent(map[string]any{"mcpServers": servers}, "", "  ")
    return out
}
```

### Fix AI-06 — PreInvocation ephemeralMessage Routing Reminder

The PreInvocation hook can fire a script that outputs an `ephemeralMessage`. This costs
zero persistent context tokens — it's transient.

```bash
#!/bin/bash
# architect-ai hook ide-preinvocation
# ⚠️ VERIFY: script must output valid JSON to stdout
# ⚠️ VERIFY: ephemeralMessage field name (may vary by IDE version)
cat << 'EOF'
{
  "injectSteps": [
    {
      "ephemeralMessage": "ROUTING: Use ctx_batch_execute for multi-grep ops. Use codegraph_context before raw file reads. context7 resolve-library-id + get-library-docs for external APIs. Stop when evidence sufficient."
    }
  ]
}
EOF
```

This replaces the injected `context-mode-routing-policy.md` text block — same
information, zero persistent tokens (ephemeral = not stored in transcript).

### Fix AI-07 — Stop Hook Sentinel

```bash
#!/bin/bash
# architect-ai hook ide-stop  
# Reads stop event from stdin, writes sentinel file, outputs decision
# ⚠️ VERIFY: terminationReason field values (model_stop, user_stop, error, etc.)
# ⚠️ VERIFY: output format {decision, reason} vs just {}

INPUT=$(cat)
REASON=$(echo "$INPUT" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('terminationReason','unknown'))" 2>/dev/null || echo "unknown")

# Write sentinel for orchestrator polling
echo "APPLY_BATCH_COMPLETE:$REASON" > .atl/state/last-stop.txt

# Allow continuation if reason is user-initiated stop during mid-phase
if [ "$REASON" = "user_stop" ]; then
  echo '{"decision": "terminate", "reason": "User requested stop"}'
else
  echo '{"decision": "terminate", "reason": "Phase complete"}'
fi
```

### Fix AI-10 — Google ADC Setup

**File:** `internal/app/install_cmd.go`

```go
// After Antigravity IDE install:
// ⚠️ VERIFY: gcloud binary availability; user may not have gcloud installed
// ⚠️ VERIFY: whether ADC is required for local MCP servers or only remote OAuth ones
func ensureGoogleADC(dryRun bool) {
    out, err := exec.Command("gcloud", "auth", "application-default",
        "print-access-token").Output()
    if err != nil || len(strings.TrimSpace(string(out))) == 0 {
        fmt.Println("  Google ADC not configured. Run:")
        fmt.Println("  gcloud auth application-default login")
        if !dryRun {
            // Optionally prompt user to run it now
        }
    }
}
```

---

## 7. David Kim Context Engineering — Coverage per IDE Agent

### AI-11: Protocol Shell Format

Add Protocol Shell headers to all IDE phase protocols:

```markdown
<!-- internal/assets/antigravity-ide/sdd-phase-protocols/sdd-explore.md -->
/sdd.explore{
  intent="Map change impact surface via semantic graph + targeted research",
  input={change_topic, project, codegraph_available},
  process=[
    /adr_check{action="mem_search arch/_global/decision"},
    /semantic_map{action="codegraph_context OR rg fallback"},
    /external_docs{action="context7 get-library-docs", condition="dependency change"},
    /verify{action="empirical proof — run script or check file"}
  ],
  output={executive_summary, impact_surface, open_questions, empirical_proof, status}
}
```

### AI-12: Token Budget Tracking

IDE has 1M token window — context saturation is NOT an acute concern. However,
transcript bloat (large tool outputs) still matters for readability and cost.

```yaml
# .atl/config.yaml — IDE budget section
token_budget:
  model: gemini                 # ⚠️ VERIFY: Gemini model variant used by IDE
  total: 1048576                # 1M context confirmed from status JSON
  active_layers:
    global_directives: 500
    cognitive_posture: 300
    phase_protocol: 900
    task: 1000
    codegraph_context: 3000
    engram_carryforward: 500
  reserve: 1040376              # ~99% headroom vs Claude's 25%
  # Note: budget tracking still matters for ctx_batch_execute usage discipline
```

### AI-13: Self-Refinement Quality Loop

Add to `sdd-verify.md` (IDE version):

```markdown
## Quality Gate (Self-Refinement — David Kim CE)

Before persisting to Engram, score the proposal artifact:
  relevance_score:    Does every section address the change_topic?     (0–1)
  completeness_score: Are all impacted modules covered?                (0–1)
  coherence_score:    Is the design internally consistent?             (0–1)
  efficiency_score:   Are there redundant steps or over-engineering?   (0–1)

IF overall_score < 0.85: iterate with targeted improvements (max 2 retries)
THEN persist to Engram with scores in metadata.
```

### AI-14: Context7 Targeted Topic Fetch

Add to `sdd-explore.md` (IDE version):

```markdown
## External Docs (Context7 MCP)

When a dependency change is detected:
1. mcp__context7__resolve-library-id(libraryName: "{detected_library}")
2. mcp__context7__get-library-docs(
     context7CompatibleLibraryID: "{resolved_id}",
     topic: "{specific_aspect}",    ← ALWAYS specify topic; never fetch full docs
     tokens: 5000                   ← cap; IDE has 1M window but targeted is better
   )

Do NOT fetch entire library docs without a topic filter.
Context7 serves targeted sections — 5–15K tokens vs 100K+ for full docs.
```

---

## 8. sequential-thinking Detection & Configuration (IDE)

```go
// internal/agents/antigravity_ide/adapter.go (new)
// ⚠️ ALL PATHS REQUIRE VERIFICATION against actual IDE install locations

type IDEMCPAvailability struct {
    NPXPath         string  // ⚠️ VERIFY: may not be in IDE's PATH
    ContextModePath string  // ⚠️ VERIFY: binary name
    CodeGraphPath   string  // ⚠️ VERIFY: package name
    EngramPath      string  // ⚠️ VERIFY: from $ENGRAM_BIN env var
}

func DetectIDEMCPs() IDEMCPAvailability {
    npx, _ := exec.LookPath("npx")           // ⚠️ May fail in IDE environment
    cm,  _ := exec.LookPath("context-mode")  // ⚠️ Verify binary name
    eg,  _ := exec.LookPath("engram")        // ⚠️ Check $ENGRAM_BIN first

    return IDEMCPAvailability{
        NPXPath:         npx,
        ContextModePath: cm,
        EngramPath:      firstNonEmpty(os.Getenv("ENGRAM_BIN"), eg),
    }
}

func firstNonEmpty(a, b string) string {
    if a != "" { return a }
    return b
}
```

---

## 9. context-mode Detection & Configuration (IDE)

```go
// ⚠️ VERIFY: Does context-mode --mcp work in the IDE's environment?
// ⚠️ VERIFY: Does the IDE support stdio-based MCP servers or only HTTP?

func DetectContextModeForIDE() (string, string, error) {
    // Check binary
    path, err := exec.LookPath("context-mode")  // ⚠️ Verify binary name
    if err != nil {
        // Try npm global
        out, err2 := exec.Command("npm", "root", "-g").Output()
        if err2 == nil {
            npmRoot := strings.TrimSpace(string(out))
            altPath := filepath.Join(npmRoot, ".bin", "context-mode")
            if _, statErr := os.Stat(altPath); statErr == nil {
                path = altPath
            }
        }
    }
    if path == "" {
        return "", "", fmt.Errorf("context-mode not found")
    }
    out, err := exec.Command(path, "--version").Output()  // ⚠️ Verify --version flag
    return path, strings.TrimSpace(string(out)), err
}
```

---

## 10. Improvement Roadmap

| Week | Task | Verify Before |
|------|------|--------------|
| 1 | AI-01: Fix hooks.json named-group format | Named-group format against `antigravity.google/docs/ide-hooks` |
| 1 | AI-08: Split antigravity → antigravity-ide dir | Confirm no CLI users rely on current dir |
| 1 | AI-02: Fix MCP array format | `serverUrl` vs `url` key; local vs remote distinction |
| 2 | AI-03: Add sequential-thinking to IDE MCP | npx availability in IDE environment |
| 2 | AI-05: Add codegraph to IDE MCP | Package name `@colbymchenry/codegraph`; `--mcp` flag |
| 2 | AI-06: PreInvocation ephemeral routing reminder | `ephemeralMessage` field in IDE's PreInvocation |
| 2 | AI-07: Stop hook sentinel | `decision` output schema; binary invocation |
| 3 | AI-04: context-mode as MCP server | stdio vs HTTP support in IDE |
| 3 | AI-10: Google ADC setup | gcloud availability; when ADC is required |
| 3 | AI-11/12: Protocol Shell + token budget | Model name, context_window_size in IDE |
| 4 | AI-13: Self-refinement quality gate | Quality threshold values |
| 4 | AI-14: Context7 targeted topic fetch | topic parameter behavior |
