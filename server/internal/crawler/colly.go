package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/temoto/robotstxt"
)

const defaultUserAgent = "AccessibleContentBot/0.1 (+https://github.com/sergkoch/HantatonDosCon)"

// Options настраивает краулер.
type Options struct {
	MaxPages       int
	Parallelism    int
	Delay          time.Duration
	RequestTimeout time.Duration
	UserAgent      string
}

// Colly — реализация Crawler на основе Colly.
type Colly struct {
	opts Options
	http *http.Client
}

// New создаёт краулер с разумными значениями по умолчанию.
func New(opts Options) *Colly {
	if opts.MaxPages <= 0 {
		opts.MaxPages = 50
	}
	if opts.Parallelism <= 0 || opts.Parallelism > 4 {
		opts.Parallelism = 4
	}
	if opts.Delay <= 0 {
		opts.Delay = 200 * time.Millisecond
	}
	if opts.RequestTimeout <= 0 {
		opts.RequestTimeout = 30 * time.Second
	}
	if opts.UserAgent == "" {
		opts.UserAgent = defaultUserAgent
	}

	return &Colly{
		opts: opts,
		http: &http.Client{Timeout: opts.RequestTimeout},
	}
}

// Crawl обходит сайт в пределах домена и собирает страницы.
func (c *Colly) Crawl(ctx context.Context, startURL string, depth int) ([]Page, error) {
	start, err := url.Parse(startURL)
	if err != nil || start.Host == "" || (start.Scheme != "http" && start.Scheme != "https") {
		return nil, fmt.Errorf("crawler: неверный стартовый URL")
	}

	if depth < 1 {
		depth = 1
	}

	host := start.Host
	robots := c.fetchRobots(ctx, start)

	collector := colly.NewCollector(
		colly.AllowedDomains(start.Hostname()),
		colly.MaxDepth(depth),
		colly.Async(true),
		colly.UserAgent(c.opts.UserAgent),
	)
	collector.SetRequestTimeout(c.opts.RequestTimeout)

	if err := collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: c.opts.Parallelism,
		Delay:       c.opts.Delay,
	}); err != nil {
		return nil, err
	}

	var mu sync.Mutex
	pages := make(map[string]*Page)
	order := make([]string, 0, c.opts.MaxPages)
	limitReached := false

	collector.OnHTML("html", func(e *colly.HTMLElement) {
		pageURL := e.Request.URL.String()

		mu.Lock()
		if _, exists := pages[pageURL]; exists {
			mu.Unlock()
			return
		}
		if len(pages) >= c.opts.MaxPages {
			limitReached = true
			mu.Unlock()
			return
		}
		mu.Unlock()

		page := &Page{
			URL:   pageURL,
			Title: strings.TrimSpace(e.ChildText("title")),
		}

		e.ForEach("img", func(_ int, el *colly.HTMLElement) {
			src := el.Request.AbsoluteURL(strings.TrimSpace(el.Attr("src")))
			if src == "" {
				return
			}
			page.Images = append(page.Images, Image{Src: src, Alt: el.Attr("alt")})
		})

		e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
			href := el.Request.AbsoluteURL(strings.TrimSpace(el.Attr("href")))
			if href == "" {
				return
			}
			text := strings.TrimSpace(el.Text)
			if strings.HasSuffix(strings.ToLower(href), ".pdf") {
				page.PDFs = append(page.PDFs, Link{Href: href, Text: text})
			} else {
				page.Links = append(page.Links, Link{Href: href, Text: text})
			}
		})

		mu.Lock()
		if _, exists := pages[pageURL]; !exists && len(pages) < c.opts.MaxPages {
			pages[pageURL] = page
			order = append(order, pageURL)
		}
		mu.Unlock()
	})

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		mu.Lock()
		reached := limitReached || len(pages) >= c.opts.MaxPages
		mu.Unlock()
		if reached {
			return
		}

		href := e.Request.AbsoluteURL(strings.TrimSpace(e.Attr("href")))
		if !c.allow(href, host, robots) {
			return
		}

		// Ошибки отдельных страниц не должны прерывать обход.
		_ = e.Request.Visit(href)
	})

	if err := collector.Visit(startURL); err != nil {
		return nil, err
	}
	collector.Wait()

	result := make([]Page, 0, len(order))
	mu.Lock()
	for _, pageURL := range order {
		result = append(result, *pages[pageURL])
	}
	mu.Unlock()

	return result, nil
}

func (c *Colly) fetchRobots(ctx context.Context, start *url.URL) *robotstxt.RobotsData {
	robotsURL := fmt.Sprintf("%s://%s/robots.txt", start.Scheme, start.Host)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, robotsURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", c.opts.UserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	data, err := robotstxt.FromResponse(resp)
	if err != nil {
		return nil
	}

	return data
}

func (c *Colly) allow(rawURL, host string, robots *robotstxt.RobotsData) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host != host {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	if strings.HasSuffix(strings.ToLower(parsed.Path), ".pdf") {
		return false
	}
	if robots != nil {
		group := robots.FindGroup(c.opts.UserAgent)
		if group != nil && !group.Test(parsed.Path) {
			return false
		}
	}

	return true
}
