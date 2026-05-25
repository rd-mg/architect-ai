package gate

// Dimensions represents the 5 observable dimensions of Adaptive Reasoning Gate v3
type Dimensions struct {
	D1Complexity      int
	D2Uncertainty     int
	D3ErrorPressure   int
	D4ContextPressure int
	D5Security        int
}

// Score takes D1-D5 dimension scores (0-3 each) and returns the appropriate Mode (1-3)
// and posture list (1-2 postures). The routing matrix matches the production format
// used by all orchestrator skill injections.
//
// D5 (Security) is typically 0. Non-zero values override posture selection to
// include +++Adversarial:
//   - D5 == 2 → forces at least Mode 2, +++Adversarial + +++Critical/Systemic
//   - D5 == 3 → forces Mode 3, +++Adversarial + +++Forensic
func Score(dims [5]int) (int, []string) {
	d1, d2, d3, d4, d5 := dims[0], dims[1], dims[2], dims[3], dims[4]

	// Determine Mode
	var mode int
	switch {
	case d3 >= 2 || d4 >= 3 || d5 == 3:
		mode = 3
	case d5 == 2 || d1+d2 >= 3 || d3 >= 1:
		mode = 2
	default:
		mode = 1
	}

	// Determine Postures (max 2)
	var postures []string

	// D5 >= 2 overrides posture selection — must include +++Adversarial
	if d5 >= 2 {
		postures = append(postures, "+++Adversarial")
		switch {
		case d5 == 3:
			postures = append(postures, "+++Forensic")
		case d1 >= 2:
			postures = append(postures, "+++Systemic")
		default:
			postures = append(postures, "+++Critical")
		}
		return mode, postures
	}

	switch mode {
	case 1:
		postures = append(postures, "+++Pragmatic")
	case 2:
		postures = append(postures, "+++Critical")
		if d1 >= 2 {
			postures = append(postures, "+++Systemic")
		} else if d2 >= 2 {
			postures = append(postures, "+++Socratic")
		} else if d3 == 1 {
			postures = append(postures, "+++Forensic")
		}
	case 3:
		priority := posturePriority(d1, d3, d4)
		postures = append(postures, priority...)
	}

	return mode, postures
}

// posturePriority selects the appropriate postures for Mode 3 based on dimension context.
func posturePriority(d1, d3, d4 int) []string {
	switch {
	case d4 >= 3:
		// Context saturated — stabilize
		return []string{"+++Pragmatic"}
	case d1 >= 3:
		// Architectural/paradigm change — full impact analysis
		return []string{"+++Systemic", "+++Adversarial"}
	case d3 >= 2:
		// Error pressure — forensic investigation
		return []string{"+++Forensic", "+++Pragmatic"}
	default:
		// Default Mode 3 — adversarial + systemic review
		return []string{"+++Adversarial", "+++Systemic"}
	}
}
