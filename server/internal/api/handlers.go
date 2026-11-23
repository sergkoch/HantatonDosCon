package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"accessible-content/server/internal/checks"
	"accessible-content/server/internal/crawler"
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

	go h.runScan(scan)

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

func (h *handler) runScan(initial store.Scan) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	scan, err := h.store.Get(ctx, initial.ID)
	if err != nil {
		return
	}

	scan.Status = store.StatusRunning
	if err := h.store.Update(ctx, scan); err != nil {
		h.log.Error("update running scan", "error", err)
		return
	}

	if h.crawler == nil {
		h.fail(ctx, scan, "но не настроен")
		return
	}

	pages, err := h.crawler.Crawl(ctx, scan.URL, scan.Depth)
	if err != nil {
		h.log.Error("crawl failed", "url", scan.URL, "error", err)
		h.fail(ctx, scan, "краулер завершился ошибкой")
		return
	}

	altChecker := checks.AltChecker{}
	linkChecker := checks.LinkChecker{}

	checkFindings := make([]checks.Finding, 0)
	links := make([]crawler.Link, 0)
	for _, page := range pages {
		checkFindings = append(checkFindings, altChecker.Check(page)...)
		checkFindings = append(checkFindings, linkChecker.Check(page)...)
		links = append(links, page.Links...)
		links = append(links, page.PDFs...)
	}

	checkFindings = append(checkFindings, checks.BrokenLinks(ctx, h.httpClient, links, 20)...)

	critical, warning, ok := checks.Stats(checkFindings)
	score := checks.Score(checkFindings)

	scan.Findings = toStoreFindings(checkFindings)
	scan.Stats = store.Stats{Critical: critical, Warning: warning, OK: ok}
	scan.Score = &score
	scan.Status = store.StatusDone

	if err := h.store.Update(ctx, scan); err != nil {
		h.log.Error("update finished scan", "error", err)
	}
}

func (h *handler) fail(ctx context.Context, scan store.Scan, reason string) {
	h.log.Error("scan failed", "id", scan.ID, "reason", reason)
	scan.Status = store.StatusFailed
	if err := h.store.Update(ctx, scan); err != nil {
		h.log.Error("update failed scan", "error", err)
	}
}

func toStoreFindings(findings []checks.Finding) []store.Finding {
	result := make([]store.Finding, 0, len(findings))
	for i, f := range findings {
		result = append(result, store.Finding{
			ID:             fmt.Sprintf("f%d", i+1),
			Kind:           kindFromCode(f.Code),
			Severity:       store.Severity(f.Severity),
			Title:          f.Title,
			Element:        f.Element,
			Recommendation: f.Recommendation,
			WCAG:           f.WCAG,
		})
	}

	return result
}

func kindFromCode(code string) string {
	switch {
	case strings.HasPrefix(code, "IMG_"):
		return "image"
	case strings.HasPrefix(code, "PDF_"):
		return "pdf"
	case strings.HasPrefix(code, "LINK_"):
		return "link"
	default:
		return "image"
	}
}
