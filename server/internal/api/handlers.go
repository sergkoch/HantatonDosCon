package api

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"accessible-content/server/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type healthResponse struct {
	Status  string    `json:"status"`
	Version string    `json:"version"`
	AI      string    `json:"ai"`
	Time    time.Time `json:"time"`
}

func (h *handler) health(c *gin.Context) {
	ai := "missing"
	if h.aiConfigured {
		ai = "configured"
	}

	c.JSON(http.StatusOK, healthResponse{
		Status:  "ok",
		Version: "0.1.0",
		AI:      ai,
		Time:    time.Now().UTC(),
	})
}

type createScanRequest struct {
	URL   string `json:"url" binding:"required"`
	Depth int    `json:"depth"`
}

func (h *handler) createScan(c *gin.Context) {
	var request createScanRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "detail": err.Error()})
		return
	}

	parsed, err := url.ParseRequestURI(request.URL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "detail": "url must be http or https"})
		return
	}

	if request.Depth == 0 {
		request.Depth = 1
	}

	if request.Depth < 1 || request.Depth > 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "detail": "depth must be between 1 and 3"})
		return
	}

	scan := store.Scan{
		ID:        uuid.NewString(),
		URL:       parsed.String(),
		Depth:     request.Depth,
		Status:    store.StatusQueued,
		Findings:  []store.Finding{},
		CreatedAt: time.Now().UTC(),
	}

	if err := h.store.Create(c.Request.Context(), scan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "store_error", "detail": err.Error()})
		return
	}

	go h.runMockScan(scan.ID)

	c.JSON(http.StatusCreated, scan)
}

func (h *handler) getScan(c *gin.Context) {
	scan, err := h.store.Get(c.Request.Context(), c.Param("id"))
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "store_error", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, scan)
}

func (h *handler) runMockScan(id string) {
	time.Sleep(2 * time.Second)

	ctx := context.Background()
	scan, err := h.store.Get(ctx, id)
	if err != nil {
		return
	}

	score := 42
	scan.Status = store.StatusDone
	scan.Score = &score
	scan.Stats = store.Stats{Critical: 2, Warning: 1, OK: 1}
	scan.Findings = []store.Finding{
		{
			ID:             "f1",
			Kind:           "image",
			Severity:       store.SeverityCritical,
			Title:          "Нет alt у изображения",
			Element:        `<img src="photo.svg">`,
			Recommendation: "Добавьте описание, что на картинке.",
			WCAG:           "1.1.1",
		},
		{
			ID:             "f2",
			Kind:           "pdf",
			Severity:       store.SeverityCritical,
			Title:          "PDF без текстового слоя",
			Element:        `<a href="report.pdf">Скачать отчёт</a>`,
			Recommendation: "Опубликуйте текстовую версию или PDF с текстовым слоем.",
			WCAG:           "1.3.1",
		},
		{
			ID:             "f3",
			Kind:           "link",
			Severity:       store.SeverityWarning,
			Title:          "Ссылка «подробнее»",
			Element:        `<a href="/news/1">подробнее</a>`,
			Recommendation: "Напишите, что откроется: «Положение о конкурсе».",
			WCAG:           "2.4.4",
		},
	}

	if err := h.store.Update(ctx, scan); err != nil {
		h.log.Error("update mock scan", "error", err)
	}
}
