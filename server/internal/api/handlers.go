package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "invalid_input", "detail": "request body too large"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "detail": err.Error()})
		return
	}

	parsed, err := url.ParseRequestURI(request.URL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "detail": "url must be http or https"})
		return
	}

	if h.policy != nil {
		validateCtx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		validateErr := h.policy.ValidateURL(validateCtx, parsed.String())
		cancel()
		if validateErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "detail": "url must be public"})
			return
		}
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

	checkFindings = append(checkFindings, h.checkPDFs(ctx, pages)...)

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

const (
	maxPDFsPerScan = 10
	maxPDFSize     = 20 << 20
)

func (h *handler) checkPDFs(ctx context.Context, pages []crawler.Page) []checks.Finding {
	if h.pdfAnalyzer == nil {
		return nil
	}

	findings := make([]checks.Finding, 0)
	seen := make(map[string]struct{})

	for _, page := range pages {
		for _, link := range page.PDFs {
			if len(seen) >= maxPDFsPerScan {
				return findings
			}
			if _, ok := seen[link.Href]; ok {
				continue
			}
			seen[link.Href] = struct{}{}
			findings = append(findings, h.checkPDF(ctx, link)...)
		}
	}

	return findings
}

func (h *handler) checkPDF(ctx context.Context, link crawler.Link) []checks.Finding {
	requestCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, link.Href, nil)
	if err != nil {
		return nil
	}

	response, err := h.pdfClient.Do(request)
	if err != nil {
		h.log.Warn("pdf download failed", "href", link.Href, "error", err)
		return nil
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil
	}

	data, err := io.ReadAll(io.LimitReader(response.Body, maxPDFSize+1))
	if err != nil {
		h.log.Warn("pdf read failed", "href", link.Href, "error", err)
		return nil
	}
	if len(data) == 0 || len(data) > maxPDFSize {
		h.log.Warn("pdf skipped", "href", link.Href, "size", len(data))
		return nil
	}

	result, err := h.pdfAnalyzer.Analyze(requestCtx, bytes.NewReader(data), int64(len(data)))
	if err != nil {
		h.log.Warn("pdf analysis failed", "href", link.Href, "error", err)
		return nil
	}

	element := fmt.Sprintf(`<a href="%s">%s</a>`, link.Href, link.Text)
	findings := make([]checks.Finding, 0, 5)

	if !result.TextLayer {
		findings = append(findings, checks.Finding{
			Code:           "PDF_NO_TEXT",
			Severity:       "critical",
			Title:          "PDF без текстового слоя",
			Element:        element,
			Recommendation: "Опубликуйте PDF с текстовым слоем или HTML-версию: сканы без текста плохо читаются скринридером.",
			WCAG:           "1.3.1",
		})
	}

	if !result.Tagged {
		findings = append(findings, checks.Finding{
			Code:           "PDF_NOT_TAGGED",
			Severity:       "warning",
			Title:          "PDF без тегов структуры",
			Element:        element,
			Recommendation: "Пересохраните PDF с включённой разметкой структуры (Tagged PDF), чтобы скринридер понимал порядок чтения.",
			WCAG:           "1.3.1",
		})
	} else if result.AltTexts == 0 {
		findings = append(findings, checks.Finding{
			Code:           "PDF_ALT_MISSING",
			Severity:       "warning",
			Title:          "В PDF не найден /Alt",
			Element:        element,
			Recommendation: "Добавьте /Alt к смысловым изображениям и схемам в PDF.",
			WCAG:           "1.1.1",
		})
	} else {
		findings = append(findings, checks.Finding{
			Code:           "PDF_ALT_OK",
			Severity:       "ok",
			Title:          "В PDF есть теги и /Alt",
			Element:        element,
			Recommendation: "Ничего не нужно: структура и альтернативный текст на месте.",
			WCAG:           "1.1.1",
		})
	}

	if result.Title == "" {
		findings = append(findings, checks.Finding{
			Code:           "PDF_TITLE_MISSING",
			Severity:       "warning",
			Title:          "PDF без заголовка",
			Element:        element,
			Recommendation: "Заполните Title в свойствах документа: он показывается во вкладке и в списке файлов.",
			WCAG:           "2.4.2",
		})
	}

	if result.Lang == "" {
		findings = append(findings, checks.Finding{
			Code:           "PDF_LANG_MISSING",
			Severity:       "warning",
			Title:          "PDF без языка документа",
			Element:        element,
			Recommendation: "Укажите язык документа в свойствах PDF, чтобы скринридер выбрал правильное произношение.",
			WCAG:           "3.1.1",
		})
	}

	return findings
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
