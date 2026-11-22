package store

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("scan not found")

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
	SeverityOK       Severity = "ok"
)

type Status string

const (
	StatusQueued  Status = "queued"
	StatusRunning Status = "running"
	StatusDone    Status = "done"
	StatusFailed  Status = "failed"
)

type Finding struct {
	ID             string   `json:"id"`
	Kind           string   `json:"kind"`
	Severity       Severity `json:"severity"`
	Title          string   `json:"title"`
	Element        string   `json:"element"`
	Recommendation string   `json:"recommendation"`
	WCAG           string   `json:"wcag"`
}

type Stats struct {
	Critical int `json:"critical"`
	Warning  int `json:"warning"`
	OK       int `json:"ok"`
}

type Scan struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Depth     int       `json:"depth"`
	Status    Status    `json:"status"`
	Score     *int      `json:"score"`
	Stats     Stats     `json:"stats"`
	Findings  []Finding `json:"findings"`
	CreatedAt time.Time `json:"createdAt"`
}

type Store interface {
	Create(ctx context.Context, scan Scan) error
	Get(ctx context.Context, id string) (Scan, error)
	Update(ctx context.Context, scan Scan) error
}
