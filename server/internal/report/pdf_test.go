package report

import (
	"bytes"
	"testing"
	"time"

	"accessible-content/server/internal/store"
	"accessible-content/server/internal/texts"
)

func TestWritePDF(t *testing.T) {
	score := 42
	delta := 12
	scan := store.Scan{
		ID:        "scan-1",
		URL:       "https://example.ru",
		Depth:     2,
		Status:    store.StatusDone,
		Version:   2,
		Score:     &score,
		Pages:     3,
		Stats:     store.Stats{Critical: 1, Warning: 1, OK: 1},
		CreatedAt: time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC),
		Changes: &store.Changes{
			Added:      []store.Finding{{Title: "Новая проблема"}},
			Fixed:      []store.Finding{{Title: "Исправленная проблема"}},
			Remaining:  1,
			ScoreDelta: &delta,
		},
		Findings: []store.Finding{
			{
				Code:           "IMG_NO_ALT",
				Kind:           "image",
				Severity:       store.SeverityCritical,
				Title:          "Нет alt у изображения",
				Element:        `<img src="photo.svg">`,
				Recommendation: "Добавьте описание изображения.",
				WCAG:           "1.1.1",
			},
			{
				Code:           "LINK_BROKEN",
				Kind:           "link",
				Severity:       store.SeverityWarning,
				Title:          "Битая ссылка",
				Element:        `<a href="/old">old</a>`,
				Recommendation: "Исправьте адрес.",
				WCAG:           "2.4.4",
			},
		},
	}

	var buffer bytes.Buffer
	if err := WritePDF(&buffer, scan, texts.RU); err != nil {
		t.Fatalf("WritePDF: %v", err)
	}

	if !bytes.HasPrefix(buffer.Bytes(), []byte("%PDF")) {
		t.Fatal("expected PDF header")
	}
	if buffer.Len() < 1000 {
		t.Fatalf("PDF is too small: %d bytes", buffer.Len())
	}
}
