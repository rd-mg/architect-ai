# Fase 3: Metering con Desglose por Fase SDD

**Objetivo:** Resolver OBS-01 (metering ciego en runtime) y OBS-06 (sin desglose por fase). Extiende `SessionStats` y `Hook` con tracking por fase SDD sin romper ninguna API existente. Todos los callers de `Add()`, `Record()`, `PrintBanner()` siguen funcionando sin cambios.

---

## Paso 1: Extender `SessionStats` con tracking por fase

**Archivo a modificar:** `internal/metering/session_stats.go`

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
// SessionStats accumulates usage across a single CLI/IDE session.
type SessionStats struct {
	mu sync.Mutex

	AgentID      string
	SessionStart time.Time
	SessionID    string

	PromptTokens     int64
	CompletionTokens int64
	CachedTokens     int64
	CacheCreated     int64
	RequestCount     int

	// Per-model breakdown for pricing calculations.
	perModel map[string]*modelStats
}
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
// PhaseRecord tracks token usage for a single SDD phase within the session.
type PhaseRecord struct {
	Phase            string
	StartedAt        time.Time
	CompletedAt      time.Time
	PromptTokens     int64
	CompletionTokens int64
	CachedTokens     int64
}

// BudgetAlert is called when a phase exceeds its token budget.
// Runs in a separate goroutine; must be goroutine-safe.
type BudgetAlert func(phase string, used, limit int64)

// SessionStats accumulates usage across a single CLI/IDE session.
type SessionStats struct {
	mu sync.Mutex

	AgentID      string
	SessionStart time.Time
	SessionID    string

	PromptTokens     int64
	CompletionTokens int64
	CachedTokens     int64
	CacheCreated     int64
	RequestCount     int

	// Per-model breakdown for pricing calculations.
	perModel map[string]*modelStats

	// Per-phase breakdown for SDD cost observability (v0.3+).
	perPhase     map[string]*PhaseRecord
	currentPhase string
	phaseLimit   int64       // 0 = no limit
	onBudget     BudgetAlert // nil = no alert
}
```

---

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
// NewSessionStats returns a fresh SessionStats for the given agent.
func NewSessionStats(agentID, sessionID string) *SessionStats {
	return &SessionStats{
		AgentID:      agentID,
		SessionID:    sessionID,
		SessionStart: time.Now(),
		perModel:     make(map[string]*modelStats),
	}
}
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
// NewSessionStats returns a fresh SessionStats for the given agent.
func NewSessionStats(agentID, sessionID string) *SessionStats {
	return &SessionStats{
		AgentID:      agentID,
		SessionID:    sessionID,
		SessionStart: time.Now(),
		perModel:     make(map[string]*modelStats),
		perPhase:     make(map[string]*PhaseRecord),
	}
}

// WithBudgetAlert configures a per-phase token budget and the callback
// invoked (in a new goroutine) when any phase exceeds it.
// Call before any RecordPhaseStart. Not threadsafe with concurrent Add calls;
// call at session initialization only.
func (s *SessionStats) WithBudgetAlert(limitPerPhase int64, fn BudgetAlert) *SessionStats {
	s.phaseLimit = limitPerPhase
	s.onBudget = fn
	return s
}

// RecordPhaseStart marks the beginning of a new SDD phase for per-phase metering.
// Subsequent Add() calls attribute tokens to this phase until RecordPhaseEnd().
// Safe for concurrent callers.
func (s *SessionStats) RecordPhaseStart(phase string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.perPhase == nil {
		s.perPhase = make(map[string]*PhaseRecord)
	}
	s.currentPhase = phase
	if _, exists := s.perPhase[phase]; !exists {
		s.perPhase[phase] = &PhaseRecord{
			Phase:     phase,
			StartedAt: time.Now(),
		}
	}
}

// RecordPhaseEnd marks the completion of an SDD phase.
// Safe for concurrent callers.
func (s *SessionStats) RecordPhaseEnd(phase string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if pr, ok := s.perPhase[phase]; ok {
		if pr.CompletedAt.IsZero() {
			pr.CompletedAt = time.Now()
		}
	}
	if s.currentPhase == phase {
		s.currentPhase = ""
	}
}

// PhaseBreakdown returns a snapshot map of phase → total tokens (prompt + completion).
// Safe for concurrent callers.
func (s *SessionStats) PhaseBreakdown() map[string]int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]int64, len(s.perPhase))
	for phase, pr := range s.perPhase {
		out[phase] = pr.PromptTokens + pr.CompletionTokens
	}
	return out
}

// EngramPhaseCostContent returns a compact, human-readable block suitable for
// persisting to Engram under "metering/{project}/{change}/phase-costs".
// Call at session end, after all phases have been recorded.
func (s *SessionStats) EngramPhaseCostContent(project, change, sessionID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	total := s.PromptTokens + s.CompletionTokens
	cacheRate := 0.0
	if s.PromptTokens > 0 {
		cacheRate = 100.0 * float64(s.CachedTokens) / float64(s.PromptTokens)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "project: %s\n", project)
	fmt.Fprintf(&b, "change: %s\n", change)
	fmt.Fprintf(&b, "session_id: %s\n", sessionID)
	fmt.Fprintf(&b, "total_tokens: %d\n", total)
	fmt.Fprintf(&b, "cache_hit_rate_pct: %.1f\n", cacheRate)
	fmt.Fprintf(&b, "phases:\n")
	for phase, pr := range s.perPhase {
		phaseTotal := pr.PromptTokens + pr.CompletionTokens
		dur := ""
		if !pr.CompletedAt.IsZero() {
			dur = pr.CompletedAt.Sub(pr.StartedAt).Round(time.Second).String()
		}
		fmt.Fprintf(&b, "  %s: {tokens: %d, cached: %d, duration: %q}\n",
			phase, phaseTotal, pr.CachedTokens, dur)
	}
	return b.String()
}
```

---

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
// Add folds a single UsageDelta into the running totals.
// Safe for concurrent callers.
func (s *SessionStats) Add(d UsageDelta) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.PromptTokens += d.PromptTokens
	s.CompletionTokens += d.CompletionTokens
	s.CachedTokens += d.CachedTokens
	s.CacheCreated += d.CacheCreated
	s.RequestCount++

	model := d.Model
	if model == "" {
		model = "unknown"
	}
	ms, ok := s.perModel[model]
	if !ok {
		ms = &modelStats{}
		s.perModel[model] = ms
	}
	ms.PromptTokens += d.PromptTokens
	ms.CompletionTokens += d.CompletionTokens
	ms.CachedTokens += d.CachedTokens
	ms.CacheCreated += d.CacheCreated
	ms.Requests++
}
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
// Add folds a single UsageDelta into the running totals.
// Also attributes tokens to the current SDD phase (if any).
// Safe for concurrent callers.
func (s *SessionStats) Add(d UsageDelta) {
	s.mu.Lock()

	s.PromptTokens += d.PromptTokens
	s.CompletionTokens += d.CompletionTokens
	s.CachedTokens += d.CachedTokens
	s.CacheCreated += d.CacheCreated
	s.RequestCount++

	model := d.Model
	if model == "" {
		model = "unknown"
	}
	ms, ok := s.perModel[model]
	if !ok {
		ms = &modelStats{}
		s.perModel[model] = ms
	}
	ms.PromptTokens += d.PromptTokens
	ms.CompletionTokens += d.CompletionTokens
	ms.CachedTokens += d.CachedTokens
	ms.CacheCreated += d.CacheCreated
	ms.Requests++

	var phaseTotal int64
	var shouldAlert bool
	if s.currentPhase != "" {
		pr, exists := s.perPhase[s.currentPhase]
		if !exists {
			pr = &PhaseRecord{Phase: s.currentPhase, StartedAt: time.Now()}
			s.perPhase[s.currentPhase] = pr
		}
		pr.PromptTokens += d.PromptTokens
		pr.CompletionTokens += d.CompletionTokens
		pr.CachedTokens += d.CachedTokens
		phaseTotal = pr.PromptTokens + pr.CompletionTokens
		if s.phaseLimit > 0 && phaseTotal > s.phaseLimit && s.onBudget != nil {
			shouldAlert = true
		}
	}
	phase := s.currentPhase
	limit := s.phaseLimit
	alert := s.onBudget

	s.mu.Unlock()

	if shouldAlert {
		go alert(phase, phaseTotal, limit)
	}
}
```

---

## Paso 2: Extender `hook.go` con anotación de fase

**Archivo a modificar:** `internal/metering/hook.go`

**Acción:** Modificar

**Código a reemplazar — BUSCAR EXACTAMENTE:**
```go
// Record folds a single UsageDelta into the session stats.
// Safe for concurrent callers; no-op if the hook was not registered.
func (h *Hook) Record(d UsageDelta) {
	if h == nil {
		return
	}
	h.mu.Lock()
	stats := h.stats
	h.mu.Unlock()
	stats.Add(d)
}
```

**Código de reemplazo — INSERTAR EXACTAMENTE:**
```go
// Record folds a single UsageDelta into the session stats.
// Safe for concurrent callers; no-op if the hook was not registered.
func (h *Hook) Record(d UsageDelta) {
	if h == nil {
		return
	}
	h.mu.Lock()
	stats := h.stats
	h.mu.Unlock()
	stats.Add(d)
}

// PhaseStart marks the beginning of an SDD phase for per-phase token tracking.
// Call from sdd-orchestrator adapters before delegating to sub-agents.
// No-op if the hook was not registered.
func (h *Hook) PhaseStart(phase string) {
	if h == nil {
		return
	}
	h.mu.Lock()
	stats := h.stats
	h.mu.Unlock()
	if stats != nil {
		stats.RecordPhaseStart(phase)
	}
}

// PhaseEnd marks the completion of an SDD phase.
// No-op if the hook was not registered.
func (h *Hook) PhaseEnd(phase string) {
	if h == nil {
		return
	}
	h.mu.Lock()
	stats := h.stats
	h.mu.Unlock()
	if stats != nil {
		stats.RecordPhaseEnd(phase)
	}
}

// PhaseBreakdown returns token usage per phase. Used by orchestrator adapters
// to persist cost records to Engram at session end.
func (h *Hook) PhaseBreakdown() map[string]int64 {
	if h == nil {
		return nil
	}
	h.mu.Lock()
	stats := h.stats
	h.mu.Unlock()
	if stats == nil {
		return nil
	}
	return stats.PhaseBreakdown()
}
```

---

## Paso 3: Crear tests de metering con fase

**Archivo a crear:** `internal/metering/phase_test.go`

**Acción:** Crear

```go
package metering

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRecordPhaseStart_TracksTokens(t *testing.T) {
	s := NewSessionStats("sdd-orchestrator", "test-session-1")
	s.RecordPhaseStart("sdd-explore")
	s.Add(UsageDelta{Model: "claude-sonnet-4", PromptTokens: 5000, CompletionTokens: 1000})
	s.RecordPhaseEnd("sdd-explore")

	s.RecordPhaseStart("sdd-apply")
	s.Add(UsageDelta{Model: "claude-sonnet-4", PromptTokens: 15000, CompletionTokens: 8000})
	s.RecordPhaseEnd("sdd-apply")

	bd := s.PhaseBreakdown()
	if bd["sdd-explore"] != 6000 {
		t.Errorf("sdd-explore: got %d, want 6000", bd["sdd-explore"])
	}
	if bd["sdd-apply"] != 23000 {
		t.Errorf("sdd-apply: got %d, want 23000", bd["sdd-apply"])
	}
	if s.TotalTokens() != 29000 {
		t.Errorf("TotalTokens: got %d, want 29000", s.TotalTokens())
	}
}

func TestRecordPhaseStart_NoPhase_GlobalStillAccumulates(t *testing.T) {
	s := NewSessionStats("general", "test-session-2")
	s.Add(UsageDelta{Model: "claude-haiku-4", PromptTokens: 1000, CompletionTokens: 200})
	if s.TotalTokens() != 1200 {
		t.Errorf("global total: got %d, want 1200", s.TotalTokens())
	}
	bd := s.PhaseBreakdown()
	if len(bd) != 0 {
		t.Errorf("no phases started — breakdown should be empty, got %v", bd)
	}
}

func TestBudgetAlert_FiresWhenExceeded(t *testing.T) {
	alerted := make(chan string, 1)
	s := NewSessionStats("sdd-orchestrator", "test-session-3")
	s.WithBudgetAlert(5000, func(phase string, used, limit int64) {
		alerted <- phase
	})

	s.RecordPhaseStart("sdd-apply")
	s.Add(UsageDelta{Model: "claude-sonnet-4", PromptTokens: 6000})

	select {
	case phase := <-alerted:
		if phase != "sdd-apply" {
			t.Errorf("expected alert for sdd-apply, got %s", phase)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("budget alert should have fired within 200ms")
	}
}

func TestBudgetAlert_DoesNotFireBelowLimit(t *testing.T) {
	fired := false
	s := NewSessionStats("sdd-orchestrator", "test-session-4")
	s.WithBudgetAlert(10000, func(phase string, used, limit int64) {
		fired = true
	})

	s.RecordPhaseStart("sdd-spec")
	s.Add(UsageDelta{Model: "claude-sonnet-4", PromptTokens: 3000, CompletionTokens: 500})
	time.Sleep(50 * time.Millisecond)
	if fired {
		t.Error("alert should not fire when under budget limit")
	}
}

func TestEngramPhaseCostContent_Format(t *testing.T) {
	s := NewSessionStats("sdd-orchestrator", "test-session-5")
	s.RecordPhaseStart("sdd-spec")
	s.Add(UsageDelta{Model: "claude-sonnet-4", PromptTokens: 3000, CompletionTokens: 800, CachedTokens: 500})
	s.RecordPhaseEnd("sdd-spec")

	content := s.EngramPhaseCostContent("my-project", "add-payment", "test-session-5")

	for _, required := range []string{
		"project: my-project",
		"change: add-payment",
		"session_id: test-session-5",
		"total_tokens: 3800",
		"sdd-spec:",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("EngramPhaseCostContent missing %q\nContent:\n%s", required, content)
		}
	}
}

func TestAdd_ConcurrentSafe(t *testing.T) {
	s := NewSessionStats("agent", "concurrent-test")
	s.RecordPhaseStart("sdd-apply")

	const goroutines = 50
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Add(UsageDelta{Model: "claude-sonnet-4", PromptTokens: 100, CompletionTokens: 20})
		}()
	}
	wg.Wait()

	if s.TotalTokens() != int64(goroutines)*120 {
		t.Errorf("concurrent Add: got %d, want %d", s.TotalTokens(), int64(goroutines)*120)
	}
	bd := s.PhaseBreakdown()
	if bd["sdd-apply"] != int64(goroutines)*120 {
		t.Errorf("phase breakdown concurrent: got %d, want %d", bd["sdd-apply"], int64(goroutines)*120)
	}
}

func TestHookPhaseStart_NoopWhenNil(t *testing.T) {
	var h *Hook
	h.PhaseStart("sdd-verify")
	h.PhaseEnd("sdd-verify")
	bd := h.PhaseBreakdown()
	if bd != nil {
		t.Error("nil hook PhaseBreakdown should return nil")
	}
}

func TestHookPhaseStart_Works(t *testing.T) {
	h := &Hook{
		stats: NewSessionStats("sdd-orchestrator", "hook-test-1"),
	}
	h.PhaseStart("sdd-spec")
	h.Record(UsageDelta{Model: "claude-sonnet-4", PromptTokens: 2000})
	h.PhaseEnd("sdd-spec")

	bd := h.PhaseBreakdown()
	if bd["sdd-spec"] != 2000 {
		t.Errorf("hook phase breakdown: got %d, want 2000", bd["sdd-spec"])
	}
}
```

---

## Verificación de Fase

```bash
# 1. Compilar el paquete metering
go build ./internal/metering/...

# 2. Tests existentes de metering (asegurar no regresiones)
go test ./internal/metering/... -v -count=1

# 3. Tests nuevos de fase
go test ./internal/metering/... -v -count=1 -run TestRecordPhase
go test ./internal/metering/... -v -count=1 -run TestBudgetAlert
go test ./internal/metering/... -v -count=1 -run TestEngram
go test ./internal/metering/... -v -count=1 -run TestHookPhase

# 4. Race detector
go test -race ./internal/metering/... -count=1

# 5. Compilar todo para asegurar que los callers de Hook no se rompen
go build ./...
```

