package checks

import "testing"

func TestScore(t *testing.T) {
	findings := []Finding{
		{Severity: "critical"},
		{Severity: "critical"},
		{Severity: "warning"},
	}

	if got := Score(findings); got != 77 {
		t.Fatalf("ожидали 77, получили %d", got)
	}

	many := make([]Finding, 20)
	for i := range many {
		many[i] = Finding{Severity: "critical"}
	}
	if got := Score(many); got != 0 {
		t.Fatalf("оценка не должна быть ниже 0, получили %d", got)
	}

	if got := Score(nil); got != 100 {
		t.Fatalf("без находок ожидали 100, получили %d", got)
	}
}

func TestStats(t *testing.T) {
	findings := []Finding{
		{Severity: "critical"},
		{Severity: "warning"},
		{Severity: "ok"},
	}

	c, w, o := Stats(findings)
	if c != 1 || w != 1 || o != 1 {
		t.Fatalf("ожидали 1/1/1, получили %d/%d/%d", c, w, o)
	}
}
