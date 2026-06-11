# Fase 4: Corrección del Adaptive Reasoning Gate — Estado Compuesto D3+D4

**Objetivo:** Resolver SM-03 (estado (D3>=2, D4>=3) manejado por la rama D4>=3 sola, eliminando +++Forensic bajo incidente de producción). Corregir `posturePriority` y `ValidateDecision` en `internal/reasoning/gate/` sin romper ningún test existente.

---

## Paso 1: Corregir `posturePriority` en `classify.go`

**Archivo a modificar:** `internal/reasoning/gate/classify.go`

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
// posturePriority selects the appropriate postures for Mode 3 based on dimension context.
func posturePriority(d1, d3, d4 int) []string {
	switch {
	case d4 >= 3:
		// Context saturated — stabilize
		return []string{"+++Pragmatic"}
	case d1 >= 3:
		// Architectural/paradigm change — full impact analysis
		return []string{"+++Systemic", "+++Adversarial"}
	case d3 >= 2:
		// Error pressure — forensic investigation
		return []string{"+++Forensic", "+++Pragmatic"}
	default:
		// Default Mode 3 — adversarial + systemic review
		return []string{"+++Adversarial", "+++Systemic"}
	}
}
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
// posturePriority selects the appropriate postures for Mode 3 based on dimension context.
//
// State machine invariant: (D3>=2 AND D4>=3) is a joint critical state that
// requires +++Forensic (investigate root cause) and +++Pragmatic (avoid aggravation).
// This case must be evaluated BEFORE the plain D4>=3 branch, which only uses
// +++Pragmatic and would suppress forensic investigation during a production incident.
func posturePriority(d1, d3, d4 int) []string {
	switch {
	case d3 >= 2 && d4 >= 3:
		// Joint critical: production incident (D3>=2) + context saturated (D4>=3).
		// +++Forensic investigates; +++Pragmatic prevents aggravating the incident
		// with unnecessary work under context pressure.
		return []string{"+++Forensic", "+++Pragmatic"}
	case d4 >= 3:
		// Context saturated only — stabilize output, defer heavy analysis.
		return []string{"+++Pragmatic"}
	case d1 >= 3:
		// Architectural/paradigm change — full impact analysis required.
		return []string{"+++Systemic", "+++Adversarial"}
	case d3 >= 2:
		// Production error pressure — forensic investigation primary.
		return []string{"+++Forensic", "+++Pragmatic"}
	default:
		// Default Mode 3 — adversarial + systemic review.
		return []string{"+++Adversarial", "+++Systemic"}
	}
}
```

---

## Paso 2: Extender `ValidateDecision` con los invariantes faltantes

**Archivo a modificar:** `internal/reasoning/gate/validator.go`

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
// ValidateDecision checks that the Mode and derived postures are consistent with D-scores.
// It only validates dimensions + mode consistency; posture validation is done by the
// orchestrator via Score() comparison.
func ValidateDecision(h *GateHeader) []string {
	var issues []string

	// D3 >= 2 must be Mode 3
	if h.D3 >= 2 && h.Mode < 3 {
		issues = append(issues, fmt.Sprintf("D3=%d requires Mode 3, got Mode %d", h.D3, h.Mode))
	}

	// Mode 3 requires D3 >= 2 or D4 >= 3
	if h.Mode == 3 && h.D3 < 2 && h.D4 < 3 {
		issues = append(issues, fmt.Sprintf("Mode 3 requires D3 >= 2 or D4 >= 3, got D3=%d, D4=%d", h.D3, h.D4))
	}

	return issues
}
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
// ValidateDecision checks that the Mode and derived postures are consistent with D-scores.
// It only validates dimensions + mode consistency; posture validation is done by the
// orchestrator via Score() comparison.
//
// Invariants enforced:
//   - D3 >= 2 → Mode must be 3
//   - Mode 3 → D3 >= 2 OR D4 >= 3
//   - Mode >= 2 → Rationale must be non-empty
//   - D3 >= 2 AND D4 >= 3 → posturePriority must include "+++Forensic"
//   - All D-scores must be 0–3 (enforced by headerPattern regex, documented here)
func ValidateDecision(h *GateHeader) []string {
	var issues []string

	// Invariant: D3 >= 2 must be Mode 3
	if h.D3 >= 2 && h.Mode < 3 {
		issues = append(issues, fmt.Sprintf("D3=%d requires Mode 3, got Mode %d", h.D3, h.Mode))
	}

	// Invariant: Mode 3 requires D3 >= 2 or D4 >= 3
	if h.Mode == 3 && h.D3 < 2 && h.D4 < 3 {
		issues = append(issues, fmt.Sprintf(
			"Mode 3 requires D3 >= 2 or D4 >= 3, got D3=%d, D4=%d", h.D3, h.D4))
	}

	// Invariant: Mode >= 2 should have non-empty rationale
	if h.Mode >= 2 && strings.TrimSpace(h.Rationale) == "" {
		issues = append(issues, fmt.Sprintf(
			"Mode %d requires a non-empty rationale after the bracket — found none", h.Mode))
	}

	// Invariant: joint critical state (D3>=2, D4>=3) must produce +++Forensic posture.
	// We validate by calling Score() and checking the returned postures.
	if h.D3 >= 2 && h.D4 >= 3 {
		_, postures := Score([5]int{h.D1, h.D2, h.D3, h.D4, 0})
		hasForensic := false
		for _, p := range postures {
			if p == "+++Forensic" {
				hasForensic = true
				break
			}
		}
		if !hasForensic {
			issues = append(issues, fmt.Sprintf(
				"D3=%d AND D4=%d (joint critical) requires +++Forensic posture, got %v",
				h.D3, h.D4, postures))
		}
	}

	return issues
}
```

---

## Paso 3: Agregar tests para el nuevo caso de estado conjunto

**Archivo a modificar:** `internal/reasoning/gate/classify_test.go`

**Acción:** Modificar — añadir al final del archivo (antes del último `}` de cierre del paquete si existe, o como nuevas funciones top-level)

**Comando previo:**
```bash
# Verificar contenido actual del archivo de test
tail -30 internal/reasoning/gate/classify_test.go
```

**Código a añadir al final de `classify_test.go`:**
```go
func TestPosturePriority_JointCritical_D3AndD4(t *testing.T) {
	// D3=2, D4=3: joint critical — must return +++Forensic, not just +++Pragmatic
	mode, postures := Score([5]int{0, 0, 2, 3, 0})
	if mode != 3 {
		t.Errorf("D3=2, D4=3: expected Mode 3, got %d", mode)
	}
	hasForensic := false
	hasPragmatic := false
	for _, p := range postures {
		if p == "+++Forensic" {
			hasForensic = true
		}
		if p == "+++Pragmatic" {
			hasPragmatic = true
		}
	}
	if !hasForensic {
		t.Errorf("D3=2, D4=3: expected +++Forensic in postures, got %v", postures)
	}
	if !hasPragmatic {
		t.Errorf("D3=2, D4=3: expected +++Pragmatic in postures, got %v", postures)
	}
}

func TestPosturePriority_D4Only_NoPressure(t *testing.T) {
	// D4=3, D3=0: context saturated only — must be +++Pragmatic only (no +++Forensic)
	mode, postures := Score([5]int{0, 0, 0, 3, 0})
	if mode != 3 {
		t.Errorf("D4=3, D3=0: expected Mode 3, got %d", mode)
	}
	if len(postures) != 1 || postures[0] != "+++Pragmatic" {
		t.Errorf("D4=3, D3=0: expected [+++Pragmatic], got %v", postures)
	}
}

func TestPosturePriority_D3Only_NoContextSaturation(t *testing.T) {
	// D3=2, D4=0: production incident, no context pressure — +++Forensic + +++Pragmatic
	mode, postures := Score([5]int{0, 0, 2, 0, 0})
	if mode != 3 {
		t.Errorf("D3=2, D4=0: expected Mode 3, got %d", mode)
	}
	hasForensic := false
	for _, p := range postures {
		if p == "+++Forensic" {
			hasForensic = true
		}
	}
	if !hasForensic {
		t.Errorf("D3=2, D4=0: expected +++Forensic, got %v", postures)
	}
}

func TestValidateDecision_JointCritical_RequiresForensic(t *testing.T) {
	h := &GateHeader{Mode: 3, D1: 0, D2: 0, D3: 2, D4: 3, Rationale: "production incident and context saturated"}
	issues := ValidateDecision(h)
	if len(issues) != 0 {
		t.Errorf("D3=2,D4=3 Mode 3 with rationale: expected no issues, got %v", issues)
	}
}

func TestValidateDecision_Mode2_EmptyRationale_Flagged(t *testing.T) {
	h := &GateHeader{Mode: 2, D1: 2, D2: 1, D3: 0, D4: 1, Rationale: ""}
	issues := ValidateDecision(h)
	found := false
	for _, iss := range issues {
		if strings.Contains(iss, "rationale") {
			found = true
		}
	}
	if !found {
		t.Errorf("Mode 2 with empty rationale should flag a rationale issue, got %v", issues)
	}
}

func TestValidateDecision_Mode1_EmptyRationale_OK(t *testing.T) {
	// Mode 1 does not require rationale
	h := &GateHeader{Mode: 1, D1: 0, D2: 0, D3: 0, D4: 0, Rationale: ""}
	issues := ValidateDecision(h)
	for _, iss := range issues {
		if strings.Contains(iss, "rationale") {
			t.Errorf("Mode 1 should not require rationale, got issue: %s", iss)
		}
	}
}
```

**Agregar el import de `"strings"` en el test file si no está presente:**

```bash
grep -n '"strings"' internal/reasoning/gate/classify_test.go
```

Si no está, añadirlo al bloque de imports del archivo de test.

---

## Verificación de Fase

```bash
# 1. Compilar el paquete gate
go build ./internal/reasoning/gate/...

# 2. Todos los tests existentes deben pasar
go test ./internal/reasoning/gate/... -v -count=1

# 3. Tests nuevos de estado compuesto
go test ./internal/reasoning/gate/... -v -count=1 -run TestPosturePriority_Joint
go test ./internal/reasoning/gate/... -v -count=1 -run TestValidateDecision_Joint
go test ./internal/reasoning/gate/... -v -count=1 -run TestValidateDecision_Mode2_Empty

# 4. Verificar que el caso (D3=2, D4=3) ya no devuelve solo +++Pragmatic
go test ./internal/reasoning/gate/... -v -run TestPosturePriority_D4Only

# 5. Compilar todo (asegurar compatibilidad de callers)
go build ./...

# 6. Vet
go vet ./internal/reasoning/gate/...
```

