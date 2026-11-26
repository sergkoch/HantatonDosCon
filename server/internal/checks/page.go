package checks

import (
	"fmt"
	"strings"

	"accessible-content/server/internal/crawler"
	"accessible-content/server/internal/texts"
)

// PageChecker проверяет атрибуты страницы: язык, заголовок и структуру h1–h6.
type PageChecker struct {
	Lang texts.Lang
}

func (p PageChecker) Check(page crawler.Page) []Finding {
	findings := make([]Finding, 0)

	if strings.TrimSpace(page.Lang) == "" {
		findings = append(findings, Finding{
			Code:           "PAGE_LANG_MISSING",
			Severity:       "warning",
			Title:          texts.T(p.Lang, texts.PageLangTitle),
			Element:        "<html>",
			Recommendation: texts.T(p.Lang, texts.PageLangRec),
			WCAG:           "3.1.1",
		})
	}

	if strings.TrimSpace(page.Title) == "" {
		findings = append(findings, Finding{
			Code:           "PAGE_TITLE_MISSING",
			Severity:       "warning",
			Title:          texts.T(p.Lang, texts.PageTitleTitle),
			Element:        "<head>",
			Recommendation: texts.T(p.Lang, texts.PageTitleRec),
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
				Title:          texts.Tf(p.Lang, texts.PageHeadingSkipTitle, previous, heading.Level),
				Element:        headingElement(heading),
				Recommendation: texts.T(p.Lang, texts.PageHeadingSkipRec),
				WCAG:           "1.3.1",
			})
		}
		previous = heading.Level
	}

	if !hasH1 {
		findings = append(findings, Finding{
			Code:           "PAGE_H1_MISSING",
			Severity:       "warning",
			Title:          texts.T(p.Lang, texts.PageH1Title),
			Element:        "<body>",
			Recommendation: texts.T(p.Lang, texts.PageH1Rec),
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
type FormChecker struct {
	Lang texts.Lang
}

func (f FormChecker) Check(page crawler.Page) []Finding {
	findings := make([]Finding, 0)

	for _, input := range page.Inputs {
		if input.HasLabel {
			continue
		}

		findings = append(findings, Finding{
			Code:           "FORM_LABEL_MISSING",
			Severity:       "critical",
			Title:          texts.T(f.Lang, texts.FormLabelTitle),
			Element:        inputElement(input),
			Recommendation: texts.T(f.Lang, texts.FormLabelRec),
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
type TableChecker struct {
	Lang texts.Lang
}

func (t TableChecker) Check(page crawler.Page) []Finding {
	findings := make([]Finding, 0)

	for i, table := range page.Tables {
		if table.HasHeaders {
			continue
		}

		findings = append(findings, Finding{
			Code:           "TABLE_HEADERS_MISSING",
			Severity:       "warning",
			Title:          texts.Tf(t.Lang, texts.TableHeadersTitle, i+1),
			Element:        "<table>",
			Recommendation: texts.T(t.Lang, texts.TableHeadersRec),
			WCAG:           "1.3.1",
		})
	}

	return findings
}
