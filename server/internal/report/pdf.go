package report

import (
	"fmt"
	"io"
	"strings"

	"accessible-content/server/internal/store"
	"accessible-content/server/internal/texts"

	"github.com/go-pdf/fpdf"
	"golang.org/x/image/font/gofont/goregular"
)

// PDFWriter implements Writer for PDF reports.
type PDFWriter struct{}

func (PDFWriter) Write(w io.Writer, scan store.Scan, format Format) error {
	if format != FormatPDF {
		return fmt.Errorf("report: unsupported format %q", format)
	}
	return WritePDF(w, scan, texts.Parse(scan.Lang))
}

// WritePDF renders a human-readable accessibility report in the requested
// language. Findings keep the scan language plus a translation, so any of the
// supported languages can be exported.
func WritePDF(w io.Writer, scan store.Scan, lang texts.Lang) error {
	lang = texts.Parse(string(lang))
	scanLang := texts.Parse(scan.Lang)

	pick := func(primary, alternative string) string {
		if lang != scanLang && alternative != "" {
			return alternative
		}
		return primary
	}

	// Старые сканы могли сохранить находки вперемешку: сортируем копию.
	findings := make([]store.Finding, 0, len(scan.Findings))
	for _, finding := range scan.Findings {
		if finding.Severity != store.SeverityOK {
			findings = append(findings, finding)
		}
	}
	store.SortBySeverity(findings)

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.SetTitle(texts.T(lang, texts.DocTitle), true)
	pdf.SetAuthor(texts.T(lang, texts.DocAuthor), true)
	pdf.SetCreator(texts.T(lang, texts.DocAuthor), true)

	pdf.AddUTF8FontFromBytes("go", "", goregular.TTF)
	pdf.AddUTF8FontFromBytes("go", "B", goregular.TTF)

	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont("go", "", 9)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(0, 5, texts.Tf(lang, texts.PDFPage, pdf.PageNo()), "", 0, "C", false, 0, "")
		pdf.SetTextColor(0, 0, 0)
	})

	pdf.AddPage()
	pdf.SetFont("go", "B", 18)
	pdf.MultiCell(0, 9, texts.T(lang, texts.DocTitle), "", "L", false)
	pdf.Ln(2)

	pdf.SetFont("go", "", 11)
	writeLine(pdf, texts.T(lang, texts.PDFSite), scan.URL)
	writeLine(pdf, texts.T(lang, texts.PDFScan), texts.Tf(lang, texts.PDFScanValue,
		scan.Version, scan.CreatedAt.Format(texts.T(lang, texts.PDFDateFormat))))
	writeLine(pdf, texts.T(lang, texts.PDFDepth), fmt.Sprintf("%d", scan.Depth))
	writeLine(pdf, texts.T(lang, texts.PDFScore), scoreText(scan.Score, lang))
	writeLine(pdf, texts.T(lang, texts.PDFFindings), texts.Tf(lang, texts.PDFStatsValue,
		scan.Stats.Critical, scan.Stats.Warning, scan.Stats.OK))
	if scan.Pages > 0 || scan.CheckedLinks > 0 {
		writeLine(pdf, texts.T(lang, texts.PDFCoverage), texts.Tf(lang, texts.PDFCoverageValue,
			scan.Pages, scan.CheckedLinks, scan.UnverifiedLinks))
	}

	if scan.Changes != nil {
		pdf.Ln(3)
		pdf.SetFont("go", "B", 13)
		pdf.CellFormat(0, 7, texts.T(lang, texts.PDFChanges), "", 1, "L", false, 0, "")
		pdf.SetFont("go", "", 11)
		delta := ""
		if scan.Changes.ScoreDelta != nil {
			delta = texts.Tf(lang, texts.PDFDelta, *scan.Changes.ScoreDelta)
		}
		writeLine(pdf, texts.T(lang, texts.PDFAdded), fmt.Sprintf("%d", len(scan.Changes.Added)))
		writeLine(pdf, texts.T(lang, texts.PDFFixed), fmt.Sprintf("%d", len(scan.Changes.Fixed)))
		writeLine(pdf, texts.T(lang, texts.PDFRemaining), fmt.Sprintf("%d%s", scan.Changes.Remaining, delta))
	}

	pdf.Ln(3)
	pdf.SetFont("go", "B", 13)
	pdf.CellFormat(0, 7, texts.T(lang, texts.PDFFindings), "", 1, "L", false, 0, "")
	pdf.SetFont("go", "", 9)
	pdf.SetTextColor(120, 120, 120)
	pdf.MultiCell(0, 4.5, texts.T(lang, texts.PDFProblemsOnly), "", "L", false)
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(1)

	if len(findings) == 0 {
		pdf.SetFont("go", "", 11)
		pdf.MultiCell(0, 6, texts.T(lang, texts.PDFNoFindings), "", "L", false)
	} else {
		for i, finding := range findings {
			if pdf.GetY() > 255 {
				pdf.AddPage()
			}

			pdf.SetFont("go", "B", 11)
			pdf.MultiCell(0, 6, fmt.Sprintf("%d. %s · %s", i+1, severityText(finding.Severity, lang), pick(finding.Title, finding.TitleAlt)), "", "L", false)
			pdf.SetFont("go", "", 10)

			if finding.Kind != "" {
				pdf.MultiCell(0, 5, texts.T(lang, texts.PDFType)+kindText(finding.Kind, lang), "", "L", false)
			}
			if finding.Element != "" {
				pdf.MultiCell(0, 5, texts.T(lang, texts.PDFElement)+finding.Element, "", "L", false)
			}
			if finding.Recommendation != "" {
				pdf.MultiCell(0, 5, texts.T(lang, texts.PDFRecommendation)+pick(finding.Recommendation, finding.RecommendationAlt), "", "L", false)
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
	pdf.MultiCell(0, 4.5, texts.T(lang, texts.PDFFootnote), "", "L", false)

	return pdf.Output(w)
}

func writeLine(pdf *fpdf.Fpdf, label, value string) {
	pdf.SetFont("go", "B", 11)
	pdf.CellFormat(42, 6, label+":", "", 0, "L", false, 0, "")
	pdf.SetFont("go", "", 11)
	pdf.MultiCell(0, 6, value, "", "L", false)
}

func scoreText(score *int, lang texts.Lang) string {
	if score == nil {
		return "—"
	}
	return texts.Tf(lang, texts.PDFScoreOf, *score)
}

func severityText(severity store.Severity, lang texts.Lang) string {
	switch severity {
	case store.SeverityCritical:
		return texts.T(lang, texts.SeverityCriticalLabel)
	case store.SeverityWarning:
		return texts.T(lang, texts.SeverityWarningLabel)
	case store.SeverityOK:
		return texts.T(lang, texts.SeverityOKLabel)
	default:
		return strings.TrimSpace(string(severity))
	}
}

func kindText(kind string, lang texts.Lang) string {
	switch kind {
	case "image":
		return texts.T(lang, texts.KindImage)
	case "pdf":
		return texts.T(lang, texts.KindPDF)
	case "link":
		return texts.T(lang, texts.KindLink)
	case "page":
		return texts.T(lang, texts.KindPage)
	case "form":
		return texts.T(lang, texts.KindForm)
	case "table":
		return texts.T(lang, texts.KindTable)
	default:
		return kind
	}
}
