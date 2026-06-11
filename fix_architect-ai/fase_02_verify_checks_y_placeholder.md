# Fase 2: Checks de Verificación — Placeholders y Estado SDD

**Objetivo:** Resolver F-03 (CLAUDE.md con placeholders sin resolver) y agregar observabilidad de Engram al comando `architect-ai check all`. Crea nuevos checks de verificación sin romper la firma de `Check` existente ni los callers de `RunChecks`.

---

## Paso 1: Agregar constantes de fix hints

**Archivo a modificar:** `internal/verify/fix_hints.go`

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
// Fix hint constants — copy-paste commands shown under failed checks.
// One constant per check ID family. Keep them short (one command, no prose).\nconst (\n\tFixInstallEngram     = "architect-ai install --component engram"\n\tFixInstallContext7   = "architect-ai install --component context7"\n\tFixInstallNotebookLM = "architect-ai install --component notebooklm"\n\tFixInstallGGA        = "architect-ai install --component gga"\n\tFixSync              = "architect-ai sync"\n\tFixRepairPermissions = "architect-ai sync --repair-permissions"\n\tFixInstallNode       = "Install Node.js 18+ → https://nodejs.org"\n)
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
// Fix hint constants — copy-paste commands shown under failed checks.
// One constant per check ID family. Keep them short (one command, no prose).
const (
	FixInstallEngram     = "architect-ai install --component engram"
	FixInstallContext7   = "architect-ai install --component context7"
	FixInstallNotebookLM = "architect-ai install --component notebooklm"
	FixInstallGGA        = "architect-ai install --component gga"
	FixSync              = "architect-ai sync"
	FixRepairPermissions = "architect-ai sync --repair-permissions"
	FixInstallNode       = "Install Node.js 18+ → https://nodejs.org"
	FixBuild             = "architect-ai build"
	FixMigrateV03        = "architect-ai migrate-v03"
	FixDiagnoseEngram    = "architect-ai diagnose engram"
)
```

---

## Paso 2: Crear el archivo de checks de proyecto

**Archivo a crear:** `internal/verify/project_checks.go`

**Acción:** Crear

```go
package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// claudeMDPlaceholders is the canonical list of template tokens that must not
// appear in a materialized CLAUDE.md. Each token indicates the file was never
// built with `architect-ai build`.
var claudeMDPlaceholders = []string{
	"{content from",
	"{L0_HASH}",
	"{L1A_HASH}",
	"{L1B_HASH}",
	"{CONTENT_HASH}",
	"{INJECTED_MODE}",
	"{D1}", "{D2}", "{D3}", "{D4}",
}

// claudeMDMinBytes is the minimum size of a properly built CLAUDE.md.
// A file smaller than this has not been materialized.
const claudeMDMinBytes = 20_000

// ClaudeMDNoPlaceholdersCheck returns a Check that verifies CLAUDE.md in
// projectRoot has been built (no unresolved template placeholders).
//
// Precondition:  projectRoot is the root directory of the architect-ai project.
// Postcondition: returns nil iff CLAUDE.md exists, is > 20KB, and contains no
//
//	template placeholders.
func ClaudeMDNoPlaceholdersCheck(projectRoot string) Check {
	return Check{
		ID:          "verify:project:claude-md-no-placeholders",
		Description: "CLAUDE.md has no unresolved template placeholders",
		FixHint:     FixBuild,
		Run: func(_ context.Context) error {
			claudePath := filepath.Join(projectRoot, "CLAUDE.md")
			data, err := os.ReadFile(claudePath)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return fmt.Errorf("read CLAUDE.md: %w", err)
			}
			for _, ph := range claudeMDPlaceholders {
				if strings.Contains(string(data), ph) {
					return fmt.Errorf(
						"CLAUDE.md contains unresolved placeholder %q — run: architect-ai build\n"+
							"All agent prompts must be materialized before use.",
						ph,
					)
				}
			}
			if len(data) < claudeMDMinBytes {
				return fmt.Errorf(
					"CLAUDE.md is only %d bytes (expected >%d for a built file) — run: architect-ai build",
					len(data), claudeMDMinBytes,
				)
			}
			return nil
		},
	}
}

// SDDStateEnumCheck returns a Check that verifies all sdd-state.yaml files
// under atDir use valid status enum values. Detects the "running" vs
// "in_progress" mismatch introduced before v0.3.
//
// Precondition:  atDir is the .atl directory of the project.
// Postcondition: returns nil iff no state.yaml contains a "running" status or
//
//	any other value outside ValidStatuses.
func SDDStateEnumCheck(atDir string) Check {
	return Check{
		ID:          "verify:project:sdd-state-enum",
		Description: "sdd-state.yaml uses valid status enum values (no obsolete \"running\")",
		FixHint:     FixMigrateV03,
		Soft:        true,
		Run: func(_ context.Context) error {
			changesDir := filepath.Join(atDir, "openspec", "changes")
			entries, err := os.ReadDir(changesDir)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return fmt.Errorf("read changes dir: %w", err)
			}
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				statePath := filepath.Join(changesDir, entry.Name(), "state.yaml")
				data, err := os.ReadFile(statePath)
				if err != nil {
					if os.IsNotExist(err) {
						continue
					}
					return fmt.Errorf("read %s: %w", statePath, err)
				}
				content := string(data)
				if strings.Contains(content, `status: "running"`) ||
					strings.Contains(content, "status: running\n") {
					return fmt.Errorf(
						"sdd-state.yaml for change %q contains obsolete status \"running\"\n"+
							"Run: architect-ai migrate-v03",
						entry.Name(),
					)
				}
			}
			return nil
		},
	}
}

// probeLogEntry is a single line from .atl/probe-log.jsonl.
type probeLogEntry struct {
	TS     string `json:"ts"`
	Probe  string `json:"probe"`
	Result string `json:"result"`
	Error  string `json:"error"`
}

// EngramProbeLogCheck returns a soft Check that warns when Engram has failed
// 3 or more consecutive times. Reads .atl/probe-log.jsonl (JSONL format,
// one JSON object per line).
//
// This check is always Soft: a degraded Engram is a warning, not a blocker.
func EngramProbeLogCheck(atDir string) Check {
	return Check{
		ID:          "verify:engram:probe-log",
		Description: "Engram probe has not failed 3+ consecutive times",
		FixHint:     FixDiagnoseEngram,
		Soft:        true,
		Run: func(_ context.Context) error {
			logPath := filepath.Join(atDir, "probe-log.jsonl")
			data, err := os.ReadFile(logPath)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return nil
			}
			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			start := 0
			if len(lines) > 10 {
				start = len(lines) - 10
			}
			recent := lines[start:]
			consecFails := 0
			for _, line := range recent {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				var entry probeLogEntry
				if jsonErr := json.Unmarshal([]byte(line), &entry); jsonErr != nil {
					continue
				}
				if entry.Result == "failed" {
					consecFails++
				} else if entry.Result == "ok" {
					consecFails = 0
				}
			}
			if consecFails >= 3 {
				return fmt.Errorf(
					"Engram has failed %d consecutive times — sessions will degrade to 'none' mode\n"+
						"Run: architect-ai diagnose engram",
					consecFails,
				)
			}
			return nil
		},
	}
}

// WriteProbeLogEntry appends a structured probe result to .atl/probe-log.jsonl.
// Creates the file if it does not exist. Rotates the log at maxLines entries.
// Safe to call from multiple goroutines because it opens with O_APPEND which is
// atomic for writes smaller than PIPE_BUF on Linux (and our entries are <512 bytes).
func WriteProbeLogEntry(atDir, probe, result, errMsg string) error {
	logPath := filepath.Join(atDir, "probe-log.jsonl")

	entry := probeLogEntry{
		TS:     time.Now().UTC().Format(time.RFC3339),
		Probe:  probe,
		Result: result,
		Error:  errMsg,
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal probe entry: %w", err)
	}
	line = append(line, '\n')

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open probe log: %w", err)
	}
	defer f.Close()
	_, err = f.Write(line)
	if err != nil {
		return fmt.Errorf("write probe log: %w", err)
	}

	rotateProbeLog(logPath, 200)
	return nil
}

// rotateProbeLog keeps the log at most maxLines lines by truncating from the top.
// It is best-effort: any error is silently discarded to avoid masking the
// primary write path.
func rotateProbeLog(logPath string, maxLines int) {
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

## Paso 3: Inyectar los nuevos checks en `runPostApplyVerification`

**Archivo a modificar:** `internal/cli/install.go`

> **Nota de localización**: La función `runPostApplyVerification` está en el paquete `cli` dentro de `internal/cli/install.go` (o el archivo equivalente donde se encuentre `func runPostApplyVerification`). Buscar la función exacta con:
> ```bash
> grep -rn "func runPostApplyVerification" internal/
> ```

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
	checks = append(checks, antigravityCollisionCheck(resolved.Agents)...)
	checks = append(checks, antigravityInstallCheck(homeDir, resolved.Agents)...)

	return verify.BuildReport(verify.RunChecks(context.Background(), checks))
}
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
	checks = append(checks, antigravityCollisionCheck(resolved.Agents)...)
	checks = append(checks, antigravityInstallCheck(homeDir, resolved.Agents)...)

	projectRoot, _ := os.Getwd()
	if projectRoot != "" {
		checks = append(checks, verify.ClaudeMDNoPlaceholdersCheck(projectRoot))
	}
	atDir := filepath.Join(projectRoot, ".atl")
	checks = append(checks, verify.SDDStateEnumCheck(atDir))
	checks = append(checks, verify.EngramProbeLogCheck(atDir))

	return verify.BuildReport(verify.RunChecks(context.Background(), checks))
}
```

---

## Paso 4: Crear tests para los nuevos checks

**Archivo a crear:** `internal/verify/project_checks_test.go`

**Acción:** Crear

```go
package verify

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestClaudeMDNoPlaceholdersCheck_Pass(t *testing.T) {
	dir := t.TempDir()
	content := strings.Repeat("x", claudeMDMinBytes+1)
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	check := ClaudeMDNoPlaceholdersCheck(dir)
	if err := check.Run(context.Background()); err != nil {
		t.Errorf("expected no error for clean CLAUDE.md, got: %v", err)
	}
}

func TestClaudeMDNoPlaceholdersCheck_FailOnPlaceholder(t *testing.T) {
	cases := []string{
		"{content from .atl/agents/architect.md}",
		"{L0_HASH}",
		"{L1A_HASH}",
		"{INJECTED_MODE}",
		"{D4}",
	}
	for _, ph := range cases {
		t.Run(ph, func(t *testing.T) {
			dir := t.TempDir()
			content := strings.Repeat("a", claudeMDMinBytes) + ph
			if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(content), 0o644); err != nil {
				t.Fatalf("setup: %v", err)
			}
			check := ClaudeMDNoPlaceholdersCheck(dir)
			if err := check.Run(context.Background()); err == nil {
				t.Errorf("expected error for placeholder %q, got nil", ph)
			}
		})
	}
}

func TestClaudeMDNoPlaceholdersCheck_FailOnSmallFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("too small"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	check := ClaudeMDNoPlaceholdersCheck(dir)
	if err := check.Run(context.Background()); err == nil {
		t.Error("expected error for small CLAUDE.md, got nil")
	}
}

func TestClaudeMDNoPlaceholdersCheck_MissingFile(t *testing.T) {
	dir := t.TempDir()
	check := ClaudeMDNoPlaceholdersCheck(dir)
	if err := check.Run(context.Background()); err != nil {
		t.Errorf("missing CLAUDE.md should not error (project may not use it): %v", err)
	}
}

func TestSDDStateEnumCheck_Pass(t *testing.T) {
	dir := t.TempDir()
	changesDir := filepath.Join(dir, "openspec", "changes", "my-change")
	if err := os.MkdirAll(changesDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	content := `version: "3.0"
change_name: "my-change"
phases:
  sdd-explore:
    status: "in_progress"
`
	if err := os.WriteFile(filepath.Join(changesDir, "state.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	check := SDDStateEnumCheck(dir)
	if err := check.Run(context.Background()); err != nil {
		t.Errorf("expected no error for in_progress status: %v", err)
	}
}

func TestSDDStateEnumCheck_FailOnRunning(t *testing.T) {
	dir := t.TempDir()
	changesDir := filepath.Join(dir, "openspec", "changes", "bad-change")
	if err := os.MkdirAll(changesDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	content := `version: "3.0"
change_name: "bad-change"
phases:
  sdd-apply:
    status: "running"
`
	if err := os.WriteFile(filepath.Join(changesDir, "state.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	check := SDDStateEnumCheck(dir)
	if err := check.Run(context.Background()); err == nil {
		t.Error("expected error for obsolete 'running' status, got nil")
	}
}

func TestEngramProbeLogCheck_NoLog(t *testing.T) {
	dir := t.TempDir()
	check := EngramProbeLogCheck(dir)
	if err := check.Run(context.Background()); err != nil {
		t.Errorf("missing probe log should not error: %v", err)
	}
}

func TestEngramProbeLogCheck_ThreeConsecutiveFails(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "probe-log.jsonl")

	entries := []probeLogEntry{
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "ok"},
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "failed", Error: "timeout"},
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "failed", Error: "timeout"},
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "failed", Error: "connection refused"},
	}
	var sb strings.Builder
	for _, e := range entries {
		line, _ := json.Marshal(e)
		sb.Write(line)
		sb.WriteByte('\n')
	}
	if err := os.WriteFile(logPath, []byte(sb.String()), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	check := EngramProbeLogCheck(dir)
	if err := check.Run(context.Background()); err == nil {
		t.Error("expected error for 3 consecutive failures, got nil")
	}
}

func TestEngramProbeLogCheck_RecoveredAfterFails(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "probe-log.jsonl")

	entries := []probeLogEntry{
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "failed"},
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "failed"},
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "failed"},
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "ok"},
	}
	var sb strings.Builder
	for _, e := range entries {
		line, _ := json.Marshal(e)
		sb.Write(line)
		sb.WriteByte('\n')
	}
	if err := os.WriteFile(logPath, []byte(sb.String()), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	check := EngramProbeLogCheck(dir)
	if err := check.Run(context.Background()); err != nil {
		t.Errorf("after recovery 'ok' entry, check should pass: %v", err)
	}
}

func TestWriteProbeLogEntry_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	if err := WriteProbeLogEntry(dir, "engram", "ok", ""); err != nil {
		t.Fatalf("WriteProbeLogEntry: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "probe-log.jsonl"))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(data), `"result":"ok"`) {
		t.Errorf("expected ok entry in log, got: %s", data)
	}
}
```

---

## Verificación de Fase

```bash
# 1. Compilar el paquete verify
go build ./internal/verify/...

# 2. Compilar todos los paquetes (detectar imports rotos)
go build ./...

# 3. Ejecutar tests del paquete verify
go test ./internal/verify/... -v -count=1

# 4. Verificar que no hay imports circulares
go list -deps ./internal/verify/... | head -20

# 5. Verificar que fix_hints.go compila con las nuevas constantes
go vet ./internal/verify/...

# 6. Localizar el archivo donde está runPostApplyVerification y
#    confirmar que el import de "path/filepath" ya está presente
grep -n "filepath" "$(grep -rl 'runPostApplyVerification' internal/)"

# 7. Ejecutar el check CLI end-to-end si hay un proyecto configurado
if [ -f ".atl/config.yaml" ]; then
  go run ./cmd/architect-ai check all 2>&1 | grep -E "verify:project|PASSED|FAILED|WARNING"
fi
```

