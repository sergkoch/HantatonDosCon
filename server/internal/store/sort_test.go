package store

import "testing"

func TestSortBySeverity(t *testing.T) {
	findings := []Finding{
		{ID: "w1", Severity: SeverityWarning},
		{ID: "c1", Severity: SeverityCritical},
		{ID: "o1", Severity: SeverityOK},
		{ID: "w2", Severity: SeverityWarning},
		{ID: "c2", Severity: SeverityCritical},
	}

	SortBySeverity(findings)

	want := []string{"c1", "c2", "w1", "w2", "o1"}
	for i, id := range want {
		if findings[i].ID != id {
			t.Fatalf("position %d: want %s, got %s (%+v)", i, id, findings[i].ID, findings)
		}
	}
}
