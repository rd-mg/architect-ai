package metering

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSessionStats_AddAndTotal(t *testing.T) {
	s := NewSessionStats("claude", "session-1")

	s.Add(UsageDelta{PromptTokens: 1000, CompletionTokens: 500, CachedTokens: 300, Model: "claude-sonnet-4"})
	s.Add(UsageDelta{PromptTokens: 2000, CompletionTokens: 1000, CachedTokens: 1500, Model: "claude-sonnet-4"})

	if got := s.TotalTokens(); got != 4500 {
		t.Errorf("TotalTokens = %d, want 4500", got)
	}
	if got := s.RequestCount; got != 2 {
		t.Errorf("RequestCount = %d, want 2", got)
	}
}

func TestSessionStats_CacheHitRatio(t *testing.T) {
	s := NewSessionStats("claude", "session-1")

	// Empty session → 0
	if got := s.CacheHitRatio(); got != 0 {
		t.Errorf("empty CacheHitRatio = %v, want 0", got)
	}

	s.Add(UsageDelta{PromptTokens: 1000, CachedTokens: 400, Model: "claude-sonnet-4"})
	if got := s.CacheHitRatio(); got != 0.4 {
		t.Errorf("CacheHitRatio = %v, want 0.4", got)
	}

	s.Add(UsageDelta{PromptTokens: 1000, CachedTokens: 600, Model: "claude-sonnet-4"})
	if got := s.CacheHitRatio(); got != 0.5 {
		t.Errorf("after second add CacheHitRatio = %v, want 0.5", got)
	}
}

func TestSessionStats_EstimatedSavings(t *testing.T) {
	s := NewSessionStats("claude", "session-1")

	// 1M cached tokens on sonnet-4: saving = (3.00 - 0.30) * 1_000_000 / 1_000_000 = 2.70
	s.Add(UsageDelta{PromptTokens: 1_000_000, CachedTokens: 1_000_000, Model: "claude-sonnet-4"})

	got := s.EstimatedSavingsUSD()
	want := 2.70
	if got < want-0.01 || got > want+0.01 {
		t.Errorf("EstimatedSavingsUSD = %v, want ~%v", got, want)
	}
}

func TestSessionStats_UnknownModelNoSavings(t *testing.T) {
	s := NewSessionStats("claude", "session-1")
	s.Add(UsageDelta{PromptTokens: 1000, CachedTokens: 500, Model: "future-model-xyz"})

	if got := s.EstimatedSavingsUSD(); got != 0 {
		t.Errorf("unknown model savings = %v, want 0", got)
	}
}

func TestRenderExitBanner_ContainsKeyFields(t *testing.T) {
	s := NewSessionStats("gemini", "session-2")
	s.SessionStart = time.Now().Add(-90 * time.Second)
	s.Add(UsageDelta{PromptTokens: 50000, CompletionTokens: 12000, CachedTokens: 20000, Model: "gemini-2.5-pro"})

	banner := s.RenderExitBanner()

	for _, want := range []string{"gemini", "62,000", "20,000", "(40%)", "Session summary"} {
		if !strings.Contains(banner, want) {
			t.Errorf("banner missing %q\nbanner was:\n%s", want, banner)
		}
	}
}

func TestRenderExitBanner_ZeroUsageOmitsCacheLines(t *testing.T) {
	s := NewSessionStats("claude", "session-3")
	s.Add(UsageDelta{PromptTokens: 1000, CompletionTokens: 500, Model: "claude-sonnet-4"})

	banner := s.RenderExitBanner()

	if strings.Contains(banner, "From cache") {
		t.Errorf("banner should not show 'From cache' when CachedTokens=0:\n%s", banner)
	}
	if strings.Contains(banner, "Cache created") {
		t.Errorf("banner should not show 'Cache created' when CacheCreated=0:\n%s", banner)
	}
	if strings.Contains(banner, "Est. savings") {
		t.Errorf("banner should not show 'Est. savings' when there are no savings:\n%s", banner)
	}
}

func TestSessionStats_ConcurrentAdd(t *testing.T) {
	s := NewSessionStats("claude", "session-4")

	var wg sync.WaitGroup
	const N = 100
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Add(UsageDelta{PromptTokens: 10, CompletionTokens: 5, CachedTokens: 2, Model: "claude-sonnet-4"})
		}()
	}
	wg.Wait()

	if got := s.RequestCount; got != N {
		t.Errorf("RequestCount after concurrent = %d, want %d", got, N)
	}
	if got := s.TotalTokens(); got != int64(N*15) {
		t.Errorf("TotalTokens after concurrent = %d, want %d", got, N*15)
	}
	if got := s.CachedTokens; got != int64(N*2) {
		t.Errorf("CachedTokens after concurrent = %d, want %d", got, N*2)
	}
}

func TestLookupPricing_LongestPrefix(t *testing.T) {
	// "claude-sonnet-4-20250514" should resolve to "claude-sonnet-4"
	p, ok := LookupPricing("claude-sonnet-4-20250514")
	if !ok {
		t.Fatal("claude-sonnet-4-20250514 did not resolve")
	}
	if p.PromptPer1M != 3.00 {
		t.Errorf("prefix match returned wrong pricing: %v", p)
	}
}

func TestLookupPricing_ExactBeatsPrefix(t *testing.T) {
	p, ok := LookupPricing("claude-opus-4-7")
	if !ok {
		t.Fatal("exact lookup failed")
	}
	if p.PromptPer1M != 15.00 {
		t.Errorf("exact match returned wrong pricing: %v", p)
	}
}

func TestLookupPricing_UnknownReturnsFalse(t *testing.T) {
	_, ok := LookupPricing("some-model-that-does-not-exist")
	if ok {
		t.Error("unknown model should return ok=false")
	}
}

func TestHook_PrintBannerIdempotent(t *testing.T) {
	// Reset global hook for test isolation
	globalMu.Lock()
	globalHook = nil
	globalMu.Unlock()

	var sink strings.Builder
	h := Register("claude", "session-5", &sink)
	h.Record(UsageDelta{PromptTokens: 100, CompletionTokens: 50, Model: "claude-sonnet-4"})

	h.PrintBanner()
	first := sink.Len()

	h.PrintBanner() // Should be no-op due to sync.Once
	second := sink.Len()

	if first != second {
		t.Errorf("PrintBanner not idempotent: first=%d second=%d", first, second)
	}
	if first == 0 {
		t.Error("PrintBanner produced no output")
	}
}

func TestHumanizeTokens(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0"},
		{42, "42"},
		{999, "999"},
		{1000, "1,000"},
		{10000, "10,000"},
		{1234567, "1,234,567"},
	}
	for _, tc := range cases {
		if got := humanizeTokens(tc.in); got != tc.want {
			t.Errorf("humanizeTokens(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
