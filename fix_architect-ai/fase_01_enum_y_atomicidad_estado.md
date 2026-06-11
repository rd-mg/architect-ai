# Fase 1: Corrección del Enum `running` y Atomicidad del Estado SDD

**Objetivo:** Resolver F-04-A (TOCTOU en escritura de estado) y F-04-B (enum `"running"` inválido que bloquea `Save()` y el rollback harness). Esta fase desbloquea TODO el pipeline SDD en modos `openspec` e `hybrid`.

---

## Paso 1: Agregar nuevos valores al enum `ValidStatuses` y corregir `Save()`

**Archivo a modificar:** `internal/components/openspec/state.go`

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
// ValidStatuses is the closed set of phase status values.
var ValidStatuses = []string{
	"pending", "in_progress", "completed", "skipped", "failed",
}
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
// ValidStatuses is the closed set of phase status values.
// NOTE: "running" was removed in v0.3. All prompts must use "in_progress".
var ValidStatuses = []string{
	"pending", "in_progress", "completed", "skipped", "failed",
	"abandoned", "infrastructure_blocked", "partially_completed",
	"rollback_in_progress",
}
```

---

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
// Save writes state.yaml atomically (tmp + rename).
// Touches UpdatedAt to now (UTC) before writing.
func Save(path string, s *State) error {
	s.UpdatedAt = time.Now().UTC()
	parent := filepath.Base(filepath.Dir(path))
	if err := Validate(s, parent); err != nil {
		return err
	}
	out, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	// fsync is best-effort; wrap in separate call for platforms that support it.
	if f, err := os.OpenFile(tmp, os.O_RDWR, 0o644); err == nil {
		_ = f.Sync()
		_ = f.Close()
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
// Save writes state.yaml atomically (unique-tmp + rename + O_EXCL lock).
//
// Atomicity guarantee:
//   - Uses a per-PID+nanosecond .tmp path so concurrent writers never share
//     the same temporary file (eliminates the TOCTOU collision on .tmp).
//   - Uses O_CREATE|O_EXCL for the lockfile so only one writer can hold
//     the lock at a time (kernel-level atomicity, no Stat+Write race).
//   - Rename is atomic on all POSIX filesystems (Linux, macOS).
//
// Precondition:  path is writable; s is a valid *State.
// Postcondition: path contains the marshalled s; lockfile removed.
func Save(path string, s *State) error {
	s.UpdatedAt = time.Now().UTC()
	parent := filepath.Base(filepath.Dir(path))
	if err := Validate(s, parent); err != nil {
		return err
	}
	out, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	lockPath := path + ".lock"
	lf, lockErr := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if lockErr != nil {
		if os.IsExist(lockErr) {
			info, statErr := os.Stat(lockPath)
			if statErr == nil && time.Since(info.ModTime()) > 30*time.Second {
				_ = os.Remove(lockPath)
				return Save(path, s)
			}
			return fmt.Errorf("state file locked by another process (%s)", lockPath)
		}
		return fmt.Errorf("acquire lock: %w", lockErr)
	}
	_, _ = fmt.Fprintf(lf, "%d", os.Getpid())
	_ = lf.Close()
	defer func() { _ = os.Remove(lockPath) }()

	tmp := fmt.Sprintf("%s.%d.%d.tmp", path, os.Getpid(), time.Now().UnixNano())
	defer func() { _ = os.Remove(tmp) }()

	if err := os.WriteFile(tmp, out, 0o644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if f, ferr := os.OpenFile(tmp, os.O_RDWR, 0o644); ferr == nil {
		_ = f.Sync()
		_ = f.Close()
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}
```

---

**Añadir import `"fmt"` si no está presente — VERIFICAR imports actuales:**

El import `"fmt"` ya está presente en el archivo (`fmt.Errorf`). No se requiere cambio.

---

## Paso 2: Corregir el enum en `phase-dag-enforcement.md`

**Archivo a modificar:** `internal/assets/_shared/phase-dag-enforcement.md`

**Acción:** Modificar

**Comando previo (si aplica):**
```bash
grep -n "running" internal/assets/_shared/phase-dag-enforcement.md
```

**Código a reemplazar — BUSCAR EXACTAMENTE** (ajustar contexto si el archivo difiere en espaciado):

Buscar todas las ocurrencias del patrón `status.*\`running\`` y `status: "running"` en el archivo:

```bash
sed -i \
  "s/status.*\`running\`/status: \`in_progress\`/g; \
   s/status: \"running\"/status: \"in_progress\"/g; \
   s/update status to \`running\`/update status to \`in_progress\`/g; \
   s/→ \"running\"/→ \"in_progress\"/g" \
  internal/assets/_shared/phase-dag-enforcement.md
```

**Verificar que no quedan ocurrencias:**
```bash
grep -n "\"running\"\|status.*running" internal/assets/_shared/phase-dag-enforcement.md
```
La salida debe estar vacía.

---

## Paso 3: Corregir el rollback harness

**Archivo a modificar:** `internal/assets/_shared/rollback-harness.md`

**Acción:** Modificar

**Comando previo:**
```bash
grep -n "running" internal/assets/_shared/rollback-harness.md
```

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```
r'(  sdd-apply:.*?status: )\"(running|failed)\"'
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```
r'(  sdd-apply:.*?status: )\"(in_progress|failed)\"'
```

**Comando de aplicación:**
```bash
sed -i \
  "s/'(running|failed)'/'(in_progress|failed)'/g; \
   s/\"running|failed\"/\"in_progress|failed\"/g" \
  internal/assets/_shared/rollback-harness.md
```

**Verificar:**
```bash
grep -n "\"running\"" internal/assets/_shared/rollback-harness.md
```
La salida debe estar vacía.

---

## Paso 4: Crear test de concurrencia con `-race`

**Archivo a crear:** `internal/components/openspec/concurrent_test.go`

**Acción:** Crear

```go
package openspec

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestSave_ConcurrentWriters_NoCorruption(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	changeName := "test-change"
	stateDir := filepath.Join(dir, changeName)
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(stateDir, "state.yaml")

	makeState := func(n int) *State {
		now := time.Now().UTC()
		return &State{
			SchemaVersion: SchemaVersion,
			ChangeName:    changeName,
			CreatedAt:     now,
			UpdatedAt:     now,
			ArtifactStore: "openspec",
			Phases: map[string]*Phase{
				"sdd-explore": {
					Status: "in_progress",
					StartedAt: func() *time.Time {
						t := time.Now().UTC()
						return &t
					}(),
					Meta: map[string]any{"writer": n},
				},
			},
		}
	}

	const writers = 20
	var wg sync.WaitGroup
	errs := make([]error, writers)

	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			errs[n] = Save(path, makeState(n))
		}(i)
	}
	wg.Wait()

	// Count successes vs lock-contention errors (both are acceptable).
	successes := 0
	for _, err := range errs {
		if err == nil {
			successes++
		} else if err.Error() == "" {
			t.Errorf("unexpected empty error")
		}
	}
	if successes == 0 {
		t.Fatal("at least one writer must succeed")
	}

	// The final file must be valid, parseable YAML — not corrupted.
	final, err := Load(path)
	if err != nil {
		t.Fatalf("final state.yaml corrupted after concurrent writes: %v", err)
	}
	if final.ChangeName != changeName {
		t.Errorf("change_name corrupted: got %q, want %q", final.ChangeName, changeName)
	}
}

func TestSave_EnumInProgress_Accepted(t *testing.T) {
	dir := t.TempDir()
	changeName := "enum-test"
	stateDir := filepath.Join(dir, changeName)
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(stateDir, "state.yaml")

	now := time.Now().UTC()
	s := &State{
		SchemaVersion: SchemaVersion,
		ChangeName:    changeName,
		CreatedAt:     now,
		UpdatedAt:     now,
		ArtifactStore: "openspec",
		Phases: map[string]*Phase{
			"sdd-explore": {
				Status: "in_progress",
				StartedAt: func() *time.Time {
					t := time.Now().UTC()
					return &t
				}(),
			},
		},
	}
	if err := Save(path, s); err != nil {
		t.Fatalf("Save with in_progress should succeed: %v", err)
	}
}

func TestSave_EnumRunning_Rejected(t *testing.T) {
	dir := t.TempDir()
	changeName := "enum-running"
	stateDir := filepath.Join(dir, changeName)
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(stateDir, "state.yaml")

	now := time.Now().UTC()
	s := &State{
		SchemaVersion: SchemaVersion,
		ChangeName:    changeName,
		CreatedAt:     now,
		UpdatedAt:     now,
		ArtifactStore: "openspec",
		Phases: map[string]*Phase{
			"sdd-explore": {Status: "running"},
		},
	}
	if err := Save(path, s); err == nil {
		t.Fatal(`Save with status "running" must fail — "running" is not a valid status`)
	}
}

func TestSave_NewStatuses_Accepted(t *testing.T) {
	newStatuses := []string{
		"abandoned", "infrastructure_blocked",
		"partially_completed", "rollback_in_progress",
	}
	for _, status := range newStatuses {
		t.Run(fmt.Sprintf("status_%s", status), func(t *testing.T) {
			dir := t.TempDir()
			changeName := "status-test"
			stateDir := filepath.Join(dir, changeName)
			if err := os.MkdirAll(stateDir, 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			path := filepath.Join(stateDir, "state.yaml")

			now := time.Now().UTC()
			ph := &Phase{Status: status}
			if status == "in_progress" {
				ph.StartedAt = func() *time.Time {
					t := time.Now().UTC()
					return &t
				}()
			}
			if status == "completed" {
				ph.CompletedAt = func() *time.Time {
					t := time.Now().UTC()
					return &t
				}()
			}
			if status == "failed" {
				ph.Error = "test error"
			}
			s := &State{
				SchemaVersion: SchemaVersion,
				ChangeName:    changeName,
				CreatedAt:     now,
				UpdatedAt:     now,
				ArtifactStore: "openspec",
				Phases:        map[string]*Phase{"sdd-explore": ph},
			}
			if err := Save(path, s); err != nil {
				t.Errorf("Save with status %q should succeed: %v", status, err)
			}
		})
	}
}
```

---

## Paso 5: Crear script de migración para cambios existentes con estado inválido

**Archivo a crear:** `internal/components/openspec/migrate_v03.go`

**Acción:** Crear

```go
// Package openspec — migrate_v03.go
// Migrates sdd-state.yaml files that contain the obsolete "running" status
// (written by prompts before v0.3) to the canonical "in_progress" status.
// Safe to run multiple times (idempotent).
package openspec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MigrateV03 scans all state.yaml files under atDir and replaces the obsolete
// "running" status with "in_progress". Writes atomically via Save().
// Returns the list of files that were migrated and any error.
func MigrateV03(atDir string) (migrated []string, err error) {
	changesDir := filepath.Join(atDir, "openspec", "changes")
	entries, readErr := os.ReadDir(changesDir)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return nil, nil
		}
		return nil, fmt.Errorf("read changes dir: %w", readErr)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		statePath := filepath.Join(changesDir, entry.Name(), "state.yaml")
		data, readErr := os.ReadFile(statePath)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				continue
			}
			return migrated, fmt.Errorf("read %s: %w", statePath, readErr)
		}

		content := string(data)
		if !strings.Contains(content, `status: "running"`) &&
			!strings.Contains(content, "status: running") {
			continue
		}

		fixed := strings.ReplaceAll(content, `status: "running"`, `status: "in_progress"`)
		fixed = strings.ReplaceAll(fixed, "status: running", "status: in_progress")

		state, parseErr := func() (*State, error) {
			var s State
			if err := func() error {
				return loadYAML([]byte(fixed), &s)
			}(); err != nil {
				return nil, err
			}
			return &s, nil
		}()
		if parseErr != nil {
			return migrated, fmt.Errorf("parse migrated %s: %w", statePath, parseErr)
		}

		if saveErr := Save(statePath, state); saveErr != nil {
			return migrated, fmt.Errorf("save migrated %s: %w", statePath, saveErr)
		}
		migrated = append(migrated, statePath)
	}
	return migrated, nil
}

// loadYAML is a thin wrapper around yaml.Unmarshal kept private to the package.
func loadYAML(data []byte, v any) error {
	import_yaml_v3 := func() error {
		return _yamlUnmarshal(data, v)
	}
	return import_yaml_v3()
}
```

**Nota**: la función `_yamlUnmarshal` referencia el mismo `yaml.Unmarshal` ya importado en `state.go`. Dado que ambos están en `package openspec`, el import de `gopkg.in/yaml.v3` ya está disponible. Reemplazar el contenido de `loadYAML` con:

**Archivo a modificar:** `internal/components/openspec/migrate_v03.go`

**Acción:** Crear (versión final sin wrapper intermedio)

```go
package openspec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// MigrateV03 scans all state.yaml files under atDir and replaces the obsolete
// "running" status with "in_progress". Writes atomically via Save().
// Returns the list of migrated file paths and any first error encountered.
func MigrateV03(atDir string) (migrated []string, err error) {
	changesDir := filepath.Join(atDir, "openspec", "changes")
	entries, readErr := os.ReadDir(changesDir)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return nil, nil
		}
		return nil, fmt.Errorf("read changes dir: %w", readErr)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		statePath := filepath.Join(changesDir, entry.Name(), "state.yaml")
		data, readErr := os.ReadFile(statePath)
		if readErr != nil {
			if os.IsNotExist(readErr) {
				continue
			}
			return migrated, fmt.Errorf("read %s: %w", statePath, readErr)
		}

		content := string(data)
		if !strings.Contains(content, `status: "running"`) &&
			!strings.Contains(content, "status: running\n") {
			continue
		}

		fixed := strings.ReplaceAll(content, `status: "running"`, `status: "in_progress"`)
		fixed = strings.ReplaceAll(fixed, "status: running\n", "status: in_progress\n")

		var s State
		if parseErr := yaml.Unmarshal([]byte(fixed), &s); parseErr != nil {
			return migrated, fmt.Errorf("parse migrated %s: %w", statePath, parseErr)
		}

		if saveErr := Save(statePath, &s); saveErr != nil {
			return migrated, fmt.Errorf("save migrated %s: %w", statePath, saveErr)
		}
		migrated = append(migrated, statePath)
	}
	return migrated, nil
}
```

---

## Verificación de Fase

```bash
# 1. Compilar el paquete
go build ./internal/components/openspec/...

# 2. Tests existentes deben seguir pasando
go test ./internal/components/openspec/... -v -count=1

# 3. Test de concurrencia con race detector
go test -race ./internal/components/openspec/... -v -count=1 -run TestSave_Concurrent

# 4. Test del enum correcto
go test ./internal/components/openspec/... -v -count=1 -run TestSave_Enum

# 5. Verificar que no queda "running" en assets de prompts
grep -rn '"running"\|status.*running' internal/assets/_shared/ \
  | grep -v "\.go:" \
  | grep -v "#"

# 6. Ejecutar migración en el proyecto actual si existe .atl/
if [ -d ".atl/openspec/changes" ]; then
  go run ./cmd/architect-ai migrate-v03 2>&1
fi
```

