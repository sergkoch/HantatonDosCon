package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"accessible-content/server/internal/ai"
	"accessible-content/server/internal/checks"
	"accessible-content/server/internal/crawler"
	"accessible-content/server/internal/report"
	"accessible-content/server/internal/store"
	"accessible-content/server/internal/texts"

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

// clientIDHeader carries an anonymous browser identity. History and reports
// are scoped to it, so every visitor sees only their own scans.
const clientIDHeader = "X-Client-ID"

// clientID returns the caller identity or an empty string when the header is
// missing or malformed. Empty identity means "no client" and keeps its scans
// separate from every real browser.
func clientID(c *gin.Context) string {
	raw := strings.TrimSpace(c.GetHeader(clientIDHeader))
	if len(raw) < 8 || len(raw) > 64 {
		return ""
	}
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return ""
		}
	}

	return raw
}

func isOwner(scan store.Scan, owner string) bool {
	return scan.Owner == owner
}

// requestLang выбирает язык ответа из заголовков запроса.
func requestLang(c *gin.Context) texts.Lang {
	return texts.FromHeaders(c.GetHeader("X-Lang"), c.GetHeader("Accept-Language"))
}

func (h *handler) createScan(c *gin.Context) {
	lang := requestLang(c)

	var request createScanRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "invalid_input", "detail": texts.T(lang, texts.ErrBodyTooLarge)})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "detail": texts.T(lang, texts.ErrInvalidJSON)})
		return
	}

	parsed, err := url.ParseRequestURI(request.URL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "detail": texts.T(lang, texts.ErrURLScheme)})
		return
	}

	if h.policy != nil {
		validateCtx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		validateErr := h.policy.ValidateURL(validateCtx, parsed.String())
		cancel()
		if validateErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "detail": texts.T(lang, texts.ErrURLPublic)})
			return
		}
	}

	if request.Depth == 0 {
		request.Depth = 1
	}

	if request.Depth < 1 || request.Depth > 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "detail": texts.T(lang, texts.ErrDepthRange)})
		return
	}

	owner := clientID(c)

	version := 1
	previousID := ""
	if previous, findErr := h.store.LatestDone(c.Request.Context(), parsed.String(), owner); findErr == nil {
		version = previous.Version + 1
		previousID = previous.ID
	}

	scan := store.Scan{
		ID:         uuid.NewString(),
		Owner:      owner,
		URL:        parsed.String(),
		Depth:      request.Depth,
		Status:     store.StatusQueued,
		Lang:       string(lang),
		Version:    version,
		PreviousID: previousID,
		Findings:   []store.Finding{},
		CreatedAt:  time.Now().UTC(),
	}

	if err := h.store.Create(c.Request.Context(), scan); err != nil {
		h.log.Error("create scan", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "store_error", "detail": texts.T(lang, texts.ErrStore)})
		return
	}

	go h.runScan(scan)

	c.JSON(http.StatusCreated, scan)
}

func (h *handler) listScans(c *gin.Context) {
	targetURL := strings.TrimSpace(c.Query("url"))
	limit := 20
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}

	scans, err := h.store.List(c.Request.Context(), targetURL, clientID(c), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "store_error", "detail": texts.T(requestLang(c), texts.ErrStore)})
		return
	}
	if scans == nil {
		scans = []store.Scan{}
	}

	c.JSON(http.StatusOK, scans)
}

func (h *handler) downloadPDF(c *gin.Context) {
	lang := requestLang(c)

	scan, err := h.store.Get(c.Request.Context(), c.Param("id"))
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "detail": texts.T(lang, texts.ErrNotFound)})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "store_error", "detail": texts.T(lang, texts.ErrStore)})
		return
	}
	if !isOwner(scan, clientID(c)) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "detail": texts.T(lang, texts.ErrNotFound)})
		return
	}

	var buffer bytes.Buffer
	if err := report.WritePDF(&buffer, scan); err != nil {
		h.log.Error("pdf report failed", "id", scan.ID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "report_error", "detail": texts.T(lang, texts.ErrReport)})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="report-%s.pdf"`, scan.ID))
	c.Data(http.StatusOK, "application/pdf", buffer.Bytes())
}

func (h *handler) getScan(c *gin.Context) {
	lang := requestLang(c)

	scan, err := h.store.Get(c.Request.Context(), c.Param("id"))
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "detail": texts.T(lang, texts.ErrNotFound)})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "store_error", "detail": texts.T(lang, texts.ErrStore)})
		return
	}

	if !isOwner(scan, clientID(c)) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "detail": texts.T(lang, texts.ErrNotFound)})
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

	lang := texts.Parse(scan.Lang)

	if h.crawler == nil {
		h.fail(ctx, scan, texts.T(lang, texts.ScanNoCrawler))
		return
	}

	pages, err := h.crawler.Crawl(ctx, scan.URL, scan.Depth)
	if err != nil {
		h.log.Error("crawl failed", "url", scan.URL, "error", err)
		h.fail(ctx, scan, texts.T(lang, texts.ScanCrawlFailed))
		return
	}

	if len(pages) == 0 {
		h.fail(ctx, scan, texts.T(lang, texts.ScanNoPages))
		return
	}

	scan.Pages = len(pages)

	altChecker := checks.AltChecker{Lang: lang}
	linkChecker := checks.LinkChecker{Lang: lang}
	pageChecker := checks.PageChecker{Lang: lang}
	formChecker := checks.FormChecker{Lang: lang}
	tableChecker := checks.TableChecker{Lang: lang}

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

	checkFindings = append(checkFindings, h.checkPDFs(ctx, pages, lang)...)

	if h.aiClient != nil {
		checkFindings = append(checkFindings, h.checkImages(ctx, pages, lang)...)
		h.runAILinkChecks(ctx, pages, &checkFindings, lang)
	}

	brokenLinks, checkedLinks, unverifiedLinks := checks.BrokenLinks(ctx, h.httpClient, links, 50, lang)
	checkFindings = append(checkFindings, brokenLinks...)

	critical, warning, ok := checks.Stats(checkFindings)
	score := checks.Score(checkFindings)

	scan.Findings = toStoreFindings(checkFindings)
	scan.Stats = store.Stats{Critical: critical, Warning: warning, OK: ok}
	scan.CheckedLinks = checkedLinks
	scan.UnverifiedLinks = unverifiedLinks
	scan.Score = &score
	scan.Status = store.StatusDone

	if scan.PreviousID != "" {
		previous, previousErr := h.store.Get(ctx, scan.PreviousID)
		if previousErr != nil {
			h.log.Warn("previous scan not found", "id", scan.PreviousID, "error", previousErr)
		} else {
			changes := store.Diff(previous.Findings, scan.Findings)
			if previous.Score != nil {
				delta := score - *previous.Score
				changes.ScoreDelta = &delta
			}
			scan.Changes = changes
		}
	}

	if err := h.store.Update(ctx, scan); err != nil {
		h.log.Error("update finished scan", "error", err)
	}
}

const (
	maxPDFsPerScan = 10
	maxPDFSize     = 20 << 20
)

func (h *handler) checkPDFs(ctx context.Context, pages []crawler.Page, lang texts.Lang) []checks.Finding {
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
			findings = append(findings, h.checkPDF(ctx, link, lang)...)
		}
	}

	return findings
}

func (h *handler) checkPDF(ctx context.Context, link crawler.Link, lang texts.Lang) []checks.Finding {
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
			Title:          texts.T(lang, texts.PDFNoTextTitle),
			Element:        element,
			Recommendation: texts.T(lang, texts.PDFNoTextRec),
			WCAG:           "1.3.1",
		})
	}

	if !result.Tagged {
		findings = append(findings, checks.Finding{
			Code:           "PDF_NOT_TAGGED",
			Severity:       "warning",
			Title:          texts.T(lang, texts.PDFNotTaggedTitle),
			Element:        element,
			Recommendation: texts.T(lang, texts.PDFNotTaggedRec),
			WCAG:           "1.3.1",
		})
	} else if result.AltTexts == 0 {
		findings = append(findings, checks.Finding{
			Code:           "PDF_ALT_MISSING",
			Severity:       "warning",
			Title:          texts.T(lang, texts.PDFAltMissingTitle),
			Element:        element,
			Recommendation: texts.T(lang, texts.PDFAltMissingRec),
			WCAG:           "1.1.1",
		})
	} else {
		findings = append(findings, checks.Finding{
			Code:           "PDF_ALT_OK",
			Severity:       "ok",
			Title:          texts.T(lang, texts.PDFAltOKTitle),
			Element:        element,
			Recommendation: texts.T(lang, texts.PDFAltOKRec),
			WCAG:           "1.1.1",
		})
	}

	if result.Title == "" {
		findings = append(findings, checks.Finding{
			Code:           "PDF_TITLE_MISSING",
			Severity:       "warning",
			Title:          texts.T(lang, texts.PDFTitleMissingTitle),
			Element:        element,
			Recommendation: texts.T(lang, texts.PDFTitleMissingRec),
			WCAG:           "2.4.2",
		})
	}

	if result.Lang == "" {
		findings = append(findings, checks.Finding{
			Code:           "PDF_LANG_MISSING",
			Severity:       "warning",
			Title:          texts.T(lang, texts.PDFLangMissingTitle),
			Element:        element,
			Recommendation: texts.T(lang, texts.PDFLangMissingRec),
			WCAG:           "3.1.1",
		})
	}

	return findings
}

const (
	maxImagesPerScan = 6
	maxImageSize     = 5 << 20
)

func (h *handler) checkImages(ctx context.Context, pages []crawler.Page, lang texts.Lang) []checks.Finding {
	if h.aiClient == nil || !h.aiVision {
		return nil
	}

	findings := make([]checks.Finding, 0)
	seen := make(map[string]struct{})

	for _, page := range pages {
		for _, image := range page.Images {
			if len(seen) >= maxImagesPerScan {
				return findings
			}
			if !image.HasAlt || strings.TrimSpace(image.Alt) == "" {
				continue
			}

			src := strings.TrimSpace(image.Src)
			parsed, err := url.Parse(src)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
				continue
			}
			if _, exists := seen[src]; exists {
				continue
			}
			seen[src] = struct{}{}

			findings = append(findings, h.checkImage(ctx, page.URL, image, lang)...)
		}
	}

	return findings
}

func (h *handler) checkImage(ctx context.Context, pageURL string, image crawler.Image, lang texts.Lang) []checks.Finding {
	requestCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(requestCtx, http.MethodGet, image.Src, nil)
	if err != nil {
		return nil
	}

	response, err := h.pdfClient.Do(request)
	if err != nil {
		h.log.Warn("image download failed", "src", image.Src, "error", err)
		return nil
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil
	}

	contentType := strings.TrimSpace(strings.Split(response.Header.Get("Content-Type"), ";")[0])
	if !strings.HasPrefix(contentType, "image/") || contentType == "image/svg+xml" {
		return nil
	}

	data, err := io.ReadAll(io.LimitReader(response.Body, maxImageSize+1))
	if err != nil || len(data) == 0 || len(data) > maxImageSize {
		h.log.Warn("image skipped", "src", image.Src, "size", len(data))
		return nil
	}

	result, err := h.aiClient.CheckImage(requestCtx, ai.ImageRequest{
		Image:    data,
		MimeType: contentType,
		Alt:      image.Alt,
		Context:  pageURL,
		Language: string(lang),
	})
	if err != nil {
		h.log.Warn("ai image check failed", "src", image.Src, "error", err)
		return nil
	}

	element := fmt.Sprintf(`<img src="%s" alt="%s">`, image.Src, image.Alt)

	switch strings.ToLower(strings.TrimSpace(result.Verdict)) {
	case "mismatch":
		recommendation := texts.T(lang, texts.AIMismatchRec)
		if strings.TrimSpace(result.Reason) != "" {
			recommendation += " " + strings.TrimSpace(result.Reason)
		}
		if strings.TrimSpace(result.SuggestedAlt) != "" {
			recommendation += texts.Tf(lang, texts.AISuggestedAlt, strings.TrimSpace(result.SuggestedAlt))
		}

		return []checks.Finding{{
			Code:           "IMG_ALT_AI_MISMATCH",
			Severity:       "critical",
			Title:          texts.T(lang, texts.AIMismatchTitle),
			Element:        element,
			Recommendation: recommendation,
			WCAG:           "1.1.1",
		}}
	case "match":
		recommendation := texts.T(lang, texts.AIMatchRec)
		if strings.TrimSpace(result.Reason) != "" {
			recommendation += " " + strings.TrimSpace(result.Reason)
		}

		return []checks.Finding{{
			Code:           "IMG_ALT_AI_OK",
			Severity:       "ok",
			Title:          texts.T(lang, texts.AIMatchTitle),
			Element:        element,
			Recommendation: recommendation,
			WCAG:           "1.1.1",
		}}
	default:
		return nil
	}
}

func (h *handler) runAILinkChecks(ctx context.Context, pages []crawler.Page, findings *[]checks.Finding, lang texts.Lang) {
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

			res, err := h.aiClient.CheckLink(ctx, ai.LinkRequest{Text: link.Text, URL: link.Href, Language: string(lang)})
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
				Title:          texts.T(lang, texts.AILinkTitle),
				Element:        fmt.Sprintf(`<a href="%s">%s</a>`, link.Href, link.Text),
				Recommendation: texts.Tf(lang, texts.AILinkRec, res.SuggestedText),
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
			Code:           f.Code,
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
