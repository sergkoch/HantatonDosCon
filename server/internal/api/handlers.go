package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"accessible-content/server/internal/ai"
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

	if len(pages) == 0 {
		h.fail(ctx, scan, "не удалось обойти ни одной страницы: сайт недоступен, закрыт robots.txt или отдаёт контент только через JavaScript")
		return
	}

	scan.Pages = len(pages)

	altChecker := checks.AltChecker{}
	linkChecker := checks.LinkChecker{}
	pageChecker := checks.PageChecker{}
	formChecker := checks.FormChecker{}
	tableChecker := checks.TableChecker{}

	checkFindings := make([]checks.Finding, 0)
	links := make([]crawler.Link, 0)
	for _, page := range pages {
		checkFindings = append(checkFindings, altChecker.Check(page)...)
		checkFindings = append(checkFindings, linkChecker.Check(page)...)
		checkFindings = append(checkFindings, pageChecker.Check(page)...)
		checkFindings = append(checkFindings, formChecker.Check(page)...)
		checkFindings = append(checkFindings, tableChecker.Check(page)...)
		links = append(links, page.Links...)
		links = append(links, page.PDFs...)
	}

	if h.aiClient != nil {
		h.runAILinkChecks(ctx, pages, &checkFindings)
	}

	brokenLinks, checkedLinks, unverifiedLinks := checks.BrokenLinks(ctx, h.httpClient, links, 50)
	checkFindings = append(checkFindings, brokenLinks...)

	critical, warning, ok := checks.Stats(checkFindings)
	score := checks.Score(checkFindings)

	scan.Findings = toStoreFindings(checkFindings)
	scan.Stats = store.Stats{Critical: critical, Warning: warning, OK: ok}
	scan.CheckedLinks = checkedLinks
	scan.UnverifiedLinks = unverifiedLinks
	scan.Score = &score
	scan.Status = store.StatusDone

	if err := h.store.Update(ctx, scan); err != nil {
		h.log.Error("update finished scan", "error", err)
	}
}

func (h *handler) runAILinkChecks(ctx context.Context, pages []crawler.Page, findings *[]checks.Finding) {
	const budget = 8
	seen := make(map[string]struct{})
	left := budget

	for _, page := range pages {
		links := make([]crawler.Link, 0, len(page.Links)+len(page.PDFs))
		links = append(links, page.Links...)
		links = append(links, page.PDFs...)

		for _, link := range links {
			if left <= 0 {
				return
			}
			if !checks.IsGenericLinkText(link.Text) {
				continue
			}
			if _, ok := seen[link.Href]; ok {
				continue
			}
			seen[link.Href] = struct{}{}
			left--

			res, err := h.aiClient.CheckLink(ctx, ai.LinkRequest{Text: link.Text, URL: link.Href})
			if err != nil {
				h.log.Warn("ai link check failed", "href", link.Href, "error", err)
				continue
			}
			if res.SuggestedText == "" {
				continue
			}

			*findings = append(*findings, checks.Finding{
				Code:           "LINK_TEXT_MISMATCH",
				Severity:       "warning",
				Title:          "Текст ссылки не соответствует цели",
				Element:        fmt.Sprintf(`<a href="%s">%s</a>`, link.Href, link.Text),
				Recommendation: fmt.Sprintf("Предлагаемый текст: «%s».", res.SuggestedText),
				WCAG:           "2.4.4",
			})
		}
	}
}

func (h *handler) fail(ctx context.Context, scan store.Scan, reason string) {
	h.log.Error("scan failed", "id", scan.ID, "reason", reason)
	scan.Status = store.StatusFailed
	scan.Error = reason
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
	case strings.HasPrefix(code, "PAGE_"):
		return "page"
	case strings.HasPrefix(code, "FORM_"):
		return "form"
	case strings.HasPrefix(code, "TABLE_"):
		return "table"
	default:
		return "image"
	}
}
