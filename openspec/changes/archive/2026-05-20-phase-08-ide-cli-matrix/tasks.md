# Tasks: Phase 8 — IDE/CLI Full Adapter Matrix v2

> **Source of Truth:** `/home/rdmachadog/Documents/fix_achitect-ai/08-phase-ide-cli-full-matrix.md`

## Status: COMPLETED

## Implementation Order: TEST-FIRST

> Per user directive: write tests first, then implementation. All code MUST match SOT exactly.

---

## T0 — Go Tests (Write First)

- [x] Create `internal/install/adapter/` package directory
- [x] **Write `internal/install/adapter/injector_test.go` FIRST** (SOT §8.7)
  - [x] `TestContentHash_Deterministic` — same input → same hash; different input → different hash
  - [x] `TestInjectSection_SkipsWhenUpToDate` — second injection with same content returns `injected=false`
  - [x] `TestInjectSection_UpdatesWhenContentChanges` — changed content triggers re-injection; old content absent
  - [x] `TestInjectSection_NoMarkerDuplication` — inject 4 times (v1/v2/v3/final) → exactly 1 `architect-ai:L0:start` marker
  - [x] `TestAllPlatforms_HaveConfig` — all 5 platforms (opencode, claude, cursor, antigravity, gemini) in `Supported` map
  - [x] `TestOpenCode_HasNoL2DelegationRead` — L2 agents have no `delegation_read` (manual verification log)
- [x] Confirm tests compile but fail (no implementation yet)

---

## T1 — Go Implementation (After Tests)

- [x] **Write `internal/install/adapter/injector.go`** (SOT §8.7)
  - [x] Import: `crypto/sha256`, `fmt`, `os`, `path/filepath`, `regexp`, `strings`
  - [x] `Platform` struct with fields: `ID`, `EntryFile`, `SupportsRealSubagents`, `SupportsParallel`, `SupportsNativeMCP`, `CompressCommand`, `HasDegradedMode`
  - [x] `Supported` map — all 5 platforms, values from SOT §8.7 table:
    - `opencode`: EntryFile=`opencode.json`, real sub-agents ✅, parallel ✅, native MCP ✅, compress=`/compact`
    - `claude`: EntryFile=`CLAUDE.md`, real sub-agents ✅, parallel ✅, native MCP ✅, compress=`/compact`
    - `cursor`: EntryFile=`.github/copilot-instructions.md`, all false, HasDegradedMode=true
    - `antigravity`: EntryFile=`.antigravity/agent.md`, all false, HasDegradedMode=true
    - `gemini`: EntryFile=`GEMINI.md`, real sub-agents ✅, parallel ✅, native MCP ✅, compress=`/compress`
  - [x] `contentHash(content string) string` — SHA256, first 4 bytes as 8 hex chars
  - [x] Compile regexps: `sectionStartRe`, `sectionEndRe`
  - [x] `InjectSection(filePath, sectionName, content string) (bool, error)`:
    - [x] Start marker includes `hash:{contentHash(content)}`
    - [x] If existing file already contains exact startMarker → return `false, nil` (idempotent)
    - [x] If section exists (any hash variant) → replace using regexp `FindStringIndex`
    - [x] If section absent → append
    - [x] Write via temp file + `os.Rename` (atomic)
  - [x] `ValidateInstallation(platformID, projectDir string) []string`:
    - [x] Unknown platform → `["unknown platform: X"]`
    - [x] Check entry file exists
    - [x] Check 6 required `.atl/` files: `sdd-state.yaml`, `skill-manifest.yaml`, `_generated/foundation.md`, `agents/architect.md`, `agents/sdd-orchestrator.md`, `agents/general-orchestrator.md`
    - [x] WARN (not error) if: `!SupportsRealSubagents`, `!SupportsNativeMCP`, `CompressCommand == ""`
- [x] Run tests — all 6 must pass

---

## T2 — OpenCode Configuration (SOT §8.2)

- [x] **Write `internal/assets/opencode/opencode.json`** — v2 full schema:
  - [x] Top-level keys: `$schema`, `plugin`, `permission`, `agent`, `mcp`
  - [x] `plugin`: `[".atl/plugins/background-agents.ts", "opencode-gemini-auth@latest"]`
  - [x] `permission.bash` deny rules: `curl *|*bash`, `curl *|*sh`, `wget *|*bash`, `eval *`, `exec *`, `> /dev/`, `mkfs`; ask rules: git commit/push/rebase/reset, rm -rf, sudo, chmod 777
  - [x] `permission.read` deny: `.env`, `.env.*`, `credentials.json`, `secrets/**`, `.ssh/**`
  - [x] L0 `architect` agent: mode=primary, model=anthropic:claude-opus-4-5, tools={bash,read,edit,write,delegate,delegation_list,delegation_read}
  - [x] L1a `sdd-orchestrator`: mode=primary, model=anthropic:claude-opus-4-5, tools with delegation_read/list, task permission `sdd-*` only
  - [x] L1b `general-orchestrator`: mode=primary, model=anthropic:claude-sonnet-4-5, tools with delegation_read/list, task permission allow researcher/solver/ideator/generalist
  - [x] L2 agents (hidden, subagent, NO delegation_read): sdd-init (haiku), sdd-onboard (sonnet), sdd-explore (sonnet), sdd-propose (opus), sdd-spec (sonnet), sdd-design (opus), sdd-tasks (sonnet), sdd-apply (sonnet), sdd-verify (sonnet), sdd-archive (haiku), researcher (haiku), solver (sonnet), ideator (sonnet), generalist (haiku)
  - [x] MCP block: context7 (remote https://mcp.context7.com/mcp), context-mode (local npx), engram (local ${ENGRAM_BIN}), sequential-thinking (local npx)

---

## T3 — Claude Code Configuration (SOT §8.3)

- [x] **Write `internal/assets/claude/.claude/settings.json`** — exact allow/deny lists from SOT:
  - [x] Allow: Bash(git status), Bash(git log*), Bash(git diff*), Bash(rg*), Bash(find*), Bash(cat*), Bash(echo*), Bash(ls*), Bash(pwd), Read(**), Write(.atl/**), Write(openspec/**), Edit(**), Task(**), mcp__engram__mem_*, mcp__context7__*, mcp__sequential_thinking__*, mcp__context_mode__*
  - [x] Deny: Bash(curl * | *bash), Bash(curl * | *sh), Bash(eval *), Bash(sudo *), Bash(rm -rf *), Read(**/.env), Read(**/.env.*), Read(**/credentials.json), Read(**/.ssh/**)
  - [x] MCP servers: engram (stdio, ${ENGRAM_BIN:-engram} mcp --tools=agent), context7 (npx -y @upstash/context7-mcp@latest), sequential_thinking (npx), context_mode (npx -y @mksglu/context-mode)
- [x] **Document CLAUDE.md section structure** (hash-based markers):
  - [x] Outer header: `<!-- AUTO-GENERATED by architect-ai sync v2 — hash:{CONTENT_HASH} -->`
  - [x] Sections: L0 (hash), L1a (hash), L1b (hash), foundation (hash), state-dag (no hash)
  - [x] `InjectSection` skips re-injection when hash matches (idempotent)

---

## T4 — VSCode Copilot Degraded Mode (SOT §8.4)

- [x] **Write `internal/assets/cursor/.github/copilot-instructions.md`**:
  - [x] Header: `<!-- architect-ai:generated v2 -->`, `<!-- PLATFORM: vscode-copilot -->`, `<!-- DEGRADED MODE: MCP not natively available -->`
  - [x] Degraded mode table: Engram (→ .atl/ YAML), sequential-thinking (→ inline template), context-mode (→ manual truncation), context7 (→ NOT AVAILABLE, use rg), Sub-agents (→ logical simulation)
  - [x] Degraded Engram alternative: cat .atl/sdd-state.yaml, .atl/apply-progress.yaml, .atl/session.yaml
  - [x] Sequential Thinking inline template: `[SEQUENTIAL THINKING — inline]` / Branch A / Branch B / Decision (trigger: D1+D2 >= 5)
  - [x] L0/L1a/L1b/foundation sections (no hash in VSCode version)

---

## T5 — Antigravity Single-Thread Adapter (SOT §8.5)

- [x] **Write `internal/assets/antigravity/.antigravity/agent.md`**:
  - [x] Header: `<!-- architect-ai:generated v2 -->`, `<!-- PLATFORM: antigravity -->`, `<!-- RUNTIME: single-thread -->`
  - [x] RUNTIME NOTICE: no real sub-agents, sequential inline, ULTRA framing + identity clear
  - [x] Inline sequential thinking template (D1+D2 >= 5, include Branch C if D5>=2)
  - [x] Simulated Delegation Protocol: 6-step (ULTRA emit from→to, load rules, execute, emit to→from, clear, resume)
  - [x] Context Management: D4>=2 → save checkpoint → mem_save → if no Engram emit LITE fallback
  - [x] Phase DAG Enforcement: check STATE=`.atl/sdd-state.yaml` before any SDD phase
  - [x] L0 section (single-thread version), foundation section

---

## T6 — Gemini CLI Configuration (SOT §8.6)

- [x] **Write `internal/assets/gemini/GEMINI.md`**:
  - [x] Header: `<!-- architect-ai:generated v2 -->`, `<!-- PLATFORM: gemini-cli -->`
  - [x] L0, L1a, L1b sections with run_subagent delegation
  - [x] Gemini CLI Specifics: entry, run_subagent, compress /compress, MCP .gemini/settings.json
  - [x] Mode A (inline simple), Mode B (SDD delegation), Mode C (General delegation)
  - [x] Sequential Thinking section: MCP available, fallback if server unavailable
  - [x] foundation section
- [x] **Write `internal/assets/gemini/.gemini/settings.json`** — exact from SOT §8.6:
  - [x] general: defaultApprovalMode auto_edit
  - [x] ide: enabled true
  - [x] mcpServers: context7 (httpUrl only, trust:false), context-mode (npx, timeout 15000), engram (${ENGRAM_BIN:-engram}), sequential-thinking (npx, timeout 30000, trust:true)
  - [x] model, security.auth.selectedType=oauth-personal, ui settings

---

## T7 — Verification

- [x] Run `go test ./internal/install/adapter/...` — all 7 tests pass
- [x] **Test 1**: rg "delegation_read" opencode.json — only L0/L1 entries
- [x] **Test 2**: CLAUDE.md idempotency — 2 syncs → 1 start marker
- [x] **Test 3**: VSCode degraded mode — inline template present, no MCP block
- [x] **Test 4**: Antigravity identity isolation — ULTRA framing protocol present
- [x] **Test 5**: Gemini context7 — httpUrl only, no command/args
- [x] **Test 6**: ValidateInstallation empty dir → len(issues) >= 6

---

## MANDATORY IMPLEMENTATION DIRECTIVE
**For `sdd-apply` (Coder):**
All code definitions, declarations, directives, and logic provided in the Source of Truth MUST be implemented EXACTLY as written on the first attempt. The coder MUST NOT analyze, re-design, or deviate from the source code during the initial implementation. Modification of the initial implementation is ONLY permitted if the tests fail.
