package checks

// Score считает оценку по формуле: 100 - critical*10 - warning*3, в диапазоне 0..100.
func Score(findings []Finding) int {
	critical, warning := 0, 0
	for _, f := range findings {
		switch f.Severity {
		case "critical":
			critical++
		case "warning":
			warning++
		}
	}

	score := 100 - critical*10 - warning*3
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// Stats считает количество находок по критичности.
func Stats(findings []Finding) (critical, warning, ok int) {
	for _, f := range findings {
		switch f.Severity {
		case "critical":
			critical++
		case "warning":
			warning++
		default:
			ok++
		}
	}

	return critical, warning, ok
}
