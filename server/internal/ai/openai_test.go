package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestCheckLinkParsesSuggestion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("неожиданный путь: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("нет ключа авторизации")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"{\"meaningful\":false,\"matches\":false,\"reason\":\"непонятно вне контекста\",\"suggested_text\":\"Положение о конкурсе\",\"confidence\":0.9}"}}]}`)
	}))
	defer server.Close()

	client := NewOpenAI(Options{
		BaseURL:   server.URL,
		APIKey:    "test-key",
		TextModel: "test-model",
		Timeout:   2 * time.Second,
	})

	res, err := client.CheckLink(context.Background(), LinkRequest{
		Text: "подробнее",
		URL:  "https://example.ru/contest.pdf",
	})
	if err != nil {
		t.Fatalf("CheckLink: %v", err)
	}
	if res.Meaningful {
		t.Errorf("ожидали meaningful=false")
	}
	if res.SuggestedText != "Положение о конкурсе" {
		t.Errorf("ожидали suggested_text, получили %q", res.SuggestedText)
	}
}

func TestCheckLinkCaches(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"{\"meaningful\":true,\"matches\":true,\"reason\":\"\",\"suggested_text\":\"\",\"confidence\":0.8}"}}]}`)
	}))
	defer server.Close()

	client := NewOpenAI(Options{BaseURL: server.URL, APIKey: "k", TextModel: "m", Timeout: 2 * time.Second})
	req := LinkRequest{Text: "О нас", URL: "https://example.ru/about"}

	if _, err := client.CheckLink(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if _, err := client.CheckLink(context.Background(), req); err != nil {
		t.Fatal(err)
	}

	if got := calls.Load(); got != 1 {
		t.Fatalf("ожидали 1 запрос (кэш), получили %d", got)
	}
}

func TestCheckImageParsesMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		messages, _ := body["messages"].([]any)
		if len(messages) != 2 {
			t.Errorf("expected system and user messages, got %d", len(messages))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"{\"verdict\":\"mismatch\",\"reason\":\"на картинке красный круг\",\"suggested_alt\":\"Красный круг\",\"confidence\":0.9}"}}]}`)
	}))
	defer server.Close()

	client := NewOpenAI(Options{
		BaseURL:     server.URL,
		APIKey:      "test-key",
		TextModel:   "text-model",
		VisionModel: "vision-model",
		Timeout:     2 * time.Second,
	})

	result, err := client.CheckImage(context.Background(), ImageRequest{
		Image:    []byte("not-a-real-png"),
		MimeType: "image/png",
		Alt:      "Синий квадрат",
		Context:  "https://example.ru/",
	})
	if err != nil {
		t.Fatalf("CheckImage: %v", err)
	}
	if result.Verdict != "mismatch" {
		t.Fatalf("expected mismatch, got %q", result.Verdict)
	}
	if result.SuggestedAlt != "Красный круг" {
		t.Fatalf("unexpected suggested alt %q", result.SuggestedAlt)
	}
}

func TestCheckLinkDisabled(t *testing.T) {
	client := NewOpenAI(Options{})
	if _, err := client.CheckLink(context.Background(), LinkRequest{Text: "x"}); err == nil {
		t.Fatal("без ключа CheckLink должен вернуть ошибку")
	}
}
