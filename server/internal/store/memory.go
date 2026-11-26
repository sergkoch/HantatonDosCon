package store

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

const maxMemoryScans = 1000

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

	if len(m.scans) >= maxMemoryScans {
		m.removeOldestLocked()
	}

	m.scans[scan.ID] = cloneScan(scan)
	return nil
}

func (m *Memory) Get(_ context.Context, id string) (Scan, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	scan, exists := m.scans[id]
	if !exists {
		return Scan{}, ErrNotFound
	}

	return cloneScan(scan), nil
}

func (m *Memory) Update(_ context.Context, scan Scan) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.scans[scan.ID]; !exists {
		return ErrNotFound
	}

	m.scans[scan.ID] = cloneScan(scan)
	return nil
}

func (m *Memory) List(_ context.Context, targetURL, owner string, limit int) ([]Scan, error) {
	if limit <= 0 {
		limit = 20
	}

	m.mu.RLock()
	scans := make([]Scan, 0, len(m.scans))
	for _, scan := range m.scans {
		if scan.Owner != owner {
			continue
		}
		if targetURL != "" && scan.URL != targetURL {
			continue
		}
		scans = append(scans, cloneScan(scan))
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

func (m *Memory) LatestDone(_ context.Context, targetURL, owner string) (Scan, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var latest Scan
	found := false

	for _, scan := range m.scans {
		if scan.URL != targetURL || scan.Status != StatusDone || scan.Owner != owner {
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

	return cloneScan(latest), nil
}

func (m *Memory) removeOldestLocked() {
	var oldestID string
	var oldest time.Time
	found := false

	for id, scan := range m.scans {
		if !found || scan.CreatedAt.Before(oldest) {
			oldestID = id
			oldest = scan.CreatedAt
			found = true
		}
	}

	if found {
		delete(m.scans, oldestID)
	}
}

// DeleteByOwner удаляет все сканы одного клиента.
func (m *Memory) DeleteByOwner(_ context.Context, owner string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	deleted := 0
	for id, scan := range m.scans {
		if scan.Owner == owner {
			delete(m.scans, id)
			deleted++
		}
	}

	return deleted, nil
}

func cloneScan(scan Scan) Scan {
	scan.Findings = append([]Finding(nil), scan.Findings...)

	if scan.Changes != nil {
		changes := *scan.Changes
		changes.Added = append([]Finding(nil), scan.Changes.Added...)
		changes.Fixed = append([]Finding(nil), scan.Changes.Fixed...)
		scan.Changes = &changes
	}

	return scan
}
