package api

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"accessible-content/server/internal/store"
)

func TestDownloadPDF(t *testing.T) {
	memory := store.NewMemory()
	scan := store.Scan{
		ID:        "scan-pdf",
		URL:       "https://example.ru",
		Status:    store.StatusDone,
		Version:   1,
		Pages:     1,
		Stats:     store.Stats{Critical: 1},
		CreatedAt: time.Now().UTC(),
		Findings: []store.Finding{
			{
				Code:           "IMG_NO_ALT",
				Kind:           "image",
				Severity:       store.SeverityCritical,
				Title:          "Нет alt у изображения",
				Element:        `<img src="photo.svg">`,
				Recommendation: "Добавьте описание.",
				WCAG:           "1.1.1",
			},
		},
	}
	if err := memory.Create(context.Background(), scan); err != nil {
		t.Fatal(err)
	}

	router := NewRouter(Options{Store: memory, Log: slog.Default()})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/scans/scan-pdf/pdf", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/pdf" {
		t.Fatalf("unexpected content type %q", contentType)
	}
	if !bytes.HasPrefix(recorder.Body.Bytes(), []byte("%PDF")) {
		t.Fatal("expected PDF body")
	}
}

func TestDownloadPDFInBothLanguages(t *testing.T) {
	memory := store.NewMemory()
	scan := store.Scan{
		ID:        "scan-pdf-lang",
		URL:       "https://example.ru",
		Status:    store.StatusDone,
		Lang:      "ru",
		Version:   1,
		Pages:     1,
		Score:     new(int),
		CreatedAt: time.Now().UTC(),
		Findings: []store.Finding{
			{
				Code:              "IMG_ALT_MISSING",
				Kind:              "image",
				Severity:          store.SeverityCritical,
				Title:             "Нет alt у изображения",
				Element:           `<img src="photo.svg">`,
				Recommendation:    "Добавьте описание.",
				WCAG:              "1.1.1",
				TitleAlt:          "Image has no alt text",
				RecommendationAlt: "Describe the image.",
			},
		},
	}
	if err := memory.Create(context.Background(), scan); err != nil {
		t.Fatal(err)
	}

	router := NewRouter(Options{Store: memory, Log: slog.Default()})

	for _, lang := range []string{"ru", "en"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/scans/scan-pdf-lang/pdf?lang="+lang, nil)
		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d: %s", lang, recorder.Code, recorder.Body.String())
		}
		if !bytes.HasPrefix(recorder.Body.Bytes(), []byte("%PDF")) {
			t.Fatalf("%s: expected PDF body", lang)
		}
	}
}
