# Common & General Implementation Plan — architect-ai V4

**Scope:** Cross-agent issues from the Upgrade Blueprint + David Kim Context Engineering
  coverage + Context7 coverage + code quality/verification standards  
**Priority:** This plan PRECEDES all per-agent plans — these fixes propagate to all agents  
**Guiding principle:** Pareto — fix once in shared code, benefit all 14 agents

> ⚠️ **UNIVERSAL CODE VERIFICATION POLICY** (applies to this plan AND all agent plans):
> Every code snippet in every `.md` plan is **PSEUDOCODE / REFERENCE ONLY**.
> Before implementing any snippet:
> - Verify binary names exist on target OS (`which <binary>`)
> - Verify config file paths against vendor documentation
> - Verify JSON/TOML/YAML field names against actual schema docs
> - Verify MCP package names via `npm view <package>` or `npx --yes <package> --version`
> - Verify Go API signatures against the current `go.mod` version
> - Run with `--dry-run` first; check outputs before applying
> - Test on a clean machine (not developer machine with pre-installed tools)

---

## Table of Contents

1. [B1: Pipeline Parallelization](#b1-pipeline-parallelization)
2. [B2: Registry Serial I/O](#b2-registry-serial-io)
3. [B3: Double Tool Probe](#b3-double-tool-probe)
4. [B5: Context Saturation → Dynamic Assembler](#b5-context-saturation)
5. [B6: State Race Condition](#b6-state-race-condition)
6. [CodeGraph MCP — Universal Integration](#codegraph-mcp-universal)
7. [sequential-thinking — Universal Standard](#sequential-thinking-universal)
8. [context-mode — Universal Standard](#context-mode-universal)
9. [David Kim CE — Universal Coverage](#david-kim-ce-universal)
10. [Context7 — Universal Coverage](#context7-universal)
11. [Tool Policy YAML — Universal Gate](#tool-policy-yaml)
12. [L3: Hot-reload Skill Registry](#l3-hot-reload)
13. [L4: Engram TTL Configuration](#l4-engram-ttl)
14. [L5: Archive Cleanup Sidecar](#l5-archive-cleanup)
15. [eintegrate — New Checks](#eintegrate-new-checks)

---

## B1: Pipeline Parallelization

**Source:** Blueprint §3.6, Audit B1 — 65% latency waste from serial pipeline

**File:** `internal/pipeline/runner.go`

```go
// ⚠️ VERIFY: errgroup import path — "golang.org/x/sync/errgroup"
// ⚠️ VERIFY: current Step interface definition before extending
// ⚠️ VERIFY: context cancellation behavior — all steps should cancel on first error

import "golang.org/x/sync/errgroup"

type GroupMode int
const (
    Sequential GroupMode = iota
    Parallel
)

type StepGroup struct {
    Steps []Step
    Mode  GroupMode
}

// RunGroup executes steps sequentially or in parallel depending on Mode.
// MUST be called AFTER B6 mutex fix — parallel runners write shared state.
func (r *Runner) RunGroup(ctx context.Context, g StepGroup) error {
    if g.Mode == Sequential {
        return r.runSequential(ctx, g.Steps)
    }
    // Parallel execution
    // ⚠️ VERIFY: errgroup.WithContext cancels remaining goroutines on first error
    // ⚠️ VERIFY: step.Run() is safe to call concurrently (no shared mutable state)
    eg, egCtx := errgroup.WithContext(ctx)
    for _, step := range g.Steps {
        step := step // capture loop variable
        eg.Go(func() error {
            return r.runStep(egCtx, step)
        })
    }
    return eg.Wait()
}
```

**Test requirement (strict TDD — no merge without passing):**

```bash
# ⚠️ VERIFY: -race flag supported on CI platform
go test -race ./internal/pipeline/... -count=3 -timeout=60s
```

**Dependency:** B6 (mutex) MUST be merged first. B1 without B6 = data race.

---

## B2: Registry Serial I/O

**Source:** Blueprint §2.1, Audit B2 — 3 serial filesystem walks

**File:** `internal/app/skills_cmd.go` (or wherever `WriteLocalSkillRegistry` lives)

```go
// ⚠️ VERIFY: exact function name — may be WriteLocalSkillRegistry or similar
// ⚠️ VERIFY: three walk sources: user, project, overlay — confirm all three
// ⚠️ VERIFY: merge order matters for precedence (project > user > overlay?)

func WriteLocalSkillRegistryParallel(dirs []string, dest string) error {
    type result struct {
        dir     string
        content string
        err     error
    }
    results := make([]result, len(dirs))
    var wg sync.WaitGroup
    for i, dir := range dirs {
        i, dir := i, dir
        wg.Add(1)
        go func() {
            defer wg.Done()
            content, err := walkSkillDir(dir)  // ⚠️ VERIFY: function signature
            results[i] = result{dir, content, err}
        }()
    }
    wg.Wait()
    // Merge in defined precedence order
    return mergeAndWrite(results, dest)  // ⚠️ VERIFY: precedence order
}
```

---

## B3: Double Tool Probe

**Source:** Blueprint §2.1, Audit B3 — 4–7 redundant RPCs per cold start

**Files:** General Orchestrator + SDD Orchestrator templates (all 14 agents)

**Fix:** Session-scoped tool-availability cache in Engram.

```markdown
<!-- Add to ALL sdd-orchestrator.md files — Section: Tool Availability -->

## Tool Availability (Session Cache)

**On FIRST invocation of this session:**
1. Run tool availability probe (existing logic)
2. Cache result: mem_save(
     title: "tool-availability-{session_id}",
     topic_key: "session/{session_id}/tools",
     content: {json of available tools},
     ttl_minutes: 5    ← ⚠️ VERIFY: Engram supports TTL on individual records?
   )

**On SUBSEQUENT invocations:**
1. mem_search(query: "session/{session_id}/tools")
2. IF cache hit AND age < 5 minutes: use cached result, SKIP probe
3. IF cache miss OR expired: run probe, update cache

**SDD_INTENT fast path:**
IF first message matches SDD keywords (use sdd / sdd- / /sdd):
  SKIP General Orchestrator setup entirely
  Load SDD Orchestrator directly
  Run tool probe once → cache result
```

> ⚠️ **VERIFY**: Engram's `mem_save` TTL support — if not supported natively,
> store timestamp in content and check age manually in orchestrator.

---

## B5: Context Saturation

**Source:** Blueprint §3.3 + §3.4, Audit B5 — ~10,800 wasted tokens/session

### Dynamic Context Assembler

**File:** `internal/protoshell/assembler.go` (new package)

```go
// ⚠️ VERIFY: tokenize function — which tokenizer matches the target model?
// ⚠️ VERIFY: keyword extraction from skill-registry — SKILL.md frontmatter format
// ⚠️ VERIFY: "trigger" field name in skill-registry index

package protoshell

type AssemblyContext struct {
    PhaseName    string
    TaskKeywords []string
    Technologies []string
    BudgetConfig BudgetLayer
}

// AssembleContext returns only the skills relevant to the current phase+task.
// ⚠️ VERIFY: skill keyword extraction logic matches actual SKILL.md format
func AssembleContext(registry SkillRegistry, ctx AssemblyContext) (string, int, error) {
    var injected []string
    var totalTokens int

    for _, skill := range registry.Skills {
        if skillMatches(skill, ctx) {
            content := skill.Content
            tokens := estimateTokens(content) // ⚠️ VERIFY: estimation method
            if totalTokens+tokens > ctx.BudgetConfig.ProjectStandards {
                break // Budget cap reached
            }
            injected = append(injected, content)
            totalTokens += tokens
        }
    }
    return strings.Join(injected, "\n\n"), totalTokens, nil
}

func skillMatches(skill Skill, ctx AssemblyContext) bool {
    for _, kw := range skill.TriggerKeywords {
        if containsAny(ctx.PhaseName, ctx.TaskKeywords, ctx.Technologies, kw) {
            return true
        }
    }
    return false
}
```

### Skill Registry Keyword Index

**Addition to `.atl/skill-registry.md` generator:**

```markdown
<!-- SKILL-REGISTRY-INDEX-V4 -->
<!-- ⚠️ VERIFY: JSON format is valid; keywords match actual SKILL.md content -->
```json
{
  "ripgrep":              ["search", "grep", "rg", "sdd-explore", "find"],
  "bash-expert":          ["shell", "bash", "script", "sdd-apply", "run"],
  "context-guardian":     ["context", "token", "window", "saturation", "budget"],
  "codegraph":            ["explore", "trace", "callers", "impact", "ast", "graph"],
  "adaptive-reasoning":   ["mode", "circuit-breaker", "posture", "D1", "D2"],
  "mcp-notebooklm":      ["research", "notebooklm", "academic", "paper"],
  "ctx_fetch_and_index":  ["web", "fetch", "url", "http", "context-mode"],
  "sequential-thinking":  ["plan", "analyze", "think", "decompose", "D1"],
  "engram":               ["memory", "persist", "recall", "session", "history"]
}
```
<!-- SKILL-REGISTRY-INDEX-END -->
```

> ⚠️ **VERIFY**: Skill ID strings match actual SKILL.md file names exactly.

---

## B6: State Race Condition

**Source:** Blueprint §3.6, Audit B6 — no mutex on file-backed state

**File:** `internal/state/state.go`

```go
// ⚠️ VERIFY: current Store struct field names
// ⚠️ VERIFY: all callers of Read/Write — may have additional methods needing mutex

type Store struct {
    mu   sync.RWMutex  // ← ADD THIS FIELD
    path string
    // ... existing fields
}

// Read acquires a read lock before accessing state file.
// ⚠️ VERIFY: existing function signature matches
func (s *Store) Read() (*State, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    // existing ReadFile logic unchanged
    return readFromDisk(s.path)  // ⚠️ VERIFY: actual implementation
}

// Write acquires an exclusive lock before writing state file.
func (s *Store) Write(state *State) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    // existing WriteFile logic unchanged
    return writeToDisk(s.path, state)  // ⚠️ VERIFY: actual implementation
}
```

**Test:**

```bash
go test -race ./internal/state/... -count=5 -timeout=30s
```

**MUST merge before B1.** This is the foundation.

---

## CodeGraph MCP — Universal Integration {#codegraph-mcp-universal}

**Source:** Blueprint §3.7, eintegrate E-07

### Component File

**File:** `internal/components/mcp/codegraph.go` (new)

```go
// ⚠️ VERIFY: @colbymchenry/codegraph is the correct npm package name
//   Run: npm view @colbymchenry/codegraph to confirm existence
// ⚠️ VERIFY: "serve --mcp" is the correct codegraph subcommand
//   Run: npx @colbymchenry/codegraph --help to confirm
// ⚠️ VERIFY: codegraph init -i --quiet is the correct init command
//   Run: codegraph init --help to confirm flags

package mcp

const CodeGraphNPXPackage = "@colbymchenry/codegraph"  // ⚠️ VERIFY

// CodeGraphMCPCommand returns the command to start codegraph MCP server.
func CodeGraphMCPCommand() (string, []string) {
    if path, err := exec.LookPath("codegraph"); err == nil {
        return path, []string{"serve", "--mcp"}  // ⚠️ VERIFY subcommand
    }
    return "npx", []string{"-y", CodeGraphNPXPackage, "serve", "--mcp"}
}

// CodeGraphInitCommand returns the project initialization command.
func CodeGraphInitCommand(projectDir string) *exec.Cmd {
    cmd := exec.Command("codegraph", "init", "-i", "--quiet")
    cmd.Dir = projectDir
    return cmd
    // ⚠️ VERIFY: -i and --quiet flags exist; may be --interactive and --silent
}

// IsAvailable checks whether codegraph can be run.
func CodeGraphIsAvailable() bool {
    _, errBin := exec.LookPath("codegraph")
    _, errNPX := exec.LookPath("npx")
    return errBin == nil || errNPX == nil
}
```

### Universal MCP Entry (JSON format — most agents)

```json
"codegraph": {
  "command": "npx",
  "args": ["-y", "@colbymchenry/codegraph", "serve", "--mcp"]
}
```

> ⚠️ **VERIFY ALL**: package name, `serve --mcp` subcommand, HTTP transport
> availability for VS Code, stdio stability for all other agents.

### Post-install Hook

```go
// ⚠️ VERIFY: codegraph init creates a .codegraph/ directory — confirm location
// ⚠️ VERIFY: -i flag is not interactive mode (would block CI)
func runCodeGraphInit(projectDir string, dryRun bool) error {
    if dryRun {
        fmt.Printf("  DRY RUN: would run codegraph init -i --quiet in %s\n", projectDir)
        return nil
    }
    cmd := CodeGraphInitCommand(projectDir)
    cmd.Stdout = os.Stdout  // ⚠️ suppress if --quiet works
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        // Non-fatal: codegraph may not support all languages in project
        fmt.Fprintf(os.Stderr, "WARNING: codegraph init: %v (non-fatal)\n", err)
    }
    return nil
}
```

### eintegrate Check E-07 (update)

```go
// ⚠️ VERIFY: "codegraph_callers" is the correct tool name in skill-registry
if !checkFile(".atl/skill-registry.md", "codegraph_callers") {
    errs = append(errs, "E-07: codegraph_callers missing from skill-registry.md")
}
if !checkFile(".atl/skill-registry.md", "codegraph_context") {
    errs = append(errs, "E-07b: codegraph_context missing from skill-registry.md")
}
```

---

## sequential-thinking — Universal Standard {#sequential-thinking-universal}

**Source:** README (mandatory requirement), per-agent plans

### Detection (shared function)

```go
// internal/components/mcp/sequential_thinking.go
// ⚠️ VERIFY: npm package is @modelcontextprotocol/server-sequential-thinking
//   Run: npm view @modelcontextprotocol/server-sequential-thinking

package mcp

const SeqThinkNPXPackage = "@modelcontextprotocol/server-sequential-thinking"

type SeqThinkAvailability struct {
    NPXFound   bool
    NPXPath    string
    PreCached  bool    // true if already in npm global cache
}

func DetectSequentialThinking() SeqThinkAvailability {
    npxPath, err := exec.LookPath("npx")  // ⚠️ VERIFY: npx vs npx.cmd on Windows
    if err != nil {
        return SeqThinkAvailability{}
    }
    // Check if pre-cached (avoids network on first run)
    // ⚠️ VERIFY: npm list -g output format varies by npm version
    out, _ := exec.Command("npm", "list", "-g", "--depth=0",
        SeqThinkNPXPackage).Output()
    cached := strings.Contains(string(out), "server-sequential-thinking")

    return SeqThinkAvailability{
        NPXFound:  true,
        NPXPath:   npxPath,
        PreCached: cached,
    }
}
```

### Config Formats by Agent

```
JSON mcpServers (object):  "sequential-thinking": { "command": "npx", "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"] }
JSON mcpServers (array):   { "name": "sequential-thinking", "command": "npx", "args": ["-y", "@modelcontextprotocol/server-sequential-thinking"] }
TOML:                      [mcp_servers.sequential-thinking]\ncommand="npx"\nargs=["-y","@modelcontextprotocol/server-sequential-thinking"]
Gemini CLI:                gemini mcp add --scope user sequential-thinking npx -y @modelcontextprotocol/server-sequential-thinking
VS Code (HTTP proxy):      { "type": "http", "url": "http://localhost:3001/mcp" }
```

> ⚠️ **VERIFY PER AGENT**: Which JSON structure does each agent use?
> Some use object mcpServers (Claude, OpenCode, Cursor, Windsurf, Kiro, Kilocode, Codex)
> vs array mcpServers (Antigravity IDE, GGA).
> Gemini uses CLI command, not JSON.

### Prompt-Level Fallback (ALL agents, ALL phase protocols)

Add to `internal/assets/_shared/adaptive-reasoning-gate-v2.md` and propagate:

```markdown
### sequential_thinking Tool Fallback

IF `sequential_thinking` tool is unavailable (tool call returns error or tool not listed):

Write an explicit reasoning block:
---
THINKING:
1. Problem restatement: {one sentence}
2. D1 classification: {Simple/Moderate/Complex/Expert}
3. D2 ambiguity: {Low/Medium/High}
4. D3 risk: {Low/Medium/High}
5. D4 novelty: {Low/Medium/High}
6. Three candidate approaches:
   A) {approach} — trade-off: {trade-off}
   B) {approach} — trade-off: {trade-off}
   C) {approach} — trade-off: {trade-off}
7. Selected approach: {A|B|C} — rationale: {why}
---
This satisfies the sequential-thinking protocol without the MCP tool.
```

> ⚠️ **VERIFY**: Confirm this fallback block propagates through the `_shared/` include mechanism.

---

## context-mode — Universal Standard {#context-mode-universal}

**Source:** `internal/assets/generic/context-mode-routing-policy.md`, all agent configs

### Detection (shared function)

```go
// internal/components/engram/context_mode.go (extend existing)
// ⚠️ VERIFY: binary is named "context-mode" (hyphenated) not "contextmode" or "ctx"
// ⚠️ VERIFY: npm package is "context-mode" not "@mksglu/context-mode" for global install
// ⚠️ VERIFY: --version flag is supported
// ⚠️ VERIFY: --mcp flag starts an MCP server (not all versions may support this)
// ⚠️ VERIFY: "ctx" is the shorthand binary or a separate command

type ContextModeCapability struct {
    Installed       bool
    BinaryPath      string
    Version         string
    MCPModeSupport  bool    // ⚠️ VERIFY: context-mode --mcp exists
    HookModeSupport bool    // ⚠️ VERIFY: context-mode hook <agent> pretooluse
    DoctorClean     bool
}

func DetectContextMode() ContextModeCapability {
    path, err := exec.LookPath("context-mode")  // ⚠️ VERIFY: binary name
    if err != nil {
        return ContextModeCapability{Installed: false}
    }
    out, err := exec.Command(path, "--version").Output()
    if err != nil {
        return ContextModeCapability{Installed: false}
    }
    version := strings.TrimSpace(string(out))

    // Check --mcp support
    // ⚠️ VERIFY: running context-mode --help shows --mcp flag
    helpOut, _ := exec.Command(path, "--help").Output()
    mcpSupport := strings.Contains(string(helpOut), "--mcp")

    // Check hook support  
    // ⚠️ VERIFY: "hook" subcommand exists
    hookOut, _ := exec.Command(path, "hook", "--help").Output()
    hookSupport := len(hookOut) > 0

    // Check doctor
    doctorOut, doctorErr := exec.Command(path, "doctor").Output()
    doctorClean := doctorErr == nil &&
        !strings.Contains(string(doctorOut), "ERROR")

    return ContextModeCapability{
        Installed:       true,
        BinaryPath:      path,
        Version:         version,
        MCPModeSupport:  mcpSupport,
        HookModeSupport: hookSupport,
        DoctorClean:     doctorClean,
    }
}
```

### Install

```go
func EnsureContextModeV4(dryRun bool) (ContextModeCapability, error) {
    cap := DetectContextMode()
    if cap.Installed {
        return cap, nil
    }
    // ⚠️ VERIFY: npm install -g context-mode is correct package name
    // ⚠️ VERIFY: requires npm; check npm availability first
    if dryRun {
        fmt.Println("  DRY RUN: would run: npm install -g context-mode")
        return ContextModeCapability{}, nil
    }
    cmd := exec.Command("npm", "install", "-g", "context-mode")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        return ContextModeCapability{},
            fmt.Errorf("npm install -g context-mode failed: %w\n"+
                "⚠️ Verify package name at: npm view context-mode", err)
    }
    return DetectContextMode()
}
```

### Routing Policy Update (`ctx_batch_execute` addition)

**File:** `internal/assets/generic/context-mode-routing-policy.md`

Add section:

```markdown
## ctx_batch_execute Pattern (V4)

When running ≥3 related shell commands in one phase step:

  USE: ctx_batch_execute(["cmd1", "cmd2", "cmd3"])
       → Single MCP call; all outputs compressed; returned as one block

  NOT: Three separate rg calls that each inject raw output into context

Threshold: ≥3 commands in same exploration step → batch_execute
Exception: Commands with dependencies (output of cmd1 feeds cmd2) → stay sequential

## ctx_fetch_and_index Pattern

For large web pages or external docs needed more than once:
  ctx_fetch_and_index(url)  → fetch, chunk, index into context-mode store
  ctx_search(query)         → retrieve relevant chunks by keyword
  NOT: Raw file/web content pasted into prompt
```

> ⚠️ **VERIFY**: `ctx_batch_execute`, `ctx_fetch_and_index`, `ctx_search`, `ctx_index`
> are all valid context-mode MCP tool names. Run `context-mode --list-tools` or check docs.

---

## David Kim CE — Universal Coverage {#david-kim-ce-universal}

David Kim's Context Engineering framework defines these core topics. Coverage
status across architect-ai is mapped here, with actions for each gap.

### Topic 1: Protocol Shells

**Definition:** `/verb.noun{intent, input, process, output}` — composable instruction units

**Current coverage:** Referenced in blueprint but NO phase protocol uses this format yet.

**Universal action:** Add Protocol Shell header to ALL sdd-phase-protocols across ALL agents.

```markdown
<!-- Template to prepend to every sdd-*.md file -->
/sdd.{phase_name}{
  intent="{one-line purpose of this phase}",
  input={
    change_topic: string,
    project: string,
    prior_phase_key: engram_key   // optional
  },
  process=[
    /step1{action="...", condition="...", tool="..."},
    /step2{action="...", condition="...", tool="..."}
  ],
  output={
    artifact_type: enum[explore|design|spec|...],
    engram_key: "sdd/{change_name}/{phase}",
    fields: [executive_summary, status, empirical_proof]
  }
}
```

**Files to update:** All `sdd-explore.md`, `sdd-design.md`, `sdd-verify.md`, etc.
across all 14 agent asset directories.

### Topic 2: Token Budget Tracking

**Definition:** Explicit per-layer allocation tracked before sub-agent dispatch

**Current coverage:** Caveman compression only. No explicit layer budgets.

**Universal action:**

```yaml
# .atl/config.yaml — ADD for all agents
# ⚠️ VERIFY: token counts for your specific model/agent
token_budget:
  layers:
    global_directives:  500     # ⚠️ VERIFY by counting actual chars/tokens
    cognitive_posture:  300
    adaptive_gate:      200
    project_standards:  1500    # keyword-indexed subset only
    overlay_supplement: 800     # only when overlay active
    research_routing:   300
    codegraph_context:  3000    # new V4
    phase_protocol:     900
    task:               1000
    engram_carryforward:500
  reserve:              4000    # ≥25% headroom for claude; ~99% for gemini/1M
```

### Topic 3: Self-Refinement Engine

**Definition:** Quality scoring loop on sub-agent artifacts before Engram persistence

**Current coverage:** ABSENT in all agents.

**Universal action:** Add to `sdd-verify.md` across all agents:

```markdown
## Artifact Quality Gate (Self-Refinement)

Before calling mem_save, score this phase's output:

| Dimension | Question | Score (0–1) |
|-----------|----------|-------------|
| Relevance | Does every section address the change_topic? | |
| Completeness | Are all impacted modules/files listed? | |
| Coherence | Is the design internally consistent? | |
| Efficiency | No redundant steps or over-engineering? | |

overall_score = mean(above scores)
IF overall_score < 0.85:
  → Identify lowest-scoring dimension
  → Iterate with targeted fix (max 2 retries)
  → If still < 0.85 after 2 retries: persist with quality warning

THEN: mem_save with quality_score in metadata
```

> ⚠️ **VERIFY**: threshold 0.85 — may need calibration per phase type.

### Topic 4: Dynamic Assembly

**Definition:** Context assembled dynamically based on query analysis and component weights

**Current coverage:** Partially in blueprint; not in any agent prompt.

**Universal action:** Add to all `general-orchestrator.md` and `sdd-orchestrator.md`:

```markdown
## Dynamic Context Assembly (V4)

Context injection order (in budget priority):
1. Global directives (always)
2. Cognitive posture (always, based on D1-D4)
3. Adaptive gate (always)
4. Project standards (keyword-filtered from skill-registry — NOT full registry)
5. Overlay supplement (ONLY if project overlay detected)
6. Research routing policy (always, compressed)
7. CodeGraph context pack (ONLY if codegraph_available AND explore/verify phase)
8. Phase protocol shell (always)
9. Task (always)
10. Engram carry-forward: executive summaries ONLY (full obs on demand)

Inject layer N only if sum(layers 1..N) ≤ budget.total - budget.reserve
```

### Topic 5: Few-Shot Learning in Protocols

**Definition:** Positive and negative examples within protocol definitions

**Current coverage:** Absent from all phase protocols (only instructions, no examples).

**Universal action:** Add example output blocks to `sdd-explore.md` and `sdd-design.md`:

```markdown
## Example: Correct Explore Output (few-shot positive)

executive_summary: "AuthHandler calls UserRepo.GetByID which reads from PG.
  Changing the password hash algo breaks 3 callers: login, reset, oauth-link."
impact_surface: ["auth/handler.go", "user/repo.go", "user/repo_test.go"]
empirical_proof: "rg 'GetByID' -l: 3 files confirmed"
status: complete

## Example: Incorrect Explore Output (few-shot negative — avoid)

executive_summary: "Reviewed the codebase and found some relevant files."
impact_surface: []
empirical_proof: "(none)"
status: complete  ← WRONG: no evidence = not complete
```

### Topic 6: Progressive Disclosure / Paging

**Definition:** Large sources paginated across invocations, not injected whole

**Current coverage:** Partial in `apply-continuity.md`. Absent from explore/design.

**Universal action:** Add to all `sdd-explore.md`:

```markdown
## Progressive Paging for Large Sources

IF a source (file, web page, doc) exceeds 2000 tokens of relevant content:
  1. Summarize key findings → store in Engram (`executive_summary` field)
  2. Return Engram key to orchestrator (NOT raw content)
  3. Orchestrator requests full content only if sub-agent explicitly needs it
  4. Use `ctx_fetch_and_index(url)` for web sources → `ctx_search` for retrieval
  5. Use codegraph_context (maxNodes: 25) instead of reading raw files
```

---

## Context7 — Universal Coverage {#context7-universal}

### Topic 1: Library Resolution Before Docs Fetch

All agents' `sdd-explore.md` must use Context7's two-step pattern:

```markdown
## Context7 Usage Pattern (Mandatory)

Step 1: Resolve library ID
  mcp__context7__resolve-library-id(libraryName: "{detected_library}")
  → Returns context7CompatibleLibraryID

Step 2: Fetch targeted docs
  mcp__context7__get-library-docs(
    context7CompatibleLibraryID: "{from step 1}",
    topic: "{specific aspect needed}",  ← ALWAYS specify; never omit
    tokens: 5000                         ← cap at 5000 for claude agents
  )

NEVER: Fetch full library docs without a topic filter
NEVER: Guess at the libraryID without calling resolve-library-id first
```

> ⚠️ **VERIFY**: Tool names `mcp__context7__resolve-library-id` and
> `mcp__context7__get-library-docs` against current Context7 MCP tool schema.

### Topic 2: Context7 in sdd-verify

When verifying implementation against an external API:

```markdown
## Context7 Verification Pattern

In sdd-verify, when implementation calls an external library:
1. resolve-library-id → get-library-docs(topic: "API changes in {version}")
2. Compare implementation against doc's current API signature
3. Flag any deprecated methods or signature mismatches as blocking issues
```

### Topic 3: Context7 Token Strategy

```markdown
## Context7 Token Budget

Context7 targeted fetch: cap at 5000 tokens for Claude (16K window)
Context7 targeted fetch: cap at 15000 tokens for Gemini (1M window)
context7 is NOT a substitute for Engram — use Engram for project-specific memory
context7 is NOT a substitute for codegraph — use codegraph for code relationships
context7 IS the right tool for: library API docs, framework guides, external specs
```

---

## Tool Policy YAML — Universal Gate {#tool-policy-yaml}

**Source:** Blueprint Addendum — Antigravity hook pattern analysis

**File:** `.atl/tool_policy.yaml` (generated at `architect-ai install` time)

```yaml
# ⚠️ VERIFY: All matcher values match actual tool names in each agent's tool list
# ⚠️ VERIFY: condition syntax — is "phase == sdd-apply" valid in your hook evaluator?
# ⚠️ VERIFY: "posture == production" condition availability

pre_tool_use:
  # Shell commands require human approval in apply phase
  - matcher: "run_command|Bash|Shell|execute"   # ⚠️ VERIFY: per-agent tool name
    decision: "ask"
    condition: "phase == sdd-apply AND mode != tmux"

  # Code graph queries: always allow (read-only)
  - matcher: "mcp__codegraph__.*"               # ⚠️ VERIFY: MCP tool naming pattern
    decision: "allow"

  # Engram writes: always allow
  - matcher: "mcp__engram__mem_save|mcp__engram__mem_update"
    decision: "allow"

  # Engram reads: always allow
  - matcher: "mcp__engram__mem_search|mcp__engram__mem_get_observation"
    decision: "allow"

  # External web calls: ask in production posture
  - matcher: "WebFetch|WebSearch|mcp__context7__.*"
    decision: "ask"
    condition: "posture == production"           # ⚠️ VERIFY: posture variable exists

  # Context-mode tools: always allow (sandboxed)
  - matcher: "mcp__context.mode__ctx_.*"        # ⚠️ VERIFY: context-mode MCP prefix
    decision: "allow"
```

---

## L3: Hot-reload Skill Registry {#l3-hot-reload}

**File:** `internal/app/skills_cmd.go` (or skill-registry watcher component)

```go
// ⚠️ VERIFY: fsnotify import path — "github.com/fsnotify/fsnotify"
// ⚠️ VERIFY: .atl/skill-registry.md is the correct watched path
// ⚠️ VERIFY: atomic reload doesn't break in-progress skill lookups

import "github.com/fsnotify/fsnotify"

func WatchSkillRegistry(path string, onReload func()) error {
    watcher, err := fsnotify.NewWatcher()  // ⚠️ VERIFY: API signature
    if err != nil {
        return err
    }
    go func() {
        for event := range watcher.Events {
            if event.Op&fsnotify.Write != 0 {
                // Atomic reload: build new registry, swap pointer
                onReload()
            }
        }
    }()
    return watcher.Add(path)  // ⚠️ VERIFY: Add vs Watch
}
```

---

## L4: Engram TTL Configuration {#l4-engram-ttl}

**File:** `.atl/config.yaml` + `internal/assets/_shared/engram-convention.md`

```yaml
# .atl/config.yaml — ADD
engram:
  cache_ttl_hours: 168        # 7 days — ⚠️ VERIFY: Engram honors this via mem_save?
  session_summary_ttl_hours: 48
  tool_availability_ttl_minutes: 5  # for B3 probe cache
```

> ⚠️ **VERIFY**: Engram's `mem_save` API accepts a TTL parameter. If not native,
> implement TTL check in orchestrator using stored timestamp.

---

## L5: Archive Cleanup Sidecar {#l5-archive-cleanup}

**File:** `internal/assets/antigravity-cli/sidecars/archive-cleaner.json` (new)
**File:** systemd unit or launchd plist (for non-Antigravity agents)

```json
{
  "description": "Daily cleanup of openspec/changes/archive/ entries beyond TTL",
  "builtin": "schedule",
  "args": [
    "0 2 * * *",
    "architect-ai",
    "archive",
    "--prune",
    "--max-age", "30d"
  ]
}
```

> ⚠️ **VERIFY**: `architect-ai archive --prune --max-age 30d` is a valid command.
> Check cmd/architect-ai/ for existing archive subcommand.

**For non-Antigravity agents** (via `architect-ai install --sidecar`):

```bash
# macOS launchd (⚠️ VERIFY: paths)
# ~/.local/share/architect-ai/com.architect-ai.archive-cleaner.plist

# Linux systemd (⚠️ VERIFY: paths and ExecStart)
# ~/.config/systemd/user/architect-ai-archive-cleaner.service
# [Timer] OnCalendar=daily
# [Service] ExecStart=/usr/local/bin/architect-ai archive --prune --max-age 30d
```

---

## eintegrate — New Checks {#eintegrate-new-checks}

**File:** `cmd/eintegrate/main.go`

```go
// ⚠️ VERIFY: checkFile helper reads .atl/skill-registry.md correctly
// ⚠️ VERIFY: checkAnyAgentMCPHasCodeGraph traverses correct config paths
// ⚠️ VERIFY: E-11 through E-15 don't conflict with existing E-07 through E-10

// E-11: Token budget config
if !checkFile(".atl/config.yaml", "token_budget") {
    errs = append(errs, "E-11: token_budget section missing from .atl/config.yaml")
}

// E-12: CodeGraph in at least one agent
if !checkAnyAgentMCPHasCodeGraph() {  // ⚠️ VERIFY: function impl
    errs = append(errs, "E-12: codegraph MCP not configured for any agent")
}

// E-13: sdd-explore has codegraph_context call
if !checkFileAny("sdd-explore.md", "codegraph_context") {
    errs = append(errs, "E-13: sdd-explore.md missing codegraph_context step")
}

// E-14: sequential-thinking fallback block present
if !checkFileAny("thinking-agent.md", "sequential_thinking tool unavailable") {
    errs = append(errs, "E-14: thinking-agent.md missing sequential-thinking fallback")
}

// E-15: context-mode routing policy has ctx_batch_execute pattern
if !checkFile("context-mode-routing-policy.md", "ctx_batch_execute") {
    errs = append(errs, "E-15: context-mode-routing-policy.md missing ctx_batch_execute pattern")
}

// E-16: Protocol Shell header in at least one phase protocol
if !checkFileAny("sdd-explore.md", "/sdd.explore{") {
    errs = append(errs, "E-16: sdd-explore.md missing Protocol Shell header (David Kim CE)")
}
```

---

## Implementation Sequence

```
Week 1: B6 (mutex) → B1 (parallel) → B2 (registry)
Week 2: B5 (assembler) → B3 (probe cache) → CodeGraph component
Week 3: sequential-thinking universal → context-mode universal → tool_policy.yaml
Week 4: David Kim CE: Protocol Shells → Token Budget → Self-Refinement
Week 5: Context7 universal pattern → ctx_batch_execute in all explore files
Week 6: L3 (hot-reload) → L4 (TTL config) → L5 (sidecar)
Week 7: eintegrate new checks E-11 through E-16
```

All implementations require:
1. `--dry-run` validation before real install
2. Race-condition tests (`go test -race`) for all Go changes
3. Real target-environment testing (not just developer machine)
4. External documentation verification for all binary/package names
