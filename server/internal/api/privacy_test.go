package api

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"accessible-content/server/internal/store"
)

func TestScanEndpointsRequireClientID(t *testing.T) {
	router := NewRouter(Options{Store: store.NewMemory(), Log: slog.Default()})

	cases := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/scans", ""},
		{http.MethodGet, "/api/scans/8b4d0d3c-1e5a-4f0b-9e2d-1234567890ab", ""},
		{http.MethodGet, "/api/scans/8b4d0d3c-1e5a-4f0b-9e2d-1234567890ab/pdf", ""},
		{http.MethodGet, "/api/scans/8b4d0d3c-1e5a-4f0b-9e2d-1234567890ab/gate", ""},
		{http.MethodDelete, "/api/scans", ""},
		{http.MethodPost, "/api/scans", `{"url":"https://example.ru","depth":1}`},
	}

	for _, test := range cases {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			if test.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 without X-Client-ID, got %d: %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestPrivateResponsesAreNotCached(t *testing.T) {
	memory := store.NewMemory()
	if err := memory.Create(context.Background(), store.Scan{
		ID:        "private-cache",
		Owner:     "browser-a",
		URL:       "https://example.ru",
		Status:    store.StatusDone,
		Version:   1,
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}

	router := NewRouter(Options{Store: memory, Log: slog.Default()})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/scans/private-cache", nil)
	request.Header.Set("X-Client-ID", "browser-a")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("expected Cache-Control: no-store, got %q", got)
	}
	if got := recorder.Header().Get("Vary"); !strings.Contains(got, "X-Client-ID") || !strings.Contains(got, "X-Lang") {
		t.Fatalf("expected Vary with X-Client-ID and X-Lang, got %q", got)
	}
}
