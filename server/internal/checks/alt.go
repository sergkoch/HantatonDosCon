package checks

import (
	"fmt"
	"path"
	"strings"
	"unicode/utf8"

	"accessible-content/server/internal/crawler"
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
		case alt == "" && img.InLink && strings.TrimSpace(img.LinkText) == "":
			findings = append(findings, Finding{
				Code:           "IMG_ALT_EMPTY_LINK",
				Severity:       "critical",
				Title:          "Изображение-ссылка без текстовой альтернативы",
				Element:        fmt.Sprintf(`<a><img src="%s" alt=""></a>`, img.Src),
				Recommendation: "Добавьте alt, который объясняет, куда ведёт ссылка, или текстовую подпись рядом с картинкой.",
				WCAG:           "1.1.1",
			})
		case alt == "":
			// Пустой alt — корректный способ пометить декоративное изображение.
		case isGenericAlt(alt):
			findings = append(findings, Finding{
				Code:           "IMG_ALT_GENERIC",
				Severity:       "warning",
				Title:          "alt не описывает изображение",
				Element:        fmt.Sprintf(`<img src="%s" alt="%s">`, img.Src, alt),
				Recommendation: "Опишите, что изображено. Не подойдут «картинка», «фото» или имя файла.",
				WCAG:           "1.1.1",
			})
		case utf8.RuneCountInString(alt) > 125:
			findings = append(findings, Finding{
				Code:           "IMG_ALT_TOO_LONG",
				Severity:       "warning",
				Title:          "Слишком длинный alt",
				Element:        fmt.Sprintf(`<img src="%s" alt="%s">`, img.Src, alt),
				Recommendation: "Сократите описание до 125 символов: короткое описание легче воспринять на слух.",
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
