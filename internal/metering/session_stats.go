// Package metering accumulates per-session token usage across all agents
// and renders a summary banner at session end. Supports prompt caching
// metrics (cache_read / cache_creation tokens) where the underlying
// provider exposes them.
package metering

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// UsageDelta is a single API response's token accounting.
// All fields are absolute counts for THAT response, not cumulative.
type UsageDelta struct {
	PromptTokens     int64
	CompletionTokens int64
	CachedTokens     int64 // cache_read_input_tokens / cachedContentTokenCount
	CacheCreated     int64 // cache_creation_input_tokens
	Model            string
}

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

type modelStats struct {
	PromptTokens     int64
	CompletionTokens int64
	CachedTokens     int64
	CacheCreated     int64
	Requests         int
}

// NewSessionStats returns a fresh SessionStats for the given agent.
func NewSessionStats(agentID, sessionID string) *SessionStats {
	return &SessionStats{
		AgentID:      agentID,
		SessionID:    sessionID,
		SessionStart: time.Now(),
		perModel:     make(map[string]*modelStats),
	}
}

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

// TotalTokens returns prompt + completion (cached and cache-created are part of prompt).
func (s *SessionStats) TotalTokens() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.PromptTokens + s.CompletionTokens
}

// CacheHitRatio returns cached / prompt tokens as a fraction in [0, 1].
// Returns 0 if no prompt tokens were recorded.
func (s *SessionStats) CacheHitRatio() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.PromptTokens == 0 {
		return 0
	}
	return float64(s.CachedTokens) / float64(s.PromptTokens)
}

// EstimatedSavingsUSD computes the dollar amount saved by prompt caching,
// summing across models. Uses the pricing table in pricing.go.
// Returns 0 if no cached tokens were recorded or no model is priced.
func (s *SessionStats) EstimatedSavingsUSD() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	var savings float64
	for model, ms := range s.perModel {
		p, ok := LookupPricing(model)
		if !ok {
			continue
		}
		// Cached tokens are billed at the cache-read rate instead of the full prompt rate.
		// Savings = CachedTokens * (PromptRate - CacheReadRate) / 1_000_000
		if ms.CachedTokens > 0 && p.CacheReadPer1M < p.PromptPer1M {
			savings += float64(ms.CachedTokens) * (p.PromptPer1M - p.CacheReadPer1M) / 1_000_000.0
		}
	}
	return savings
}

// RenderExitBanner returns the multi-line summary banner shown at session end.
// Safe to call on a zero-usage session (renders a minimal banner).
func (s *SessionStats) RenderExitBanner() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	duration := time.Since(s.SessionStart).Round(time.Second)
	total := s.PromptTokens + s.CompletionTokens
	cachePct := 0.0
	if s.PromptTokens > 0 {
		cachePct = 100.0 * float64(s.CachedTokens) / float64(s.PromptTokens)
	}

	// Compute savings outside the lock — release, compute, reacquire is messy; inline it.
	var savings float64
	for model, ms := range s.perModel {
		p, ok := LookupPricing(model)
		if !ok {
			continue
		}
		if ms.CachedTokens > 0 && p.CacheReadPer1M < p.PromptPer1M {
			savings += float64(ms.CachedTokens) * (p.PromptPer1M - p.CacheReadPer1M) / 1_000_000.0
		}
	}

	var b strings.Builder
	bar := "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	b.WriteString(bar)
	b.WriteString("\n")
	fmt.Fprintf(&b, "Session summary (%s) — %s\n", s.AgentID, duration)
	fmt.Fprintf(&b, "  Requests:         %d\n", s.RequestCount)
	fmt.Fprintf(&b, "  Total tokens:     %s\n", humanizeTokens(total))
	if s.CachedTokens > 0 {
		fmt.Fprintf(&b, "  From cache:       %s (%.0f%%)\n", humanizeTokens(s.CachedTokens), cachePct)
	}
	if s.CacheCreated > 0 {
		fmt.Fprintf(&b, "  Cache created:    %s\n", humanizeTokens(s.CacheCreated))
	}
	if savings > 0 {
		fmt.Fprintf(&b, "  Est. savings:     ~$%.2f\n", savings)
	}
	b.WriteString(bar)
	b.WriteString("\n")
	return b.String()
}

// humanizeTokens formats integers with thousands separators.
func humanizeTokens(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	in := fmt.Sprintf("%d", n)
	var out []byte
	for i, c := range []byte(in) {
		if i > 0 && (len(in)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	return string(out)
}

// EngramContent returns a compact, human-readable content block suitable
// for persisting to Engram under `metering/{project}/{session-id}`.
func (s *SessionStats) EngramContent(project string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var b strings.Builder
	fmt.Fprintf(&b, "Agent: %s\n", s.AgentID)
	fmt.Fprintf(&b, "Project: %s\n", project)
	fmt.Fprintf(&b, "SessionID: %s\n", s.SessionID)
	fmt.Fprintf(&b, "Start: %s\n", s.SessionStart.Format(time.RFC3339))
	fmt.Fprintf(&b, "Duration: %s\n", time.Since(s.SessionStart).Round(time.Second))
	fmt.Fprintf(&b, "Requests: %d\n", s.RequestCount)
	fmt.Fprintf(&b, "PromptTokens: %d\n", s.PromptTokens)
	fmt.Fprintf(&b, "CompletionTokens: %d\n", s.CompletionTokens)
	fmt.Fprintf(&b, "CachedTokens: %d\n", s.CachedTokens)
	fmt.Fprintf(&b, "CacheCreated: %d\n", s.CacheCreated)
	fmt.Fprintf(&b, "PerModel:\n")
	for model, ms := range s.perModel {
		fmt.Fprintf(&b, "  %s: req=%d prompt=%d completion=%d cached=%d cache_created=%d\n",
			model, ms.Requests, ms.PromptTokens, ms.CompletionTokens, ms.CachedTokens, ms.CacheCreated)
	}
	return b.String()
}
