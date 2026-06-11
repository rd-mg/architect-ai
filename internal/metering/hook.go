package metering

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Hook coordinates the shutdown-time banner print and optional Engram persistence.
// A single Hook is registered per process at startup; all agent adapters share it.
type Hook struct {
	mu      sync.Mutex
	stats   *SessionStats
	writer  io.Writer
	persist PersistFunc
	once    sync.Once
}

// PersistFunc is invoked at shutdown with the Engram-formatted content.
// Implementations typically call mem_save via the MCP client.
// Return error is logged but does NOT prevent the banner from printing.
type PersistFunc func(topicKey, content string) error

var (
	globalHook *Hook
	globalMu   sync.Mutex
)

// Register creates (or returns) the process-wide metering hook.
// Call exactly once, early in main(). Subsequent calls return the
// already-registered hook.
//
// Example:
//
//	func main() {
//	    hook := metering.Register("claude", sessionID, os.Stderr)
//	    defer hook.PrintBanner()
//	    // ...
//	}
func Register(agentID, sessionID string, w io.Writer) *Hook {
	globalMu.Lock()
	defer globalMu.Unlock()

	if globalHook != nil {
		return globalHook
	}

	h := &Hook{
		stats:  NewSessionStats(agentID, sessionID),
		writer: w,
	}
	globalHook = h

	// Install signal handler so ctrl+c also prints the banner.
	go h.watchSignals()

	return h
}

// Current returns the registered hook or nil if Register was never called.
// Agent adapters use this to fold UsageDelta into the shared stats.
func Current() *Hook {
	globalMu.Lock()
	defer globalMu.Unlock()
	return globalHook
}

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

// WithPersist attaches an Engram persistence callback. Call after Register.
// If not called, shutdown only prints the banner; no Engram write happens.
func (h *Hook) WithPersist(project string, fn PersistFunc) *Hook {
	if h == nil {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.persist = func(_, _ string) error {
		// Closure captures project; wrap caller's fn with the topic key.
		topicKey := fmt.Sprintf("metering/%s/%s", project, h.stats.SessionID)
		content := h.stats.EngramContent(project)
		return fn(topicKey, content)
	}
	return h
}

// PrintBanner renders and writes the session summary to the hook's writer.
// Idempotent — only fires once per hook even if called multiple times
// (defer in main + signal handler both call it; we only want one banner).
func (h *Hook) PrintBanner() {
	if h == nil {
		return
	}
	h.once.Do(func() {
		h.mu.Lock()
		stats := h.stats
		persist := h.persist
		w := h.writer
		h.mu.Unlock()

		if stats.RequestCount == 0 {
			// No requests recorded — skip banner to avoid noise.
			return
		}

		banner := stats.RenderExitBanner()
		_, _ = fmt.Fprint(w, banner)

		if persist != nil {
			if err := persist("", ""); err != nil {
				fmt.Fprintf(os.Stderr, "metering: Engram persistence failed: %v\n", err)
			}
		}
	})
}

// watchSignals runs in a goroutine; prints the banner on SIGINT/SIGTERM.
// The main-thread defer PrintBanner will no-op thanks to sync.Once.
func (h *Hook) watchSignals() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	h.PrintBanner()
	// Do NOT os.Exit here — let the program exit naturally. The signal
	// will be re-delivered and handled by the default handler if the
	// main thread hasn't set up its own handling.
	signal.Reset(syscall.SIGINT, syscall.SIGTERM)
}

// StatsSnapshot is a copy of session stats without the mutex.
type StatsSnapshot struct {
	AgentID      string
	SessionStart time.Time
	SessionID    string
	PromptTokens     int64
	CompletionTokens int64
	CachedTokens     int64
	CacheCreated     int64
	RequestCount     int
}

// Stats returns the current stats snapshot (copy).
// Intended for debugging / test introspection.
func (h *Hook) Stats() StatsSnapshot {
	if h == nil {
		return StatsSnapshot{}
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return StatsSnapshot{
		AgentID:          h.stats.AgentID,
		SessionStart:     h.stats.SessionStart,
		SessionID:        h.stats.SessionID,
		PromptTokens:     h.stats.PromptTokens,
		CompletionTokens: h.stats.CompletionTokens,
		CachedTokens:     h.stats.CachedTokens,
		CacheCreated:     h.stats.CacheCreated,
		RequestCount:     h.stats.RequestCount,
	}
}
