package checks

import (
	"fmt"
	"strings"

	"accessible-content/server/internal/crawler"
)

// PageChecker проверяет атрибуты страницы: язык, заголовок и структуру h1–h6.
type PageChecker struct{}

func (PageChecker) Check(page crawler.Page) []Finding {
	findings := make([]Finding, 0)

	if strings.TrimSpace(page.Lang) == "" {
		findings = append(findings, Finding{
			Code:           "PAGE_LANG_MISSING",
			Severity:       "warning",
			Title:          "Не указан язык страницы",
			Element:        "<html>",
			Recommendation: "Добавьте атрибут lang, например <html lang=\"ru\">: от него зависит произношение в скринридере.",
			WCAG:           "3.1.1",
		})
	}

	if strings.TrimSpace(page.Title) == "" {
		findings = append(findings, Finding{
			Code:           "PAGE_TITLE_MISSING",
			Severity:       "warning",
			Title:          "Нет заголовка страницы",
			Element:        "<head>",
			Recommendation: "Добавьте <title>: по нему ориентируются во вкладках и в списке страниц скринридера.",
			WCAG:           "2.4.2",
		})
	}

	hasH1 := false
	previous := 0

	for _, heading := range page.Headings {
		if heading.Level == 1 {
			hasH1 = true
		}
		if previous != 0 && heading.Level > previous+1 {
			findings = append(findings, Finding{
				Code:           "PAGE_HEADING_SKIP",
				Severity:       "warning",
				Title:          fmt.Sprintf("Пропущен уровень заголовка: h%d → h%d", previous, heading.Level),
				Element:        headingElement(heading),
				Recommendation: "Не пропускайте уровни: после h1 идёт h2, затем h3. Иначе структура страницы читается неверно.",
				WCAG:           "1.3.1",
			})
		}
		previous = heading.Level
	}

	if !hasH1 {
		findings = append(findings, Finding{
			Code:           "PAGE_H1_MISSING",
			Severity:       "warning",
			Title:          "Нет заголовка h1",
			Element:        "<body>",
			Recommendation: "Добавьте один h1 — главный заголовок страницы.",
			WCAG:           "1.3.1",
		})
	}

	return findings
}

func headingElement(heading crawler.Heading) string {
	text := strings.TrimSpace(heading.Text)
	if text == "" {
		return fmt.Sprintf("<h%d>", heading.Level)
	}
	return fmt.Sprintf("<h%d>%s</h%d>", heading.Level, text, heading.Level)
}

// FormChecker проверяет, что у полей формы есть доступные подписи.
type FormChecker struct{}

func (FormChecker) Check(page crawler.Page) []Finding {
	findings := make([]Finding, 0)

	for _, input := range page.Inputs {
		if input.HasLabel {
			continue
		}

		findings = append(findings, Finding{
			Code:           "FORM_LABEL_MISSING",
			Severity:       "critical",
			Title:          "Поле формы без подписи",
			Element:        inputElement(input),
			Recommendation: "Свяжите поле с <label for=\"...\"> или добавьте aria-label. Placeholder не заменяет подпись.",
			WCAG:           "3.3.2",
		})
	}

	return findings
}

func inputElement(input crawler.Input) string {
	tag := inputTag(input.Type)
	attrs := ""

	if tag == "input" {
		attrs += fmt.Sprintf(` type="%s"`, input.Type)
	}
	if input.Name != "" {
		attrs += fmt.Sprintf(` name="%s"`, input.Name)
	}
	if input.Placeholder != "" {
		attrs += fmt.Sprintf(` placeholder="%s"`, input.Placeholder)
	}

	return fmt.Sprintf("<%s%s>", tag, attrs)
}

func inputTag(kind string) string {
	switch kind {
	case "select", "textarea":
		return kind
	default:
		return "input"
	}
}

// TableChecker проверяет, что у таблиц есть заголовки столбцов.
type TableChecker struct{}

func (TableChecker) Check(page crawler.Page) []Finding {
	findings := make([]Finding, 0)

	for i, table := range page.Tables {
		if table.HasHeaders {
			continue
		}

		findings = append(findings, Finding{
			Code:           "TABLE_HEADERS_MISSING",
			Severity:       "warning",
			Title:          fmt.Sprintf("Таблица без заголовков (№%d)", i+1),
			Element:        "<table>",
			Recommendation: "Добавьте <th scope=\"col\"> или role=\"columnheader\": без заголовков непонятно, что означает каждая ячейка.",
			WCAG:           "1.3.1",
		})
	}

	return findings
}
