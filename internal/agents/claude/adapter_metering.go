package claude

import (
	"encoding/json"

	"github.com/rd-mg/architect-ai/internal/metering"
)

// claudeUsageEnvelope mirrors the `usage` block in Claude Messages API responses.
//
//	{
//	  "usage": {
//	    "input_tokens": 2095,
//	    "cache_creation_input_tokens": 12345,
//	    "cache_read_input_tokens": 4567,
//	    "output_tokens": 503
//	  },
//	  "model": "claude-sonnet-4-20250514",
//	  ...
//	}
type claudeUsageEnvelope struct {
	Usage struct {
		InputTokens              int64 `json:"input_tokens"`
		OutputTokens             int64 `json:"output_tokens"`
		CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
		CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
	} `json:"usage"`
	Model string `json:"model"`
}

// ExtractUsage parses a raw Claude API response and returns a normalized UsageDelta.
// Returns a zero-value UsageDelta and error if parsing fails.
//
// Implements the agents.MeteringCapable interface.
func ExtractUsage(rawResponse []byte) (metering.UsageDelta, error) {
	var env claudeUsageEnvelope
	if err := json.Unmarshal(rawResponse, &env); err != nil {
		return metering.UsageDelta{}, err
	}
	return metering.UsageDelta{
		PromptTokens:     env.Usage.InputTokens,
		CompletionTokens: env.Usage.OutputTokens,
		CachedTokens:     env.Usage.CacheReadInputTokens,
		CacheCreated:     env.Usage.CacheCreationInputTokens,
		Model:            env.Model,
	}, nil
}

// SessionHookEnabled reports whether Claude's adapter should feed metering
// stats into the global session hook. Always true for Claude — the
// Messages API always returns usage blocks.
//
// Implements the agents.MeteringCapable interface.
func SessionHookEnabled() bool {
	return true
}

// RecordResponse is a convenience wrapper that parses and records in one step.
// Intended for adapter call sites that just received a raw response:
//
//	rawBody := readFullResponse(resp)
//	claude.RecordResponse(rawBody)
//
// Silently no-ops if the metering hook has not been registered (common in tests).
func RecordResponse(rawResponse []byte) {
	hook := metering.Current()
	if hook == nil {
		return
	}
	delta, err := ExtractUsage(rawResponse)
	if err != nil {
		// Malformed response — don't poison the stats.
		return
	}
	hook.Record(delta)
}
