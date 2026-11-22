package report

import (
	"io"

	"accessible-content/server/internal/store"
)

type Format string

const (
	FormatCSV Format = "csv"
	FormatPDF Format = "pdf"
)

type Writer interface {
	Write(w io.Writer, scan store.Scan, format Format) error
}
