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
		URL:       "https://example.ru",
		Status:    StatusDone,
		Version:   1,
		CreatedAt: now.Add(-time.Hour),
	}
	newer := Scan{
		ID:        "newer",
		URL:       "https://example.ru",
		Status:    StatusDone,
		Version:   2,
		CreatedAt: now,
	}
	other := Scan{
		ID:        "other",
		URL:       "https://other.ru",
		Status:    StatusDone,
		Version:   1,
		CreatedAt: now,
	}

	for _, scan := range []Scan{older, newer, other} {
		if err := memory.Create(ctx, scan); err != nil {
			t.Fatal(err)
		}
	}

	list, err := memory.List(ctx, "https://example.ru", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].ID != "newer" {
		t.Fatalf("unexpected list: %+v", list)
	}

	latest, err := memory.LatestDone(ctx, "https://example.ru")
	if err != nil {
		t.Fatal(err)
	}
	if latest.ID != "newer" {
		t.Fatalf("expected newer, got %s", latest.ID)
	}
}
