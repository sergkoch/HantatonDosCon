package checks

import "accessible-content/server/internal/crawler"

type Finding struct {
	Code           string
	Severity       string
	Title          string
	Element        string
	Recommendation string
	WCAG           string
}

type Checker interface {
	Check(page crawler.Page) []Finding
}
