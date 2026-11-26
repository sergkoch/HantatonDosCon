package checks

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"accessible-content/server/internal/crawler"
	"accessible-content/server/internal/texts"
)

const linkCheckUserAgent = "Mozilla/5.0 (compatible; AccessibleContentBot/0.1; +https://github.com/sergkoch/HantatonDosCon)"

var genericLinkText = map[string]struct{}{
	"подробнее":    {},
	"здесь":        {},
	"тут":          {},
	"ссылка":       {},
	"читать":       {},
	"читать далее": {},
	"далее":        {},
	"перейти":      {},
	"клик":         {},
	"кликните":     {},
	"click here":   {},
	"here":         {},
	"more":         {},
	"read more":    {},
	"link":         {},
}

// IsGenericLinkText сообщает, является ли текст ссылки обобщённым («подробнее», «здесь» и т.п.).
func IsGenericLinkText(text string) bool {
	_, ok := genericLinkText[strings.ToLower(strings.TrimSpace(text))]
	return ok
}

// LinkChecker проверяет текст ссылок (T13). Lang задаёт язык находок.
type LinkChecker struct {
	Lang texts.Lang
}

func (l LinkChecker) Check(page crawler.Page) []Finding {
	findings := make([]Finding, 0)

	links := make([]crawler.Link, 0, len(page.Links)+len(page.PDFs))
	links = append(links, page.Links...)
	links = append(links, page.PDFs...)

	for _, link := range links {
		text := strings.TrimSpace(link.Text)
		name := text
		if name == "" {
			name = strings.TrimSpace(link.Label)
		}

		if name == "" {
			findings = append(findings, Finding{
				Code:           "LINK_TEXT_EMPTY",
				Severity:       "warning",
				Title:          texts.T(l.Lang, texts.LinkEmptyTitle),
				Element:        fmt.Sprintf(`<a href="%s"></a>`, link.Href),
				Recommendation: texts.T(l.Lang, texts.LinkEmptyRec),
				WCAG:           "2.4.4",
			})
			continue
		}

		if text == "" {
			continue
		}

		if _, ok := genericLinkText[strings.ToLower(text)]; ok {
			findings = append(findings, Finding{
				Code:           "LINK_TEXT_GENERIC",
				Severity:       "warning",
				Title:          texts.T(l.Lang, texts.LinkGenericTitle),
				Element:        linkElement(link),
				Recommendation: texts.Tf(l.Lang, texts.LinkGenericRec, text),
				WCAG:           "2.4.4",
			})
		}
	}

	return findings
}

// BrokenLinks проверяет доступность ссылок по HTTP (T13).
// Битыми считаются только 404/410. Ответы 401/403/429/5xx и сетевые ошибки
// означают «не удалось проверить»: это часто защита от ботов, а не мёртвая ссылка.
// Возвращает находки, количество проверенных и количество непроверенных ссылок.
func BrokenLinks(ctx context.Context, client *http.Client, links []crawler.Link, max int, lang texts.Lang) ([]Finding, int, int) {
	if max <= 0 {
		max = 50
	}
	if client == nil {
		client = &http.Client{Timeout: 6 * time.Second}
	}

	seen := make(map[string]struct{}, len(links))
	candidates := make([]crawler.Link, 0, max)

	for _, link := range links {
		if len(candidates) >= max {
			break
		}

		href := strings.TrimSpace(link.Href)
		if href == "" || strings.HasPrefix(href, "#") {
			continue
		}

		lower := strings.ToLower(href)
		if strings.HasPrefix(lower, "javascript:") ||
			strings.HasPrefix(lower, "mailto:") ||
			strings.HasPrefix(lower, "tel:") {
			continue
		}

		if _, exists := seen[href]; exists {
			continue
		}
		seen[href] = struct{}{}
		candidates = append(candidates, crawler.Link{Href: href, Text: strings.TrimSpace(link.Text), Label: strings.TrimSpace(link.Label)})
	}

	type checkResult struct {
		link   crawler.Link
		status int
		err    error
	}

	results := make([]checkResult, len(candidates))
	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup

	for i, link := range candidates {
		wg.Add(1)
		go func(i int, link crawler.Link) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			status, err := statusOf(ctx, client, link.Href)
			results[i] = checkResult{link: link, status: status, err: err}
		}(i, link)
	}
	wg.Wait()

	findings := make([]Finding, 0)
	unverified := 0

	for _, result := range results {
		switch {
		case result.err != nil:
			unverified++
		case result.status == http.StatusNotFound || result.status == http.StatusGone:
			findings = append(findings, Finding{
				Code:           "LINK_BROKEN",
				Severity:       "critical",
				Title:          texts.T(lang, texts.LinkBrokenTitle),
				Element:        linkElement(result.link),
				Recommendation: texts.T(lang, texts.LinkBrokenRec),
				WCAG:           "2.4.4",
			})
		case result.status >= http.StatusBadRequest:
			unverified++
		}
	}

	return findings, len(candidates), unverified
}

func linkElement(link crawler.Link) string {
	text := strings.TrimSpace(link.Text)
	if text == "" {
		text = strings.TrimSpace(link.Label)
	}
	return fmt.Sprintf(`<a href="%s">%s</a>`, link.Href, text)
}

func statusOf(ctx context.Context, client *http.Client, href string) (int, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, href, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", linkCheckUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Range", "bytes=0-0")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}
