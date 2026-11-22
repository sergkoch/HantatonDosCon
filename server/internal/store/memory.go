package store

import (
	"context"
	"errors"
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
