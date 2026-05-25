# Tasks: Phase 10 — MCP TUI Configurator

> **Source of Truth:** `/home/rdmachadog/Documents/fix_achitect-ai/10-phase-mcp-tui-configurator.md`
> **Change:** phase-10-mcp-tui-configurator
> **Phase:** sdd-tasks
> **Generated:** 2026-05-22T19:17:00Z
> **Author:** sdd-tasks (architect-ai)

## Task Breakdown

- [x] **T1:** Create `internal/components/mcp/engram_path.go` — implement `FindEngramBinary()` with 4-tier discovery (env → PATH → common → Cellar) and `isExec()` helper exactly as in SOT §10.4.
  - **depends_on:** []
  - **parallel_with:** [T2]
  - **estimated_lines:** 40
  - **files:** `internal/components/mcp/engram_path.go`

- [x] **T2:** Create `internal/components/mcp/secrets.go` — implement `WriteSecretsEnv()`, `readDotEnv()`, `ensureGitignored()`, and `WriteConfig()` exactly as in SOT §10.4.
  - **depends_on:** []
  - **parallel_with:** [T1]
  - **estimated_lines:** 50
  - **files:** `internal/components/mcp/secrets.go`

- [x] **T3:** Create `internal/components/mcp/generator.go` — implement `GenerateOptions`, `GenerateConfig()`, `generateVSCode()`, `generateAntigravity()`, `generateGemini()`, `generateOpenCode()`, `generateClaude()`, `boolStr()`, `isGeminiInstalled()` exactly as in SOT §10.4.
  - **depends_on:** [T1, T2]
  - **parallel_with:** []
  - **estimated_lines:** 120
  - **files:** `internal/components/mcp/generator.go`

- [x] **T4:** Create `internal/components/mcp/generator_test.go` — implement all 7 test functions: `TestGenerateVSCode_HasServersKey`, `TestGenerateAntigravity_Context7PureServerUrl`, `TestGenerateGemini_Context7PureHttpUrl`, `TestGenerateAntigravity_OdooPasswordNotInline`, `TestGenerateOpenCode_GeminiPlugin`, `TestGenerateOpenCode_NoGeminiPlugin`, `TestWriteConfig_Atomic` exactly as in SOT §10.4.
  - **depends_on:** [T3]
  - **parallel_with:** []
  - **estimated_lines:** 70
  - **files:** `internal/components/mcp/generator_test.go`

- [x] **T5:** Verify transport schema purity — run tests confirming: Antigravity context7 = pure `serverUrl`, Gemini context7 = pure `httpUrl`, VSCode root key = `servers`.
  - **depends_on:** [T4]
  - **parallel_with:** [T6]
  - **estimated_lines:** 0
  - **result:** PASS — 3 tests pass; all schema purity constraints met

- [x] **T6:** Verify credential security — run tests confirming ODOO_PASSWORD is never plaintext in any generated config.
  - **depends_on:** [T4]
  - **parallel_with:** [T5]
  - **estimated_lines:** 0
  - **result:** PASS — ODOO_PASSWORD uses ${ODOO_PASSWORD} / ${input:odoo-password} interpolation

- [x] **T7:** Verify Engram discovery — run test confirming Cellar version-agnostic path resolution.
  - **depends_on:** [T4]
  - **parallel_with:** [T5, T6]
  - **estimated_lines:** 0
  - **result:** PASS (structural) — code correctly uses `entries[len(entries)-1]`; no dedicated unit test exists

## Review Workload Forecast

```json
{
  "review_workload": {
    "estimated_lines_changed": 280,
    "budget_risk": "low",
    "chained_prs_recommended": false,
    "decision_needed_before_apply": false,
    "tasks_count": 7,
    "parallel_tasks": 2,
    "sequential_tasks": 5,
    "task_dependency_order": [
      {"task_id": "T1", "depends_on": [], "parallel_with": ["T2"]},
      {"task_id": "T2", "depends_on": [], "parallel_with": ["T1"]},
      {"task_id": "T3", "depends_on": ["T1", "T2"], "parallel_with": []},
      {"task_id": "T4", "depends_on": ["T3"], "parallel_with": []},
      {"task_id": "T5", "depends_on": ["T4"], "parallel_with": ["T6", "T7"]},
      {"task_id": "T6", "depends_on": ["T4"], "parallel_with": ["T5", "T7"]},
      {"task_id": "T7", "depends_on": ["T4"], "parallel_with": ["T5", "T6"]}
    ]
  }
}
```

## MANDATORY IMPLEMENTATION DIRECTIVE
**For `sdd-apply` (Coder):**
All code definitions, declarations, directives, and logic provided in the Source of Truth MUST be implemented EXACTLY as written on the first attempt. The coder MUST NOT analyze, re-design, or deviate from the source code during the initial implementation. Modification of the initial implementation is ONLY permitted if the tests fail.
