package store

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFindingJSONExposesAlternateLanguage(t *testing.T) {
	data, err := json.Marshal(Finding{
		ID:                "f1",
		Title:             "Ссылка без текста",
		TitleAlt:          "Link has no text",
		Recommendation:    "Добавьте текст.",
		RecommendationAlt: "Add text.",
	})
	if err != nil {
		t.Fatal(err)
	}

	body := string(data)
	if !strings.Contains(body, `"titleAlt":"Link has no text"`) {
		t.Fatalf("titleAlt must be exposed, got %s", body)
	}
	if !strings.Contains(body, `"recommendationAlt":"Add text."`) {
		t.Fatalf("recommendationAlt must be exposed, got %s", body)
	}
}
