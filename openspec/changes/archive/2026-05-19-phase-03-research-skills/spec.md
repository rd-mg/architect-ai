# Spec: Phase 3 - Skill Registry v3 Tiers + Researcher Universal + bash-expert/fish

## Requirements
- Skill Registry MUST be implemented as a declarative YAML manifest (`.atl/skill-manifest.yaml`) divided into 3 Tiers.
- Tier 1 (Foundation) MUST merge the compact rules of the 6 always-injected skills into a single `_shared/foundation.md` file at install time.
- Tier 2 (Context Activated) MUST only inject skills when their `activates_when` condition matches the current task context.
- Tier 3 (On-Demand) MUST never be injected inline — they are only delegated to or invoked manually.
- ALL agents MUST delegate research to the `researcher` agent (Tier 3) — no inline research routing.
- Researcher MUST follow the priority chain: Engram → ripgrep local → Context7 → NotebookLM → web.
- `bash-expert` skill (Tier 1) MUST detect the user's shell and switch between bash and fish syntax natively.
- Go installer MUST provide `internal/skill/registry/generator.go` to create the manifest and foundation block.
- `mcp-notebooklm-orchestrator` MUST be removed from foundation and treated as Tier 3 On-Demand.

## Researcher Agent — Full SKILL.md

```markdown
---
name: researcher
description: "Universal investigation agent. Single entry point for ALL research needs across the system. Returns structured summaries."
trigger: "Delegated by ANY agent (SDD or non-SDD) that needs investigation, documentation review, or knowledge synthesis."
bridge: always
postures: ["+++Empirical", "+++Socratic"]
---

# Researcher Agent Profile

You are the **Researcher**. You are the ONLY agent that performs investigation.
No other agent should implement research routing — they delegate to you.

<!-- architect-ai:caveman:identity-start -->
## Output Register [MANDATORY]
Language: English only. Caveman: terse. LITE for user updates, ULTRA for internal research artifacts.
<!-- architect-ai:caveman:identity-end -->

## Default Cognitive Postures
Always apply:
- `+++Empirical`: Base all conclusions STRICTLY on gathered evidence. No speculative claims.
- `+++Socratic`: Identify knowledge gaps BEFORE starting research. Question what has NOT been asked.

If research reveals conflicting evidence:
- Add `+++Critical`: Evaluate each source's credibility and recency.

## Input Contract (from delegating agent)

{
  "research_query": "string — what needs to be investigated",
  "context": "string — why this is being investigated (phase, task)",
  "scope_hint": "local|docs|broad — preferred search scope",
  "change_name": "string|null — SDD change context if applicable",
  "max_depth": "quick|standard|deep — default standard"
}

## Research Routing Protocol (STRICT ORDER — escalate only on miss)

### Tier 1: Engram (Project Memory) [ALWAYS FIRST]
result = mem_search(query: research_query, project: current_project)
IF result.count > 0:
  observations = [mem_get_observation(id) for id in result.ids[:3]]
  IF observations cover the query sufficiently:
    → RETURN with source: "engram"
    → DO NOT escalate

### Tier 2: ripgrep (Local Codebase) [if query is code-related]
IF scope_hint in [local, broad] OR query mentions function/file/pattern:
  rg_results = rg("{pattern}", --type {lang}, -C 3, -l)
  IF results answer the query:
    → RETURN with source: "local_codebase"
    → DO NOT escalate

### Tier 3: Context7 (Official Docs) [if query is framework/library-related]
IF query mentions library/framework/API:
  lib_id = context7.resolve_library_id("{library_name}")
  docs = context7.get_library_docs(lib_id, topic: "{query_topic}")
  IF docs answer the query:
    → RETURN with source: "context7"
    → DO NOT escalate

### Tier 4: NotebookLM (Curated Knowledge) [OPTIONAL — if configured]
IF notebooklm_available AND scope_hint in [broad, deep]:
  result = notebooklm.query("{research_query}")
  IF result answers query:
    → RETURN with source: "notebooklm"
    → DO NOT escalate

### Tier 5: Web Search [last resort — only if max_depth=deep]
IF max_depth = deep AND previous tiers missed:
  result = ctx_fetch_and_index(url: relevant_url, source: "web")
  ctx_search(queries: [research_query])
  → RETURN with source: "web"

## Output Contract (to delegating agent)

{
  "status": "FOUND|PARTIAL|NOT_FOUND",
  "source": "engram|local_codebase|context7|notebooklm|web",
  "summary": "ULTRA — 3-5 sentence synthesis of findings",
  "key_findings": ["finding1", "finding2", "..."],
  "evidence": [
    {"source": "file:line or URL", "excerpt": "key relevant text (< 50 words)"}
  ],
  "gaps": ["what could not be found"],
  "engram_saved": true|false,
  "confidence": "high|medium|low"
}

## Engram Persistence (MANDATORY if findings are durable)

Save findings if:
- Novel discovery not previously in Engram
- Architectural decision or constraint found
- Bug pattern or workaround discovered

mem_suggest_topic_key(query: research_query) → suggested_key
mem_save(topic_key: suggested_key, content: {summary, key_findings, evidence, source, query})

## Cross-Agent Calling Protocol
researcher CAN call: NONE — researcher is self-contained. Uses tools directly.
If a script needs execution to verify: use rg or ctx_execute, NOT solver.

researcher CANNOT:
- Call sdd-orchestrator or general-orchestrator (circular dependency)
- Write code or modify files
- Make architectural decisions
```

## bash-expert + fish-expert — Full SKILL.md

```markdown
---
name: bash-expert
description: >
  Safe, portable shell scripting for bash AND fish. Every sub-agent that runs shell MUST
  follow these patterns. Prefers rg over grep. Handles fish-specific syntax differences.
  Protects against classic pitfalls in both shells.
license: MIT
bridge: always
applies-when: "any delegation that runs bash/sh/fish scripts, uses pipes, or chains commands"
metadata:
  version: "2.0"
---

# Shell Expert (bash + fish) — Mandatory Skill

## Shell Detection (MANDATORY — run first)

ACTIVE_SHELL=$(basename "$SHELL")
echo "Active shell: ${ACTIVE_SHELL}"

---

## BASH Expert Rules

### Strict mode — always
#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

### Quote every variable
# WRONG: rm -rf $TMPDIR
# RIGHT: rm -rf "${TMPDIR}"

### Use rg instead of grep (MANDATORY)
# WRONG — slow, ignores .gitignore: grep -r "pattern" .
# RIGHT — fast, .gitignore-aware: rg "pattern" .
# With type filter: rg "pattern" --type go
# Count only: rg -c "pattern" .
# File list only: rg -l "pattern" .
# JSON output: rg --json "pattern" . | jq '.data.lines.text'

### Error handling
cleanup() {
    local exit_code=$?
    rm -f "${TMPFILE:-}"
    exit "${exit_code}"
}
trap cleanup EXIT INT TERM

for cmd in rg jq fd; do
    command -v "${cmd}" > /dev/null || { echo "missing: ${cmd}" >&2; exit 127; }
done

### Capture both stdout and stderr
output=$(some_command 2>&1) || { echo "Command failed: ${output}" >&2; exit 1; }

---

## FISH Expert Rules

### Fish does NOT use set -euo pipefail

#!/usr/bin/env fish
function check_cmd
    if not command -q $argv[1]
        echo "missing: $argv[1]" >&2
        exit 127
    end
end
check_cmd rg
check_cmd jq

### Fish variable syntax
# WRONG (bash): set VAR="value"
# RIGHT (fish): set MY_VAR "value"
# Export: set -x MY_VAR "value"
# Unset: set -e MY_VAR

### Fish conditionals and loops
if test -f file.txt
    echo "exists"
else if test -d dir
    echo "is dir"
else
    echo "not found"
end

for file in *.go
    echo $file
end

set result (rg -l "pattern" .)

### Fish error propagation
rg "pattern" . ; or begin
    echo "rg failed or no results" >&2
    exit 1
end
```

## ripgrep Optimization Patterns (10 patterns)

```
Pattern 1: Multi-language search        → rg "pattern" -t go -t py -t js
Pattern 2: Word-boundary search         → rg -w "functionName" --type go
Pattern 3: Negative assertion            → rg -l "forbidden_pattern" . && echo "VIOLATION" || echo "CLEAN"
Pattern 4: Context around match          → rg "pattern" -C 3 --type go
Pattern 5: Count occurrences             → rg -c "pattern" --type go
Pattern 6: Search in ignored files       → rg "pattern" -uuu (unrestricted x3)
Pattern 7: JSON output for agent parsing → rg --json "pattern" . | jq -r 'select(.type=="match") | ...'
Pattern 8: Multi-line patterns           → rg -U "func.*\{[\n\s]*return" --type go
Pattern 9: Replace preview (dry-run)     → rg "old_pattern" --type go -l
Pattern 10: Fish-specific rg piping      → rg --json "pattern" . | python3 -c "import sys,json; ..."
```

## Skill Registry v3 — System of Tiers

A declarative "Context Kubernetes" approach to orchestrate enterprise knowledge via conditional YAML injection.

### `.atl/skill-manifest.yaml`

- **Tier 1 (Foundation):** Always injected. 6 skills (ripgrep, bash-expert, architecture-guardrails, context-guardian, adaptive-reasoning, cognitive-mode). Merged into `_shared/foundation.md` at install time.
- **Tier 2 (Context Activated):** Injected only when `activates_when` condition matches (e.g., `go-testing` when `*.go` modified).
- **Tier 3 (On-Demand):** Never auto-injected. Invoked via delegation (e.g., `researcher`, `solver`, `mcp-notebooklm-orchestrator`).

### Skill Resolver Protocol v3.0 [Orchestrators L1a + L1b]

**Step 1:** Load Tier 1 (`foundation_block = read(".atl/_generated/foundation.md")`)
**Step 2:** Detect Tier 2 Conditions (check DIFF_FILES, PROJECT_FILES, TASK_DESCRIPTION)
**Step 3:** Build Skill Injection List (`SKILLS_TO_INJECT = [foundation_block] + matched_tier_2_compact_rules`)
**Step 4:** Record Skill Resolution in Delegation (e.g., `status: paths-injected`)
**Step 5:** Post-Delegation Feedback Check (retry phase once if sub-agent fell back to registry)

## Go Installer — Skill Registry Generator

```go
// internal/skill/registry/generator.go
package registry

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"
)

type SkillTier string
const (
    TierFoundation       SkillTier = "foundation"
    TierContextActivated SkillTier = "context_activated"
    TierOnDemand         SkillTier = "on_demand"
)

type SkillEntry struct {
    Name          string
    Path          string
    Tier          SkillTier
    ActivatesWhen string
    CompactLines  int
}

// GenerateManifest creates .atl/skill-manifest.yaml based on Tiers
func GenerateManifest(atDir string, extraSkills []SkillEntry) error {
    // ... manifest generation logic ...
    return nil
}

// GenerateFoundationBlock merges compact rules of all Tier 1 skills into one file
func GenerateFoundationBlock(atDir string) error {
    // ... merges 6 Tier 1 skills into .atl/_generated/foundation.md ...
    return nil
}
```

## Orchestrator Delegation Pattern

**BEFORE** (duplicated in each orchestrator):
```
## Research Routing
STEP 1: mem_search in Engram
STEP 2: rg in local codebase
STEP 3: context7 for docs
STEP 4: notebooklm if available
```

**AFTER** (in ALL orchestrators):
```
## Research Delegation [MANDATORY]

Never implement research routing inline. Delegate ALL investigation:

WHEN research_needed:
  DELEGATE to researcher sub-agent with:
    research_query: "{specific question}"
    context: "{why — current phase/task}"
    scope_hint: "local|docs|broad"
    max_depth: "quick|standard|deep"

  AWAIT researcher summary
  USE summary findings without re-researching

researcher is self-contained. Do NOT second-guess its routing decisions.
If researcher returns NOT_FOUND, accept it and proceed with uncertainty noted.
```

## Scenarios

### Scenario 1: Researcher Receives All Research Queries
**Given** sdd-explore phase needs to understand a function's callers.
**When** research is needed.
**Then** sdd-explore MUST delegate to researcher (NOT implement its own rg search).
**And** researcher MUST return summary with `source: "local_codebase"`.
**And** NO rg calls in sdd-explore directly — only in researcher.

### Scenario 2: bash-expert Fish Detection
**Given** agent running in Fish shell environment (`SHELL=/usr/local/bin/fish`).
**When** shell detection runs.
**Then** fish syntax MUST be used.
**And** rg MUST still be used (not grep).
**And** NO bash-specific syntax (`set -euo pipefail`) in fish context.

### Scenario 3: rg vs grep Enforcement
**Given** any search task in any agent.
**When** search commands are generated.
**Then** ONLY rg commands in output.
**And** ZERO `grep -r` commands EVER.

### Scenario 4: bridge:always Injection
**Given** any sub-agent launch (researcher, solver, sdd-apply, etc.).
**When** prompt is constructed.
**Then** first 6 compact rules MUST be bridge skills.
**And** mcp-notebooklm-orchestrator MUST NOT be injected by default.

### Scenario 5: Researcher Tier Escalation
**Given** question answerable from Engram memory.
**When** `researcher("how is authentication implemented?", scope_hint="local")`.
**Then** Tier 1 (Engram) MUST hit → return WITHOUT calling Context7 or web.
**And** NO context7 or notebooklm calls made when Engram has the answer.

### Scenario 6: Researcher NOT_FOUND Handling
**Given** novel question with no local or Engram answer, `max_depth="quick"`.
**When** `researcher("how does obscure-library-xyz handle routing?", max_depth="quick")`.
**Then** Tiers 1-3 exhausted, `status: "NOT_FOUND"`, `confidence: "low"`.
**And** delegating agent MUST accept NOT_FOUND and mark as RISK (not retry).

## Expected Results

| Metric | Before | After |
|---|---|---|
| Research routing duplication | ~6 files with own routing | 1 file (researcher) |
| Fish shell support | ❌ No | ✅ Native in bash-expert v2 |
| rg enforcement | ⚠️ Only in ripgrep skill | ✅ bridge:always in all |
| mcp-notebooklm overhead | ✅ bridge (overhead in ALL) | ✅ On-demand via researcher |
| Research routing consistency | ⚠️ Drift between orchestrators | ✅ Single centralized protocol |
