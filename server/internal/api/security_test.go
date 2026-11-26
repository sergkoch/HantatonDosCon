package api

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"accessible-content/server/internal/safehttp"
	"accessible-content/server/internal/store"

	"github.com/gin-gonic/gin"
)

func newSecurityRouter(t *testing.T, policy *safehttp.Policy, rateLimit int) *gin.Engine {
	t.Helper()

	router := NewRouter(Options{
		Store:     store.NewMemory(),
		Log:       slog.Default(),
		Policy:    policy,
		RateLimit: rateLimit,
	})
	return router
}

func TestCreateScanRejectsPrivateTargets(t *testing.T) {
	policy := safehttp.NewPolicy(nil)
	policy.Resolve = func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("127.0.0.1")}, nil
	}
	router := newSecurityRouter(t, policy, 10)

	for _, target := range []string{"http://private.test/", "http://127.0.0.1/"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/scans", strings.NewReader(`{"url":"`+target+`","depth":1}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-Lang", "en")

		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d: %s", target, recorder.Code, recorder.Body.String())
		}
		if !strings.Contains(recorder.Body.String(), "url must be public") {
			t.Fatalf("%s: unexpected body %s", target, recorder.Body.String())
		}
	}
}

func TestCreateScanAllowsAllowlistedHost(t *testing.T) {
	policy := safehttp.NewPolicy([]string{"demo"})
	policy.Resolve = func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("172.18.0.2")}, nil
	}
	router := newSecurityRouter(t, policy, 10)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/scans", strings.NewReader(`{"url":"http://demo/","depth":1}`))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestCreateScanRateLimit(t *testing.T) {
	router := newSecurityRouter(t, nil, 2)

	for i := 0; i < 2; i++ {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/scans", strings.NewReader(`{"url":"https://example.ru","depth":1}`))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusCreated {
			t.Fatalf("request %d: expected 201, got %d", i+1, recorder.Code)
		}
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/scans", strings.NewReader(`{"url":"https://example.ru","depth":1}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get("Retry-After") != "60" {
		t.Fatalf("expected Retry-After header, got %q", recorder.Header().Get("Retry-After"))
	}
}

func TestCreateScanBodyLimit(t *testing.T) {
	router := newSecurityRouter(t, nil, 10)

	body := `{"url":"https://example.ru","padding":"` + strings.Repeat("x", maxRequestBodyBytes) + `"}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/scans", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", recorder.Code)
	}
}
