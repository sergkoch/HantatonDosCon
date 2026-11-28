package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"accessible-content/server/internal/store"
)

func gateRouter(t *testing.T, scan store.Scan) *httptest.ResponseRecorder {
	t.Helper()

	memory := store.NewMemory()
	if err := memory.Create(context.Background(), scan); err != nil {
		t.Fatal(err)
	}

	router := NewRouter(Options{Store: memory, Log: slog.Default()})
	recorder := httptest.NewRecorder()
	request := withClient(httptest.NewRequest(http.MethodGet, "/api/scans/"+scan.ID+"/gate?min_score=80&max_critical=0", nil))
	request.Header.Set("X-Client-ID", scan.Owner)
	router.ServeHTTP(recorder, request)

	return recorder
}

func TestGatePasses(t *testing.T) {
	score := 90
	recorder := gateRouter(t, store.Scan{
		ID:     "good",
		Owner:  "browser-a",
		URL:    "https://example.ru/",
		Status: store.StatusDone,
		Score:  &score,
		Stats:  store.Stats{Critical: 0, Warning: 2, OK: 5},
	})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var body gateResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Passed || body.Score == nil || *body.Score != 90 {
		t.Fatalf("unexpected gate body: %+v", body)
	}
	if len(body.Reasons) != 0 {
		t.Fatalf("expected no reasons, got %v", body.Reasons)
	}
	if body.Limits.MinScore != 80 || body.Limits.MaxCritical != 0 {
		t.Fatalf("unexpected limits: %+v", body.Limits)
	}
}

func TestGateFailsWithTwoReasons(t *testing.T) {
	score := 40
	recorder := gateRouter(t, store.Scan{
		ID:     "bad",
		Owner:  "browser-a",
		URL:    "https://example.ru/",
		Status: store.StatusDone,
		Score:  &score,
		Stats:  store.Stats{Critical: 3, Warning: 1, OK: 0},
	})

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var body gateResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Passed {
		t.Fatal("expected gate to fail")
	}
	if len(body.Reasons) != 2 {
		t.Fatalf("expected two reasons, got %v", body.Reasons)
	}
	if body.Critical != 3 {
		t.Fatalf("expected 3 critical, got %d", body.Critical)
	}
}

func TestGateHidesForeignScan(t *testing.T) {
	score := 100
	memory := store.NewMemory()
	if err := memory.Create(context.Background(), store.Scan{
		ID:     "private",
		Owner:  "browser-a",
		URL:    "https://example.ru/",
		Status: store.StatusDone,
		Score:  &score,
	}); err != nil {
		t.Fatal(err)
	}

	router := NewRouter(Options{Store: memory, Log: slog.Default()})
	recorder := httptest.NewRecorder()
	request := withClient(httptest.NewRequest(http.MethodGet, "/api/scans/private/gate", nil))
	request.Header.Set("X-Client-ID", "browser-b")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a foreign scan, got %d", recorder.Code)
	}
}

func TestGateNotReady(t *testing.T) {
	recorder := gateRouter(t, store.Scan{
		ID:     "running",
		Owner:  "browser-a",
		URL:    "https://example.ru/",
		Status: store.StatusRunning,
	})

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["error"] != "not_ready" {
		t.Fatalf("expected not_ready, got %v", body["error"])
	}
}

func TestGateRejectsInvalidLimit(t *testing.T) {
	score := 100
	memory := store.NewMemory()
	if err := memory.Create(context.Background(), store.Scan{
		ID:     "scan",
		Owner:  "browser-a",
		URL:    "https://example.ru/",
		Status: store.StatusDone,
		Score:  &score,
	}); err != nil {
		t.Fatal(err)
	}

	router := NewRouter(Options{Store: memory, Log: slog.Default()})

	cases := []string{
		"/api/scans/scan/gate?min_score=abc",
		"/api/scans/scan/gate?min_score=101",
		"/api/scans/scan/gate?max_critical=-1",
	}

	for _, path := range cases {
		t.Run(path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := withClient(httptest.NewRequest(http.MethodGet, path, nil))
			request.Header.Set("X-Client-ID", "browser-a")
			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}
