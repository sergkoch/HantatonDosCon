package store

import (
	"context"
	"errors"
	"sort"
	"sync"
)

type Memory struct {
	mu    sync.RWMutex
	scans map[string]Scan
}

func NewMemory() *Memory {
	return &Memory{scans: make(map[string]Scan)}
}

func (m *Memory) Create(_ context.Context, scan Scan) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.scans[scan.ID]; exists {
		return errors.New("scan already exists")
	}

	m.scans[scan.ID] = scan
	return nil
}

func (m *Memory) Get(_ context.Context, id string) (Scan, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	scan, exists := m.scans[id]
	if !exists {
		return Scan{}, ErrNotFound
	}

	return scan, nil
}

func (m *Memory) Update(_ context.Context, scan Scan) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.scans[scan.ID]; !exists {
		return ErrNotFound
	}

	m.scans[scan.ID] = scan
	return nil
}

func (m *Memory) List(_ context.Context, targetURL string, limit int) ([]Scan, error) {
	if limit <= 0 {
		limit = 20
	}

	m.mu.RLock()
	scans := make([]Scan, 0, len(m.scans))
	for _, scan := range m.scans {
		if targetURL != "" && scan.URL != targetURL {
			continue
		}
		scans = append(scans, scan)
	}
	m.mu.RUnlock()

	sort.Slice(scans, func(i, j int) bool {
		return scans[i].CreatedAt.After(scans[j].CreatedAt)
	})

	if len(scans) > limit {
		scans = scans[:limit]
	}

	return scans, nil
}

func (m *Memory) LatestDone(_ context.Context, targetURL string) (Scan, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var latest Scan
	found := false

	for _, scan := range m.scans {
		if scan.URL != targetURL || scan.Status != StatusDone {
			continue
		}
		if !found || scan.CreatedAt.After(latest.CreatedAt) {
			latest = scan
			found = true
		}
	}

	if !found {
		return Scan{}, ErrNotFound
	}

	return latest, nil
}
