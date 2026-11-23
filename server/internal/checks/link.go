package checks

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"accessible-content/server/internal/crawler"
)

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

// LinkChecker проверяет текст ссылок (T13).
type LinkChecker struct{}

func (LinkChecker) Check(page crawler.Page) []Finding {
	findings := make([]Finding, 0)

	links := make([]crawler.Link, 0, len(page.Links)+len(page.PDFs))
	links = append(links, page.Links...)
	links = append(links, page.PDFs...)

	for _, link := range links {
		text := strings.TrimSpace(link.Text)
		lower := strings.ToLower(text)

		if text == "" {
			findings = append(findings, Finding{
				Code:           "LINK_TEXT_EMPTY",
				Severity:       "warning",
				Title:          "Ссылка без текста",
				Element:        fmt.Sprintf(`<a href="%s"></a>`, link.Href),
				Recommendation: "Добавьте текст, который объясняет, куда ведёт ссылка.",
				WCAG:           "2.4.4",
			})
			continue
		}

		if _, ok := genericLinkText[lower]; ok {
			findings = append(findings, Finding{
				Code:           "LINK_TEXT_GENERIC",
				Severity:       "warning",
				Title:          "Непонятный текст ссылки",
				Element:        fmt.Sprintf(`<a href="%s">%s</a>`, link.Href, text),
				Recommendation: fmt.Sprintf("Напишите, что откроется: например «Положение о конкурсе» вместо «%s».", text),
				WCAG:           "2.4.4",
			})
		}
	}

	return findings
}

// BrokenLinks проверяет доступность ссылок по HTTP (T13).
func BrokenLinks(ctx context.Context, client *http.Client, links []crawler.Link, max int) []Finding {
	if max <= 0 {
		max = 50
	}
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	seen := make(map[string]struct{}, len(links))
	findings := make([]Finding, 0)
	checked := 0

	for _, link := range links {
		if checked >= max {
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
		checked++

		status, err := statusOf(ctx, client, href)
		if err != nil || status >= http.StatusBadRequest {
			findings = append(findings, Finding{
				Code:           "LINK_BROKEN",
				Severity:       "critical",
				Title:          "Битая ссылка",
				Element:        fmt.Sprintf(`<a href="%s">%s</a>`, href, strings.TrimSpace(link.Text)),
				Recommendation: "Исправьте адрес ссылки или удалите её — сейчас она ведёт в никуда.",
				WCAG:           "2.4.4",
			})
		}
	}

	return findings
}

func statusOf(ctx context.Context, client *http.Client, href string) (int, error) {
	headCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(headCtx, http.MethodHead, href, nil)
	if err == nil {
		resp, headErr := client.Do(req)
		if headErr == nil {
			defer resp.Body.Close()
			return resp.StatusCode, nil
		}
	}

	getCtx, getCancel := context.WithTimeout(ctx, 5*time.Second)
	defer getCancel()

	req, err = http.NewRequestWithContext(getCtx, http.MethodGet, href, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}
