package ai

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Options настраивает OpenAI-совместимый клиент.
type Options struct {
	BaseURL     string
	APIKey      string
	TextModel   string
	VisionModel string
	Timeout     time.Duration
	MaxRetries  int
}

// OpenAI — клиент для любого OpenAI-совместимого API.
type OpenAI struct {
	opts  Options
	http  *http.Client
	mu    sync.Mutex
	cache map[string]LinkResult
}

func NewOpenAI(opts Options) *OpenAI {
	if opts.Timeout <= 0 {
		opts.Timeout = 30 * time.Second
	}
	if opts.MaxRetries <= 0 {
		opts.MaxRetries = 2
	}
	opts.BaseURL = strings.TrimRight(opts.BaseURL, "/")
	if opts.BaseURL == "" {
		opts.BaseURL = "https://api.openai.com/v1"
	}

	return &OpenAI{
		opts:  opts,
		http:  &http.Client{Timeout: opts.Timeout},
		cache: make(map[string]LinkResult),
	}
}

func (c *OpenAI) Enabled() bool {
	return c.opts.APIKey != "" && c.opts.TextModel != ""
}

func (c *OpenAI) CheckLink(ctx context.Context, req LinkRequest) (LinkResult, error) {
	if !c.Enabled() {
		return LinkResult{}, errors.New("ai: клиент не настроен")
	}

	key := cacheKey(req)
	if res, ok := c.cached(key); ok {
		return res, nil
	}

	prompt := linkPrompt(req)
	var lastErr error

	for attempt := 0; attempt <= c.opts.MaxRetries; attempt++ {
		res, err := c.doLink(ctx, prompt)
		if err == nil {
			c.store(key, res)
			return res, nil
		}
		lastErr = err
		if !retryable(err) {
			break
		}
	}

	return LinkResult{}, lastErr
}

func (c *OpenAI) CheckImage(_ context.Context, _ ImageRequest) (ImageResult, error) {
	return ImageResult{}, errors.New("ai: проверка изображений не реализована (задача T16)")
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	Temperature    float64         `json:"temperature"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type httpError struct {
	Status int
	Body   string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("ai: статус %d: %s", e.Status, e.Body)
}

func retryable(err error) bool {
	var he *httpError
	if errors.As(err, &he) {
		return he.Status == http.StatusTooManyRequests || he.Status >= 500
	}
	return true
}

func (c *OpenAI) doLink(ctx context.Context, prompt string) (LinkResult, error) {
	body := chatRequest{
		Model: c.opts.TextModel,
		Messages: []chatMessage{
			{Role: "system", Content: "Ты ассистент по веб-доступности. Отвечай строго JSON без пояснений."},
			{Role: "user", Content: prompt},
		},
		Temperature:    0.2,
		ResponseFormat: &responseFormat{Type: "json_object"},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return LinkResult{}, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.opts.BaseURL+"/chat/completions",
		bytes.NewReader(payload),
	)
	if err != nil {
		return LinkResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.opts.APIKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return LinkResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return LinkResult{}, &httpError{Status: resp.StatusCode, Body: string(data)}
	}

	var decoded chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return LinkResult{}, err
	}
	if len(decoded.Choices) == 0 {
		return LinkResult{}, errors.New("ai: пустой ответ модели")
	}

	return parseLinkResult(strings.TrimSpace(decoded.Choices[0].Message.Content))
}

func parseLinkResult(content string) (LinkResult, error) {
	var res LinkResult
	if err := json.Unmarshal([]byte(content), &res); err == nil {
		return res, nil
	}

	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(content[start:end+1]), &res); err == nil {
			return res, nil
		}
	}

	return LinkResult{}, fmt.Errorf("ai: не удалось разобрать ответ: %s", content)
}

func linkPrompt(req LinkRequest) string {
	var b strings.Builder
	b.WriteString("Оцени текст ссылки для доступности (WCAG 2.4.4). ")
	b.WriteString("Верни JSON с полями: meaningful (bool), matches (bool), reason (строка), ")
	b.WriteString("suggested_text (строка), confidence (0..1).\n")
	fmt.Fprintf(&b, "Текст ссылки: %q\n", req.Text)
	fmt.Fprintf(&b, "URL: %s\n", req.URL)
	if req.TargetTitle != "" {
		fmt.Fprintf(&b, "Заголовок целевой страницы: %q\n", req.TargetTitle)
	}
	if req.TargetText != "" {
		fmt.Fprintf(&b, "Первый абзац целевой страницы: %q\n", req.TargetText)
	}
	b.WriteString("Если текст непонятен вне контекста — предложи короткий понятный текст в suggested_text.")

	return b.String()
}

func cacheKey(req LinkRequest) string {
	sum := sha256.Sum256([]byte(
		req.Text + "\x00" + req.URL + "\x00" + req.TargetTitle + "\x00" + req.TargetText,
	))
	return hex.EncodeToString(sum[:])
}

func (c *OpenAI) cached(key string) (LinkResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	res, ok := c.cache[key]
	return res, ok
}

func (c *OpenAI) store(key string, res LinkResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = res
}
