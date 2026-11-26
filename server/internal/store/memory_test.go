package store

import (
	"context"
	"testing"
	"time"
)

func TestMemoryListAndLatestDone(t *testing.T) {
	ctx := context.Background()
	memory := NewMemory()
	now := time.Now().UTC()

	older := Scan{
		ID:        "older",
		Owner:     "browser-a",
		URL:       "https://example.ru",
		Status:    StatusDone,
		Version:   1,
		CreatedAt: now.Add(-time.Hour),
	}
	newer := Scan{
		ID:        "newer",
		Owner:     "browser-a",
		URL:       "https://example.ru",
		Status:    StatusDone,
		Version:   2,
		CreatedAt: now,
	}
	other := Scan{
		ID:        "other",
		Owner:     "browser-b",
		URL:       "https://example.ru",
		Status:    StatusDone,
		Version:   1,
		CreatedAt: now,
	}

	for _, scan := range []Scan{older, newer, other} {
		if err := memory.Create(ctx, scan); err != nil {
			t.Fatal(err)
		}
	}

	list, err := memory.List(ctx, "https://example.ru", "browser-a", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].ID != "newer" {
		t.Fatalf("unexpected list: %+v", list)
	}

	foreign, err := memory.List(ctx, "https://example.ru", "browser-b", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(foreign) != 1 || foreign[0].ID != "other" {
		t.Fatalf("history of another browser leaked: %+v", foreign)
	}

	latest, err := memory.LatestDone(ctx, "https://example.ru", "browser-a")
	if err != nil {
		t.Fatal(err)
	}
	if latest.ID != "newer" {
		t.Fatalf("expected newer, got %s", latest.ID)
	}

	if _, err := memory.LatestDone(ctx, "https://example.ru", "browser-c"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound for an unknown browser, got %v", err)
	}
}
