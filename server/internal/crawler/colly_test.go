package crawler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestCrawler() *Colly {
	return New(Options{
		MaxPages:       10,
		Parallelism:    2,
		Delay:          time.Millisecond,
		RequestTimeout: 5 * time.Second,
	})
}

func TestCrawlCollectsContent(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, `<!doctype html><html><head><title>Главная</title></head><body>
			<img src="/pic.png" alt="Фото">
			<img src="/pic2.png">
			<a href="/about">О нас</a>
			<a href="/doc.pdf">Документ</a>
			<a href="https://external.example/x">чужое</a>
		</body></html>`)
	})

	mux.HandleFunc("/about", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, `<!doctype html><html><head><title>О нас</title></head><body>
			<a href="/">Назад</a>
		</body></html>`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	pages, err := newTestCrawler().Crawl(context.Background(), server.URL, 2)
	if err != nil {
		t.Fatalf("Crawl вернул ошибку: %v", err)
	}

	if len(pages) < 2 {
		t.Fatalf("ожидали минимум 2 страницы, получили %d", len(pages))
	}

	var home *Page
	for i := range pages {
		if strings.Contains(pages[i].Title, "Главная") {
			home = &pages[i]
		}
		if strings.Contains(pages[i].URL, "external.example") {
			t.Fatalf("чужой домен не должен обходиться: %s", pages[i].URL)
		}
	}

	if home == nil {
		t.Fatal("не нашли главную страницу")
	}

	if len(home.Images) != 2 {
		t.Fatalf("ожидали 2 изображения, получили %d", len(home.Images))
	}
	if home.Images[0].Alt != "Фото" {
		t.Fatalf("ожидали alt=Фото, получили %q", home.Images[0].Alt)
	}
	if home.Images[1].Alt != "" {
		t.Fatalf("у второй картинки alt должен быть пустым, получили %q", home.Images[1].Alt)
	}
	if len(home.PDFs) != 1 || !strings.HasSuffix(home.PDFs[0].Href, "doc.pdf") {
		t.Fatalf("ожидали один PDF doc.pdf, получили %+v", home.PDFs)
	}
}

func TestCrawlRespectsRobots(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "User-agent: *\nDisallow: /private\n")
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, `<!doctype html><html><head><title>Старт</title></head><body>
			<a href="/public">Открытая</a>
			<a href="/private">Закрытая</a>
		</body></html>`)
	})

	mux.HandleFunc("/public", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `<!doctype html><html><head><title>Публичная</title></head><body></body></html>`)
	})

	mux.HandleFunc("/private", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `<!doctype html><html><head><title>Секрет</title></head><body></body></html>`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	pages, err := newTestCrawler().Crawl(context.Background(), server.URL, 2)
	if err != nil {
		t.Fatalf("Crawl вернул ошибку: %v", err)
	}

	var sawPublic, sawPrivate bool
	for _, p := range pages {
		if strings.HasSuffix(p.URL, "/public") {
			sawPublic = true
		}
		if strings.HasSuffix(p.URL, "/private") {
			sawPrivate = true
		}
	}

	if !sawPublic {
		t.Fatal("публичная страница должна была обойтись")
	}
	if sawPrivate {
		t.Fatal("страница, закрытая в robots.txt, не должна обходиться")
	}
}

func TestCrawlDepthOne(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			http.NotFound(w, r)
			return
		}
		_, _ = io.WriteString(w, `<html><head><title>Старт</title></head><body><a href="/next">next</a></body></html>`)
	})
	mux.HandleFunc("/next", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `<html><head><title>Next</title></head><body></body></html>`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	pages, err := newTestCrawler().Crawl(context.Background(), server.URL, 1)
	if err != nil {
		t.Fatalf("Crawl вернул ошибку: %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("при глубине 1 ожидали 1 страницу, получили %d", len(pages))
	}
}
