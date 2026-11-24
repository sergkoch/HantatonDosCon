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
	Code           string   `json:"code,omitempty"`
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

// Changes describes the difference between two scans of the same URL.
type Changes struct {
	Added      []Finding `json:"added"`
	Fixed      []Finding `json:"fixed"`
	Remaining  int       `json:"remaining"`
	ScoreDelta *int      `json:"scoreDelta,omitempty"`
}

type Scan struct {
	ID              string    `json:"id"`
	URL             string    `json:"url"`
	Depth           int       `json:"depth"`
	Status          Status    `json:"status"`
	Version         int       `json:"version"`
	PreviousID      string    `json:"previousId,omitempty"`
	Score           *int      `json:"score"`
	Pages           int       `json:"pages"`
	CheckedLinks    int       `json:"checkedLinks"`
	UnverifiedLinks int       `json:"unverifiedLinks"`
	Error           string    `json:"error,omitempty"`
	Stats           Stats     `json:"stats"`
	Changes         *Changes  `json:"changes,omitempty"`
	Findings        []Finding `json:"findings"`
	CreatedAt       time.Time `json:"createdAt"`
}

type Store interface {
	Create(ctx context.Context, scan Scan) error
	Get(ctx context.Context, id string) (Scan, error)
	Update(ctx context.Context, scan Scan) error
	// List returns scans, newest first. An empty targetURL means all URLs.
	List(ctx context.Context, targetURL string, limit int) ([]Scan, error)
	// LatestDone returns the most recent completed scan for a URL.
	LatestDone(ctx context.Context, targetURL string) (Scan, error)
}
