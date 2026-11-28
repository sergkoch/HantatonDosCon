package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"accessible-content/server/internal/store"

	"github.com/gin-gonic/gin"
)

func newTestRouter() *gin.Engine {
	return NewRouter(Options{
		Store: store.NewMemory(),
		Log:   slog.Default(),
	})
}

const testClientID = "test-browser"

func withClient(request *http.Request) *http.Request {
	request.Header.Set("X-Client-ID", testClientID)
	return request
}

func TestHealth(t *testing.T) {
	router := newTestRouter()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", body["status"])
	}
}

func TestCreateScan(t *testing.T) {
	router := newTestRouter()
	recorder := httptest.NewRecorder()
	request := withClient(httptest.NewRequest(http.MethodPost, "/api/scans", strings.NewReader(`{"url":"https://example.ru","depth":2}`)))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var scan store.Scan
	if err := json.Unmarshal(recorder.Body.Bytes(), &scan); err != nil {
		t.Fatalf("decode scan: %v", err)
	}

	if scan.ID == "" {
		t.Fatal("expected scan id")
	}

	if scan.Depth != 2 {
		t.Fatalf("expected depth 2, got %d", scan.Depth)
	}

	if scan.Status != store.StatusQueued {
		t.Fatalf("expected queued, got %s", scan.Status)
	}
}

func TestInvalidURL(t *testing.T) {
	router := newTestRouter()
	recorder := httptest.NewRecorder()
	request := withClient(httptest.NewRequest(http.MethodPost, "/api/scans", strings.NewReader(`{"url":"ftp://example.ru"}`)))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestScanNotFound(t *testing.T) {
	router := newTestRouter()
	recorder := httptest.NewRecorder()
	request := withClient(httptest.NewRequest(http.MethodGet, "/api/scans/missing", nil))

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}
