package ai

import "context"

type ImageRequest struct {
	Image    []byte
	MimeType string
	Alt      string
	Context  string
}

type ImageResult struct {
	Verdict      string
	Reason       string
	SuggestedAlt string
	Confidence   float64
}

type LinkRequest struct {
	Text        string
	URL         string
	TargetTitle string
	TargetText  string
}

type LinkResult struct {
	Meaningful    bool
	Matches       bool
	Reason        string
	SuggestedText string
	Confidence    float64
}

type Client interface {
	CheckImage(ctx context.Context, req ImageRequest) (ImageResult, error)
	CheckLink(ctx context.Context, req LinkRequest) (LinkResult, error)
}
