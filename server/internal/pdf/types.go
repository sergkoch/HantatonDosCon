package pdf

import (
	"context"
	"io"
)

type Result struct {
	PageCount int
	TextLayer bool
	Tagged    bool
	Title     string
	Lang      string
	AltTexts  int
}

type Analyzer interface {
	Analyze(ctx context.Context, file io.ReaderAt, size int64) (Result, error)
}
