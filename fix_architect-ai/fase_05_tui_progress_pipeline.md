# Fase 5: TUI — ProgressFunc con `p.Send()` y Install Log

**Objetivo:** Resolver F-09 (ProgressFunc no-op — UX ciega durante instalaciones) y OBS-05 (sin log persistente de instalaciones fallidas). Agrega el campo `Program *tea.Program` al `Model`, conecta el callback de progreso al event loop de BubbleTea, y persiste el resultado de cada pipeline en `~/.architect-ai/install-log.jsonl`.

---

## Paso 1: Añadir campo `Program` al struct `Model`

**Archivo a modificar:** `internal/tui/model.go`

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
	// AgentBuilder holds the transient state for the agent-builder TUI flow.
	AgentBuilder AgentBuilderState
}
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
	// AgentBuilder holds the transient state for the agent-builder TUI flow.
	AgentBuilder AgentBuilderState

	// Program is the running tea.Program instance. It is set by main after
	// construction via SetProgram(). It enables goroutines (pipeline, generation)
	// to send StepProgressMsg back to the Update loop in real time.
	// Reads are protected by the fact that SetProgram is called before Run()
	// and never mutated again; no mutex is needed.
	Program *tea.Program
}

// SetProgram stores a reference to the running tea.Program so that goroutines
// launched by startInstalling and startGeneration can send messages back.
// Must be called after NewModel() and before p.Run().
func (m *Model) SetProgram(p *tea.Program) {
	m.Program = p
}
```

---

## Paso 2: Corregir `startInstalling` para enviar `StepProgressMsg` en tiempo real

**Archivo a modificar:** `internal/tui/model.go`

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
	return m, tea.Batch(tickCmd(), func() tea.Msg {
		onProgress := func(event pipeline.ProgressEvent) {
			// NOTE: ProgressFunc is called synchronously from the pipeline goroutine.
			// We cannot use p.Send() here because we don't have a reference to the
			// tea.Program. Instead, these events are collected in the ExecutionResult
			// and the PipelineDoneMsg handles the final state. For real-time updates,
			// we rely on the pipeline calling this synchronously from each step.
		}

		result := executeFn(selection, resolved, detection, onProgress)
		return PipelineDoneMsg{Result: result}
	})
}
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
	// Capture program reference for the closure. program may be nil during
	// tests that do not call SetProgram — guarded below before use.
	program := m.Program

	return m, tea.Batch(tickCmd(), func() tea.Msg {
		onProgress := func(event pipeline.ProgressEvent) {
			if program == nil {
				return
			}
			program.Send(StepProgressMsg{
				StepID: event.StepID,
				Status: pipeline.StepStatus(event.Status),
				Err:    event.Err,
			})
		}

		result := executeFn(selection, resolved, detection, onProgress)
		writeInstallLog(result)
		return PipelineDoneMsg{Result: result}
	})
}

// writeInstallLog appends a structured install result entry to
// ~/.architect-ai/install-log.jsonl (max 500 lines, rotated on write).
// Best-effort: any error is silently discarded so it never blocks the UI.
func writeInstallLog(result pipeline.ExecutionResult) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	logDir := filepath.Join(homeDir, ".architect-ai")
	if mkErr := os.MkdirAll(logDir, 0o755); mkErr != nil {
		return
	}
	logPath := filepath.Join(logDir, "install-log.jsonl")

	type failedStep struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Err    string `json:"error,omitempty"`
	}
	type logEntry struct {
		Timestamp   string       `json:"ts"`
		Status      string       `json:"status"`
		DurationSec float64      `json:"duration_s"`
		FailedSteps []failedStep `json:"failed_steps,omitempty"`
	}

	overallStatus := "ok"
	var failed []failedStep
	for _, step := range result.Steps {
		if step.Status == pipeline.StepStatusFailed {
			overallStatus = "failed"
			errMsg := ""
			if step.Err != nil {
				errMsg = step.Err.Error()
			}
			failed = append(failed, failedStep{
				ID:     step.ID,
				Status: string(step.Status),
				Err:    errMsg,
			})
		}
	}

	entry := logEntry{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Status:      overallStatus,
		DurationSec: result.Duration.Seconds(),
		FailedSteps: failed,
	}
	line, jsonErr := json.Marshal(entry)
	if jsonErr != nil {
		return
	}
	line = append(line, '\n')

	f, openErr := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if openErr != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(line)

	rotateInstallLog(logPath, 500)
}

// rotateInstallLog keeps install-log.jsonl at most maxLines lines.
// Best-effort: errors are silently discarded.
func rotateInstallLog(logPath string, maxLines int) {
	data, err := os.ReadFile(logPath)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) <= maxLines {
		return
	}
	lines = lines[len(lines)-maxLines:]
	trimmed := strings.Join(lines, "\n")
	tmp := logPath + ".rotate.tmp"
	if writeErr := os.WriteFile(tmp, []byte(trimmed), 0o644); writeErr != nil {
		return
	}
	_ = os.Rename(tmp, logPath)
}
```

---

## Paso 3: Añadir imports necesarios en `model.go`

**Archivo a modificar:** `internal/tui/model.go`

**Acción:** Modificar — extender el bloque de imports existente

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
```

---

## Paso 4: Actualizar `main.go` para llamar a `SetProgram`

**Archivo a modificar:** `cmd/architect-ai/main.go`

**Acción:** Modificar

**Comando previo:**
```bash
grep -n "tea.NewProgram\|tui.NewModel\|p.Run\|\.Run()" cmd/architect-ai/main.go | head -10
```

**Código a reemplazar — BUSCAR EXACTAMENTE** (el patrón exacto depende del main.go; buscar la secuencia NewModel → NewProgram → Run):

```go
	m := tui.NewModel(detection, version)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
	m := tui.NewModel(detection, version)
	p := tea.NewProgram(m, tea.WithAltScreen())
	m.SetProgram(p)
	if _, err := p.Run(); err != nil {
```

> **Nota**: Si `NewModel` retorna un valor (no puntero), añadir `SetProgram` al inicio del model antes de pasar a `NewProgram`:
> ```go
> m := tui.NewModel(detection, version)
> p := tea.NewProgram(m, tea.WithAltScreen())
> // SetProgram requires a pointer receiver — update m via the program's model:
> // Use the pattern: store program in a package-level var accessible from the closure
> ```
> La implementación más simple es cambiar el campo `Program` en el model a ser accedido desde la closure directamente via una variable de cierre capturada de `main`. Ver el Paso 2 — el closure ya captura `program := m.Program`. Si `SetProgram` no está disponible en el flujo de `main`, capturar `p` en una variable antes de `p.Run()` y pasarla al model via un setter es equivalente.
>
> **Implementación alternativa robusta** si `NewModel` retorna value (no pointer) — modificar la creación así:
> ```go
> m := tui.NewModel(detection, version)
> var p *tea.Program
> p = tea.NewProgram(m, tea.WithAltScreen())
> m.SetProgram(p)
> if _, err := p.Run(); err != nil {
> ```

---

## Paso 5: Crear tests para `writeInstallLog` y el modelo con Program

**Archivo a crear:** `internal/tui/install_log_test.go`

**Acción:** Crear

```go
package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rd-mg/architect-ai/internal/pipeline"
)

func TestWriteInstallLog_CreatesFile(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	result := pipeline.ExecutionResult{
		Duration: 2 * time.Second,
		Steps: []pipeline.StepResult{
			{ID: "install-deps", Status: pipeline.StepStatusSucceeded},
		},
	}
	writeInstallLog(result)

	logPath := filepath.Join(homeDir, ".architect-ai", "install-log.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("install log not created: %v", err)
	}
	if !strings.Contains(string(data), `"status":"ok"`) {
		t.Errorf("expected ok status in log, got: %s", data)
	}
}

func TestWriteInstallLog_FailedStepsRecorded(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	result := pipeline.ExecutionResult{
		Duration: 5 * time.Second,
		Steps: []pipeline.StepResult{
			{ID: "install-deps", Status: pipeline.StepStatusSucceeded},
			{ID: "configure-agents", Status: pipeline.StepStatusFailed,
				Err: func() error {
					return &testError{"connection timeout"}
				}()},
		},
	}
	writeInstallLog(result)

	logPath := filepath.Join(homeDir, ".architect-ai", "install-log.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("install log not created: %v", err)
	}

	var entry struct {
		Status      string `json:"status"`
		FailedSteps []struct {
			ID  string `json:"id"`
			Err string `json:"error"`
		} `json:"failed_steps"`
	}
	line := strings.TrimSpace(string(data))
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("parse log entry: %v", err)
	}
	if entry.Status != "failed" {
		t.Errorf("expected status=failed, got %q", entry.Status)
	}
	if len(entry.FailedSteps) != 1 {
		t.Errorf("expected 1 failed step, got %d", len(entry.FailedSteps))
	}
	if entry.FailedSteps[0].ID != "configure-agents" {
		t.Errorf("expected failed step id=configure-agents, got %q", entry.FailedSteps[0].ID)
	}
}

func TestRotateInstallLog_KeepsMaxLines(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "install-log.jsonl")

	var sb strings.Builder
	for i := 0; i < 600; i++ {
		sb.WriteString(`{"ts":"2026-01-01T00:00:00Z","status":"ok"}`)
		sb.WriteByte('\n')
	}
	if err := os.WriteFile(logPath, []byte(sb.String()), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	rotateInstallLog(logPath, 500)

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read rotated log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) > 500 {
		t.Errorf("rotated log has %d lines, want <= 500", len(lines))
	}
}

func TestSetProgram_StoresReference(t *testing.T) {
	m := &Model{}
	// Verify that SetProgram stores the pointer and Program field is accessible.
	// We cannot construct a real tea.Program in unit tests without a TTY,
	// so we verify the field assignment path only.
	if m.Program != nil {
		t.Error("Program should be nil before SetProgram")
	}
	// SetProgram with nil is safe (no panic)
	m.SetProgram(nil)
	if m.Program != nil {
		t.Error("Program should be nil after SetProgram(nil)")
	}
}

// testError is a minimal error implementation for test use.
type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
```

---

## Verificación de Fase

```bash
# 1. Compilar el paquete tui
go build ./internal/tui/...

# 2. Compilar cmd/architect-ai
go build ./cmd/architect-ai/...

# 3. Tests del paquete tui (sin regresiones)
go test ./internal/tui/... -v -count=1

# 4. Tests nuevos de install log
go test ./internal/tui/... -v -count=1 -run TestWriteInstallLog
go test ./internal/tui/... -v -count=1 -run TestRotateInstallLog
go test ./internal/tui/... -v -count=1 -run TestSetProgram

# 5. Verificar que el test de StepProgressMsg existente sigue pasando
go test ./internal/tui/... -v -count=1 -run TestStepProgressMsg

# 6. Race detector sobre el paquete tui
go test -race ./internal/tui/... -count=1

# 7. Verificar import de encoding/json añadido correctamente
go vet ./internal/tui/...

# 8. Confirmar que install-log.jsonl se crea en el home del usuario
# (requiere ejecutar el binario real, no es un test automatizable sin TTY)
echo "Manual check: after running 'architect-ai' and completing an install,"
echo "verify: cat ~/.architect-ai/install-log.jsonl | head -1"
```

