package pdf

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

var testAnalyzer *PdfiumAnalyzer

func TestMain(m *testing.M) {
	analyzer, err := NewAnalyzer()
	if err != nil {
		panic("init pdf analyzer: " + err.Error())
	}
	testAnalyzer = analyzer
	code := m.Run()
	analyzer.Close()
	os.Exit(code)
}

func analyzeFixture(t *testing.T, name string) Result {
	t.Helper()

	path := filepath.Join("testdata", name)
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", name, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		t.Fatalf("stat %s: %v", name, err)
	}

	result, err := testAnalyzer.Analyze(context.Background(), file, info.Size())
	if err != nil {
		t.Fatalf("analyze %s: %v", name, err)
	}

	return result
}

func TestAnalyzeTextPDF(t *testing.T) {
	result := analyzeFixture(t, "text.pdf")

	if result.PageCount != 1 {
		t.Fatalf("expected 1 page, got %d", result.PageCount)
	}
	if !result.TextLayer {
		t.Fatal("expected text layer")
	}
	if result.Tagged {
		t.Fatal("expected untagged PDF")
	}
	if result.Title != "Sample text PDF" {
		t.Fatalf("unexpected title %q", result.Title)
	}
	if result.Lang != "ru-RU" {
		t.Fatalf("unexpected lang %q", result.Lang)
	}
}

func TestAnalyzeScannedPDF(t *testing.T) {
	result := analyzeFixture(t, "scan.pdf")

	if result.TextLayer {
		t.Fatal("expected no text layer in scanned PDF")
	}
	if result.PageCount != 1 {
		t.Fatalf("expected 1 page, got %d", result.PageCount)
	}
}

func TestAnalyzeTaggedPDF(t *testing.T) {
	result := analyzeFixture(t, "tagged.pdf")

	if !result.Tagged {
		t.Fatal("expected tagged PDF")
	}
	if result.AltTexts != 1 {
		t.Fatalf("expected 1 alt text, got %d", result.AltTexts)
	}
	if result.Title != "Tagged sample PDF" {
		t.Fatalf("unexpected title %q", result.Title)
	}
	if result.Lang != "ru-RU" {
		t.Fatalf("unexpected lang %q", result.Lang)
	}
}

func TestAnalyzeBrokenPDF(t *testing.T) {
	data := []byte("%PDF-1.4\nthis is not a real pdf")

	if _, err := testAnalyzer.Analyze(context.Background(), bytes.NewReader(data), int64(len(data))); err == nil {
		t.Fatal("expected error for broken PDF")
	}
}
