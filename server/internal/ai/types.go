package ai

import "context"

type ImageRequest struct {
	Image    []byte
	MimeType string
	Alt      string
	Context  string
}

type ImageResult struct {
	Verdict      string  `json:"verdict"`
	Reason       string  `json:"reason"`
	SuggestedAlt string  `json:"suggested_alt"`
	Confidence   float64 `json:"confidence"`
}

type LinkRequest struct {
	Text        string
	URL         string
	TargetTitle string
	TargetText  string
}

type LinkResult struct {
	Meaningful    bool    `json:"meaningful"`
	Matches       bool    `json:"matches"`
	Reason        string  `json:"reason"`
	SuggestedText string  `json:"suggested_text"`
	Confidence    float64 `json:"confidence"`
}

type Client interface {
	CheckImage(ctx context.Context, req ImageRequest) (ImageResult, error)
	CheckLink(ctx context.Context, req LinkRequest) (LinkResult, error)
}
