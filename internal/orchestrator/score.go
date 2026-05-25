package orchestrator

import (
	"github.com/rd-mg/architect-ai/internal/reasoning/gate"
)

// CognitiveScorer maps a TaskContext to D1-D5 dimension scores, then calls gate.Score()
// to determine the appropriate Mode and postures.
//
// Scoring heuristics (per design §Scoring Heuristics):
//
//	| Dim |  0            | 1             | 2               | 3                   |
//	|-----|--------------|---------------|-----------------|---------------------|
//	| D1  | 1 file, same | 2-3 files,    | >3 files OR     | Architectural       |
//	|     | pkg          | same pkg      | cross-pkg       | change              |
//	| D2  | Specs exist  | Phase=design/ | Phase=explore/  | Agent=ideator/      |
//	|     | & complete   | tasks         | propose         | researcher          |
//	| D3  | attempt=0    | attempt=1     | attempt=2       | attempt>=3          |
//	| D4  | <10KB        | 10-50KB       | 50-100KB        | >100KB              |
//	| D5  | normal       | security-     | security        | security incident   |
//	|     |              | aware         | concern         |                     |
func CognitiveScorer(ctx TaskContext) (Dims48, int, []string) {
	dims := computeDims(ctx)

	// Circuit breaker: attempt >= 2 forces D3=3 (production down)
	if ctx.AttemptCount >= 2 {
		dims[2] = 3
	}

	mode, postures := gate.Score(dims)

	// Economic/Empirical override on mode >= 2
	if mode >= 2 {
		if ctx.IsCostSensitive {
			// Replace or augment second posture with +++Economic
			// Cost concerns trump measurement (checked first)
			if len(postures) >= 2 {
				postures[1] = "+++Economic"
			} else {
				postures = append(postures, "+++Economic")
			}
		} else if ctx.IsMeasurementTask {
			if len(postures) >= 2 {
				postures[1] = "+++Empirical"
			} else {
				postures = append(postures, "+++Empirical")
			}
		}
	}

	// Enforce max 2 postures
	if len(postures) > 2 {
		postures = postures[:2]
	}

	return dims, mode, postures
}

func computeDims(ctx TaskContext) Dims48 {
	var dims Dims48

	// D1: Complexity
	switch {
	case ctx.FileCount >= 5 || ctx.CrossPackage && ctx.FileCount > 3:
		dims[0] = 3 // Architectural change
	case ctx.CrossPackage || ctx.FileCount > 3:
		dims[0] = 2 // Cross-module systemic
	case ctx.FileCount >= 2:
		dims[0] = 1 // Bounded module (2-3 files)
	default:
		dims[0] = 0 // Atomic/single file
	}

	// D2: Uncertainty
	if ctx.HasSpecs {
		dims[1] = 0 // Specs complete
	} else {
		switch ctx.PhaseName {
		case "sdd-explore", "sdd-propose", "sdd-onboard":
			dims[1] = 2 // Conflicting docs or unknown
		case "sdd-init", "":
			dims[1] = 3 // Terra incognita
		default:
			dims[1] = 1 // Partial specs
		}
	}

	// D3: Error Pressure (base — circuit breaker override in CognitiveScorer)
	switch {
	case ctx.AttemptCount >= 3:
		dims[2] = 3 // Production down / data loss risk
	case ctx.AttemptCount >= 2:
		dims[2] = 2 // Repeated failure
	case ctx.AttemptCount >= 1:
		dims[2] = 1 // Recent failure
	default:
		dims[2] = 0 // First attempt
	}

	// D4: Context Pressure
	switch {
	case ctx.ContextEstimateKB >= 100:
		dims[3] = 3 // > 100KB (Guardian active)
	case ctx.ContextEstimateKB >= 50:
		dims[3] = 2 // 50-100KB
	case ctx.ContextEstimateKB >= 10:
		dims[3] = 1 // 10-50KB
	default:
		dims[3] = 0 // < 10KB
	}

	// D5: Security (computed server-side from task context)
	if ctx.IsSecuritySensitive {
		switch {
		case ctx.AttemptCount >= 2:
			dims[4] = 3 // Security incident
		case ctx.AttemptCount >= 1:
			dims[4] = 2 // Security concern
		default:
			dims[4] = 1 // Security-aware
		}
	}

	return dims
}
