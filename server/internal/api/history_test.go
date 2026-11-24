package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"accessible-content/server/internal/store"
)

func TestCreateScanAssignsVersion(t *testing.T) {
	memory := store.NewMemory()
	previous := store.Scan{
		ID:        "previous",
		URL:       "https://example.ru/",
		Depth:     1,
		Status:    store.StatusDone,
		Version:   1,
		CreatedAt: time.Now().UTC().Add(-time.Hour),
	}
	if err := memory.Create(context.Background(), previous); err != nil {
		t.Fatal(err)
	}

	router := NewRouter(Options{Store: memory, Log: slog.Default()})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/scans", strings.NewReader(`{"url":"https://example.ru/","depth":1}`))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var scan store.Scan
	if err := json.Unmarshal(recorder.Body.Bytes(), &scan); err != nil {
		t.Fatal(err)
	}
	if scan.Version != 2 {
		t.Fatalf("expected version 2, got %d", scan.Version)
	}
	if scan.PreviousID != "previous" {
		t.Fatalf("expected previous id, got %q", scan.PreviousID)
	}
}

func TestListScans(t *testing.T) {
	memory := store.NewMemory()
	now := time.Now().UTC()
	for _, scan := range []store.Scan{
		{ID: "one", URL: "https://example.ru", Status: store.StatusDone, Version: 1, CreatedAt: now.Add(-time.Hour)},
		{ID: "two", URL: "https://example.ru", Status: store.StatusDone, Version: 2, CreatedAt: now},
		{ID: "other", URL: "https://other.ru", Status: store.StatusDone, Version: 1, CreatedAt: now},
	} {
		if err := memory.Create(context.Background(), scan); err != nil {
			t.Fatal(err)
		}
	}

	router := NewRouter(Options{Store: memory, Log: slog.Default()})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/scans?url=https://example.ru&limit=10", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var scans []store.Scan
	if err := json.Unmarshal(recorder.Body.Bytes(), &scans); err != nil {
		t.Fatal(err)
	}
	if len(scans) != 2 || scans[0].ID != "two" {
		t.Fatalf("unexpected scans: %+v", scans)
	}
}
