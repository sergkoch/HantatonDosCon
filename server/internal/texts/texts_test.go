package texts

import "testing"

func TestParse(t *testing.T) {
	cases := map[string]Lang{
		"":                RU,
		"ru":              RU,
		"ru-RU":           RU,
		"en":              EN,
		"en-US":           EN,
		"EN":              EN,
		"de":              RU,
		"en,ru;q=0.8":     EN,
	}

	for raw, want := range cases {
		if got := Parse(raw); got != want {
			t.Errorf("Parse(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestFromHeadersPrefersExplicit(t *testing.T) {
	if got := FromHeaders("en", "ru-RU,ru;q=0.9"); got != EN {
		t.Errorf("explicit language must win, got %q", got)
	}
	if got := FromHeaders("", "en-GB,en;q=0.9"); got != EN {
		t.Errorf("Accept-Language must be used, got %q", got)
	}
	if got := FromHeaders("", ""); got != RU {
		t.Errorf("default must be Russian, got %q", got)
	}
}

func TestCatalogsHaveSameKeys(t *testing.T) {
	for key := range catalog[RU] {
		if _, ok := catalog[EN][key]; !ok {
			t.Errorf("missing English translation for %q", key)
		}
	}
	for key := range catalog[EN] {
		if _, ok := catalog[RU][key]; !ok {
			t.Errorf("missing Russian translation for %q", key)
		}
	}
}

func TestFallbacks(t *testing.T) {
	if got := T("de", AltMissingTitle); got != catalog[RU][AltMissingTitle] {
		t.Errorf("unknown language must fall back to Russian, got %q", got)
	}
	if got := T(EN, Key("unknown.key")); got != "unknown.key" {
		t.Errorf("unknown key must be returned as is, got %q", got)
	}
	if got := Tf(EN, PDFPage, 3); got != "Page 3" {
		t.Errorf("Tf must format, got %q", got)
	}
}
