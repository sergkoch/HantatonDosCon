package checks

import (
	"fmt"
	"path"
	"strings"
	"unicode/utf8"

	"accessible-content/server/internal/crawler"
	"accessible-content/server/internal/texts"
)

var genericAlt = map[string]struct{}{
	"картинка":    {},
	"изображение": {},
	"фото":        {},
	"рисунок":     {},
	"image":       {},
	"img":         {},
	"picture":     {},
	"photo":       {},
}

var imageExt = map[string]struct{}{
	".png":  {},
	".jpg":  {},
	".jpeg": {},
	".gif":  {},
	".svg":  {},
	".webp": {},
	".bmp":  {},
	".avif": {},
}

// AltChecker проверяет описания изображений (T12). Lang задаёт язык находок.
type AltChecker struct {
	Lang texts.Lang
}

func (a AltChecker) Check(page crawler.Page) []Finding {
	findings := make([]Finding, 0)

	for _, img := range page.Images {
		alt := strings.TrimSpace(img.Alt)

		switch {
		case !img.HasAlt:
			findings = append(findings, Finding{
				Code:           "IMG_ALT_MISSING",
				Severity:       "critical",
				Title:          texts.T(a.Lang, texts.AltMissingTitle),
				Element:        fmt.Sprintf(`<img src="%s">`, img.Src),
				Recommendation: texts.T(a.Lang, texts.AltMissingRec),
				WCAG:           "1.1.1",
			})
		case alt == "" && img.InLink && strings.TrimSpace(img.LinkText) == "":
			findings = append(findings, Finding{
				Code:           "IMG_ALT_EMPTY_LINK",
				Severity:       "critical",
				Title:          texts.T(a.Lang, texts.AltEmptyLinkTitle),
				Element:        fmt.Sprintf(`<a><img src="%s" alt=""></a>`, img.Src),
				Recommendation: texts.T(a.Lang, texts.AltEmptyLinkRec),
				WCAG:           "1.1.1",
			})
		case alt == "":
			// Пустой alt — корректный способ пометить декоративное изображение.
		case isGenericAlt(alt):
			findings = append(findings, Finding{
				Code:           "IMG_ALT_GENERIC",
				Severity:       "warning",
				Title:          texts.T(a.Lang, texts.AltGenericTitle),
				Element:        fmt.Sprintf(`<img src="%s" alt="%s">`, img.Src, alt),
				Recommendation: texts.T(a.Lang, texts.AltGenericRec),
				WCAG:           "1.1.1",
			})
		case utf8.RuneCountInString(alt) > 125:
			findings = append(findings, Finding{
				Code:           "IMG_ALT_TOO_LONG",
				Severity:       "warning",
				Title:          texts.T(a.Lang, texts.AltTooLongTitle),
				Element:        fmt.Sprintf(`<img src="%s" alt="%s">`, img.Src, alt),
				Recommendation: texts.T(a.Lang, texts.AltTooLongRec),
				WCAG:           "1.1.1",
			})
		}
	}

	return findings
}

func isGenericAlt(alt string) bool {
	lower := strings.ToLower(strings.TrimSpace(alt))

	if _, ok := genericAlt[lower]; ok {
		return true
	}

	if _, ok := imageExt[path.Ext(lower)]; ok {
		return true
	}

	return false
}
