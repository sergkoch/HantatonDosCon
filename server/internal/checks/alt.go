package checks

import (
	"fmt"
	"path"
	"strings"

	"accessible-content/server/internal/crawler"
)

var genericAlt = map[string]struct{}{
	"картинка":   {},
	"изображение": {},
	"фото":       {},
	"рисунок":    {},
	"image":      {},
	"img":        {},
	"picture":    {},
	"photo":      {},
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

// AltChecker проверяет описания изображений (T12).
type AltChecker struct{}

func (AltChecker) Check(page crawler.Page) []Finding {
	findings := make([]Finding, 0)

	for _, img := range page.Images {
		alt := strings.TrimSpace(img.Alt)

		switch {
		case !img.HasAlt:
			findings = append(findings, Finding{
				Code:           "IMG_ALT_MISSING",
				Severity:       "critical",
				Title:          "Нет alt у изображения",
				Element:        fmt.Sprintf(`<img src="%s">`, img.Src),
				Recommendation: "Добавьте описание, что на картинке. Если она декоративная — alt=\"\" и role=\"presentation\".",
				WCAG:           "1.1.1",
			})
		case alt == "":
			findings = append(findings, Finding{
				Code:           "IMG_ALT_EMPTY",
				Severity:       "critical",
				Title:          "Пустой alt у изображения",
				Element:        fmt.Sprintf(`<img src="%s" alt="">`, img.Src),
				Recommendation: "Добавьте описание картинки или отметьте её декоративной через role=\"presentation\".",
				WCAG:           "1.1.1",
			})
		case isGenericAlt(alt):
			findings = append(findings, Finding{
				Code:           "IMG_ALT_GENERIC",
				Severity:       "warning",
				Title:          "alt не описывает изображение",
				Element:        fmt.Sprintf(`<img src="%s" alt="%s">`, img.Src, alt),
				Recommendation: "Опишите, что изображено. Не подойдут «картинка», «фото» или имя файла.",
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
