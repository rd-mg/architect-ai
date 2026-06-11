# Fase 12: Persistence Contract — HybridWriter Transaccional

**Objetivo:** Resolver DBC-01 (postcondición rota en hybrid mode — una escritura puede fallar silenciosamente y retornar `completed`). Implementa un `HybridWriter` Go con semántica de transacción: ambas escrituras deben completarse o el resultado es `partially_completed`. Modifica el `persistence-contract.md` para eliminar el fallback silencioso.

---

## Paso 1: Crear `internal/components/openspec/hybrid_writer.go`

**Archivo a crear:** `internal/components/openspec/hybrid_writer.go`

**Acción:** Crear

```go
package openspec

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HybridWriteStatus represents the outcome of a hybrid (Engram + filesystem) write.
type HybridWriteStatus string

const (
	// HybridWriteCompleted means both Engram and filesystem writes succeeded.
	HybridWriteCompleted HybridWriteStatus = "completed"
	// HybridWritePartial means exactly one store succeeded. The caller must
	// not advance to the next phase — user decision required.
	HybridWritePartial HybridWriteStatus = "partially_completed"
	// HybridWriteBlocked means both stores failed. Phase is blocked.
	HybridWriteBlocked HybridWriteStatus = "blocked"
)

// HybridWriteResult is returned by HybridWriter.Write.
// It satisfies the result-contract v0.3 postcondition.
type HybridWriteResult struct {
	Status       HybridWriteStatus `json:"status"`
	EngramOK     bool              `json:"engram_ok"`
	FilesystemOK bool              `json:"filesystem_ok"`
	EngramErr    error             `json:"engram_error,omitempty"`
	FsErr        error             `json:"fs_error,omitempty"`
	// BlockedReason is non-empty whenever Status != completed.
	BlockedReason string `json:"blocked_reason,omitempty"`
}

// HybridArtifact describes what to write in both stores.
type HybridArtifact struct {
	// TopicKey is the Engram mem_save key.
	TopicKey string
	// FilePath is the absolute filesystem path (.atl/openspec/changes/…).
	FilePath string
	// Content is the artifact payload (YAML, Markdown, or plain text).
	Content string
	// Title is the Engram mem_save title (displayed in mem_search results).
	Title string
	// ArtifactType is the Engram type field (e.g. "architecture", "spec").
	ArtifactType string
}

// EngramWriteFn is the function signature for writing to Engram.
// Implementations call the appropriate MCP tool (mem_save).
type EngramWriteFn func(ctx context.Context, topicKey, title, artifactType, content string) error

// FilesystemWriteFn is the function signature for atomic filesystem writes.
// Implementations must write atomically (tmp + rename).
type FilesystemWriteFn func(ctx context.Context, path, content string) error

// HybridWriter executes atomic writes to both Engram and filesystem in parallel.
//
// Invariant (DBC-01 postcondition):
//   - If both succeed: Status == HybridWriteCompleted
//   - If exactly one succeeds: Status == HybridWritePartial (NOT completed)
//   - If both fail: Status == HybridWriteBlocked
//   - Status == HybridWriteCompleted is the ONLY valid value for advancing to the next phase.
type HybridWriter struct {
	EngramFn     EngramWriteFn
	FilesystemFn FilesystemWriteFn
	// Timeout is the deadline for each individual store write.
	// Default: 30 seconds.
	Timeout time.Duration
}

// Write executes both store writes in parallel and returns a HybridWriteResult.
//
// Preconditions:
//   - artifact.TopicKey is non-empty
//   - artifact.FilePath is an absolute path
//   - artifact.Content is non-empty
//   - hw.EngramFn is non-nil
//   - hw.FilesystemFn is non-nil
//
// Postcondition:
//   - result.Status reflects the exact combination of successes and failures
//   - result.EngramOK and result.FilesystemOK are always set (never both uncertain)
func (hw *HybridWriter) Write(ctx context.Context, artifact HybridArtifact) HybridWriteResult {
	if artifact.TopicKey == "" {
		return HybridWriteResult{
			Status:        HybridWriteBlocked,
			BlockedReason: "hybrid_write_precondition: TopicKey must not be empty",
		}
	}
	if artifact.FilePath == "" {
		return HybridWriteResult{
			Status:        HybridWriteBlocked,
			BlockedReason: "hybrid_write_precondition: FilePath must not be empty",
		}
	}
	if artifact.Content == "" {
		return HybridWriteResult{
			Status:        HybridWriteBlocked,
			BlockedReason: "hybrid_write_precondition: Content must not be empty",
		}
	}

	timeout := hw.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	writeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		engramErr error
		fsErr     error
		wg        sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		engramErr = hw.EngramFn(writeCtx,
			artifact.TopicKey,
			artifact.Title,
			artifact.ArtifactType,
			artifact.Content,
		)
	}()

	go func() {
		defer wg.Done()
		fsErr = hw.FilesystemFn(writeCtx, artifact.FilePath, artifact.Content)
	}()

	wg.Wait()

	engramOK := engramErr == nil
	fsOK := fsErr == nil

	switch {
	case engramOK && fsOK:
		return HybridWriteResult{
			Status:       HybridWriteCompleted,
			EngramOK:     true,
			FilesystemOK: true,
		}
	case engramOK && !fsOK:
		return HybridWriteResult{
			Status:        HybridWritePartial,
			EngramOK:      true,
			FilesystemOK:  false,
			FsErr:         fsErr,
			BlockedReason: fmt.Sprintf("filesystem_unavailable: %v", fsErr),
		}
	case !engramOK && fsOK:
		return HybridWriteResult{
			Status:        HybridWritePartial,
			EngramOK:      false,
			FilesystemOK:  true,
			EngramErr:     engramErr,
			BlockedReason: fmt.Sprintf("engram_unavailable: %v", engramErr),
		}
	default:
		return HybridWriteResult{
			Status:        HybridWriteBlocked,
			EngramOK:      false,
			FilesystemOK:  false,
			EngramErr:     engramErr,
			FsErr:         fsErr,
			BlockedReason: fmt.Sprintf("both_stores_failed: engram=%v fs=%v", engramErr, fsErr),
		}
	}
}

// IsComplete returns true only when both stores succeeded.
// Use this to gate advancing to the next SDD phase.
func (r HybridWriteResult) IsComplete() bool {
	return r.Status == HybridWriteCompleted
}

// ResultContractStatus maps the HybridWriteResult to the result-contract v0.3 status field.
func (r HybridWriteResult) ResultContractStatus() string {
	switch r.Status {
	case HybridWriteCompleted:
		return "completed"
	case HybridWritePartial:
		return "partially_completed"
	default:
		return "blocked"
	}
}
```

---

## Paso 2: Implementar `AtomicFilesystemWrite` helper

**Archivo a modificar:** `internal/components/openspec/hybrid_writer.go`

**Acción:** Modificar — añadir al final del archivo

```go
// AtomicFilesystemWrite is a FilesystemWriteFn implementation that writes
// content to path atomically via tmp+rename. Creates parent directories if needed.
// Safe to use as HybridWriter.FilesystemFn.
func AtomicFilesystemWrite(ctx context.Context, path, content string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create parent dirs for %s: %w", path, err)
	}

	tmp := fmt.Sprintf("%s.%d.%d.tmp", path, os.Getpid(), time.Now().UnixNano())
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write tmp %s: %w", tmp, err)
	}

	if f, err := os.OpenFile(tmp, os.O_RDWR, 0o644); err == nil {
		_ = f.Sync()
		_ = f.Close()
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename %s → %s: %w", tmp, path, err)
	}
	return nil
}
```

**Añadir los imports necesarios al archivo:**

```go
import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)
```

---

## Paso 3: Crear tests del HybridWriter

**Archivo a crear:** `internal/components/openspec/hybrid_writer_test.go`

**Acción:** Crear

```go
package openspec

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var errEngram = errors.New("engram: connection refused")
var errFS = errors.New("filesystem: permission denied")

func successEngram(_ context.Context, _, _, _, _ string) error { return nil }
func failEngram(_ context.Context, _, _, _, _ string) error    { return errEngram }
func successFS(_ context.Context, _, _ string) error           { return nil }
func failFS(_ context.Context, _, _ string) error              { return errFS }

func validArtifact() HybridArtifact {
	return HybridArtifact{
		TopicKey:     "sdd/test-change/spec",
		FilePath:     "/tmp/test-spec.md",
		Content:      "# Spec content",
		Title:        "Test Spec",
		ArtifactType: "spec",
	}
}

func TestHybridWriter_BothSucceed_Completed(t *testing.T) {
	hw := &HybridWriter{EngramFn: successEngram, FilesystemFn: successFS}
	result := hw.Write(context.Background(), validArtifact())

	if result.Status != HybridWriteCompleted {
		t.Errorf("both succeed: expected Completed, got %q", result.Status)
	}
	if !result.IsComplete() {
		t.Error("IsComplete() should be true when both succeed")
	}
	if result.ResultContractStatus() != "completed" {
		t.Errorf("ResultContractStatus: got %q, want completed", result.ResultContractStatus())
	}
}

func TestHybridWriter_EngramFails_Partial(t *testing.T) {
	hw := &HybridWriter{EngramFn: failEngram, FilesystemFn: successFS}
	result := hw.Write(context.Background(), validArtifact())

	if result.Status != HybridWritePartial {
		t.Errorf("engram fails: expected Partial, got %q", result.Status)
	}
	if result.IsComplete() {
		t.Error("IsComplete() must be false when Engram failed")
	}
	if result.FilesystemOK != true {
		t.Error("FilesystemOK should be true")
	}
	if result.EngramOK != false {
		t.Error("EngramOK should be false")
	}
	if !strings.Contains(result.BlockedReason, "engram_unavailable") {
		t.Errorf("BlockedReason should mention engram_unavailable: %q", result.BlockedReason)
	}
}

func TestHybridWriter_FsFails_Partial(t *testing.T) {
	hw := &HybridWriter{EngramFn: successEngram, FilesystemFn: failFS}
	result := hw.Write(context.Background(), validArtifact())

	if result.Status != HybridWritePartial {
		t.Errorf("fs fails: expected Partial, got %q", result.Status)
	}
	if result.EngramOK != true {
		t.Error("EngramOK should be true")
	}
	if result.FilesystemOK != false {
		t.Error("FilesystemOK should be false")
	}
}

func TestHybridWriter_BothFail_Blocked(t *testing.T) {
	hw := &HybridWriter{EngramFn: failEngram, FilesystemFn: failFS}
	result := hw.Write(context.Background(), validArtifact())

	if result.Status != HybridWriteBlocked {
		t.Errorf("both fail: expected Blocked, got %q", result.Status)
	}
	if result.IsComplete() {
		t.Error("IsComplete() must be false when both fail")
	}
	if result.ResultContractStatus() != "blocked" {
		t.Errorf("ResultContractStatus: got %q, want blocked", result.ResultContractStatus())
	}
}

func TestHybridWriter_EmptyTopicKey_Blocked(t *testing.T) {
	hw := &HybridWriter{EngramFn: successEngram, FilesystemFn: successFS}
	artifact := validArtifact()
	artifact.TopicKey = ""
	result := hw.Write(context.Background(), artifact)

	if result.Status != HybridWriteBlocked {
		t.Errorf("empty TopicKey: expected Blocked, got %q", result.Status)
	}
}

func TestHybridWriter_EmptyContent_Blocked(t *testing.T) {
	hw := &HybridWriter{EngramFn: successEngram, FilesystemFn: successFS}
	artifact := validArtifact()
	artifact.Content = ""
	result := hw.Write(context.Background(), artifact)

	if result.Status != HybridWriteBlocked {
		t.Errorf("empty Content: expected Blocked, got %q", result.Status)
	}
}

func TestHybridWriter_ContextCancelled_Blocked(t *testing.T) {
	slowEngram := func(ctx context.Context, _, _, _, _ string) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			return nil
		}
	}
	hw := &HybridWriter{EngramFn: slowEngram, FilesystemFn: failFS, Timeout: 50 * time.Millisecond}
	result := hw.Write(context.Background(), validArtifact())

	if result.IsComplete() {
		t.Error("should not be complete when context is cancelled by timeout")
	}
}

func TestAtomicFilesystemWrite_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "artifact.yaml")
	content := "version: 3.0\nchange_name: test\n"

	if err := AtomicFilesystemWrite(context.Background(), path, content); err != nil {
		t.Fatalf("AtomicFilesystemWrite: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(data) != content {
		t.Errorf("content mismatch: got %q, want %q", string(data), content)
	}
}

func TestAtomicFilesystemWrite_NoStaleTemp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "artifact.yaml")

	if err := AtomicFilesystemWrite(context.Background(), path, "content"); err != nil {
		t.Fatalf("write: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("stale tmp file found: %s", e.Name())
		}
	}
}
```

---

## Paso 4: Actualizar `persistence-contract.md` para eliminar el fallback silencioso

**Archivo a modificar:** `internal/assets/_shared/persistence-contract.md`

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```markdown
### hybrid
Write to BOTH Engram and `.atl/openspec/changes/{change}/`:
- mem_save artifact to Engram (primary)
- write artifact to filesystem (secondary)
- If Engram unavailable: fall back to `none` mode silently, note in return envelope
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```markdown
### hybrid (v0.3 — transactional write)

Write to BOTH Engram and `.atl/openspec/changes/{change}/` atomically:

1. Execute both writes **in parallel** (concurrent, not sequential).
2. Evaluate both results:

| Engram | Filesystem | Result Contract Status | Next Phase? |
|--------|------------|------------------------|-------------|
| ✓ | ✓ | `completed` | YES — advance |
| ✗ | ✓ | `partially_completed` | NO — user decides |
| ✓ | ✗ | `partially_completed` | NO — user decides |
| ✗ | ✗ | `blocked` | NO — blocked |

**REMOVED in v0.3**: `"fall back to none mode silently"` — this behavior violated the
hybrid postcondition and caused silent data loss. Sub-agents MUST NOT return `completed`
when either store write failed.

When `partially_completed`:
- Include in `blocked_reason`: which store succeeded and which failed
- Include in `risks`: "artifact_not_in_{engram|filesystem} — sync required"
- Orchestrator presents user options:
  - `[1]` Retry Engram sync: `/sdd-sync-engram {change}`
  - `[2]` Accept filesystem-only and continue (documents exception)

### Hybrid Read Protocol (v0.3)

When reading artifacts in hybrid mode:

1. `mem_search(query)` → get IDs from Engram
2. `mem_get_observation(id)` with explicit timeout
   - **If timeout/error**: DO NOT use the 300-char preview as substitute
   - **If timeout/error**: IMMEDIATELY fall back to filesystem:
     `.atl/openspec/changes/{change}/{artifact}.yaml`
3. If filesystem fallback also fails:
   - Return `status: blocked`, `blocked_reason: "artifact_unavailable_both_stores"`

**CRITICAL**: A `mem_search` preview (300 chars) is NEVER a valid substitute for
`mem_get_observation`. A preview-only read is a MISS, not a hit.
```

---

## Verificación de Fase

```bash
# 1. Compilar el paquete openspec con el nuevo HybridWriter
go build ./internal/components/openspec/...

# 2. Tests del HybridWriter
go test ./internal/components/openspec/... -v -count=1 -run TestHybridWriter
go test ./internal/components/openspec/... -v -count=1 -run TestAtomicFilesystem

# 3. Race detector (escrituras paralelas del HybridWriter)
go test -race ./internal/components/openspec/... -count=1 \
  -run TestHybridWriter_BothSucceed

# 4. Verificar que IsComplete() es el único gate correcto
go test ./internal/components/openspec/... -v -count=1 \
  -run TestHybridWriter_EngramFails_Partial

# 5. Verificar que persistence-contract.md no tiene el fallback silencioso
python3 -c "
with open('internal/assets/_shared/persistence-contract.md') as f:
    content = f.read()
assert 'fall back to none mode silently' not in content, \
    'FAIL: silent fallback still present'
assert 'partially_completed' in content, \
    'FAIL: partially_completed not documented'
assert 'transactional' in content, \
    'FAIL: transactional not mentioned'
print('persistence-contract.md validated')
"

# 6. Compilar todo
go build ./...

# 7. Vet
go vet ./internal/components/openspec/...
```

