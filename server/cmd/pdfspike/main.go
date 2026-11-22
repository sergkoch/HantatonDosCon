// Command pdfspike is a temporary T14 spike.
//
// It compares what pdfcpu and go-pdfium can do with a PDF:
// page count, text layer, tags, /Alt, title, language and page rendering.
//
// Run from the server directory:
//
//	go run ./cmd/pdfspike internal/pdf/testdata/text.pdf
//
// Remove this command after T15 implements pdf.Analyzer.
package main

import (
	"context"
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/references"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/webassembly"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func main() {
	outputDir := flag.String("out", "pdfspike-out", "directory for rendered PNG files")
	dpi := flag.Int("dpi", 150, "render DPI")
	flag.Parse()

	paths := flag.Args()
	if len(paths) == 0 {
		paths = []string{
			filepath.Join("internal", "pdf", "testdata", "text.pdf"),
			filepath.Join("internal", "pdf", "testdata", "scan.pdf"),
			filepath.Join("internal", "pdf", "testdata", "tagged.pdf"),
		}
	}

	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "create output dir:", err)
		os.Exit(1)
	}

	pool, err := webassembly.Init(webassembly.Config{
		MinIdle:  1,
		MaxIdle:  1,
		MaxTotal: 1,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "init pdfium:", err)
		os.Exit(1)
	}
	defer pool.Close()

	instance, err := pool.GetInstance(30 * time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, "get pdfium instance:", err)
		os.Exit(1)
	}
	defer instance.Close()

	for _, path := range paths {
		fmt.Println("========================================")
		analyze(instance, path, *outputDir, *dpi)
	}
}

func analyze(instance pdfium.Pdfium, path, outputDir string, dpi int) {
	fmt.Println("file:", filepath.Base(path))

	ctx, err := api.ReadContextFile(context.Background(), path)
	if err != nil {
		fmt.Println("  pdfcpu error:", err)
	} else {
		fmt.Printf("  pdfcpu: pages=%d tagged=%v title=%q lang=%v\n",
			ctx.PageCount, ctx.Tagged, ctx.Title, rootLang(ctx))
	}

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("  read error:", err)
		return
	}

	doc, err := instance.OpenDocument(&requests.OpenDocument{File: &data})
	if err != nil {
		fmt.Println("  pdfium open error:", err)
		return
	}
	defer instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})

	pageCount, err := instance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc.Document})
	if err != nil {
		fmt.Println("  pdfium page count error:", err)
		return
	}

	page := requests.Page{
		ByIndex: &requests.PageByIndex{
			Document: doc.Document,
			Index:    0,
		},
	}

	textChars := 0
	textPage, err := instance.FPDFText_LoadPage(&requests.FPDFText_LoadPage{Page: page})
	if err == nil {
		count, countErr := instance.FPDFText_CountChars(&requests.FPDFText_CountChars{TextPage: textPage.TextPage})
		if countErr == nil {
			textChars = count.Count
		}
		instance.FPDFText_ClosePage(&requests.FPDFText_ClosePage{TextPage: textPage.TextPage})
	}

	fmt.Printf("  pdfium: pages=%d textChars=%d altTexts=%d\n",
		pageCount.PageCount, textChars, countAlts(instance, doc.Document, 0))

	rendered, err := instance.RenderPageInDPI(&requests.RenderPageInDPI{
		DPI:  dpi,
		Page: page,
	})
	if err != nil {
		fmt.Println("  pdfium render error:", err)
		return
	}
	defer rendered.Cleanup()

	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)) + ".png"
	outPath := filepath.Join(outputDir, name)
	f, err := os.Create(outPath)
	if err != nil {
		fmt.Println("  save render error:", err)
		return
	}
	defer f.Close()

	if err := png.Encode(f, rendered.Result.Image); err != nil {
		fmt.Println("  encode render error:", err)
		return
	}

	bounds := rendered.Result.Image.Bounds()
	fmt.Printf("  pdfium: render=%dx%d saved=%s\n", bounds.Dx(), bounds.Dy(), outPath)
}

func rootLang(ctx *model.Context) any {
	obj, ok := ctx.RootDict["Lang"]
	if !ok {
		return nil
	}
	switch v := obj.(type) {
	case types.StringLiteral:
		return strings.Trim(string(v), "()")
	case types.Name:
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func countAlts(instance pdfium.Pdfium, doc references.FPDF_DOCUMENT, pageIndex int) int {
	tree, err := instance.FPDF_StructTree_GetForPage(&requests.FPDF_StructTree_GetForPage{
		Page: requests.Page{
			ByIndex: &requests.PageByIndex{
				Document: doc,
				Index:    pageIndex,
			},
		},
	})
	if err != nil {
		return 0
	}
	defer instance.FPDF_StructTree_Close(&requests.FPDF_StructTree_Close{StructTree: tree.StructTree})

	count, err := instance.FPDF_StructTree_CountChildren(&requests.FPDF_StructTree_CountChildren{StructTree: tree.StructTree})
	if err != nil {
		return 0
	}

	total := 0
	for i := 0; i < count.Count; i++ {
		child, err := instance.FPDF_StructTree_GetChildAtIndex(&requests.FPDF_StructTree_GetChildAtIndex{
			StructTree: tree.StructTree,
			Index:      i,
		})
		if err == nil {
			total += walkElement(instance, child.StructElement, 0)
		}
	}
	return total
}

func walkElement(instance pdfium.Pdfium, element references.FPDF_STRUCTELEMENT, depth int) int {
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
			total += walkElement(instance, child.StructElement, depth+1)
		}
	}
	return total
}
