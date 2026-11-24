package pdf

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/references"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/webassembly"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

const (
	// MaxPages limits how many pages are inspected during a single scan.
	MaxPages = 50
	// analyzeTimeout bounds one PDF analysis.
	analyzeTimeout = 30 * time.Second
)

// Analyzer inspects PDF accessibility: text layer, structure tags, /Alt,
// title and language. PDFium runs as WebAssembly, so no CGO is required.
type PdfiumAnalyzer struct {
	pool     pdfium.Pool
	instance pdfium.Pdfium
	mu       sync.Mutex
}

func NewAnalyzer() (*PdfiumAnalyzer, error) {
	pool, err := webassembly.Init(webassembly.Config{
		MinIdle:  1,
		MaxIdle:  1,
		MaxTotal: 1,
	})
	if err != nil {
		return nil, err
	}

	instance, err := pool.GetInstance(analyzeTimeout)
	if err != nil {
		_ = pool.Close()
		return nil, err
	}

	return &PdfiumAnalyzer{pool: pool, instance: instance}, nil
}

func (a *PdfiumAnalyzer) Close() {
	if a.instance != nil {
		_ = a.instance.Close()
	}
	if a.pool != nil {
		_ = a.pool.Close()
	}
}

// Analyze implements Analyzer. It is safe for concurrent use, but serializes
// PDFium calls because the WebAssembly pool has one worker.
func (a *PdfiumAnalyzer) Analyze(ctx context.Context, file io.ReaderAt, size int64) (Result, error) {
	if file == nil || size <= 0 {
		return Result{}, errors.New("pdf: empty file")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	doc, err := a.instance.OpenDocument(&requests.OpenDocument{
		FileReader:     io.NewSectionReader(file, 0, size),
		FileReaderSize: size,
	})
	if err != nil {
		return Result{}, err
	}
	defer a.instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})

	pageCount, err := a.instance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc.Document})
	if err != nil {
		return Result{}, err
	}

	result := Result{PageCount: pageCount.PageCount}

	pages := result.PageCount
	if pages > MaxPages {
		pages = MaxPages
	}

	for pageIndex := 0; pageIndex < pages; pageIndex++ {
		if err := ctx.Err(); err != nil {
			break
		}

		page := requests.Page{
			ByIndex: &requests.PageByIndex{
				Document: doc.Document,
				Index:    pageIndex,
			},
		}

		textPage, err := a.instance.FPDFText_LoadPage(&requests.FPDFText_LoadPage{Page: page})
		if err == nil {
			count, countErr := a.instance.FPDFText_CountChars(&requests.FPDFText_CountChars{TextPage: textPage.TextPage})
			if countErr == nil && count.Count > 0 {
				result.TextLayer = true
			}
			a.instance.FPDFText_ClosePage(&requests.FPDFText_ClosePage{TextPage: textPage.TextPage})
		}

		altTexts, tagged := countAltTexts(a.instance, doc.Document, pageIndex)
		result.AltTexts += altTexts
		if tagged {
			result.Tagged = true
		}
	}

	title, err := a.instance.FPDF_GetMetaText(&requests.FPDF_GetMetaText{
		Document: doc.Document,
		Tag:      "Title",
	})
	if err == nil {
		result.Title = title.Value
	}

	// pdfcpu is used for the document language and as a fallback for the
	// tagged flag and title.
	metaCtx, err := api.ReadContext(ctx, io.NewSectionReader(file, 0, size), model.NewDefaultConfiguration())
	if err == nil {
		if result.Title == "" {
			result.Title = metaCtx.Title
		}
		result.Lang = rootLang(metaCtx)
		if metaCtx.Tagged {
			result.Tagged = true
		}
	}

	return result, nil
}

func rootLang(ctx *model.Context) string {
	obj, ok := ctx.RootDict["Lang"]
	if !ok {
		return ""
	}
	switch v := obj.(type) {
	case types.StringLiteral:
		return string(v)
	case types.Name:
		return string(v)
	default:
		return ""
	}
}

func countAltTexts(instance pdfium.Pdfium, doc references.FPDF_DOCUMENT, pageIndex int) (int, bool) {
	tree, err := instance.FPDF_StructTree_GetForPage(&requests.FPDF_StructTree_GetForPage{
		Page: requests.Page{
			ByIndex: &requests.PageByIndex{
				Document: doc,
				Index:    pageIndex,
			},
		},
	})
	if err != nil {
		return 0, false
	}
	defer instance.FPDF_StructTree_Close(&requests.FPDF_StructTree_Close{StructTree: tree.StructTree})

	count, err := instance.FPDF_StructTree_CountChildren(&requests.FPDF_StructTree_CountChildren{StructTree: tree.StructTree})
	if err != nil {
		return 0, true
	}

	total := 0
	for i := 0; i < count.Count; i++ {
		child, err := instance.FPDF_StructTree_GetChildAtIndex(&requests.FPDF_StructTree_GetChildAtIndex{
			StructTree: tree.StructTree,
			Index:      i,
		})
		if err == nil {
			total += walkStructElement(instance, child.StructElement, 0)
		}
	}

	return total, true
}

func walkStructElement(instance pdfium.Pdfium, element references.FPDF_STRUCTELEMENT, depth int) int {
	if depth > 50 {
		return 0
	}

	total := 0
	alt, err := instance.FPDF_StructElement_GetAltText(&requests.FPDF_StructElement_GetAltText{StructElement: element})
	if err == nil && alt.AltText != "" {
		total++
	}

	count, err := instance.FPDF_StructElement_CountChildren(&requests.FPDF_StructElement_CountChildren{StructElement: element})
	if err != nil {
		return total
	}

	for i := 0; i < count.Count; i++ {
		child, err := instance.FPDF_StructElement_GetChildAtIndex(&requests.FPDF_StructElement_GetChildAtIndex{
			StructElement: element,
			Index:         i,
		})
		if err == nil {
			total += walkStructElement(instance, child.StructElement, depth+1)
		}
	}

	return total
}
