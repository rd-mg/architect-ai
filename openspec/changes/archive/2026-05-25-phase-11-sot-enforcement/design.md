# Design: Phase 11 — Source of Truth Enforcement

> **Source of Truth:** `/home/rdmachadog/Documents/fix_achitect-ai/11-phase-sot-enforcement.md`
> **Change:** phase-11-sot-enforcement
> **Phase:** sdd-design
> **Generated:** 2026-05-22T19:17:00Z
> **Author:** sdd-design (architect-ai)

## Architecture

Phase 11 introduces a Validation Middleware pattern across the SDD pipeline with 3 mechanisms: Zero-Deviation protocol, Traceability injection, and Assumption Linting — all executable and verifiable.

## FMEA Matrix

| Component | Failure Mode | Effect | P | S | RPN | Mitigation |
|---|---|---|---|---|---|---|
| Zero-Deviation Rule | Spec has syntax error | sdd-apply implements it; tests fail | 3 | 2 | 6 | Tests catch; attempt 2 authorizes fix |
| Assumption Linter | False positive on "probably" in comment | Build halts unnecessarily | 2 | 2 | 4 | Scope rg to text blocks, exclude code fences |
| Hard-Stop Protocol | Agent halts too often on minor ambiguity | High user friction | 4 | 3 | 12 | --force-assume flag for minor items |
| Traceability Header | Agent forgets to inject | Traceability lost silently | 3 | 4 | 12 | sdd-verify Step 0b checks all artifacts |
| PendingDecision Rule | "PendingDecision" inside valid code string | False BLOCK | 2 | 2 | 4 | Exclude lines inside code blocks |
| SOT loading | Engram unavailable | Agent invents spec | 2 | 5 | 10 | Fallback: read from openspec/ files |

## Component 1: Zero-Deviation Coder Protocol

### Patch: `internal/assets/sdd/sdd-apply.md`

```markdown
## Zero-Deviation Coder Protocol [MANDATORY — Identity Block for sdd-apply]

You are a **Zero-Deviation Coder**. Your job is to transcribe the spec and design
into working code. You are NOT an architect. You are NOT a refactorer.

### Source of Truth Loading [Step 0 — before writing any code]

```bash
CHANGE="${CHANGE_NAME}"
STATE=".atl/sdd-state.yaml"
DESIGN_STATUS=$(grep -A3 "sdd-design:" "${STATE}" | grep "status:" | awk '{print $2}' | tr -d '"')
TASKS_STATUS=$(grep -A3 "sdd-tasks:" "${STATE}" | grep "status:" | awk '{print $2}' | tr -d '"')
[ "${DESIGN_STATUS}" != "completed" ] && { echo "BLOCKED: sdd-design not completed" >&2; exit 1; }
[ "${TASKS_STATUS}" != "completed" ] && { echo "BLOCKED: sdd-tasks not completed" >&2; exit 1; }
```

Load from Engram:
```
spec_doc   = mem_get_observation("sdd/${CHANGE}/spec")
design_doc = mem_get_observation("sdd/${CHANGE}/design")
tasks_doc  = mem_get_observation("sdd/${CHANGE}/tasks")
```

### First Attempt Constraint [STRICT]

On attempt_number = 1:
- Implement ALL code from spec+design EXACTLY as written
- DO NOT analyze, re-design, or improve
- DO NOT add abstractions not in design
- DO NOT change function signatures
- DO NOT rename variables, fields, or methods
- If spec code has syntax error → implement it exactly, let tests fail

### Modification Authorization Table

| Condition | Authorized? | Action |
|---|---|---|
| First attempt, spec clear | ❌ NO | Transcribe exactly |
| First attempt, syntax error | ❌ NO | Transcribe exactly; tests catch |
| Tests fail (attempt 2+) | ✅ YES | Fix specific failure only |
| Design ambiguous | ❌ NO | HALT; emit BLOCKED |
| "Better way" known | ❌ NO | Implement as spec; note in risks |
| YAGNI spotted | ❌ NO | Implement as spec; log concern |

### Implementation Workflow

For each task (status: "pending"):
1. Read task from apply-progress.yaml
2. Find corresponding code block in design_doc
3. Implement VERBATIM — copy interfaces, signatures, field names exactly
4. If tdd_mode=true: RED → GREEN → TRIANGULATE → REFACTOR commits
5. If tdd_mode=false: implementation + tests in same commit
6. Update apply-progress.yaml: status → "completed"
```

## Component 2: Traceability Auto-Injection

### `internal/assets/_shared/traceability-injection.md`

Every SDD artifact MUST contain after the `# Title` line:

```markdown
> **Source of Truth:** {absolute_path_or_engram_key}
> **Change:** {change_name}
> **Phase:** {sdd-phase-name}
> **Generated:** {ISO8601_timestamp}
> **Author:** {agent_name} (architect-ai)
```

Phase-specific sources:
- sdd-propose: Source = user request summary or ticket URL
- sdd-spec: Source = `openspec/changes/{change}/proposal.md`
- sdd-design: Source = `openspec/changes/{change}/spec.md`
- sdd-tasks: Source = `openspec/changes/{change}/design.md`

Validation in sdd-verify:
```bash
for artifact in openspec/changes/${CHANGE}/*.md; do
  if ! grep -q "^\> \*\*Source of Truth:\*\*" "${artifact}"; then
    echo "TRACEABILITY MISSING: ${artifact}" >&2
    VIOLATIONS=$((VIOLATIONS + 1))
  fi
done
[ "${VIOLATIONS}" -gt 0 ] && exit 1
```

## Component 3: Assumption Linter

### `internal/assets/scripts/assumption-linter.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail
CHANGE="${1:?change_name required}"
ARTIFACT_DIR="openspec/changes/${CHANGE}"
VIOLATIONS=0
TOTAL_CHECKED=0

[ ! -d "${ARTIFACT_DIR}" ] && { echo "WARN: ${ARTIFACT_DIR} not found"; exit 0; }

echo "Running Assumption Linter on: ${ARTIFACT_DIR}"
echo "──────────────────────────────────────────"

# Rule 1: No PendingDecision
echo "Checking Rule 1: No PendingDecision..."
while IFS= read -r -d '' file; do
  TOTAL_CHECKED=$((TOTAL_CHECKED + 1))
  if grep -nE '\bPendingDecision\b|\bTo Be Determined\b' "${file}" 2>/dev/null | grep -v '```' | grep -q .; then
    echo "  VIOLATION: PendingDecision found in ${file}"
    grep -nE '\bPendingDecision\b|\bTo Be Determined\b' "${file}" | head -3
    VIOLATIONS=$((VIOLATIONS + 1))
  fi
done < <(find "${ARTIFACT_DIR}" -name "*.md" -print0)

# Rule 2: No FixMe/To-Do
echo "Checking Rule 2: No FixMe/To-Do..."
while IFS= read -r -d '' file; do
  if grep -nE '\bFixMe\b|\bTo-Do\b' "${file}" 2>/dev/null | grep -q .; then
    echo "  VIOLATION: FixMe/To-Do found in ${file}"
    VIOLATIONS=$((VIOLATIONS + 1))
  fi
done < <(find "${ARTIFACT_DIR}" -name "*.md" -print0)

# Rule 3: Implicit assumptions
echo "Checking Rule 3: No implicit assumptions..."
ASSUMPTION_PATTERNS=('assuming that' 'it is assumed' 'we assume' 'should work' 'probably' 'might need' 'not sure' 'unclear')
while IFS= read -r -d '' file; do
  for pattern in "${ASSUMPTION_PATTERNS[@]}"; do
    if grep -inE "${pattern}" "${file}" 2>/dev/null | grep -q .; then
      echo "  WARN: Implicit assumption in ${file}: '${pattern}'"
    fi
  done
done < <(find "${ARTIFACT_DIR}" -name "*.md" -print0)

# Rule 4: Traceability headers
echo "Checking Rule 4: Traceability headers..."
while IFS= read -r -d '' file; do
  if ! grep -q "Source of Truth:" "${file}" 2>/dev/null; then
    echo "  VIOLATION: Missing traceability in ${file}"
    VIOLATIONS=$((VIOLATIONS + 1))
  fi
done < <(find "${ARTIFACT_DIR}" -name "*.md" -print0)

# Rule 5: BDD in spec
echo "Checking Rule 5: BDD assertions in spec..."
SPEC="${ARTIFACT_DIR}/spec.md"
if [ -f "${SPEC}" ] && ! grep -qiE 'given|when|then' "${SPEC}" 2>/dev/null; then
  echo "  VIOLATION: spec.md has no BDD assertions"
  VIOLATIONS=$((VIOLATIONS + 1))
fi

echo "──────────────────────────────────────────"
echo "Checked: ${TOTAL_CHECKED} files | Violations: ${VIOLATIONS}"
[ "${VIOLATIONS}" -gt 0 ] && { echo "FAIL: ${VIOLATIONS} violation(s)"; exit 1; }
echo "PASS"
```

## Component 4: Hard-Stop Protocol

### `internal/assets/_shared/hard-stop-protocol.md`

HALT conditions:
- `source_of_truth is null` → HARD STOP
- `required_function_not_in_design` → HARD STOP
- `field_conflict_detected` → HARD STOP
- `task_dependency_not_met` → HARD STOP
- `contradictory_requirements` → HARD STOP

Output format:
```json
{
  "status": "blocked",
  "phase": "sdd-apply",
  "executive_summary": "Hard stop: {reason}",
  "blocked_reason": "{specific reason}",
  "clarification_needed": "{question for user}",
  "attempt_number": 1
}
```

--force-assume: document assumption in apply-progress.yaml, add to risks, mark task "assumed", flag in verify. NOT for: missing SOT, field conflicts, broken dependencies.

## Component 5: Gap Detector — sdd-verify Patch

### Steps 0a-0d in `sdd-verify`

```bash
# Step 0a: Assumption Linter
bash .atl/scripts/assumption-linter.sh "${CHANGE_NAME}"
[ $? -ne 0 ] && { echo "BLOCKED: Assumption linter failed"; exit 1; }

# Step 0b: Traceability Check
VIOLATIONS=0
for artifact in openspec/changes/${CHANGE_NAME}/*.md; do
  [ -f "${artifact}" ] || continue
  grep -q "Source of Truth:" "${artifact}" || { echo "TRACEABILITY MISSING: ${artifact}"; VIOLATIONS=$((VIOLATIONS+1)); }
done
[ "${VIOLATIONS}" -gt 0 ] && { echo "BLOCKED: ${VIOLATIONS} missing headers"; exit 1; }

# Step 0c: Semantic Audit (Phase 02)
# Step 0d: Validate prior phases
for phase in sdd-spec sdd-design sdd-tasks; do
  STATUS=$(grep -A3 "  ${phase}:" .atl/sdd-state.yaml | grep "status:" | awk '{print $2}' | tr -d '"')
  [ "${STATUS}" != "completed" ] && { echo "BLOCKED: ${phase} is '${STATUS}'"; exit 1; }
done
```

## Component 6: Go Installer — Linter Scripts

### `internal/sdd/linter/installer.go`

```go
package linter

import (
	"fmt"
	"os"
	"path/filepath"
)

const assumptionLinterScript = `#!/usr/bin/env bash
set -euo pipefail
CHANGE="${1:?change_name required}"
ARTIFACT_DIR="openspec/changes/${CHANGE}"
VIOLATIONS=0
[ ! -d "${ARTIFACT_DIR}" ] && { echo "WARN: ${ARTIFACT_DIR} not found"; exit 0; }
while IFS= read -r -d '' file; do
  grep -nqE '\bPendingDecision\b|\bTo Be Determined\b' "${file}" 2>/dev/null && {
    echo "VIOLATION PendingDecision: ${file}"; VIOLATIONS=$((VIOLATIONS+1)); }
  grep -nqE '\bFixMe\b|\bTo-Do\b' "${file}" 2>/dev/null && {
    echo "VIOLATION FixMe/To-Do: ${file}"; VIOLATIONS=$((VIOLATIONS+1)); }
  grep -q "Source of Truth:" "${file}" 2>/dev/null || {
    echo "VIOLATION TRACE: ${file}"; VIOLATIONS=$((VIOLATIONS+1)); }
done < <(find "${ARTIFACT_DIR}" -name "*.md" -print0 2>/dev/null)
[ "${VIOLATIONS}" -gt 0 ] && { echo "FAIL: ${VIOLATIONS} violation(s)"; exit 1; }
echo "PASS"
`

const validateContractScript = `#!/usr/bin/env bash
set -euo pipefail
RESULT=$(cat)
python3 -c "
import json, sys
data = json.loads(sys.argv[1])
required = ['status','phase','executive_summary','artifacts','next_recommended','risks','skill_resolution']
missing = [f for f in required if f not in data]
if missing:
    print('INVALID: missing fields:', missing)
    sys.exit(1)
valid_statuses = ['completed','failed','blocked','abandoned']
if data.get('status') not in valid_statuses:
    print('INVALID: status must be one of', valid_statuses)
    sys.exit(1)
sr = data.get('skill_resolution', {})
if not isinstance(sr, dict) or 'status' not in sr:
    print('INVALID: skill_resolution.status missing')
    sys.exit(1)
print('VALID')
" \"${RESULT}\"
`

func InstallScripts(atDir string) error {
	scriptsDir := filepath.Join(atDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		return fmt.Errorf("create scripts dir: %w", err)
	}
	scripts := map[string]string{
		"assumption-linter.sh":        assumptionLinterScript,
		"validate-result-contract.sh": validateContractScript,
	}
	for name, content := range scripts {
		path := filepath.Join(scriptsDir, name)
		if err := os.WriteFile(path, []byte(content), 0755); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}
	return nil
}
```

### `internal/sdd/linter/installer_test.go`

```go
package linter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallScripts_CreatesFiles(t *testing.T) {
	dir := t.TempDir()
	if err := InstallScripts(dir); err != nil { t.Fatal(err) }
	expected := []string{"scripts/assumption-linter.sh", "scripts/validate-result-contract.sh"}
	for _, f := range expected {
		path := filepath.Join(dir, f)
		info, err := os.Stat(path)
		if os.IsNotExist(err) { t.Errorf("not created: %s", f); continue }
		if info.Mode()&0111 == 0 { t.Errorf("not executable: %s", f) }
	}
}

func TestInstallScripts_Idempotent(t *testing.T) {
	dir := t.TempDir()
	if err := InstallScripts(dir); err != nil { t.Fatal(err) }
	if err := InstallScripts(dir); err != nil { t.Fatal("second should succeed") }
}
```

## Key Decisions

- **Zero-Deviation = transcriptor, not architect**: Prevents LLM creative drift on first attempt.
- **Attempt-gated modification**: Only test failures unlock code changes.
- **Assumption linter before verify**: Catches gaps early, not at the end.
- **Hard-stop with BLOCKED JSON**: Structured output the orchestrator can parse.

## MANDATORY IMPLEMENTATION DIRECTIVE
**For `sdd-apply` (Coder):**
All code definitions, declarations, directives, and logic provided in the Source of Truth MUST be implemented EXACTLY as written on the first attempt. The coder MUST NOT analyze, re-design, or deviate from the source code during the initial implementation. Modification of the initial implementation is ONLY permitted if the tests fail.
