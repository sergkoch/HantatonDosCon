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
	// TitleAlt and RecommendationAlt hold the same text in the other supported
	// language, so a report can be exported in either one. They are not exposed
	// through the JSON API: the UI shows the scan language.
	TitleAlt          string `json:"-"`
	RecommendationAlt string `json:"-"`
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
	ID string `json:"id"`
	// Owner is the anonymous client that started the scan. It is never exposed
	// in API responses: history and reports stay private to that client.
	Owner           string    `json:"-"`
	URL             string    `json:"url"`
	Depth           int       `json:"depth"`
	Status          Status    `json:"status"`
	Lang            string    `json:"lang"`
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
	// List returns scans of one owner, newest first. An empty targetURL means
	// all URLs; an empty owner means scans created without a client identity.
	List(ctx context.Context, targetURL, owner string, limit int) ([]Scan, error)
	// LatestDone returns the most recent completed scan of a URL for one owner.
	LatestDone(ctx context.Context, targetURL, owner string) (Scan, error)
}
