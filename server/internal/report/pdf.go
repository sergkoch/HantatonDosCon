package report

import (
	"fmt"
	"io"
	"strings"

	"accessible-content/server/internal/store"

	"github.com/go-pdf/fpdf"
	"golang.org/x/image/font/gofont/goregular"
)

// PDFWriter implements Writer for PDF reports.
type PDFWriter struct{}

func (PDFWriter) Write(w io.Writer, scan store.Scan, format Format) error {
	if format != FormatPDF {
		return fmt.Errorf("report: unsupported format %q", format)
	}
	return WritePDF(w, scan)
}

// WritePDF renders a human-readable accessibility report.
func WritePDF(w io.Writer, scan store.Scan) error {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.SetTitle("Отчёт о доступности", true)
	pdf.SetAuthor("Доступный контент", true)
	pdf.SetCreator("Доступный контент", true)

	pdf.AddUTF8FontFromBytes("go", "", goregular.TTF)
	pdf.AddUTF8FontFromBytes("go", "B", goregular.TTF)

	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont("go", "", 9)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(0, 5, fmt.Sprintf("Страница %d", pdf.PageNo()), "", 0, "C", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
	})

	pdf.AddPage()
	pdf.SetFont("go", "B", 18)
	pdf.MultiCell(0, 9, "Отчёт о доступности", "", "L", false)
	pdf.Ln(2)

	pdf.SetFont("go", "", 11)
	writeLine(pdf, "Сайт", scan.URL)
	writeLine(pdf, "Проверка", fmt.Sprintf("v%d от %s", scan.Version, scan.CreatedAt.Format("02.01.2006 15:04")))
	writeLine(pdf, "Глубина", fmt.Sprintf("%d", scan.Depth))
	writeLine(pdf, "Оценка", scoreText(scan.Score))
	writeLine(pdf, "Находки", fmt.Sprintf("критично: %d, предупреждений: %d, в порядке: %d",
		scan.Stats.Critical, scan.Stats.Warning, scan.Stats.OK))
	if scan.Pages > 0 || scan.CheckedLinks > 0 {
		writeLine(pdf, "Покрытие", fmt.Sprintf("страниц: %d, ссылок проверено: %d, не удалось: %d",
			scan.Pages, scan.CheckedLinks, scan.UnverifiedLinks))
	}

	if scan.Changes != nil {
		pdf.Ln(3)
		pdf.SetFont("go", "B", 13)
		pdf.CellFormat(0, 7, "Что изменилось с прошлой проверки", "", 1, "L", false, 0, "")
		pdf.SetFont("go", "", 11)
		delta := ""
		if scan.Changes.ScoreDelta != nil {
			delta = fmt.Sprintf(", изменение оценки: %+d", *scan.Changes.ScoreDelta)
		}
		writeLine(pdf, "Новых", fmt.Sprintf("%d", len(scan.Changes.Added)))
		writeLine(pdf, "Исправлено", fmt.Sprintf("%d", len(scan.Changes.Fixed)))
		writeLine(pdf, "Осталось", fmt.Sprintf("%d%s", scan.Changes.Remaining, delta))
	}

	pdf.Ln(3)
	pdf.SetFont("go", "B", 13)
	pdf.CellFormat(0, 7, "Находки", "", 1, "L", false, 0, "")

	if len(scan.Findings) == 0 {
		pdf.SetFont("go", "", 11)
		pdf.MultiCell(0, 6, "Находок нет.", "", "L", false)
	} else {
		for i, finding := range scan.Findings {
			if pdf.GetY() > 255 {
				pdf.AddPage()
			}

			pdf.SetFont("go", "B", 11)
			pdf.MultiCell(0, 6, fmt.Sprintf("%d. %s · %s", i+1, severityText(finding.Severity), finding.Title), "", "L", false)
			pdf.SetFont("go", "", 10)

			if finding.Kind != "" {
				pdf.MultiCell(0, 5, "Тип: "+kindText(finding.Kind), "", "L", false)
			}
			if finding.Element != "" {
				pdf.MultiCell(0, 5, "Элемент: "+finding.Element, "", "L", false)
			}
			if finding.Recommendation != "" {
				pdf.MultiCell(0, 5, "Рекомендация: "+finding.Recommendation, "", "L", false)
			}
			if finding.WCAG != "" {
				pdf.MultiCell(0, 5, "WCAG: "+finding.WCAG, "", "L", false)
			}
			pdf.Ln(2)
		}
	}

	pdf.Ln(3)
	pdf.SetFont("go", "", 9)
	pdf.SetTextColor(120, 120, 120)
	pdf.MultiCell(0, 4.5, "Проверка не заменяет ручной аудит. Контраст, клавиатура и ARIA проверяются отдельно.", "", "L", false)

	return pdf.Output(w)
}

func writeLine(pdf *fpdf.Fpdf, label, value string) {
	pdf.SetFont("go", "B", 11)
	pdf.CellFormat(42, 6, label+":", "", 0, "L", false, 0, "")
	pdf.SetFont("go", "", 11)
	pdf.MultiCell(0, 6, value, "", "L", false)
}

func scoreText(score *int) string {
	if score == nil {
		return "—"
	}
	return fmt.Sprintf("%d из 100", *score)
}

func severityText(severity store.Severity) string {
	switch severity {
	case store.SeverityCritical:
		return "Критично"
	case store.SeverityWarning:
		return "Предупреждение"
	case store.SeverityOK:
		return "В порядке"
	default:
		return strings.TrimSpace(string(severity))
	}
}

func kindText(kind string) string {
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
