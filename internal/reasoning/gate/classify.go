package gate

// Dimensions represents the 5 dimensions of Adaptive Reasoning Gate v3
type Dimensions struct {
	D1Complexity      int
	D2Uncertainty     int
	D3ErrorPressure   int
	D4ContextPressure int
	D5Security        int
}

// Score takes D1-D5 dimension scores and returns the appropriate Mode (1-3) and postures list (up to 2).
func Score(dims [5]int) (int, []string) {
	d := Dimensions{
		D1Complexity:      dims[0],
		D2Uncertainty:     dims[1],
		D3ErrorPressure:   dims[2],
		D4ContextPressure: dims[3],
		D5Security:        dims[4],
	}

	// Determine Mode
	var mode int
	if d.D3ErrorPressure >= 2 || d.D4ContextPressure >= 3 {
		mode = 3
	} else if d.D5Security == 3 {
		mode = 3
	} else if d.D5Security == 2 {
		mode = 2
	} else if d.D1Complexity+d.D2Uncertainty <= 2 && d.D3ErrorPressure+d.D4ContextPressure <= 2 {
		mode = 1
	} else if d.D1Complexity+d.D2Uncertainty >= 3 || d.D3ErrorPressure >= 1 {
		mode = 2
	} else {
		mode = 1
	}

	// Determine Postures
	var postures []string
	if d.D5Security >= 2 {
		postures = append(postures, "+++Adversarial")
		if mode == 3 {
			postures = append(postures, "+++Forensic")
		} else {
			postures = append(postures, "+++Critical")
		}
		return mode, postures
	}

	switch mode {
	case 1:
		postures = append(postures, "+++Pragmatic")
	case 2:
		postures = append(postures, "+++Critical")
		if d.D1Complexity >= 2 {
			postures = append(postures, "+++Systemic")
		} else if d.D2Uncertainty >= 2 {
			postures = append(postures, "+++Socratic")
		} else if d.D3ErrorPressure == 1 {
			postures = append(postures, "+++Forensic")
		}
	case 3:
		if d.D1Complexity >= 3 {
			postures = append(postures, "+++Systemic", "+++Adversarial")
		} else {
			postures = append(postures, "+++Forensic", "+++Pragmatic")
		}
	}

	return mode, postures
}
