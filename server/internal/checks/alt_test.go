package checks

import (
	"testing"

	"accessible-content/server/internal/crawler"
)

func TestAltChecker(t *testing.T) {
	page := crawler.Page{
		Images: []crawler.Image{
			{Src: "/noalt.png", HasAlt: false},
			{Src: "/empty.png", Alt: "", HasAlt: true},
			{Src: "/generic.png", Alt: "image2.png", HasAlt: true},
			{Src: "/generic2.png", Alt: "Фото", HasAlt: true},
			{Src: "/good.png", Alt: "Собака на прогулке в парке", HasAlt: true},
		},
	}

	findings := AltChecker{}.Check(page)
	if len(findings) != 4 {
		t.Fatalf("ожидали 4 находки, получили %d: %+v", len(findings), findings)
	}

	codes := make(map[string]int)
	for _, f := range findings {
		codes[f.Code]++
	}

	if codes["IMG_ALT_MISSING"] != 1 {
		t.Errorf("ожидали IMG_ALT_MISSING, получили %+v", codes)
	}
	if codes["IMG_ALT_EMPTY"] != 1 {
		t.Errorf("ожидали IMG_ALT_EMPTY, получили %+v", codes)
	}
	if codes["IMG_ALT_GENERIC"] != 2 {
		t.Errorf("ожидали 2 IMG_ALT_GENERIC, получили %+v", codes)
	}
}
