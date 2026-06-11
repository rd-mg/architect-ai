# Fase 11: Circuit Breaker — Separación Domain vs Infrastructure + Reset Protocol

**Objetivo:** Resolver F-06 (GATE_ERROR contamina el CB de dominio), F-07 (verify:abandoned sin arco de salida), y SM-07 (reseteos manuales sin justificación ni límite). Modifica el prompt `circuit-breaker.md`, el `sdd-orchestrator.md`, y el validador de resultado en Go.

---

## Paso 1: Reemplazar el contenido completo de `circuit-breaker.md`

**Archivo a modificar:** `internal/assets/_shared/circuit-breaker.md`

**Acción:** Modificar (reemplazar contenido completo del archivo)

```bash
cat > internal/assets/_shared/circuit-breaker.md << 'MDEOF'
# Circuit Breaker Protocol v0.3

Prevents recursive loops (Ralph Loops) where an agent repeatedly fails a phase
and retries indefinitely. v0.3 introduces separate counters for domain failures
vs infrastructure failures, a per-phase max_attempts override, and a reset audit trail.

## Error Classification (MANDATORY — must classify BEFORE incrementing any counter)

Every `blocked` or `failed` result contract MUST include `error_type`:

| `error_type`       | Definition | Counter incremented |
|--------------------|------------|---------------------|
| `domain`           | Spec/logic/code error requiring design fix or user input | `attempt_counts[phase]` |
| `infrastructure`   | MCP down, Engram unavailable, gate template not injected, skill file truncated | `infra_attempt_counts[phase]` |
| `none`             | Successful result | neither |

**Infrastructure errors NEVER increment `attempt_counts[phase]`.**
They use a separate `infra_attempt_counts[phase]` counter with its own limit (default: 5).

## Counters and Limits

```yaml
circuit_breaker:
  enabled: true
  max_attempts: 3          # domain errors limit (default, overridable per phase)
  attempt_counts:          # domain error counter per phase
    sdd-spec: 0
    sdd-apply: 0
  infra_attempt_counts:    # infrastructure error counter per phase (NEW v0.3)
    sdd-spec: 0
    sdd-apply: 0
  reset_events: []         # audit trail of manual resets (NEW v0.3)
  abandoned_phases: []
```

## Per-Phase max_attempts Override

Default: `max_attempts: 3` for all phases.

**Overrides injected by orchestrator in sub-agent launch prompt header:**

```
## Circuit Breaker: max_attempts={N} infra_max=5
```

| Phase | Condition | max_attempts |
|-------|-----------|--------------|
| sdd-verify | `tdd_mode: true` | 7 |
| sdd-apply  | `len(tasks) > 10` | 5 |
| sdd-explore | always | 2 |
| all others | default | 3 |

## Escalation — Domain Errors

- **Attempt 1 failed (domain)**: Update approach, increment to 2, retry.
- **Attempt 2 failed (domain)**: Request user context, increment to 3, retry.
- **Attempt 3 failed (domain)**: Circuit Breaker trips:
  - Write `status: abandoned` for phase in `.atl/sdd-state.yaml`
  - Add phase to `circuit_breaker.abandoned_phases`
  - Emit Result Contract: `status: abandoned`, `error_type: domain`
  - Exit Code 2

## Escalation — Infrastructure Errors

- **Attempts 1-4 (infrastructure)**: Retry automatically WITHOUT user prompt.
  - Increment `infra_attempt_counts[phase]` only.
  - Emit Result Contract: `status: infrastructure_blocked`, `error_type: infrastructure`
  - Orchestrator retries after brief pause (do not advance CB domain counter).
- **Attempt 5 (infrastructure)**: Infrastructure CB trips:
  - Escalate to user: "[INFRA BLOCKED] {phase} failed {N} times due to infrastructure issues."
  - Emit diagnosis hint: "Run: architect-ai diagnose engram"
  - Halt phase. Do NOT advance to next phase.

## Orchestrator on Domain CB Trip (Exit Code 2)

1. Intercept Exit Code 2.
2. Save logs and state to `.atl/openspec/changes/{change}/cb-diagnostic.yaml`.
3. Emit CB Deadlock Menu to user:

```
[CIRCUIT BREAKER] Phase {phase} abandoned after {N} domain failures.

Options:
  [1] /sdd-hotfix        — targeted fix (≤ 3 files, scope: {phase} artifacts only)
  [2] /sdd-design --amend — go back to design phase (for structural issues)
  [3] /sdd-archive --status=abandoned — close change as abandoned
  [4] /sdd-ff {phase} --force — skip phase, document exception, continue pipeline
  [5] /sdd-reset {phase} "{justification}" — reset CB counter with audit trail
```

**Phase `verify:abandoned` — Forward Arc (previously missing):**

When `sdd-verify` is abandoned:
- Option [1] `/sdd-hotfix` — patch the specific failing assertion (scope: test files only)
- Option [4] `/sdd-ff verify --force` — skip verify, emit CONDITIONALLY_APPROVED with documented exception
- Option [2] `/sdd-design --amend` — redesign if the issue is structural

The archive phase (`sdd-archive`) accepts `verify:abandoned` when invoked via Option [3]:
`/sdd-archive --status=abandoned` bypasses the `verify:completed` prerequisite
and records the abandonment in the deviation log.

## Reset Protocol (NEW v0.3)

When user invokes `/sdd-reset {phase} "{justification}"` or `/sdd-continue` with context:

1. **Justification required**: non-empty string describing why the failure was resolved.
   - If empty or omitted: REFUSE reset, prompt: "Provide justification: /sdd-reset {phase} 'fixed X'"
2. **Reset action**:
   - Set `attempt_counts[phase] = 0`
   - Preserve `infra_attempt_counts[phase]` (infrastructure issues may be unresolved)
   - Remove phase from `abandoned_phases`
   - Set phase `status: pending`
   - Append to `reset_events`:
     ```yaml
     - phase: {phase}
       timestamp: {UTC ISO-8601}
       justification: "{user-provided text}"
       reset_by: "manual"
       prior_attempt_count: {N}
     ```
3. **Reset limit**: maximum 3 manual resets per phase per change.
   - After 3 resets: REFUSE further resets, emit:
     "[CB RESET LIMIT] {phase} has been reset {N} times. Run /sdd-design --amend to redesign before retrying."
MDEOF
echo "circuit-breaker.md written"
```

---

## Paso 2: Añadir `infra_attempt_counts` y `reset_events` al `InitialState` YAML

**Archivo a modificar:** `internal/sdd/state/writer.go`

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
func InitialState(changeName, project, artifactStore, executionMode, deliveryStrategy string) string {
```

**Localizar el bloque completo con:**
```bash
grep -n "func InitialState\|circuit_breaker:\|attempt_counts:\|abandoned_phases:" \
  internal/sdd/state/writer.go | head -20
```

**Código a reemplazar — BUSCAR EXACTAMENTE** (el bloque YAML del circuit_breaker dentro de InitialState):
```go
circuit_breaker:
  enabled: true
  max_attempts: 3
  attempt_counts: {}
  abandoned_phases: []
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
circuit_breaker:
  enabled: true
  max_attempts: 3
  attempt_counts: {}
  infra_attempt_counts: {}
  reset_events: []
  abandoned_phases: []
```

---

## Paso 3: Añadir `InfraAttemptCounts` y `ResetEvents` a las structs Go

**Archivo a modificar:** `internal/components/openspec/state.go`

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
type CircuitBreaker struct {
	Enabled         bool           `yaml:"enabled"`
	MaxAttempts     int            `yaml:"max_attempts"`
	AttemptCounts   map[string]int `yaml:"attempt_counts"`
	AbandonedPhases []string       `yaml:"abandoned_phases"`
}
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
// ResetEvent records a manual circuit-breaker reset for audit purposes.
// Preserved across resets so the full history is available for retrospectives.
type ResetEvent struct {
	Phase             string `yaml:"phase"`
	Timestamp         string `yaml:"timestamp"`
	Justification     string `yaml:"justification"`
	ResetBy           string `yaml:"reset_by"`
	PriorAttemptCount int    `yaml:"prior_attempt_count"`
}

// CircuitBreaker tracks attempt counts and state for the CB protocol.
// InfraAttemptCounts is separate from AttemptCounts so infrastructure
// failures do not consume domain retry budget.
type CircuitBreaker struct {
	Enabled              bool           `yaml:"enabled"`
	MaxAttempts          int            `yaml:"max_attempts"`
	AttemptCounts        map[string]int `yaml:"attempt_counts"`
	InfraAttemptCounts   map[string]int `yaml:"infra_attempt_counts"`
	ResetEvents          []ResetEvent   `yaml:"reset_events"`
	AbandonedPhases      []string       `yaml:"abandoned_phases"`
}
```

---

## Paso 4: Añadir `RecordInfraAttempt` y `ResetPhase` al state package

**Archivo a modificar:** `internal/components/openspec/state.go`

**Acción:** Modificar — añadir las funciones a continuación de las funciones existentes de `CircuitBreaker`

**Código a insertar DESPUÉS de la definición del struct `CircuitBreaker`:**

```go
// RecordDomainAttempt increments the domain attempt counter for phase.
// Returns tripped=true and sets phase status to "abandoned" when
// attempt_counts[phase] reaches cb.MaxAttempts.
//
// Precondition: s.CircuitBreaker.Enabled == true
// Postcondition: if tripped, phase.Status == "abandoned" and phase is in AbandonedPhases
func (s *State) RecordDomainAttempt(phase string) (tripped bool) {
	if s.CircuitBreaker.AttemptCounts == nil {
		s.CircuitBreaker.AttemptCounts = make(map[string]int)
	}
	s.CircuitBreaker.AttemptCounts[phase]++

	maxAttempts := s.CircuitBreaker.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	if s.CircuitBreaker.Enabled && s.CircuitBreaker.AttemptCounts[phase] >= maxAttempts {
		if ph, ok := s.Phases[phase]; ok {
			ph.Status = "abandoned"
			s.Phases[phase] = ph
		}
		s.CircuitBreaker.AbandonedPhases = append(s.CircuitBreaker.AbandonedPhases, phase)
		return true
	}
	return false
}

// RecordInfraAttempt increments the infrastructure attempt counter for phase.
// Returns infraTripped=true when infra_attempt_counts[phase] reaches infraMax (default 5).
// Domain attempt_counts[phase] is NEVER modified by this function.
//
// Precondition: error_type == "infrastructure"
// Postcondition: infra_attempt_counts[phase]++ only; attempt_counts unchanged
func (s *State) RecordInfraAttempt(phase string, infraMax int) (infraTripped bool) {
	if s.CircuitBreaker.InfraAttemptCounts == nil {
		s.CircuitBreaker.InfraAttemptCounts = make(map[string]int)
	}
	if infraMax <= 0 {
		infraMax = 5
	}
	s.CircuitBreaker.InfraAttemptCounts[phase]++
	return s.CircuitBreaker.InfraAttemptCounts[phase] >= infraMax
}

// ResetPhase resets the domain CB counter for phase with a mandatory audit trail.
//
// Preconditions:
//   - justification must be non-empty
//   - prior manual resets for phase must be < 3
//
// Postconditions:
//   - attempt_counts[phase] == 0
//   - phase.Status == "pending"
//   - reset_events has one new entry with the justification
//   - infra_attempt_counts[phase] is preserved (not reset)
func (s *State) ResetPhase(phase, justification, resetBy string) error {
	if strings.TrimSpace(justification) == "" {
		return fmt.Errorf(
			"reset requires a non-empty justification describing why the failure was resolved")
	}

	priorResets := 0
	for _, ev := range s.CircuitBreaker.ResetEvents {
		if ev.Phase == phase {
			priorResets++
		}
	}
	if priorResets >= 3 {
		return fmt.Errorf(
			"phase %q has been manually reset %d times — "+
				"run /sdd-design --amend to redesign before retrying",
			phase, priorResets,
		)
	}

	prior := s.CircuitBreaker.AttemptCounts[phase]
	s.CircuitBreaker.ResetEvents = append(s.CircuitBreaker.ResetEvents, ResetEvent{
		Phase:             phase,
		Timestamp:         time.Now().UTC().Format(time.RFC3339),
		Justification:     strings.TrimSpace(justification),
		ResetBy:           resetBy,
		PriorAttemptCount: prior,
	})

	if s.CircuitBreaker.AttemptCounts == nil {
		s.CircuitBreaker.AttemptCounts = make(map[string]int)
	}
	s.CircuitBreaker.AttemptCounts[phase] = 0

	var newAbandoned []string
	for _, p := range s.CircuitBreaker.AbandonedPhases {
		if p != phase {
			newAbandoned = append(newAbandoned, p)
		}
	}
	s.CircuitBreaker.AbandonedPhases = newAbandoned

	if ph, ok := s.Phases[phase]; ok {
		ph.Status = "pending"
		s.Phases[phase] = ph
	}
	return nil
}
```

**Agregar import de `"strings"` si no está en `state.go`:**
```bash
grep '"strings"' internal/components/openspec/state.go
```

---

## Paso 5: Crear tests del Circuit Breaker v0.3

**Archivo a crear:** `internal/components/openspec/circuit_breaker_test.go`

**Acción:** Crear

```go
package openspec

import (
	"testing"
	"time"
)

func makeTestState(phase string) *State {
	now := time.Now().UTC()
	return &State{
		SchemaVersion: SchemaVersion,
		ChangeName:    "test-change",
		CreatedAt:     now,
		UpdatedAt:     now,
		ArtifactStore: "openspec",
		CircuitBreaker: CircuitBreaker{
			Enabled:     true,
			MaxAttempts: 3,
		},
		Phases: map[string]*Phase{
			phase: {Status: "in_progress"},
		},
	}
}

func TestRecordDomainAttempt_TripsAtMax(t *testing.T) {
	s := makeTestState("sdd-verify")
	s.CircuitBreaker.MaxAttempts = 3

	if tripped := s.RecordDomainAttempt("sdd-verify"); tripped {
		t.Error("attempt 1 should not trip CB")
	}
	if tripped := s.RecordDomainAttempt("sdd-verify"); tripped {
		t.Error("attempt 2 should not trip CB")
	}
	if tripped := s.RecordDomainAttempt("sdd-verify"); !tripped {
		t.Error("attempt 3 should trip CB (max=3)")
	}

	ph := s.Phases["sdd-verify"]
	if ph.Status != "abandoned" {
		t.Errorf("phase should be abandoned after CB trip, got %q", ph.Status)
	}
	if len(s.CircuitBreaker.AbandonedPhases) != 1 || s.CircuitBreaker.AbandonedPhases[0] != "sdd-verify" {
		t.Errorf("sdd-verify should be in abandoned_phases: %v", s.CircuitBreaker.AbandonedPhases)
	}
}

func TestRecordInfraAttempt_DoesNotIncrementDomainCounter(t *testing.T) {
	s := makeTestState("sdd-spec")

	for i := 0; i < 4; i++ {
		infraTripped := s.RecordInfraAttempt("sdd-spec", 5)
		if infraTripped {
			t.Errorf("infra attempt %d should not trip (max=5)", i+1)
		}
	}

	if s.CircuitBreaker.AttemptCounts["sdd-spec"] != 0 {
		t.Errorf("domain attempt_counts should remain 0 after infra attempts, got %d",
			s.CircuitBreaker.AttemptCounts["sdd-spec"])
	}
	if s.CircuitBreaker.InfraAttemptCounts["sdd-spec"] != 4 {
		t.Errorf("infra_attempt_counts should be 4, got %d",
			s.CircuitBreaker.InfraAttemptCounts["sdd-spec"])
	}
}

func TestRecordInfraAttempt_TripsAtInfraMax(t *testing.T) {
	s := makeTestState("sdd-apply")

	var lastTripped bool
	for i := 0; i < 5; i++ {
		lastTripped = s.RecordInfraAttempt("sdd-apply", 5)
	}
	if !lastTripped {
		t.Error("5th infra attempt should trip infrastructure CB")
	}
	// Domain counter must still be 0
	if s.CircuitBreaker.AttemptCounts["sdd-apply"] != 0 {
		t.Error("domain CB should not be affected by infra trips")
	}
}

func TestResetPhase_RequiresJustification(t *testing.T) {
	s := makeTestState("sdd-verify")
	s.CircuitBreaker.AttemptCounts = map[string]int{"sdd-verify": 3}
	s.CircuitBreaker.AbandonedPhases = []string{"sdd-verify"}
	s.Phases["sdd-verify"] = &Phase{Status: "abandoned"}

	if err := s.ResetPhase("sdd-verify", "", "manual"); err == nil {
		t.Error("reset with empty justification should fail")
	}
}

func TestResetPhase_Success(t *testing.T) {
	s := makeTestState("sdd-verify")
	s.CircuitBreaker.AttemptCounts = map[string]int{"sdd-verify": 3}
	s.CircuitBreaker.AbandonedPhases = []string{"sdd-verify"}
	s.Phases["sdd-verify"] = &Phase{Status: "abandoned"}

	if err := s.ResetPhase("sdd-verify", "Fixed nil pointer in auth.go:127", "manual"); err != nil {
		t.Fatalf("reset with justification should succeed: %v", err)
	}

	if s.CircuitBreaker.AttemptCounts["sdd-verify"] != 0 {
		t.Error("attempt_count should be 0 after reset")
	}
	if len(s.CircuitBreaker.AbandonedPhases) != 0 {
		t.Error("sdd-verify should be removed from abandoned_phases after reset")
	}
	if s.Phases["sdd-verify"].Status != "pending" {
		t.Errorf("phase status should be pending after reset, got %q",
			s.Phases["sdd-verify"].Status)
	}
	if len(s.CircuitBreaker.ResetEvents) != 1 {
		t.Errorf("expected 1 reset event, got %d", len(s.CircuitBreaker.ResetEvents))
	}
	if s.CircuitBreaker.ResetEvents[0].PriorAttemptCount != 3 {
		t.Errorf("prior_attempt_count should be 3, got %d",
			s.CircuitBreaker.ResetEvents[0].PriorAttemptCount)
	}
}

func TestResetPhase_MaxResetsEnforced(t *testing.T) {
	s := makeTestState("sdd-verify")
	s.CircuitBreaker.AttemptCounts = map[string]int{"sdd-verify": 0}
	s.CircuitBreaker.ResetEvents = []ResetEvent{
		{Phase: "sdd-verify", Justification: "fix 1"},
		{Phase: "sdd-verify", Justification: "fix 2"},
		{Phase: "sdd-verify", Justification: "fix 3"},
	}

	if err := s.ResetPhase("sdd-verify", "fix 4", "manual"); err == nil {
		t.Error("4th reset should be refused — max 3 manual resets per phase")
	}
}

func TestResetPhase_PreservesInfraCounter(t *testing.T) {
	s := makeTestState("sdd-apply")
	s.CircuitBreaker.AttemptCounts = map[string]int{"sdd-apply": 2}
	s.CircuitBreaker.InfraAttemptCounts = map[string]int{"sdd-apply": 3}

	if err := s.ResetPhase("sdd-apply", "fixed MCP timeout", "manual"); err != nil {
		t.Fatalf("reset should succeed: %v", err)
	}

	if s.CircuitBreaker.InfraAttemptCounts["sdd-apply"] != 3 {
		t.Errorf("infra_attempt_counts should be preserved after domain reset, got %d",
			s.CircuitBreaker.InfraAttemptCounts["sdd-apply"])
	}
}

func TestCircuitBreaker_YAMLRoundtrip(t *testing.T) {
	dir := t.TempDir()
	s := makeTestState("sdd-spec")
	s.CircuitBreaker.InfraAttemptCounts = map[string]int{"sdd-spec": 2}
	s.CircuitBreaker.ResetEvents = []ResetEvent{
		{
			Phase:             "sdd-spec",
			Timestamp:         time.Now().UTC().Format(time.RFC3339),
			Justification:     "fixed engram timeout",
			ResetBy:           "manual",
			PriorAttemptCount: 2,
		},
	}

	import_path := dir + "/test-change/state.yaml"
	if mkErr := mkdirFor(import_path); mkErr != nil {
		t.Fatalf("mkdir: %v", mkErr)
	}
	if err := Save(import_path, s); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(import_path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.CircuitBreaker.InfraAttemptCounts["sdd-spec"] != 2 {
		t.Errorf("infra_attempt_counts not preserved in YAML roundtrip: %v",
			loaded.CircuitBreaker.InfraAttemptCounts)
	}
	if len(loaded.CircuitBreaker.ResetEvents) != 1 {
		t.Errorf("reset_events not preserved in YAML roundtrip: %v",
			loaded.CircuitBreaker.ResetEvents)
	}
}

// mkdirFor creates all directories needed for the given file path.
func mkdirFor(path string) error {
	import_os := func() error {
		return nil
	}
	_ = import_os
	dir := path[:len(path)-len("/state.yaml")]
	return os.MkdirAll(dir, 0o755)
}
```

**Añadir import de `"os"` al test file:**

```go
package openspec

import (
	"os"
	"testing"
	"time"
)
```

---

## Verificación de Fase

```bash
# 1. Compilar el paquete openspec
go build ./internal/components/openspec/...

# 2. Tests existentes (sin regresiones)
go test ./internal/components/openspec/... -v -count=1

# 3. Tests nuevos del CB v0.3
go test ./internal/components/openspec/... -v -count=1 -run TestRecordDomainAttempt
go test ./internal/components/openspec/... -v -count=1 -run TestRecordInfraAttempt
go test ./internal/components/openspec/... -v -count=1 -run TestResetPhase
go test ./internal/components/openspec/... -v -count=1 -run TestCircuitBreaker_YAML

# 4. Race detector
go test -race ./internal/components/openspec/... -count=1

# 5. Compilar sdd/state para verificar InitialState actualizado
go build ./internal/sdd/state/...

# 6. Verificar que circuit-breaker.md tiene las secciones nuevas
python3 -c "
with open('internal/assets/_shared/circuit-breaker.md') as f:
    content = f.read()
required = [
    'infra_attempt_counts',
    'InfraAttemptCounts',
    'reset_events',
    'infrastructure',
    'verify:abandoned',
    'Forward Arc',
    'Reset Protocol',
    'justification',
]
for r in required:
    assert r in content, f'MISSING: {r}'
print('circuit-breaker.md validated')
"

# 7. Verificar que InitialState YAML incluye los nuevos campos
grep -n "infra_attempt_counts\|reset_events" internal/sdd/state/writer.go

# 8. Compilar todo
go build ./...

# 9. Vet
go vet ./internal/components/openspec/... ./internal/sdd/state/...
```

