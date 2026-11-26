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
		Owner:     "browser-a",
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
	request.Header.Set("X-Client-ID", "browser-a")

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
		{ID: "one", Owner: "browser-a", URL: "https://example.ru", Status: store.StatusDone, Version: 1, CreatedAt: now.Add(-time.Hour)},
		{ID: "two", Owner: "browser-a", URL: "https://example.ru", Status: store.StatusDone, Version: 2, CreatedAt: now},
		{ID: "other", Owner: "browser-b", URL: "https://example.ru", Status: store.StatusDone, Version: 1, CreatedAt: now},
	} {
		if err := memory.Create(context.Background(), scan); err != nil {
			t.Fatal(err)
		}
	}

	router := NewRouter(Options{Store: memory, Log: slog.Default()})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/scans?url=https://example.ru&limit=10", nil)
	request.Header.Set("X-Client-ID", "browser-a")

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

	foreign := httptest.NewRecorder()
	foreignRequest := httptest.NewRequest(http.MethodGet, "/api/scans?limit=10", nil)
	foreignRequest.Header.Set("X-Client-ID", "browser-b")
	router.ServeHTTP(foreign, foreignRequest)

	var foreignScans []store.Scan
	if err := json.Unmarshal(foreign.Body.Bytes(), &foreignScans); err != nil {
		t.Fatal(err)
	}
	if len(foreignScans) != 1 || foreignScans[0].ID != "other" {
		t.Fatalf("history of another browser leaked: %+v", foreignScans)
	}
}

func TestGetScanIsPrivateToOwner(t *testing.T) {
	memory := store.NewMemory()
	scan := store.Scan{
		ID:        "private-scan",
		Owner:     "browser-a",
		URL:       "https://example.ru",
		Status:    store.StatusDone,
		Version:   1,
		CreatedAt: time.Now().UTC(),
	}
	if err := memory.Create(context.Background(), scan); err != nil {
		t.Fatal(err)
	}

	router := NewRouter(Options{Store: memory, Log: slog.Default()})

	cases := []struct {
		name   string
		header string
		want   int
	}{
		{name: "owner", header: "browser-a", want: http.StatusOK},
		{name: "stranger", header: "browser-b", want: http.StatusNotFound},
		{name: "no identity", header: "", want: http.StatusNotFound},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/scans/private-scan", nil)
			if test.header != "" {
				request.Header.Set("X-Client-ID", test.header)
			}

			router.ServeHTTP(recorder, request)

			if recorder.Code != test.want {
				t.Fatalf("expected %d, got %d: %s", test.want, recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestClientIDRejectsMalformedValue(t *testing.T) {
	router := NewRouter(Options{Store: store.NewMemory(), Log: slog.Default()})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/scans", nil)
	request.Header.Set("X-Client-ID", "bad identity!")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var scans []store.Scan
	if err := json.Unmarshal(recorder.Body.Bytes(), &scans); err != nil {
		t.Fatal(err)
	}
	if len(scans) != 0 {
		t.Fatalf("malformed identity must not see any history: %+v", scans)
	}
}
