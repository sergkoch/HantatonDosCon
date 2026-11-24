package store

import "testing"

func finding(code, element string) Finding {
	return Finding{Code: code, Kind: "image", Severity: SeverityWarning, Title: code, Element: element}
}

func TestDiff(t *testing.T) {
	previous := []Finding{
		finding("IMG_NO_ALT", `<img src="a.png">`),
		finding("LINK_BROKEN", `<a href="/old">old</a>`),
	}
	current := []Finding{
		finding("IMG_NO_ALT", `<img src="a.png">`),
		finding("IMG_ALT_GENERIC", `<img src="b.png" alt="image2.png">`),
	}

	changes := Diff(previous, current)

	if len(changes.Added) != 1 || changes.Added[0].Code != "IMG_ALT_GENERIC" {
		t.Fatalf("unexpected added: %+v", changes.Added)
	}
	if len(changes.Fixed) != 1 || changes.Fixed[0].Code != "LINK_BROKEN" {
		t.Fatalf("unexpected fixed: %+v", changes.Fixed)
	}
	if changes.Remaining != 1 {
		t.Fatalf("expected 1 remaining, got %d", changes.Remaining)
	}
}

func TestDiffWithoutCodes(t *testing.T) {
	previous := []Finding{{Kind: "page", Title: "old", Element: "<h3>"}}
	current := []Finding{{Kind: "page", Title: "new", Element: "<h3>"}}

	changes := Diff(previous, current)
	if len(changes.Added) != 1 || len(changes.Fixed) != 1 {
		t.Fatalf("expected one added and one fixed, got %+v", changes)
	}
}
