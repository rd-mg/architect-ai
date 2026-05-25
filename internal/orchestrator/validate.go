package orchestrator

import (
	"fmt"
	"log"

	"github.com/rd-mg/architect-ai/internal/reasoning/gate"
)

// ValidateSubAgentResponse validates a sub-agent's response against the expected mode
// computed by CognitiveScorer.
//
// It extracts the first line of the response, parses the Adaptive Reasoning gate header,
// and compares the declared mode against the expected mode.
//
// Validation rules:
//   - If header parse fails (missing/malformed), result.Err is set and RePromptCount is 0
//     (caller should re-prompt once before falling back).
//   - If declared mode matches expected mode, Matched=true.
//   - If declared mode differs, mismatch is logged.
//   - Circuit breaker: if attemptCount >= 2 and response is invalid, ForceMode3=true.
func ValidateSubAgentResponse(response string, expectedDims Dims48, expectedMode int, attemptCount int) ValidationResult {
	res := ValidationResult{
		ExpectedMode: expectedMode,
		ExpectedDims: expectedDims,
	}

	// Step 1: Extract first non-empty line
	firstLine := gate.ExtractFirstLine(response)
	if firstLine == "" {
		res.Err = fmt.Errorf("empty response, cannot extract gate header")
		if attemptCount >= 2 {
			res.ForceMode3 = true
		}
		return res
	}

	// Step 2: Parse header
	header, err := gate.ParseHeader(firstLine)
	if err != nil {
		res.Err = fmt.Errorf("header parse error: %w", err)
		if attemptCount >= 2 {
			res.ForceMode3 = true
		}
		return res
	}

	// Step 3: Record declared values
	res.DeclaredMode = header.Mode
	res.DeclaredDims = Dims48{header.D1, header.D2, header.D3, header.D4, 0}

	// Step 4: Compare declared mode vs expected mode
	if header.Mode == expectedMode {
		res.Matched = true
	} else {
		res.Matched = false
		log.Printf(
			"[CognitiveMode] Mode mismatch: declared=%d, expected=%d (change=%s, dims=%v)",
			header.Mode, expectedMode, "", res.ExpectedDims,
		)
	}

	// Step 5: Circuit breaker — if attempt >= 2 and mismatch, force Mode 3
	if attemptCount >= 2 && !res.Matched {
		res.ForceMode3 = true
	}

	return res
}
