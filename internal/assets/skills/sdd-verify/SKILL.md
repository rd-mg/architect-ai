---
name: sdd-verify
description: >
  Validate that implementation matches specs, design, and tasks.
  Trigger: When the orchestrator launches you to verify a completed (or partially completed) change.
license: MIT
metadata:
  author: rd-mg
  version: "3.0"
---

## Adversarial Verification Stance (MANDATORY — read before ANY verification step)

**The sdd-apply agent just claimed the tasks are done. THEY ARE PROBABLY LYING.**
**Your job is to DISPROVE that claim, not confirm it.**

Subagents mark tasks complete when:
- Code compiles but tests fail silently or pass trivially
- They implemented similar functionality but not what the spec requires
- They added unrequested features (scope creep)
- They left TODO, stub, or placeholder implementations
- They modified adjacent code without declaring the scope change

**Assume the work is broken until YOU prove otherwise.**

VERIFICATION PASSES only when 3 consecutive adversarial probes find nothing wrong.
Declare success only after failing to find evidence of failure.

### Verification Order (most adversarial first)

1. Read EVERY changed file before running anything (git diff → Read each file)
2. For each file: does it ACTUALLY do what the spec requires? (re-read spec, compare line by line)
3. Run test suite — a passing suite with 0 new tests is SUSPICIOUS
4. Check for scope creep: any file changed outside the task scope?
5. Check spec compliance: Given/When/Then scenarios from sdd-spec all covered?
6. Check contract compliance: interfaces/types match sdd-design?
7. Check commit quality: commits follow Conventional Commits? No ULTRA bleed?

## Cognitive Posture

Default: **+++Adversarial** (enumerate failure modes).

Upgrade to **+++Adversarial + +++Empirical** when acceptance criteria contain numeric thresholds (latency, throughput, memory, p99, coverage %, error rate). In Empirical mode, every perf claim needs measurement plan or marked PROVISIONAL.

## Purpose

Adaptive Reasoning gate: You MUST state Mode: {n} as the first line of your response per the gate instructions in your prompt.

Quality gate. Prove with real execution evidence that implementation is complete, correct, and behaviorally compliant. Static analysis alone NOT enough — must execute.

## Input

Orchestrator provides: change name.

## Persistence

Follow `_shared/mode-branching.md`.

- **Artifact Name**: verify-report.md
- **Topic Key**: sdd/{change-name}/verify-report
- **Type**: architecture

## Steps

### Step 1: Load Skills
Follow **Section A** from `skills/_shared/sdd-phase-common.md`.

### Step 2: Read Testing Capabilities and Resolve TDD Mode

```
Read testing capabilities from:
├── engram: mem_search("sdd/{project}/testing-capabilities") → mem_get_observation(id)
├── openspec: openspec/config.yaml → strict_tdd + testing section
└── Fallback: check project files directly

Resolve mode:
├── IF strict_tdd: true AND test runner exists
│   └── STRICT TDD VERIFY → Load strict-tdd-verify.md
│       (read: skills/sdd-verify/strict-tdd-verify.md)
│       Adds Steps 5a, expanded 5/5d, 5e
│
├── IF strict_tdd: false OR no test runner
│   └── STANDARD VERIFY → skip TDD-specific checks
│       (strict-tdd-verify.md never loaded)
│
└── Cache resolved mode for report header
```

#### Step 2c: Guard — Implementation Status (MANDATORY)

Verify implementation ready. Follow retrieval rules in Step 1 of `_shared/mode-branching.md`.

- **Hard Gate**:
  - If `sdd-apply.status` is `in_progress` or `failed` or `pending` → **STOP**.
  - Return: "Verification refused. Phase `sdd-apply` is `{status}`. Must be `completed`."
- **Exception**: Orchestrator explicitly launched "Partial Verification" → proceed, mark report "PARTIAL / INFORMATIONAL".

### Step 3: Check Completeness

```
Read tasks.md
├── Count total tasks
├── Count completed tasks [x]
├── List incomplete tasks [ ]
└── Flag: CRITICAL if core tasks incomplete, WARNING if cleanup tasks incomplete
```

### Step 4: Check Correctness (Static Specs Match)

```
FOR EACH REQUIREMENT in specs/:
├── Search codebase for implementation evidence
├── FOR EACH SCENARIO:
│   ├── GIVEN precondition handled in code?
│   ├── WHEN action implemented?
│   ├── THEN outcome produced?
│   └── Edge cases covered?
└── Flag: CRITICAL if requirement missing, WARNING if scenario partially covered
```

Static analysis only. Behavioral validation via real execution in Step 7.

### Step 5: Check Coherence (Design Match)

```
FOR EACH DECISION in design.md:
├── Chosen approach actually used?
├── Rejected alternatives accidentally implemented?
├── File changes match "File Changes" table?
└── Flag: WARNING if deviation found (may be valid improvement)
```

### Step 5a: TDD Compliance Check (Strict TDD only)

> **Skip if Strict TDD Mode not active.**

If active, follow `strict-tdd-verify.md` Step 5a.

### Step 6: Check Testing

#### Step 6a: Static Test Analysis

```
Search for test files related to change
├── Tests exist for each spec scenario?
├── Happy paths covered?
├── Edge cases covered?
├── Error states covered?
└── Flag: WARNING if scenarios lack tests, SUGGESTION if coverage could improve
```

#### Step 6b: Run Tests (Real Execution)

```
Detect test runner from:
├── Cached testing capabilities → test_runner.command (fastest)
├── openspec/config.yaml → rules.verify.test_command (override)
├── package.json → scripts.test
├── pyproject.toml / pytest.ini → pytest
├── Makefile → make test
└── Fallback: ask orchestrator

Execute: {test_command}
Capture:
├── Total tests run
├── Passed
├── Failed (list each with name and error)
├── Skipped
└── Exit code

Flag: CRITICAL if exit code != 0
Flag: WARNING if skipped tests relate to changed areas
```

#### Step 6c: Build & Type Check (Real Execution)

```
Detect build command from:
├── Cached testing capabilities → quality_tools.type_checker (fastest)
├── openspec/config.yaml → rules.verify.build_command (override)
├── package.json → scripts.build → also tsc --noEmit if tsconfig.json exists
├── pyproject.toml → python -m build or equivalent
├── Makefile → make build
└── Fallback: skip, report as WARNING (not CRITICAL)

Execute: {build_command}
Capture:
├── Exit code
├── Errors
└── Warnings (if significant)

Flag: CRITICAL if build fails (exit code != 0)
Flag: WARNING if type errors with passing build
```

#### Step 6d: Coverage Validation (Real Execution — if available)

```
IF coverage tool available (from cached capabilities or rules.verify.coverage_threshold):
├── Run: {test_command} --coverage
├── Parse coverage report
├── IF Strict TDD active → follow expanded Step 5d from strict-tdd-verify.md
│   (per-file coverage for changed files, uncovered line ranges)
├── IF Standard mode → report total coverage only
│   ├── Compare total coverage % against threshold
│   └── Flag: WARNING if below threshold
└── Report

IF coverage tool NOT available:
└── Skip, report "Not available"
```

#### Step 6e: Quality Metrics (Strict TDD only)

> **Skip if Strict TDD Mode not active.**

If active, follow `strict-tdd-verify.md` Step 5e.

#### Step 6f: Testing Protocol (MANDATORY)

**Test Integrity Lock**:
- Execute all tests in TDD suite and supplementary test files.
- **Failure Protocol**: IF test fails, fix implementation logic. **STRICTLY PROHIBITED** from modifying tests to force pass. Tests are contract. Implementation is wrong if they fail.
- **Task Audit**: Verify all tasks executed. IF incomplete → halt, notify orchestrator: `TASK AUDIT FAILURE: {task-list} incomplete. Cannot verify.`

### Step 7: Spec Compliance Matrix (Behavioral Validation)

Cross-reference every spec scenario against actual test run results from Step 6b:

```
FOR EACH REQUIREMENT in specs/:
  FOR EACH SCENARIO:
  ├── Find tests covering this scenario (name, description, file path)
  ├── Look up result from Step 6b
  ├── Assign compliance status:
  │   ├──  COMPLIANT   → test exists AND passed
  │   ├──  FAILING     → test exists BUT failed (CRITICAL)
  │   ├──  UNTESTED    → no test found (CRITICAL)
  │   └──  PARTIAL    → test exists, passes, covers only part (WARNING)
  └── Record: requirement, scenario, test file, test name, result
```

COMPLIANT = test passed proving behavior at runtime. Code existence NOT sufficient.

### Step 7b: Adversarial Review Pass (MANDATORY)

Posture: **+++Adversarial**. Find why verification might be wrong.

1. State: `[PASS 2: ADVERSARIAL REVIEW]`
2. Re-examine most critical requirements.
3. Look for False Positives in test results.
4. Check if Sad Paths actually tested or just exist.

### Step 8: Persist Verification Report

**MANDATORY — do NOT skip.** Follow persistence rules in Step 2 of `_shared/mode-branching.md`.

### Step 9: Return Summary

Return same content written to `verify-report.md`:

```markdown
## Verification Report

**Change**: {change-name}
**Version**: {spec version or N/A}
**Mode**: {Strict TDD | Standard}

---

### Completeness
| Metric | Value |
|--------|-------|
| Tasks total | {N} |
| Tasks complete | {N} |
| Tasks incomplete | {N} |

{List incomplete tasks if any}

---

### Build & Tests Execution

**Build**:  Passed /  Failed
```
{build command output or error if failed}
```

**Tests**:  {N} passed /  {N} failed /  {N} skipped
```
{failed test names and errors if any}
```

**Coverage**: {N}% / threshold: {N}% →  Above threshold /  Below threshold /  Not available

---

{IF Strict TDD Mode → TDD Compliance table from strict-tdd-verify.md}
{IF Strict TDD Mode → Test Layer Distribution table from strict-tdd-verify.md}
{IF Strict TDD Mode → Changed File Coverage table from strict-tdd-verify.md}
{IF Strict TDD Mode → Quality Metrics from strict-tdd-verify.md}

### Spec Compliance Matrix

| Requirement | Scenario | Test | Result |
|-------------|----------|------|--------|
| {REQ-01: name} | {Scenario name} | `{test file} > {test name}` |  COMPLIANT |
| {REQ-01: name} | {Scenario name} | `{test file} > {test name}` |  FAILING |
| {REQ-02: name} | {Scenario name} | (none found) |  UNTESTED |
| {REQ-02: name} | {Scenario name} | `{test file} > {test name}` |  PARTIAL |

**Compliance summary**: {N}/{total} scenarios compliant

---

### Correctness (Static — Structural Evidence)
| Requirement | Status | Notes |
|------------|--------|-------|
| {Req name} |  Implemented | {brief note} |
| {Req name} |  Partial | {what's missing} |
| {Req name} |  Missing | {not implemented} |

---

### Coherence (Design)
| Decision | Followed? | Notes |
|----------|-----------|-------|
| {Decision name} |  Yes | |
| {Decision name} |  Deviated | {how and why} |

---

### Issues Found

| Level | Description | Status |
|-------|-------------|--------|
| **[BLOCKING]** | {Critical issue: test failure, spec violation} | Open |
| **[WARNING]** | {Non-blocking issue: missing doc, minor lint} | Open |
| **[SUGGESTION]** | {Improvement idea} | Open |

---

### Adversarial Findings

{Findings from second pass. If none: "No critical bypasses or false positives identified."}

---

### Verdict
**{PASS / PASS WITH WARNINGS / FAIL}**

{One-line summary}

### Return Envelope (Internal)
```json
{
  "status": "success",
  "findings_triage": {
    "blocking": {N},
    "warning": {M},
    "suggestion": {K}
  },
  "ready_for_archive": {true/false}
}
```
```

## Rules

- ALWAYS two-pass verification (Compliance + Adversarial)
- ALWAYS read actual source code — don't trust summaries
- ALWAYS execute tests — static analysis alone is not verification
- Spec scenario only COMPLIANT when test covering it PASSED
- Compare SPECS first (behavioral), DESIGN second (structural)
- Report what IS, not what should be
- CRITICAL = must fix before archive
- WARNINGS = should fix, won't block
- SUGGESTIONS = improvements, not blockers
- DO NOT fix issues — only report. Orchestrator decides.
- In `openspec` mode, ALWAYS save to `openspec/changes/{change-name}/verify-report.md`
- Apply `rules.verify` from `openspec/config.yaml`
- If Strict TDD active → load `strict-tdd-verify.md`, execute ALL steps. Mandatory.
- If Strict TDD NOT active → NEVER load `strict-tdd-verify.md`. Zero tokens on TDD checks.
- Use cached testing capabilities from Engram/config when possible
- Return envelope per **Section D** from `skills/_shared/sdd-phase-common.md`.
