# Fase 13: Makefile, Script de Verificación Global y Migrate CLI

**Objetivo:** Proporcionar los comandos de entrada para aplicar y verificar todo el hardening v0.3 de una sola ejecución. Crea el `Makefile` con los targets de hardening, el script de verificación completo, y el subcomando `architect-ai migrate-v03` para migrar estado de proyectos existentes.

---

## Paso 1: Crear el target `hardening` en el `Makefile`

**Archivo a modificar:** `Makefile`

**Acción:** Modificar — añadir los targets al final del Makefile existente

**Comando previo:**
```bash
tail -20 Makefile
```

**Código a añadir al final del `Makefile`:**
```makefile
# ─── architect-ai Hardening v0.3 ─────────────────────────────────────────────

.PHONY: hardening hardening-verify hardening-prompts hardening-go test-race migrate-v03

## hardening-prompts: Apply all prompt-level patches (Markdown assets)
hardening-prompts:
	@echo "Applying prompt patches for hardening v0.3..."
	@# H-01: Fix enum running→in_progress in phase-dag-enforcement.md
	@sed -i \
	  "s/update status to \`running\`/update status to \`in_progress\`/g; \
	   s/status: \"running\"/status: \"in_progress\"/g; \
	   s/status: running$$/status: in_progress/g" \
	  internal/assets/_shared/phase-dag-enforcement.md
	@echo "  ✓ H-01: enum fix applied to phase-dag-enforcement.md"
	@sed -i \
	  "s/(running|failed)/(in_progress|failed)/g" \
	  internal/assets/_shared/rollback-harness.md
	@echo "  ✓ H-01: enum fix applied to rollback-harness.md"
	@# H-03: Remove Mode A from Gemini template
	@python3 scripts/patch_gemini_mode_a.py
	@echo "  ✓ H-03: Mode A removed from GEMINI.md"
	@echo "Prompt patches complete."

## hardening-go: Build and test all Go changes for hardening v0.3
hardening-go:
	@echo "Building all Go packages..."
	@go build ./... 2>&1 || (echo "BUILD FAILED" && exit 1)
	@echo "  ✓ Build passed"
	@echo "Running tests..."
	@go test ./internal/components/openspec/... -count=1 -timeout 120s 2>&1 || (echo "openspec TESTS FAILED" && exit 1)
	@go test ./internal/metering/... -count=1 -timeout 60s 2>&1 || (echo "metering TESTS FAILED" && exit 1)
	@go test ./internal/reasoning/gate/... -count=1 -timeout 60s 2>&1 || (echo "gate TESTS FAILED" && exit 1)
	@go test ./internal/verify/... -count=1 -timeout 60s 2>&1 || (echo "verify TESTS FAILED" && exit 1)
	@go test ./internal/components/engram/engramkeys/... -count=1 -timeout 60s 2>&1 || (echo "engramkeys TESTS FAILED" && exit 1)
	@go test ./internal/components/filemerge/... -count=1 -timeout 60s 2>&1 || (echo "filemerge TESTS FAILED" && exit 1)
	@go test ./internal/components/mcp/... -count=1 -timeout 60s 2>&1 || (echo "mcp TESTS FAILED" && exit 1)
	@go test ./internal/tui/... -count=1 -timeout 60s 2>&1 || (echo "tui TESTS FAILED" && exit 1)
	@go test ./internal/cli/... -count=1 -timeout 60s 2>&1 || (echo "cli TESTS FAILED" && exit 1)
	@echo "  ✓ All tests passed"

## hardening-verify: Verify all hardening patches are correctly applied
hardening-verify:
	@bash .atl/scripts/verify-hardening.sh

## test-race: Run all tests with the race detector
test-race:
	@go test -race \
	  ./internal/components/openspec/... \
	  ./internal/metering/... \
	  ./internal/reasoning/gate/... \
	  ./internal/tui/... \
	  -count=1 -timeout 180s

## hardening: Apply all patches, build, test, and verify (full pipeline)
hardening: hardening-prompts hardening-go hardening-verify
	@echo ""
	@echo "════════════════════════════════════════"
	@echo "architect-ai hardening v0.3 complete ✓"
	@echo "════════════════════════════════════════"

## migrate-v03: Migrate existing project sdd-state.yaml files to v0.3 enum
migrate-v03:
	@go run ./cmd/architect-ai migrate-v03
```

---

## Paso 2: Crear `scripts/patch_gemini_mode_a.py`

**Archivo a crear:** `scripts/patch_gemini_mode_a.py`

**Acción:** Crear

```python
#!/usr/bin/env python3
"""
patch_gemini_mode_a.py
Removes Mode A (inline execution) from internal/assets/gemini/GEMINI.md
and inserts the IMMUTABILITY RULE section in its place.
Idempotent: safe to run multiple times.
"""

import os
import sys

GEMINI_PATH = "internal/assets/gemini/GEMINI.md"

MODE_A_MARKER = "## Mode A (Gemini inline"
IMMUTABILITY_MARKER = "## IMMUTABILITY RULE"

IMMUTABILITY_BLOCK = """## IMMUTABILITY RULE — L0 is a Pure Router (no exceptions)

L0 NEVER executes any tool directly. Zero exceptions.
This includes: bash, shell commands, file reads, file writes, grep, find, cat,
any MCP call, any network request.

Rationale: L0 executing inline accumulates operational context that should be
isolated in L1. This artificially elevates D4 in subsequent delegations and
violates the Strict Isolation Rule between L1a and L1b.

If you find yourself about to call bash, read_file, or any tool: STOP.
Route to general-orchestrator via run_subagent instead.
Even "git status" goes through general-orchestrator (Mode C).

"""


def patch(path: str) -> bool:
    with open(path, "r", encoding="utf-8") as f:
        content = f.read()

    if IMMUTABILITY_MARKER in content:
        print(f"  Already patched: {path}")
        return False

    if MODE_A_MARKER not in content:
        print(f"  Mode A not found — may already be removed: {path}")
        return False

    # Find and remove the Mode A section
    lines = content.split("\n")
    new_lines = []
    skip = False
    inserted = False

    for i, line in enumerate(lines):
        if line.startswith("## Mode A (Gemini inline"):
            skip = True
            if not inserted:
                new_lines.append(IMMUTABILITY_BLOCK.rstrip())
                inserted = True
            continue
        if skip:
            # Skip until the next ## section
            if line.startswith("## ") and not line.startswith("## Mode A"):
                skip = False
                new_lines.append(line)
            # else: continue skipping Mode A content
        else:
            new_lines.append(line)

    patched = "\n".join(new_lines)

    tmp_path = path + ".patch.tmp"
    with open(tmp_path, "w", encoding="utf-8") as f:
        f.write(patched)
    os.rename(tmp_path, path)
    print(f"  Patched: {path}")
    return True


def verify(path: str) -> bool:
    with open(path, "r", encoding="utf-8") as f:
        content = f.read()
    if MODE_A_MARKER in content:
        print(f"  FAIL: Mode A still present in {path}", file=sys.stderr)
        return False
    if IMMUTABILITY_MARKER not in content:
        print(f"  FAIL: IMMUTABILITY RULE not found in {path}", file=sys.stderr)
        return False
    print(f"  PASS: {path} verified")
    return True


if __name__ == "__main__":
    if not os.path.exists(GEMINI_PATH):
        print(f"SKIP: {GEMINI_PATH} not found", file=sys.stderr)
        sys.exit(0)
    patch(GEMINI_PATH)
    if not verify(GEMINI_PATH):
        sys.exit(1)
```

**Comando para hacerlo ejecutable:**
```bash
chmod +x scripts/patch_gemini_mode_a.py
```

---

## Paso 3: Crear el script `verify-hardening.sh` completo

**Archivo a crear:** `.atl/scripts/verify-hardening.sh`

**Acción:** Crear

```bash
#!/usr/bin/env bash
# .atl/scripts/verify-hardening.sh
# Verifies all architect-ai hardening v0.3 patches are applied.
# Exit 0 = all checks pass. Exit 1 = one or more checks failed.
# Usage: bash .atl/scripts/verify-hardening.sh
# Or via Makefile: make hardening-verify

set -euo pipefail

PASS=0
FAIL=0
WARN=0

pass() { echo "  ✓ $1"; ((PASS++)) || true; }
fail() { echo "  ✗ FAIL: $1"; ((FAIL++)) || true; }
warn() { echo "  ⚠ WARN: $1"; ((WARN++)) || true; }

check_file_not_contains() {
  local file="$1" pattern="$2" label="$3"
  if [ ! -f "$file" ]; then
    warn "$label — file not found: $file (skip)"
    return
  fi
  if grep -qP "$pattern" "$file" 2>/dev/null; then
    fail "$label"
  else
    pass "$label"
  fi
}

check_file_contains() {
  local file="$1" pattern="$2" label="$3"
  if [ ! -f "$file" ]; then
    warn "$label — file not found: $file (skip)"
    return
  fi
  if grep -qP "$pattern" "$file" 2>/dev/null; then
    pass "$label"
  else
    fail "$label"
  fi
}

check_go_contains() {
  local file="$1" pattern="$2" label="$3"
  if [ ! -f "$file" ]; then
    fail "$label — Go file not found: $file"
    return
  fi
  if grep -qF "$pattern" "$file" 2>/dev/null; then
    pass "$label"
  else
    fail "$label"
  fi
}

echo "═══ architect-ai Hardening Verification v0.3 ═══"
echo ""
echo "── FASE 01: Enum + Atomicidad ──"

check_file_not_contains \
  "internal/assets/_shared/phase-dag-enforcement.md" \
  'status.*"running"' \
  "H-01A: 'running' removed from phase-dag-enforcement.md"

check_file_contains \
  "internal/assets/_shared/phase-dag-enforcement.md" \
  "in_progress" \
  "H-01B: 'in_progress' present in phase-dag-enforcement.md"

check_file_not_contains \
  "internal/assets/_shared/rollback-harness.md" \
  '"running\|failed"' \
  "H-01C: rollback-harness targets in_progress not running"

check_go_contains \
  "internal/components/openspec/state.go" \
  "O_CREATE|os.O_EXCL" \
  "H-02A: O_EXCL lockfile in state.go"

check_go_contains \
  "internal/components/openspec/state.go" \
  "os.Getpid()" \
  "H-02B: PID-unique .tmp path in state.go"

check_go_contains \
  "internal/components/openspec/state.go" \
  "infrastructure_blocked" \
  "H-02C: infrastructure_blocked in ValidStatuses"

echo ""
echo "── FASE 02: Verify Checks ──"

check_go_contains \
  "internal/verify/project_checks.go" \
  "claude-md-no-placeholders" \
  "H-04A: placeholder check function exists"

check_go_contains \
  "internal/verify/project_checks.go" \
  "sdd-state-enum" \
  "H-04B: enum check function exists"

check_go_contains \
  "internal/verify/fix_hints.go" \
  "FixBuild" \
  "H-04C: FixBuild constant defined"

echo ""
echo "── FASE 03: Metering ──"

check_go_contains \
  "internal/metering/session_stats.go" \
  "PhaseRecord" \
  "H-OBS01A: PhaseRecord struct defined"

check_go_contains \
  "internal/metering/session_stats.go" \
  "RecordPhaseStart" \
  "H-OBS01B: RecordPhaseStart method exists"

check_go_contains \
  "internal/metering/hook.go" \
  "PhaseStart" \
  "H-OBS01C: Hook.PhaseStart method exists"

echo ""
echo "── FASE 04: Gate Classifier ──"

check_go_contains \
  "internal/reasoning/gate/classify.go" \
  "d3 >= 2 && d4 >= 3" \
  "H-SM03A: joint critical D3+D4 case in posturePriority"

check_go_contains \
  "internal/reasoning/gate/validator.go" \
  "Forensic" \
  "H-SM03B: Forensic posture validated in ValidateDecision"

echo ""
echo "── FASE 05: TUI Progress ──"

check_go_contains \
  "internal/tui/model.go" \
  "Program *tea.Program" \
  "H-F09A: Program field in TUI Model"

check_go_contains \
  "internal/tui/model.go" \
  "prog.Send(StepProgressMsg" \
  "H-F09B: prog.Send() in startInstalling closure"

check_go_contains \
  "internal/tui/model.go" \
  "writeInstallLog" \
  "H-OBS05: writeInstallLog function exists"

echo ""
echo "── FASE 06: FilemergeJSON + Secrets ──"

check_go_contains \
  "internal/components/filemerge/json_merge.go" \
  "ErrBaseMalformed" \
  "H-F17A: ErrBaseMalformed defined"

check_go_contains \
  "internal/components/mcp/secrets.go" \
  "gitignoreCovers" \
  "H-F14A: gitignoreCovers function exists"

echo ""
echo "── FASE 07: Context Guardian ──"

check_file_contains \
  "internal/assets/skills/context-guardian/SKILL.md" \
  "Cooldown Rule" \
  "H-F13A: Cooldown Rule in context-guardian SKILL.md"

check_file_contains \
  "internal/assets/skills/context-guardian/SKILL.md" \
  "Dynamic Threshold" \
  "H-F13B: Dynamic Threshold in context-guardian SKILL.md"

check_file_contains \
  "internal/assets/skills/context-guardian/SKILL.md" \
  "pack-{compaction_count}" \
  "H-EPI02: versioned pack topic_key"

echo ""
echo "── FASE 08: Odoo Token Economy ──"

check_file_contains \
  "internal/assets/_shared/odoo-overlay-routing.md" \
  "Tier A" \
  "H-F12A: Tier A injection defined"

check_file_contains \
  "internal/assets/_shared/odoo-overlay-routing.md" \
  "Maximum limit: 50" \
  "H-F08A: pagination guard with 50 record limit"

check_file_contains \
  "internal/assets/_shared/odoo-overlay-routing.md" \
  "YOLO Mode Guard" \
  "H-F08B: YOLO Mode Guard defined"

echo ""
echo "── FASE 09: Gemini Mode A ──"

check_file_not_contains \
  "internal/assets/gemini/GEMINI.md" \
  "Mode A \(Gemini inline" \
  "H-F01A: Mode A removed from GEMINI.md"

check_file_contains \
  "internal/assets/gemini/GEMINI.md" \
  "IMMUTABILITY RULE" \
  "H-F01B: IMMUTABILITY RULE in GEMINI.md"

check_go_contains \
  "internal/architect/routing_rules.go" \
  "RoutingRulesMarkdown" \
  "H-EPI01: RoutingRulesMarkdown() function exists"

echo ""
echo "── FASE 10: Engram Keys ──"

check_go_contains \
  "internal/components/engram/engramkeys/keys.go" \
  "sha256.Sum256" \
  "H-F16A: SHA-256 hash in ResearchTopicKey"

check_go_contains \
  "internal/cli/diagnose_engram.go" \
  "RunDiagnoseEngram" \
  "H-OBS02: diagnose engram command exists"

echo ""
echo "── FASE 11: Circuit Breaker v0.3 ──"

check_file_contains \
  "internal/assets/_shared/circuit-breaker.md" \
  "infra_attempt_counts" \
  "H-F06A: infra_attempt_counts in circuit-breaker.md"

check_file_contains \
  "internal/assets/_shared/circuit-breaker.md" \
  "verify:abandoned" \
  "H-F07B: verify:abandoned forward arc documented"

check_go_contains \
  "internal/components/openspec/state.go" \
  "InfraAttemptCounts" \
  "H-F06B: InfraAttemptCounts in CircuitBreaker struct"

check_go_contains \
  "internal/components/openspec/state.go" \
  "ResetEvents" \
  "H-SM07: ResetEvents in CircuitBreaker struct"

echo ""
echo "── FASE 12: HybridWriter ──"

check_go_contains \
  "internal/components/openspec/hybrid_writer.go" \
  "HybridWritePartial" \
  "H-DBC01A: HybridWritePartial status defined"

check_go_contains \
  "internal/components/openspec/hybrid_writer.go" \
  "IsComplete" \
  "H-DBC01B: IsComplete() method defined"

check_file_not_contains \
  "internal/assets/_shared/persistence-contract.md" \
  "fall back to none mode silently" \
  "H-DBC01C: silent fallback removed from persistence-contract.md"

echo ""
echo "═══════════════════════════════════════════════"
echo "Results: ${PASS} passed  ${WARN} warnings  ${FAIL} failed"
echo "═══════════════════════════════════════════════"

if [ "${FAIL}" -gt 0 ]; then
  echo ""
  echo "Run: make hardening   to apply all patches"
  exit 1
fi

echo "All hardening checks passed ✓"
exit 0
```

**Comando para hacerlo ejecutable:**
```bash
mkdir -p .atl/scripts
chmod +x .atl/scripts/verify-hardening.sh
```

---

## Paso 4: Registrar `migrate-v03` como subcomando del CLI

**Archivo a modificar:** `cmd/architect-ai/main.go` (o el dispatcher equivalente)

**Acción:** Modificar — añadir el case en el switch de subcomandos

**Código a insertar** (dentro del `switch args[0]`):

```go
case "migrate-v03":
	projectRoot, _ := os.Getwd()
	atDir := filepath.Join(projectRoot, ".atl")
	migrated, err := openspec.MigrateV03(atDir)
	if err != nil {
		fmt.Fprintf(stderr, "migrate-v03: %v\n", err)
		return 1
	}
	if len(migrated) == 0 {
		fmt.Fprintln(stdout, "migrate-v03: no state files required migration (already up to date)")
	} else {
		fmt.Fprintf(stdout, "migrate-v03: migrated %d file(s):\n", len(migrated))
		for _, f := range migrated {
			fmt.Fprintf(stdout, "  %s\n", f)
		}
	}
	return 0
```

**Agregar imports necesarios en `main.go`:**
```bash
grep '"path/filepath"\|openspec"' cmd/architect-ai/main.go
```
Añadir `"github.com/rd-mg/architect-ai/internal/components/openspec"` si no está importado.

---

## Verificación de Fase (y del plan completo)

```bash
# 1. Crear el directorio de scripts si no existe
mkdir -p .atl/scripts scripts

# 2. Hacer ejecutables todos los scripts nuevos
chmod +x .atl/scripts/verify-hardening.sh
chmod +x .atl/hooks/pre_compact_hook.sh 2>/dev/null || true
chmod +x scripts/patch_gemini_mode_a.py

# 3. Ejecutar el hardening completo desde cero
make hardening 2>&1

# 4. Si make hardening pasa, ejecutar la verificación standalone
make hardening-verify 2>&1

# 5. Race detector en todos los paquetes críticos
make test-race 2>&1

# 6. Migrar proyectos existentes con estado obsoleto
make migrate-v03 2>&1

# 7. Test de humo del CLI completo
go run ./cmd/architect-ai check all 2>&1 | tail -20
go run ./cmd/architect-ai diagnose engram 2>&1
go run ./cmd/architect-ai migrate-v03 2>&1

# 8. eintegrate final (verifica E-12, E-13, y todos los checks previos)
go run ./cmd/eintegrate/... 2>&1 | grep -E "PASS|FAIL|E-[0-9]"

# 9. Resumen final
echo ""
echo "════════ HARDENING v0.3 COMPLETE ════════"
echo "Fases completadas: 1-13"
echo "Run 'make hardening-verify' en cada repo para confirmar."
```

