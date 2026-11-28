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
	// Confidence is the AI confidence for semantic findings (0..1).
	Confidence *float64 `json:"confidence,omitempty"`
	// TitleAlt and RecommendationAlt hold the same text in the other supported
	// language. The API exposes them, so the report can switch language without
	// rescanning.
	TitleAlt          string `json:"titleAlt,omitempty"`
	RecommendationAlt string `json:"recommendationAlt,omitempty"`
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

// AICoverage records how much of the semantic check the AI actually did.
type AICoverage struct {
	Enabled       bool `json:"enabled"`
	ImagesChecked int  `json:"imagesChecked"`
	LinksChecked  int  `json:"linksChecked"`
	Skipped       int  `json:"skipped"`
}

type Scan struct {
	ID string `json:"id"`
	// Owner is the anonymous client that started the scan. It is never exposed
	// in API responses: history and reports stay private to that client.
	Owner           string      `json:"-"`
	URL             string      `json:"url"`
	Depth           int         `json:"depth"`
	Status          Status      `json:"status"`
	Lang            string      `json:"lang"`
	Version         int         `json:"version"`
	PreviousID      string      `json:"previousId,omitempty"`
	Score           *int        `json:"score"`
	Pages           int         `json:"pages"`
	CheckedLinks    int         `json:"checkedLinks"`
	UnverifiedLinks int         `json:"unverifiedLinks"`
	Error           string      `json:"error,omitempty"`
	Stats           Stats       `json:"stats"`
	AI              *AICoverage `json:"ai,omitempty"`
	Changes         *Changes    `json:"changes,omitempty"`
	Findings        []Finding   `json:"findings"`
	CreatedAt       time.Time   `json:"createdAt"`
}

type Store interface {
	Create(ctx context.Context, scan Scan) error
	Get(ctx context.Context, id string) (Scan, error)
	Update(ctx context.Context, scan Scan) error
	// List returns scans of one owner, newest first. An empty targetURL means
	// all URLs. The owner is always required by the API, so an empty owner
	// never exposes a shared history.
	List(ctx context.Context, targetURL, owner string, limit int) ([]Scan, error)
	// LatestDone returns the most recent completed scan of a URL for one owner.
	LatestDone(ctx context.Context, targetURL, owner string) (Scan, error)
	// DeleteByOwner removes every scan of one client and reports how many
	// were deleted.
	DeleteByOwner(ctx context.Context, owner string) (int, error)
}
