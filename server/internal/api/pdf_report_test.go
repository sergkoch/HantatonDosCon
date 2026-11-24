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
