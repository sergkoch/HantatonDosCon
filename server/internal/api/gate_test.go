package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"accessible-content/server/internal/store"

	"github.com/gin-gonic/gin"
)

func gateTestRouter(t *testing.T, scans ...store.Scan) *gin.Engine {
	t.Helper()

	memory := store.NewMemory()
	for _, scan := range scans {
		if err := memory.Create(context.Background(), scan); err != nil {
			t.Fatal(err)
		}
	}

	return NewRouter(Options{Store: memory, Log: slog.Default()})
}

func doneScan(id, owner string, score, critical, warning int) store.Scan {
	value := score
	return store.Scan{
		ID:        id,
		Owner:     owner,
		URL:       "https://example.ru",
		Status:    store.StatusDone,
		Score:     &value,
		Stats:     store.Stats{Critical: critical, Warning: warning},
		CreatedAt: time.Now().UTC(),
	}
}

func TestGatePasses(t *testing.T) {
	router := gateTestRouter(t, doneScan("scan-ok", "browser-a", 95, 0, 2))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/scans/scan-ok/gate?min_score=80&max_critical=0", nil)
	request.Header.Set("X-Client-ID", "browser-a")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestGateFailsWithReasons(t *testing.T) {
	router := gateTestRouter(t, doneScan("scan-bad", "browser-a", 70, 2, 1))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/scans/scan-bad/gate?min_score=80&max_critical=0", nil)
	request.Header.Set("X-Client-ID", "browser-a")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var body struct {
		Passed  bool     `json:"passed"`
		Reasons []string `json:"reasons"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Passed {
		t.Fatal("expected passed=false")
	}
	if len(body.Reasons) != 2 {
		t.Fatalf("expected two reasons, got %+v", body.Reasons)
	}
}

func TestGateHidesForeignScan(t *testing.T) {
	router := gateTestRouter(t, doneScan("scan-private", "browser-a", 95, 0, 0))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/scans/scan-private/gate", nil)
	request.Header.Set("X-Client-ID", "browser-b")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestGateNotReady(t *testing.T) {
	router := gateTestRouter(t, store.Scan{
		ID:        "scan-running",
		Owner:     "browser-a",
		URL:       "https://example.ru",
		Status:    store.StatusRunning,
		CreatedAt: time.Now().UTC(),
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/scans/scan-running/gate", nil)
	request.Header.Set("X-Client-ID", "browser-a")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", recorder.Code)
	}
}

func TestGateRejectsBadParameter(t *testing.T) {
	router := gateTestRouter(t, doneScan("scan-ok", "browser-a", 95, 0, 0))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/scans/scan-ok/gate?min_score=abc", nil)
	request.Header.Set("X-Client-ID", "browser-a")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}
