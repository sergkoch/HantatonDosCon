package checks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"accessible-content/server/internal/crawler"
)

func TestLinkChecker(t *testing.T) {
	page := crawler.Page{
		Links: []crawler.Link{
			{Href: "/a", Text: "подробнее"},
			{Href: "/b", Text: "   "},
			{Href: "/c", Text: "Положение о конкурсе"},
			{Href: "/d", Text: "здесь"},
		},
	}

	findings := LinkChecker{}.Check(page)
	if len(findings) != 3 {
		t.Fatalf("ожидали 3 находки, получили %d: %+v", len(findings), findings)
	}

	codes := make(map[string]int)
	for _, f := range findings {
		codes[f.Code]++
	}
	if codes["LINK_TEXT_GENERIC"] != 2 {
		t.Errorf("ожидали 2 LINK_TEXT_GENERIC, получили %+v", codes)
	}
	if codes["LINK_TEXT_EMPTY"] != 1 {
		t.Errorf("ожидали LINK_TEXT_EMPTY, получили %+v", codes)
	}
}

func TestBrokenLinks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ok" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	links := []crawler.Link{
		{Href: server.URL + "/ok", Text: "ок"},
		{Href: server.URL + "/missing", Text: "нет"},
		{Href: server.URL + "/ok", Text: "дубль"},
	}

	findings := BrokenLinks(context.Background(), server.Client(), links, 10)
	if len(findings) != 1 {
		t.Fatalf("ожидали 1 находку, получили %d: %+v", len(findings), findings)
	}
	if findings[0].Code != "LINK_BROKEN" {
		t.Fatalf("ожидали LINK_BROKEN, получили %s", findings[0].Code)
	}
}
