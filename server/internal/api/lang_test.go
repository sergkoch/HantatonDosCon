package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"accessible-content/server/internal/store"
)

func TestCreateScanStoresLanguage(t *testing.T) {
	router := NewRouter(Options{Store: store.NewMemory(), Log: slog.Default()})

	cases := []struct {
		name   string
		header string
		want   string
	}{
		{name: "default is Russian", header: "", want: "ru"},
		{name: "explicit English", header: "en", want: "en"},
		{name: "language tag", header: "en-GB", want: "en"},
		{name: "unknown falls back", header: "de", want: "ru"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			body := strings.NewReader(`{"url":"https://example.com/","depth":1}`)
			request := httptest.NewRequest(http.MethodPost, "/api/scans", body)
			request.Header.Set("Content-Type", "application/json")
			if test.header != "" {
				request.Header.Set("X-Lang", test.header)
			}

			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusCreated {
				t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
			}

			var scan store.Scan
			if err := json.Unmarshal(recorder.Body.Bytes(), &scan); err != nil {
				t.Fatal(err)
			}
			if scan.Lang != test.want {
				t.Fatalf("expected lang %q, got %q", test.want, scan.Lang)
			}
		})
	}
}

func TestErrorsFollowRequestLanguage(t *testing.T) {
	router := NewRouter(Options{Store: store.NewMemory(), Log: slog.Default()})

	cases := []struct {
		name   string
		lang   string
		accept string
		want   string
	}{
		{name: "Russian by default", want: "адрес должен быть http или https"},
		{name: "English via header", lang: "en", want: "url must be http or https"},
		{name: "English via Accept-Language", accept: "en-US,en;q=0.9", want: "url must be http or https"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/scans", strings.NewReader(`{"url":"not-a-url","depth":1}`))
			request.Header.Set("Content-Type", "application/json")
			if test.lang != "" {
				request.Header.Set("X-Lang", test.lang)
			}
			if test.accept != "" {
				request.Header.Set("Accept-Language", test.accept)
			}

			router.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", recorder.Code)
			}
			if !strings.Contains(recorder.Body.String(), test.want) {
				t.Fatalf("expected %q in body, got %s", test.want, recorder.Body.String())
			}
		})
	}
}
