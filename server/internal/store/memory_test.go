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

func TestMemoryClonesFindings(t *testing.T) {
	ctx := context.Background()
	memory := NewMemory()
	scan := Scan{
		ID:        "clone",
		URL:       "https://example.ru",
		Status:    StatusDone,
		Findings:  []Finding{{Title: "original"}},
		CreatedAt: time.Now().UTC(),
	}
	if err := memory.Create(ctx, scan); err != nil {
		t.Fatal(err)
	}

	scan.Findings[0].Title = "mutated after create"
	loaded, err := memory.Get(ctx, "clone")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Findings[0].Title != "original" {
		t.Fatalf("memory store shares findings slice: %q", loaded.Findings[0].Title)
	}

	loaded.Findings[0].Title = "mutated get"
	again, err := memory.Get(ctx, "clone")
	if err != nil {
		t.Fatal(err)
	}
	if again.Findings[0].Title != "original" {
		t.Fatalf("memory store shares findings on get: %q", again.Findings[0].Title)
	}
}

func TestMemoryPreservesEmptySlices(t *testing.T) {
	ctx := context.Background()
	memory := NewMemory()
	scan := Scan{
		ID:       "empty-slices",
		URL:      "https://example.ru",
		Status:   StatusDone,
		Findings: []Finding{},
		Changes:  &Changes{Added: []Finding{}, Fixed: []Finding{}},
	}
	if err := memory.Create(ctx, scan); err != nil {
		t.Fatal(err)
	}

	loaded, err := memory.Get(ctx, "empty-slices")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Findings == nil {
		t.Fatal("findings must stay an empty array, not null")
	}
	if loaded.Changes == nil || loaded.Changes.Added == nil || loaded.Changes.Fixed == nil {
		t.Fatal("changes arrays must stay empty arrays, not null")
	}
}

func TestMemoryLimit(t *testing.T) {
	ctx := context.Background()
	memory := NewMemory()
	now := time.Now().UTC()

	for i := 0; i < maxMemoryScans+1; i++ {
		scan := Scan{
			ID:        string(rune('a'+i%26)) + time.Duration(i).String(),
			URL:       "https://example.ru",
			Status:    StatusDone,
			CreatedAt: now.Add(time.Duration(i) * time.Second),
		}
		if err := memory.Create(ctx, scan); err != nil {
			t.Fatal(err)
		}
	}

	scans, err := memory.List(ctx, "", "", maxMemoryScans+10)
	if err != nil {
		t.Fatal(err)
	}
	if len(scans) != maxMemoryScans {
		t.Fatalf("expected %d scans, got %d", maxMemoryScans, len(scans))
	}
}

func TestMemoryDeleteByOwner(t *testing.T) {
	ctx := context.Background()
	memory := NewMemory()
	now := time.Now().UTC()

	for _, scan := range []Scan{
		{ID: "a1", Owner: "browser-a", URL: "https://a.ru", Status: StatusDone, CreatedAt: now},
		{ID: "a2", Owner: "browser-a", URL: "https://b.ru", Status: StatusDone, CreatedAt: now},
		{ID: "b1", Owner: "browser-b", URL: "https://a.ru", Status: StatusDone, CreatedAt: now},
	} {
		if err := memory.Create(ctx, scan); err != nil {
			t.Fatal(err)
		}
	}

	deleted, err := memory.DeleteByOwner(ctx, "browser-a")
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 2 {
		t.Fatalf("expected 2 deleted, got %d", deleted)
	}

	left, err := memory.List(ctx, "", "browser-a", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) != 0 {
		t.Fatalf("expected empty history, got %+v", left)
	}

	other, err := memory.List(ctx, "", "browser-b", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(other) != 1 || other[0].ID != "b1" {
		t.Fatalf("history of another browser must stay, got %+v", other)
	}
}
