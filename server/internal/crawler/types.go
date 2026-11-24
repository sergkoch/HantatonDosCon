package crawler

import "context"

type Image struct {
	Src      string
	Alt      string
	HasAlt   bool
	InLink   bool
	LinkText string
}

type Link struct {
	Href  string
	Text  string
	Label string
}

type Heading struct {
	Level int
	Text  string
}

type Input struct {
	Type        string
	Name        string
	Placeholder string
	HasLabel    bool
}

type Table struct {
	HasHeaders bool
}

type Page struct {
	URL        string
	Title      string
	Lang       string
	TextLength int
	Images     []Image
	Links      []Link
	PDFs       []Link
	Headings   []Heading
	Inputs     []Input
	Tables     []Table
}

type Crawler interface {
	Crawl(ctx context.Context, startURL string, depth int) ([]Page, error)
}
