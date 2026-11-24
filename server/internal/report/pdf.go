package report

import (
	"fmt"

	"accessible-content/server/internal/store"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/fontrepository"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

const fontFamily = "go"

var (
	inkColor      = props.Color{Red: 15, Green: 23, Blue: 42}
	mutedColor    = props.Color{Red: 71, Green: 85, Blue: 105}
	lineColor     = props.Color{Red: 203, Green: 213, Blue: 225}
	headColor     = props.Color{Red: 241, Green: 245, Blue: 249}
	criticalColor = props.Color{Red: 185, Green: 28, Blue: 28}
	warningColor  = props.Color{Red: 180, Green: 83, Blue: 9}
	okColor       = props.Color{Red: 21, Green: 128, Blue: 61}
)

// BuildPDF собирает PDF-отчёт по готовому скану.
func BuildPDF(scan store.Scan) ([]byte, error) {
	fonts, err := fontrepository.New().
		AddUTF8FontFromBytes(fontFamily, fontstyle.Normal, goregular.TTF).
		AddUTF8FontFromBytes(fontFamily, fontstyle.Bold, gobold.TTF).
		Load()
	if err != nil {
		return nil, fmt.Errorf("report: загрузка шрифтов: %w", err)
	}

	cfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithLeftMargin(12).
		WithTopMargin(12).
		WithRightMargin(12).
		WithBottomMargin(14).
		WithDefaultFont(&props.Font{Family: fontFamily, Style: fontstyle.Normal, Size: 9, Color: &inkColor}).
		WithCustomFonts(fonts).
		WithPageNumber(props.PageNumber{
			Pattern: "Страница {current} из {total}",
			Place:   props.Bottom,
			Family:  fontFamily,
			Style:   fontstyle.Normal,
			Size:    8,
			Color:   &mutedColor,
		}).
		Build()

	m := maroto.New(cfg)

	m.AddAutoRow(col.New(12).Add(text.New("Отчёт о доступности сайта", textProps(18, fontstyle.Bold, &inkColor))))
	m.AddAutoRow(col.New(12).Add(text.New(scan.URL, textProps(11, fontstyle.Normal, &inkColor))))

	score := "—"
	if scan.Score != nil {
		score = fmt.Sprintf("%d", *scan.Score)
	}

	m.AddAutoRow(
		col.New(4).Add(text.New(fmt.Sprintf("Оценка: %s из 100", score), textProps(10, fontstyle.Bold, &inkColor))),
		col.New(4).Add(text.New(fmt.Sprintf("Находок: %d", len(scan.Findings)), textProps(9, fontstyle.Normal, &inkColor))),
		col.New(4).Add(text.New(fmt.Sprintf("Критично: %d · Предупреждений: %d", scan.Stats.Critical, scan.Stats.Warning), textProps(9, fontstyle.Normal, &inkColor))),
	)

	m.AddAutoRow(col.New(12).Add(text.New(
		fmt.Sprintf("Страниц проверено: %d · Ссылок проверено: %d · Не удалось проверить: %d",
			scan.Pages, scan.CheckedLinks, scan.UnverifiedLinks),
		textProps(8, fontstyle.Normal, &mutedColor),
	)))

	m.AddAutoRow(col.New(12).Add(text.New(
		"Сформировано: "+scan.CreatedAt.Format("02.01.2006 15:04"),
		textProps(8, fontstyle.Normal, &mutedColor),
	)))

	m.AddRows(row.New(3))

	m.AddRows(
		row.New().
			WithStyle(&props.Cell{
				BackgroundColor: &headColor,
				BorderType:      border.Bottom,
				BorderColor:     &lineColor,
				BorderThickness: 0.2,
			}).
			Add(
				col.New(2).Add(text.New("Критичность", textProps(8, fontstyle.Bold, &mutedColor))),
				col.New(2).Add(text.New("Тип", textProps(8, fontstyle.Bold, &mutedColor))),
				col.New(4).Add(text.New("Что не так", textProps(8, fontstyle.Bold, &mutedColor))),
				col.New(4).Add(text.New("Рекомендация", textProps(8, fontstyle.Bold, &mutedColor))),
			),
	)

	if len(scan.Findings) == 0 {
		m.AddAutoRow(col.New(12).Add(text.New(
			"Находок нет: по проверенным критериям проблем не найдено.",
			textProps(9, fontstyle.Normal, &inkColor),
		)))
	}

	for _, finding := range scan.Findings {
		severity := finding.Severity
		severityColor := severityColor(severity)

		detailText := finding.Title
		if finding.Element != "" {
			detailText += "\n" + finding.Element
		}
		if finding.WCAG != "" {
			detailText += "\nWCAG " + finding.WCAG
		}

		m.AddRows(
			row.New().
				WithStyle(&props.Cell{
					BorderType:      border.Bottom,
					BorderColor:     &lineColor,
					BorderThickness: 0.15,
				}).
				Add(
					col.New(2).Add(text.New(severityLabel(severity), textProps(8, fontstyle.Bold, &severityColor))),
					col.New(2).Add(text.New(kindLabel(finding.Kind), textProps(8, fontstyle.Normal, &inkColor))),
					col.New(4).Add(text.New(detailText, textProps(8, fontstyle.Normal, &inkColor))),
					col.New(4).Add(text.New(finding.Recommendation, textProps(8, fontstyle.Normal, &inkColor))),
				),
		)
	}

	m.AddRows(row.New(4))
	m.AddAutoRow(col.New(12).Add(text.New(
		"Проверка не заменяет ручной аудит: контраст, клавиатура и ARIA не проверяются.",
		textProps(8, fontstyle.Normal, &mutedColor),
	)))

	document, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("report: генерация PDF: %w", err)
	}

	return document.GetBytes(), nil
}

func textProps(size float64, style fontstyle.Type, color *props.Color) props.Text {
	return props.Text{
		Family:          fontFamily,
		Style:           style,
		Size:            size,
		Color:           color,
		Top:             1.2,
		Bottom:          1.2,
		Left:            1.5,
		Right:           1.5,
		VerticalPadding: 0.5,
	}
}

func severityLabel(severity store.Severity) string {
	switch severity {
	case store.SeverityCritical:
		return "Критично"
	case store.SeverityWarning:
		return "Предупреждение"
	default:
		return "В порядке"
	}
}

func severityColor(severity store.Severity) props.Color {
	switch severity {
	case store.SeverityCritical:
		return criticalColor
	case store.SeverityWarning:
		return warningColor
	default:
		return okColor
	}
}

func kindLabel(kind string) string {
	switch kind {
	case "image":
		return "Изображение"
	case "pdf":
		return "PDF"
	case "link":
		return "Ссылка"
	case "page":
		return "Страница"
	case "form":
		return "Форма"
	case "table":
		return "Таблица"
	default:
		return kind
	}
}
