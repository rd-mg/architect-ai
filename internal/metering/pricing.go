package metering

// Pricing holds per-model token costs in USD per 1M tokens.
// CacheReadPer1M is the discounted rate applied when a prompt cache hit
// occurs; prompts not served from cache are billed at PromptPer1M.
//
// Values here are approximate and may be OUT OF DATE. The metering
// package reports savings as ESTIMATES; do not present them as invoices.
// The banner uses these for the "Est. savings" line only.
//
// To update: check the official pricing page for each provider and
// edit this table. The table is deliberately simple — no per-org
// contract pricing, no volume discounts.
type Pricing struct {
	PromptPer1M     float64
	CompletionPer1M float64
	CacheReadPer1M  float64
	CacheWritePer1M float64
}

// pricingTable maps a model identifier (as seen in API responses) to
// approximate pricing. Keys should match the strings that appear in
// the `model` field of provider responses.
//
// Aliases: many providers return a family name (e.g. "claude-sonnet-4")
// and a dated variant (e.g. "claude-sonnet-4-20250514"). We include
// both where common; LookupPricing does longest-prefix match.
var pricingTable = map[string]Pricing{
	// Anthropic Claude family — approximate Apr 2026 USD/1M
	"claude-opus-4-7":   {PromptPer1M: 15.00, CompletionPer1M: 75.00, CacheReadPer1M: 1.50, CacheWritePer1M: 18.75},
	"claude-opus-4-6":   {PromptPer1M: 15.00, CompletionPer1M: 75.00, CacheReadPer1M: 1.50, CacheWritePer1M: 18.75},
	"claude-opus-4":     {PromptPer1M: 15.00, CompletionPer1M: 75.00, CacheReadPer1M: 1.50, CacheWritePer1M: 18.75},
	"claude-sonnet-4-6": {PromptPer1M: 3.00, CompletionPer1M: 15.00, CacheReadPer1M: 0.30, CacheWritePer1M: 3.75},
	"claude-sonnet-4":   {PromptPer1M: 3.00, CompletionPer1M: 15.00, CacheReadPer1M: 0.30, CacheWritePer1M: 3.75},
	"claude-haiku-4-5":  {PromptPer1M: 0.80, CompletionPer1M: 4.00, CacheReadPer1M: 0.08, CacheWritePer1M: 1.00},
	"claude-haiku-4":    {PromptPer1M: 0.80, CompletionPer1M: 4.00, CacheReadPer1M: 0.08, CacheWritePer1M: 1.00},

	// Google Gemini — approximate
	"gemini-2.5-pro":   {PromptPer1M: 1.25, CompletionPer1M: 5.00, CacheReadPer1M: 0.3125, CacheWritePer1M: 0},
	"gemini-2.5-flash": {PromptPer1M: 0.075, CompletionPer1M: 0.30, CacheReadPer1M: 0.01875, CacheWritePer1M: 0},
	"gemini-2.0-pro":   {PromptPer1M: 1.25, CompletionPer1M: 5.00, CacheReadPer1M: 0.3125, CacheWritePer1M: 0},
	"gemini-2.0-flash": {PromptPer1M: 0.075, CompletionPer1M: 0.30, CacheReadPer1M: 0.01875, CacheWritePer1M: 0},

	// OpenAI (for Codex / Cursor when configured)
	"gpt-5":      {PromptPer1M: 2.50, CompletionPer1M: 10.00, CacheReadPer1M: 1.25, CacheWritePer1M: 0},
	"gpt-5-mini": {PromptPer1M: 0.25, CompletionPer1M: 2.00, CacheReadPer1M: 0.125, CacheWritePer1M: 0},
	"gpt-4o":     {PromptPer1M: 2.50, CompletionPer1M: 10.00, CacheReadPer1M: 1.25, CacheWritePer1M: 0},
	"o3":         {PromptPer1M: 15.00, CompletionPer1M: 60.00, CacheReadPer1M: 7.50, CacheWritePer1M: 0},
	"o3-mini":    {PromptPer1M: 1.10, CompletionPer1M: 4.40, CacheReadPer1M: 0.55, CacheWritePer1M: 0},

	// Alibaba Qwen
	"qwen-3-max":   {PromptPer1M: 1.60, CompletionPer1M: 6.40, CacheReadPer1M: 0.16, CacheWritePer1M: 0},
	"qwen-3-coder": {PromptPer1M: 0.80, CompletionPer1M: 3.20, CacheReadPer1M: 0.08, CacheWritePer1M: 0},
}

// LookupPricing returns the pricing for a model string.
// Uses longest-prefix match so that "claude-sonnet-4-20250514" resolves to
// "claude-sonnet-4" when the dated variant isn't explicitly listed.
func LookupPricing(model string) (Pricing, bool) {
	if p, ok := pricingTable[model]; ok {
		return p, true
	}

	// Longest-prefix match.
	var bestKey string
	for key := range pricingTable {
		if len(key) > len(bestKey) && hasPrefix(model, key) {
			bestKey = key
		}
	}
	if bestKey != "" {
		return pricingTable[bestKey], true
	}
	return Pricing{}, false
}

// SetPricing injects or overrides a pricing entry at runtime.
// Intended for users whose provider pricing differs from the defaults
// (contract pricing, enterprise rates). Not threadsafe; call before
// session start.
func SetPricing(model string, p Pricing) {
	pricingTable[model] = p
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
