package checks

import (
	"testing"

	"accessible-content/server/internal/crawler"
)

func TestPageChecker(t *testing.T) {
	page := crawler.Page{
		Headings: []crawler.Heading{
			{Level: 2, Text: "Раздел"},
			{Level: 4, Text: "Подраздел"},
		},
		TextLength: 1000,
	}

	findings := PageChecker{}.Check(page)
	codes := make(map[string]int)
	for _, f := range findings {
		codes[f.Code]++
	}

	if codes["PAGE_LANG_MISSING"] != 1 {
		t.Errorf("ожидали PAGE_LANG_MISSING, получили %+v", codes)
	}
	if codes["PAGE_TITLE_MISSING"] != 1 {
		t.Errorf("ожидали PAGE_TITLE_MISSING, получили %+v", codes)
	}
	if codes["PAGE_H1_MISSING"] != 1 {
		t.Errorf("ожидали PAGE_H1_MISSING, получили %+v", codes)
	}
	if codes["PAGE_HEADING_SKIP"] != 1 {
		t.Errorf("ожидали PAGE_HEADING_SKIP, получили %+v", codes)
	}
}

func TestPageCheckerClean(t *testing.T) {
	page := crawler.Page{
		Lang:  "ru",
		Title: "Новости",
		Headings: []crawler.Heading{
			{Level: 1, Text: "Новости"},
			{Level: 2, Text: "Главное"},
			{Level: 3, Text: "Подробности"},
		},
		TextLength: 1000,
	}

	if findings := (PageChecker{}).Check(page); len(findings) != 0 {
		t.Fatalf("ожидали 0 находок, получили %+v", findings)
	}
}

func TestFormChecker(t *testing.T) {
	page := crawler.Page{
		Inputs: []crawler.Input{
			{Type: "text", Name: "name", Placeholder: "Ваше имя"},
			{Type: "text", Name: "phone", HasLabel: true},
			{Type: "select", Name: "city", HasLabel: true},
		},
	}

	findings := FormChecker{}.Check(page)
	if len(findings) != 1 {
		t.Fatalf("ожидали 1 находку, получили %+v", findings)
	}
	if findings[0].Code != "FORM_LABEL_MISSING" {
		t.Fatalf("ожидали FORM_LABEL_MISSING, получили %s", findings[0].Code)
	}
}

func TestTableChecker(t *testing.T) {
	page := crawler.Page{
		Tables: []crawler.Table{
			{HasHeaders: false},
			{HasHeaders: true},
		},
	}

	findings := TableChecker{}.Check(page)
	if len(findings) != 1 {
		t.Fatalf("ожидали 1 находку, получили %+v", findings)
	}
	if findings[0].Code != "TABLE_HEADERS_MISSING" {
		t.Fatalf("ожидали TABLE_HEADERS_MISSING, получили %s", findings[0].Code)
	}
}
