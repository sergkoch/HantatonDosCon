// Package texts holds user-facing strings of the service in every supported
// language. Findings are stored together with the language of the scan, so a
// report keeps the language it was scanned in.
package texts

import (
	"fmt"
	"strings"
)

type Lang string

const (
	RU Lang = "ru"
	EN Lang = "en"
)

// Parse нормализует язык: всё, что начинается с «en», считаем английским,
// остальное — русским.
func Parse(raw string) Lang {
	value := strings.ToLower(strings.TrimSpace(raw))
	if strings.HasPrefix(value, "en") {
		return EN
	}
	return RU
}

// FromHeaders выбирает язык запроса: сначала явный X-Lang, затем Accept-Language.
func FromHeaders(explicit, acceptLanguage string) Lang {
	if strings.TrimSpace(explicit) != "" {
		return Parse(explicit)
	}
	return Parse(acceptLanguage)
}

// Other возвращает второй поддерживаемый язык.
func Other(lang Lang) Lang {
	if Parse(string(lang)) == EN {
		return RU
	}
	return EN
}

type Key string

const (
	AltMissingTitle       Key = "alt.missing.title"
	AltMissingRec         Key = "alt.missing.rec"
	AltEmptyLinkTitle     Key = "alt.emptyLink.title"
	AltEmptyLinkRec       Key = "alt.emptyLink.rec"
	AltGenericTitle       Key = "alt.generic.title"
	AltGenericRec         Key = "alt.generic.rec"
	AltTooLongTitle       Key = "alt.tooLong.title"
	AltTooLongRec         Key = "alt.tooLong.rec"
	LinkEmptyTitle        Key = "link.empty.title"
	LinkEmptyRec          Key = "link.empty.rec"
	LinkGenericTitle      Key = "link.generic.title"
	LinkGenericRec        Key = "link.generic.rec"
	LinkBrokenTitle       Key = "link.broken.title"
	LinkBrokenRec         Key = "link.broken.rec"
	PageLangTitle         Key = "page.lang.title"
	PageLangRec           Key = "page.lang.rec"
	PageTitleTitle        Key = "page.title.title"
	PageTitleRec          Key = "page.title.rec"
	PageHeadingSkipTitle  Key = "page.heading.title"
	PageHeadingSkipRec    Key = "page.heading.rec"
	PageH1Title           Key = "page.h1.title"
	PageH1Rec             Key = "page.h1.rec"
	FormLabelTitle        Key = "form.label.title"
	FormLabelRec          Key = "form.label.rec"
	TableHeadersTitle     Key = "table.headers.title"
	TableHeadersRec       Key = "table.headers.rec"
	PDFNoTextTitle        Key = "pdf.noText.title"
	PDFNoTextRec          Key = "pdf.noText.rec"
	PDFNotTaggedTitle     Key = "pdf.notTagged.title"
	PDFNotTaggedRec       Key = "pdf.notTagged.rec"
	PDFAltMissingTitle    Key = "pdf.altMissing.title"
	PDFAltMissingRec      Key = "pdf.altMissing.rec"
	PDFAltOKTitle         Key = "pdf.altOk.title"
	PDFAltOKRec           Key = "pdf.altOk.rec"
	PDFTitleMissingTitle  Key = "pdf.titleMissing.title"
	PDFTitleMissingRec    Key = "pdf.titleMissing.rec"
	PDFLangMissingTitle   Key = "pdf.langMissing.title"
	PDFLangMissingRec     Key = "pdf.langMissing.rec"
	AIMismatchTitle       Key = "ai.mismatch.title"
	AIMismatchRec         Key = "ai.mismatch.rec"
	AIMismatchUnsureTitle Key = "ai.mismatchUnsure.title"
	AIMismatchUnsureRec   Key = "ai.mismatchUnsure.rec"
	AIConfidence          Key = "ai.confidence"
	AISuggestedAlt        Key = "ai.mismatch.suggested"
	AIMatchTitle          Key = "ai.match.title"
	AIMatchRec            Key = "ai.match.rec"
	AILinkTitle           Key = "ai.link.title"
	AILinkRec             Key = "ai.link.rec"
	AILinkRecPlain        Key = "ai.link.recPlain"
	ScanNoCrawler         Key = "scan.fail.noCrawler"
	ScanCrawlFailed       Key = "scan.fail.crawl"
	ScanNoPages           Key = "scan.fail.noPages"
	ScanTimeout           Key = "scan.fail.timeout"
	ErrURLScheme          Key = "err.urlScheme"
	ErrURLPublic          Key = "err.urlPublic"
	ErrDepthRange         Key = "err.depth"
	ErrBodyTooLarge       Key = "err.bodyTooLarge"
	ErrInvalidJSON        Key = "err.invalidJSON"
	ErrNotFound           Key = "err.notFound"
	ErrStore              Key = "err.store"
	ErrReport             Key = "err.report"
	ErrClientID           Key = "err.clientId"
	ErrGateMinScore       Key = "err.gateMinScore"
	ErrGateMaxCritical    Key = "err.gateMaxCritical"
	ErrGateNotReady       Key = "err.gateNotReady"
	ErrGateFailed         Key = "err.gateFailed"
	ErrGateScore          Key = "err.gateScore"
	ErrGateCritical       Key = "err.gateCritical"
	DocTitle              Key = "doc.title"
	DocAuthor             Key = "doc.author"
	PDFPage               Key = "pdf.page"
	PDFSite               Key = "pdf.site"
	PDFScan               Key = "pdf.scan"
	PDFScanValue          Key = "pdf.scan.value"
	PDFDepth              Key = "pdf.depth"
	PDFScore              Key = "pdf.score"
	PDFFindings           Key = "pdf.findings"
	PDFStatsValue         Key = "pdf.stats.value"
	PDFCoverage           Key = "pdf.coverage"
	PDFCoverageValue      Key = "pdf.coverage.value"
	PDFChanges            Key = "pdf.changes"
	PDFAdded              Key = "pdf.added"
	PDFFixed              Key = "pdf.fixed"
	PDFRemaining          Key = "pdf.remaining"
	PDFDelta              Key = "pdf.delta"
	PDFNoFindings         Key = "pdf.none"
	PDFProblemsOnly       Key = "pdf.problemsOnly"
	PDFType               Key = "pdf.type"
	PDFElement            Key = "pdf.element"
	PDFRecommendation     Key = "pdf.recommendation"
	PDFScoreOf            Key = "pdf.scoreOf"
	PDFFootnote           Key = "pdf.footnote"
	PDFDateFormat         Key = "pdf.dateFormat"
	SeverityCriticalLabel Key = "severity.critical"
	SeverityWarningLabel  Key = "severity.warning"
	SeverityOKLabel       Key = "severity.ok"
	KindImage             Key = "kind.image"
	KindPDF               Key = "kind.pdf"
	KindLink              Key = "kind.link"
	KindPage              Key = "kind.page"
	KindForm              Key = "kind.form"
	KindTable             Key = "kind.table"
)

var catalog = map[Lang]map[Key]string{
	RU: {
		AltMissingTitle:   "Нет alt у изображения",
		AltMissingRec:     "Добавьте описание, что на картинке. Если она декоративная — alt=\"\" и role=\"presentation\".",
		AltEmptyLinkTitle: "Изображение-ссылка без текстовой альтернативы",
		AltEmptyLinkRec:   "Добавьте alt, который объясняет, куда ведёт ссылка, или текстовую подпись рядом с картинкой.",
		AltGenericTitle:   "alt не описывает изображение",
		AltGenericRec:     "Опишите, что изображено. Не подойдут «картинка», «фото» или имя файла.",
		AltTooLongTitle:   "Слишком длинный alt",
		AltTooLongRec:     "Сократите описание до 125 символов: короткое описание легче воспринять на слух.",

		LinkEmptyTitle:   "Ссылка без текста",
		LinkEmptyRec:     "Добавьте текст, который объясняет, куда ведёт ссылка.",
		LinkGenericTitle: "Непонятный текст ссылки",
		LinkGenericRec:   "Напишите, что откроется: например «Положение о конкурсе» вместо «%s».",
		LinkBrokenTitle:  "Битая ссылка",
		LinkBrokenRec:    "Исправьте адрес ссылки или удалите её — сейчас она ведёт в никуда.",

		PageLangTitle:        "Не указан язык страницы",
		PageLangRec:          "Добавьте атрибут lang, например <html lang=\"ru\">: от него зависит произношение в скринридере.",
		PageTitleTitle:       "Нет заголовка страницы",
		PageTitleRec:         "Добавьте <title>: по нему ориентируются во вкладках и в списке страниц скринридера.",
		PageHeadingSkipTitle: "Пропущен уровень заголовка: h%d → h%d",
		PageHeadingSkipRec:   "Не пропускайте уровни: после h1 идёт h2, затем h3. Иначе структура страницы читается неверно.",
		PageH1Title:          "Нет заголовка h1",
		PageH1Rec:            "Добавьте один h1 — главный заголовок страницы.",

		FormLabelTitle:    "Поле формы без подписи",
		FormLabelRec:      "Свяжите поле с <label for=\"...\"> или добавьте aria-label. Placeholder не заменяет подпись.",
		TableHeadersTitle: "Таблица без заголовков (№%d)",
		TableHeadersRec:   "Добавьте <th scope=\"col\"> или role=\"columnheader\": без заголовков непонятно, что означает каждая ячейка.",

		PDFNoTextTitle:       "PDF без текстового слоя",
		PDFNoTextRec:         "Опубликуйте PDF с текстовым слоем или HTML-версию: сканы без текста плохо читаются скринридером.",
		PDFNotTaggedTitle:    "PDF без тегов структуры",
		PDFNotTaggedRec:      "Пересохраните PDF с включённой разметкой структуры (Tagged PDF), чтобы скринридер понимал порядок чтения.",
		PDFAltMissingTitle:   "В PDF не найден /Alt",
		PDFAltMissingRec:     "Добавьте /Alt к смысловым изображениям и схемам в PDF.",
		PDFAltOKTitle:        "В PDF есть теги и /Alt",
		PDFAltOKRec:          "Ничего не нужно: структура и альтернативный текст на месте.",
		PDFTitleMissingTitle: "PDF без заголовка",
		PDFTitleMissingRec:   "Заполните Title в свойствах документа: он показывается во вкладке и в списке файлов.",
		PDFLangMissingTitle:  "PDF без языка документа",
		PDFLangMissingRec:    "Укажите язык документа в свойствах PDF, чтобы скринридер выбрал правильное произношение.",

		AIMismatchTitle:       "alt не соответствует изображению",
		AIMismatchRec:         "Модель считает, что alt не соответствует изображению.",
		AIMismatchUnsureTitle: "ИИ не уверен, что alt не соответствует",
		AIMismatchUnsureRec:   "Модель сомневается. Проверьте изображение и alt вручную.",
		AIConfidence:          " Уверенность ИИ: %d%%.",
		AISuggestedAlt:        " Предлагаемый alt: «%s».",
		AIMatchTitle:          "alt соответствует изображению",
		AIMatchRec:            "Модель подтвердила, что alt соответствует изображению.",
		AILinkTitle:           "Текст ссылки не соответствует цели",
		AILinkRec:             "Предлагаемый текст: «%s».",
		AILinkRecPlain:        "Модель считает, что текст ссылки не соответствует цели. Проверьте формулировку вручную.",

		ScanNoCrawler:   "краулер не настроен",
		ScanCrawlFailed: "краулер завершился ошибкой",
		ScanNoPages:     "не удалось обойти ни одной страницы: сайт недоступен, закрыт robots.txt или отдаёт контент только через JavaScript",
		ScanTimeout:     "проверка не завершилась за 90 секунд: часть проверок не выполнена, оценка не сохранена",

		ErrURLScheme:       "адрес должен быть http или https",
		ErrURLPublic:       "адрес должен быть публичным",
		ErrDepthRange:      "глубина должна быть от 1 до 3",
		ErrBodyTooLarge:    "тело запроса слишком большое",
		ErrInvalidJSON:     "не удалось разобрать запрос",
		ErrNotFound:        "не найдено",
		ErrStore:           "ошибка хранилища",
		ErrReport:          "не удалось построить отчёт",
		ErrClientID:        "нужен идентификатор браузера: обновите страницу",
		ErrGateMinScore:    "min_score должен быть числом от 0 до 100",
		ErrGateMaxCritical: "max_critical должен быть числом не меньше 0",
		ErrGateNotReady:    "проверка ещё не завершена",
		ErrGateFailed:      "проверка завершилась ошибкой",
		ErrGateScore:       "оценка %d ниже порога %d",
		ErrGateCritical:    "критичных находок %d больше допустимого %d",

		DocTitle:          "Отчёт о доступности",
		DocAuthor:         "Доступный контент",
		PDFPage:           "Страница %d",
		PDFSite:           "Сайт",
		PDFScan:           "Проверка",
		PDFScanValue:      "v%d от %s",
		PDFDepth:          "Глубина",
		PDFScore:          "Оценка",
		PDFFindings:       "Находки",
		PDFStatsValue:     "критично: %d, предупреждений: %d, в порядке: %d",
		PDFCoverage:       "Покрытие",
		PDFCoverageValue:  "страниц: %d, ссылок проверено: %d, не удалось: %d",
		PDFChanges:        "Что изменилось с прошлой проверки",
		PDFAdded:          "Новых",
		PDFFixed:          "Исправлено",
		PDFRemaining:      "Осталось",
		PDFDelta:          ", изменение оценки: %+d",
		PDFNoFindings:     "Находок нет.",
		PDFProblemsOnly:   "В список включены только критические находки и предупреждения.",
		PDFType:           "Тип: ",
		PDFElement:        "Элемент: ",
		PDFRecommendation: "Рекомендация: ",
		PDFScoreOf:        "%d из 100",
		PDFFootnote:       "Проверка не заменяет ручной аудит. Контраст, клавиатура и ARIA проверяются отдельно.",
		PDFDateFormat:     "02.01.2006 15:04",

		SeverityCriticalLabel: "Критично",
		SeverityWarningLabel:  "Предупреждение",
		SeverityOKLabel:       "В порядке",
		KindImage:             "Изображение",
		KindPDF:               "PDF",
		KindLink:              "Ссылка",
		KindPage:              "Страница",
		KindForm:              "Форма",
		KindTable:             "Таблица",
	},
	EN: {
		AltMissingTitle:   "Image has no alt text",
		AltMissingRec:     "Describe what the image shows. If it is decorative, use alt=\"\" and role=\"presentation\".",
		AltEmptyLinkTitle: "Linked image has no text alternative",
		AltEmptyLinkRec:   "Add alt text that explains where the link leads, or a visible caption next to the image.",
		AltGenericTitle:   "alt does not describe the image",
		AltGenericRec:     "Describe what is shown. “Image”, “photo” or a file name are not enough.",
		AltTooLongTitle:   "alt text is too long",
		AltTooLongRec:     "Keep the description under 125 characters: short text is easier to listen to.",

		LinkEmptyTitle:   "Link has no text",
		LinkEmptyRec:     "Add text that explains where the link leads.",
		LinkGenericTitle: "Link text is unclear",
		LinkGenericRec:   "Name what opens: for example “Contest rules” instead of “%s”.",
		LinkBrokenTitle:  "Broken link",
		LinkBrokenRec:    "Fix the link target or remove the link: it currently leads nowhere.",

		PageLangTitle:        "Page language is not set",
		PageLangRec:          "Add the lang attribute, for example <html lang=\"en\">: screen readers use it for pronunciation.",
		PageTitleTitle:       "Page has no title",
		PageTitleRec:         "Add a <title>: tabs and screen reader page lists rely on it.",
		PageHeadingSkipTitle: "Heading level skipped: h%d → h%d",
		PageHeadingSkipRec:   "Do not skip heading levels: h2 follows h1, then h3. Otherwise the page structure is read incorrectly.",
		PageH1Title:          "Page has no h1 heading",
		PageH1Rec:            "Add one h1: the main heading of the page.",

		FormLabelTitle:    "Form field has no label",
		FormLabelRec:      "Associate the field with <label for=\"...\"> or add aria-label. A placeholder is not a label.",
		TableHeadersTitle: "Table has no headers (#%d)",
		TableHeadersRec:   "Add <th scope=\"col\"> or role=\"columnheader\": without headers the cells have no meaning.",

		PDFNoTextTitle:       "PDF has no text layer",
		PDFNoTextRec:         "Publish a PDF with a text layer or an HTML version: scanned pages are unreadable for screen readers.",
		PDFNotTaggedTitle:    "PDF is not tagged",
		PDFNotTaggedRec:      "Re-save the PDF with structure tags enabled (Tagged PDF) so screen readers follow the reading order.",
		PDFAltMissingTitle:   "No /Alt found in the PDF",
		PDFAltMissingRec:     "Add /Alt to meaningful images and diagrams in the PDF.",
		PDFAltOKTitle:        "PDF has tags and /Alt",
		PDFAltOKRec:          "Nothing to do: structure and alternative text are in place.",
		PDFTitleMissingTitle: "PDF has no title",
		PDFTitleMissingRec:   "Fill in the Title property: it is shown in tabs and file lists.",
		PDFLangMissingTitle:  "PDF has no document language",
		PDFLangMissingRec:    "Set the document language in the PDF properties so screen readers choose the right pronunciation.",

		AIMismatchTitle:       "alt does not match the image",
		AIMismatchRec:         "The model reports that alt does not match the image.",
		AIMismatchUnsureTitle: "AI is not sure the alt mismatches",
		AIMismatchUnsureRec:   "The model is uncertain. Check the image and alt manually.",
		AIConfidence:          " AI confidence: %d%%.",
		AISuggestedAlt:        " Suggested alt: “%s”.",
		AIMatchTitle:          "alt matches the image",
		AIMatchRec:            "The model confirmed that alt matches the image.",
		AILinkTitle:           "Link text does not match its destination",
		AILinkRec:             "Suggested text: “%s”.",
		AILinkRecPlain:        "The model reports that the link text does not match its destination. Review the wording manually.",

		ScanNoCrawler:   "crawler is not configured",
		ScanCrawlFailed: "crawler failed",
		ScanNoPages:     "no pages could be crawled: the site is unreachable, blocked by robots.txt or renders content only with JavaScript",
		ScanTimeout:     "the scan did not finish within 90 seconds: some checks were skipped and the score was not saved",

		ErrURLScheme:       "url must be http or https",
		ErrURLPublic:       "url must be public",
		ErrDepthRange:      "depth must be between 1 and 3",
		ErrBodyTooLarge:    "request body too large",
		ErrInvalidJSON:     "request could not be parsed",
		ErrNotFound:        "not found",
		ErrStore:           "store error",
		ErrReport:          "report generation failed",
		ErrClientID:        "browser identity is required: reload the page",
		ErrGateMinScore:    "min_score must be a number between 0 and 100",
		ErrGateMaxCritical: "max_critical must be a number greater than or equal to 0",
		ErrGateNotReady:    "the scan is not finished yet",
		ErrGateFailed:      "the scan failed",
		ErrGateScore:       "score %d is below the threshold %d",
		ErrGateCritical:    "critical findings %d exceed the allowed %d",

		DocTitle:          "Accessibility report",
		DocAuthor:         "Accessible Content",
		PDFPage:           "Page %d",
		PDFSite:           "Site",
		PDFScan:           "Scan",
		PDFScanValue:      "v%d, %s",
		PDFDepth:          "Depth",
		PDFScore:          "Score",
		PDFFindings:       "Findings",
		PDFStatsValue:     "critical: %d, warnings: %d, ok: %d",
		PDFCoverage:       "Coverage",
		PDFCoverageValue:  "pages: %d, links checked: %d, unverified: %d",
		PDFChanges:        "What changed since the previous scan",
		PDFAdded:          "Added",
		PDFFixed:          "Fixed",
		PDFRemaining:      "Remaining",
		PDFDelta:          ", score change: %+d",
		PDFNoFindings:     "No findings.",
		PDFProblemsOnly:   "Only critical findings and warnings are listed.",
		PDFType:           "Type: ",
		PDFElement:        "Element: ",
		PDFRecommendation: "Recommendation: ",
		PDFScoreOf:        "%d of 100",
		PDFFootnote:       "This check does not replace a manual audit. Contrast, keyboard access and ARIA are reviewed separately.",
		PDFDateFormat:     "02 Jan 2006 15:04",

		SeverityCriticalLabel: "Critical",
		SeverityWarningLabel:  "Warning",
		SeverityOKLabel:       "OK",
		KindImage:             "Image",
		KindPDF:               "PDF",
		KindLink:              "Link",
		KindPage:              "Page",
		KindForm:              "Form",
		KindTable:             "Table",
	},
}

// T возвращает строку на указанном языке. Неизвестный язык — русский,
// неизвестный ключ — сам ключ (чтобы пропуск было видно в отчёте).
func T(lang Lang, key Key) string {
	if dict, ok := catalog[lang]; ok {
		if value, ok := dict[key]; ok {
			return value
		}
	}
	if value, ok := catalog[RU][key]; ok {
		return value
	}
	return string(key)
}

// Tf подставляет аргументы в шаблон строки.
func Tf(lang Lang, key Key, args ...any) string {
	return fmt.Sprintf(T(lang, key), args...)
}
