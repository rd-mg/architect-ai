# Fase 14: Correcciones y Complementos — Gaps de Compilación

**Objetivo:** Resolver todos los gaps de compilación detectados en las fases anteriores: `TotalTokens()` faltante en `SessionStats`, `mkdirFor` helper inválido en tests del CB, `strings` import faltante en `state.go`, y `Load()` de `openspec` no definida explícitamente. Esta fase es prerequisito de la verificación final. No introduce lógica nueva — solo garantiza que el código sea compilable.

---

## Paso 1: Añadir `TotalTokens()` a `SessionStats`

El método `TotalTokens()` es referenciado en `fase_03_metering_phase_breakdown.md` (tests) pero no existe en `session_stats.go`. Se añade aquí.

**Archivo a modificar:** `internal/metering/session_stats.go`

**Acción:** Modificar — añadir el método al final del archivo, antes del último bloque de funciones de rendering

**Código a añadir** (insertar después de `EngramPhaseCostContent` y antes de `humanizeTokens` o similar):

```go
// TotalTokens returns the sum of all prompt and completion tokens
// recorded in this session, across all models and phases.
func (s *SessionStats) TotalTokens() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.PromptTokens + s.CompletionTokens
}
```

---

## Paso 2: Añadir `strings` import en `session_stats.go`

**Archivo a modificar:** `internal/metering/session_stats.go`

**Acción:** Modificar — añadir `"strings"` al bloque de imports si no está presente

**Comando de verificación:**
```bash
grep '"strings"' internal/metering/session_stats.go
```

Si no está, añadir al bloque de imports:

```go
import (
	"fmt"
	"strings"
	"sync"
	"time"
)
```

---

## Paso 3: Corregir `mkdirFor` en `circuit_breaker_test.go`

La función `mkdirFor` en `fase_11` usa un cierre que llama a `os.MkdirAll` pero importa `os` de forma incorrecta. La versión correcta y directa:

**Archivo a modificar:** `internal/components/openspec/circuit_breaker_test.go`

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
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

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
// mkdirFor creates all parent directories for the given file path.
func mkdirFor(path string) error {
	dir := path[:len(path)-len("/state.yaml")]
	return os.MkdirAll(dir, 0o755)
}
```

---

## Paso 4: Verificar que `Load()` existe en `openspec` package

`fase_11` (TestCircuitBreaker_YAMLRoundtrip) y `fase_01` usan `Load(path)`. Verificar:

```bash
grep -n "^func Load(" internal/components/openspec/state.go
```

Si `Load` no existe con esa firma exacta, añadirla:

**Archivo a modificar:** `internal/components/openspec/state.go`

**Acción:** Modificar — añadir la función si no existe

```go
// Load reads and parses a state.yaml file.
//
// Precondition:  path points to a valid YAML file.
// Postcondition: returned *State is non-nil and validated.
func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var s State
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &s, nil
}
```

---

## Paso 5: Verificar que `SchemaVersion` está definido en `openspec`

Los tests de `fase_01` usan `openspec.SchemaVersion`. Verificar:

```bash
grep -n "SchemaVersion" internal/components/openspec/state.go | head -5
```

Si no existe, añadir:

**Archivo a modificar:** `internal/components/openspec/state.go`

**Acción:** Modificar — añadir constante al inicio del archivo después del `package` y antes de los `import`s

```go
// SchemaVersion is the canonical version string for sdd-state.yaml.
// Increment this when making breaking changes to the State struct.
const SchemaVersion = "3.0"
```

---

## Paso 6: Verificar que el import `"strings"` está en `state.go` para `ResetPhase`

`ResetPhase` usa `strings.TrimSpace`. Verificar:

```bash
grep '"strings"' internal/components/openspec/state.go
```

Si no está, añadir al bloque de imports:

**Código a reemplazar — BUSCAR EXACTAMENTE** (el bloque de imports existente):
```go
import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)
```

---

## Paso 7: Verificar que `pipeline.StepStatusFailed` y `pipeline.StepStatusSucceeded` existen

`fase_05` usa `pipeline.StepStatusFailed` y `pipeline.StepStatusSucceeded`. Verificar:

```bash
grep -rn "StepStatusFailed\|StepStatusSucceeded\|StepStatus " \
  internal/pipeline/ | head -10
```

Si los nombres difieren (ej: `StepStatusError` en lugar de `StepStatusFailed`), ajustar en `model.go`:

**Archivo a modificar:** `internal/tui/model.go`

**Código a reemplazar** (si el nombre real del status de fallo es diferente):
```go
if step.Status == pipeline.StepStatusFailed {
```

**Código de reemplazo** (con el nombre correcto del enum — usar el que exista en el paquete):
```go
// Usar el valor exacto del enum que exista en internal/pipeline/
// Comando para encontrarlo:
// grep -n "StepStatus\|Failed\|Error" internal/pipeline/*.go | head -20
if step.Status == pipeline.StepStatusFailed {  // ajustar si necesario
```

**Comando de localización exacta:**
```bash
grep -rn "type StepStatus\|StepStatus[A-Z]" internal/pipeline/ | head -10
```

---

## Paso 8: Verificar que `tea.Program` no requiere puntero vs valor en `SetProgram`

`fase_05` añade `SetProgram(p *tea.Program)` a `Model`. Si `Model` es valor (no puntero) en BubbleTea, el método con receptor puntero no funciona desde el `Init()`. Verificar:

```bash
grep -n "type Model struct\|func.*Model.*Init\|func.*Model.*Update\|func.*Model.*View" \
  internal/tui/model.go | head -10
```

Si el método debe ser en value receiver (BubbleTea estándar), la solución alternativa es capturar `p` en el closure directamente desde `main.go`:

**Archivo a modificar:** `internal/tui/model.go`

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
// SetProgram stores a reference to the running tea.Program so that goroutines
// launched by startInstalling and startGeneration can send messages back.
// Must be called after NewModel() and before p.Run().
func (m *Model) SetProgram(p *tea.Program) {
	m.Program = p
}
```

**Código de reemplazo — forma que funciona con value receiver de BubbleTea:**
```go
// SetProgram stores a reference to the running tea.Program so goroutines
// (pipeline, generation) can send StepProgressMsg back to the Update loop.
// IMPORTANT: Because BubbleTea's Update/View use value receivers, Program is
// set on the Model value before passing to tea.NewProgram. The closure in
// startInstalling captures 'm.Program' at call time — this is safe because
// Program is set once before p.Run() and never mutated.
// Call sequence: m := NewModel(...) → m.Program = p → p = tea.NewProgram(m, ...) → p.Run()
func (m *Model) SetProgram(p *tea.Program) {
	m.Program = p
}
```

**Nota de implementación en `main.go`:**

El patrón correcto para BubbleTea cuando `NewModel` retorna un valor (no puntero):

```go
// main.go — patrón correcto para SetProgram con value-receiver model
m := tui.NewModel(detection, version)
// Necesitamos que m.Program apunte al program ANTES de pasarlo a NewProgram.
// Solución: usar una variable capturada por cierre en vez de SetProgram.
// Ver: https://github.com/charmbracelet/bubbletea/discussions/176

// OPCIÓN A (si NewModel retorna *Model — puntero):
p := tea.NewProgram(m, tea.WithAltScreen())
m.SetProgram(p)

// OPCIÓN B (si NewModel retorna Model — valor):
// El campo Program en Model no puede apuntar al program antes de crearlo.
// Usar una var de paquete como intermediaria:
var globalProgram *tea.Program
m := tui.NewModelWithProgramRef(&globalProgram, detection, version)
globalProgram = tea.NewProgram(m, tea.WithAltScreen())
```

**Si `NewModel` retorna `Model` (valor)**, modificar `NewModel` para aceptar un puntero a programa:

**Archivo a modificar:** `internal/tui/model.go`

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
func NewModel(detection system.PlatformDetection, version string) Model {
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
// NewModel constructs the root TUI model.
// programRef is a pointer to a *tea.Program variable that will be set by main
// after calling tea.NewProgram. The closure in startInstalling reads from this
// pointer when it needs to send progress messages. It may be nil during tests.
func NewModel(detection system.PlatformDetection, version string, programRef **tea.Program) Model {
```

**Y en el struct constructor, almacenar la referencia:**
```go
	return Model{
		// ... existing fields ...
		programRef: programRef,
	}
```

**Añadir campo `programRef` al struct `Model`:**
```go
	// programRef is an indirect reference to the tea.Program set by main.
	// Used by goroutines to send msgs without capturing a stale copy.
	programRef **tea.Program
```

**Actualizar el closure en `startInstalling`:**
```go
	return m, tea.Batch(tickCmd(), func() tea.Msg {
		onProgress := func(event pipeline.ProgressEvent) {
			if m.programRef == nil || *m.programRef == nil {
				return
			}
			(*m.programRef).Send(StepProgressMsg{
				StepID: event.StepID,
				Status: pipeline.StepStatus(event.Status),
				Err:    event.Err,
			})
		}
		result := executeFn(selection, resolved, detection, onProgress)
		writeInstallLog(result)
		return PipelineDoneMsg{Result: result}
	})
```

**Actualizar `main.go`:**
```go
var prog *tea.Program
m := tui.NewModel(detection, version, &prog)
prog = tea.NewProgram(m, tea.WithAltScreen())
if _, err := prog.Run(); err != nil {
```

---

## Paso 9: Verificar que `pipeline.ProgressEvent` tiene los campos usados

`fase_05` accede a `event.StepID`, `event.Status`, `event.Err`. Verificar:

```bash
grep -rn "type ProgressEvent\|StepID\|ProgressEvent{" internal/pipeline/ | head -10
```

Si los campos tienen nombres distintos, ajustar:

```go
// Buscar el nombre exacto con:
// grep -n "type ProgressEvent struct" internal/pipeline/*.go
// Luego ajustar el callback onProgress para usar los campos reales
```

---

## Paso 10: Parche final para `classify.go` — verificar función `Score()`

`fase_04` llama `Score([5]int{h.D1, h.D2, h.D3, h.D4, 0})` desde `ValidateDecision`. Verificar que `Score` existe y acepta `[5]int`:

```bash
grep -n "func Score\|func score" internal/reasoning/gate/classify.go | head -5
```

Si la firma es diferente (ej: `Score(d1, d2, d3, d4 int)`), ajustar la llamada en `validator.go`:

**Código a reemplazar en `validator.go`** (si `Score` acepta ints separados en lugar de array):
```go
		_, postures := Score([5]int{h.D1, h.D2, h.D3, h.D4, 0})
```

**Código de reemplazo** (si la firma es `Score(d1, d2, d3, d4 int) (int, []string)`):
```go
		_, postures := Score(h.D1, h.D2, h.D3, h.D4)
```

**Comando para encontrar la firma exacta:**
```bash
grep -n "^func Score\|^func posturePriority" internal/reasoning/gate/classify.go
```

---

## Verificación de Fase

```bash
# 1. Build completo — debe pasar sin errores
go build ./... 2>&1
if [ $? -ne 0 ]; then
  echo "BUILD FAILED — revisar imports y firmas de función"
  exit 1
fi
echo "✓ Build passed"

# 2. Verificar que TotalTokens existe
go doc github.com/rd-mg/architect-ai/internal/metering SessionStats.TotalTokens 2>/dev/null \
  && echo "✓ TotalTokens exists" \
  || echo "✗ TotalTokens missing — add to session_stats.go"

# 3. Verificar que Load existe en openspec
go doc github.com/rd-mg/architect-ai/internal/components/openspec Load 2>/dev/null \
  && echo "✓ openspec.Load exists" \
  || echo "✗ openspec.Load missing — add to state.go"

# 4. Verificar que SchemaVersion existe
go doc github.com/rd-mg/architect-ai/internal/components/openspec SchemaVersion 2>/dev/null \
  && echo "✓ SchemaVersion exists" \
  || echo "✗ SchemaVersion missing — add constant to state.go"

# 5. Tests con race detector — el conjunto completo
go test -race \
  ./internal/components/openspec/... \
  ./internal/metering/... \
  ./internal/reasoning/gate/... \
  ./internal/verify/... \
  ./internal/components/engram/engramkeys/... \
  ./internal/components/filemerge/... \
  ./internal/components/mcp/... \
  -count=1 -timeout 180s 2>&1

# 6. Verificar todos los test files de fases anteriores
go test ./internal/tui/... -count=1 -timeout 60s 2>&1
go test ./internal/cli/... -count=1 -timeout 60s 2>&1
go test ./internal/architect/... -count=1 -timeout 60s 2>&1

# 7. Ejecutar vet en todos los paquetes modificados
go vet \
  ./internal/components/openspec/... \
  ./internal/metering/... \
  ./internal/reasoning/gate/... \
  ./internal/verify/... \
  ./internal/components/engram/engramkeys/... \
  ./internal/components/filemerge/... \
  ./internal/components/mcp/... \
  ./internal/tui/... \
  ./internal/cli/... \
  ./internal/architect/... \
  ./internal/sdd/state/... \
  2>&1

echo "Fase 14 complete ✓"
```

