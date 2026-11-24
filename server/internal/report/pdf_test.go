package report

import (
	"bytes"
	"testing"
	"time"

	"accessible-content/server/internal/store"
)

func TestBuildPDF(t *testing.T) {
	score := 55
	scan := store.Scan{
		ID:              "scan-1",
		URL:             "https://example.com",
		Status:          store.StatusDone,
		Score:           &score,
		Pages:           2,
		CheckedLinks:    7,
		UnverifiedLinks: 1,
		Stats:           store.Stats{Critical: 1, Warning: 1},
		CreatedAt:       time.Date(2026, 11, 24, 12, 30, 0, 0, time.UTC),
		Findings: []store.Finding{
			{
				ID:             "f1",
				Kind:           "image",
				Severity:       store.SeverityCritical,
				Title:          "Нет alt у изображения",
				Element:        `<img src="photo.svg">`,
				Recommendation: "Добавьте описание, что на картинке.",
				WCAG:           "1.1.1",
			},
			{
				ID:             "f2",
				Kind:           "form",
				Severity:       store.SeverityWarning,
				Title:          "Поле формы без подписи",
				Element:        `<input type="text" name="phone">`,
				Recommendation: "Свяжите поле с label.",
				WCAG:           "3.3.2",
			},
		},
	}

	data, err := BuildPDF(scan)
	if err != nil {
		t.Fatalf("BuildPDF: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		t.Fatalf("ответ не похож на PDF: %q", data[:min(8, len(data))])
	}
	if len(data) < 1500 {
		t.Fatalf("PDF слишком маленький: %d байт", len(data))
	}
}

func TestBuildPDFWithoutFindings(t *testing.T) {
	score := 100
	scan := store.Scan{
		ID:        "scan-2",
		URL:       "https://example.com",
		Status:    store.StatusDone,
		Score:     &score,
		Pages:     1,
		CreatedAt: time.Now().UTC(),
	}

	data, err := BuildPDF(scan)
	if err != nil {
		t.Fatalf("BuildPDF: %v", err)
	}
	if !bytes.HasPrefix(data, []byte("%PDF")) {
		t.Fatal("ответ не похож на PDF")
	}
}
