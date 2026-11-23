package crawler

import "context"

type Image struct {
	Src    string
	Alt    string
	HasAlt bool
}

type Link struct {
	Href string
	Text string
}

type Page struct {
	URL    string
	Title  string
	Images []Image
	Links  []Link
	PDFs   []Link
}

type Crawler interface {
	Crawl(ctx context.Context, startURL string, depth int) ([]Page, error)
}
