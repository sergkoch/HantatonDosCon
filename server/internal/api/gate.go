package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"accessible-content/server/internal/store"
	"accessible-content/server/internal/texts"

	"github.com/gin-gonic/gin"
)

type gateLimits struct {
	MinScore    int `json:"minScore"`
	MaxCritical int `json:"maxCritical"`
}

type gateResponse struct {
	ID       string       `json:"id"`
	URL      string       `json:"url"`
	Status   store.Status `json:"status"`
	Passed   bool         `json:"passed"`
	Score    *int         `json:"score"`
	Critical int          `json:"critical"`
	Warning  int          `json:"warning"`
	Limits   gateLimits   `json:"limits"`
	Reasons  []string     `json:"reasons"`
}

// gateInt читает целочисленный query-параметр. Пустое значение — значение
// по умолчанию; нечисловое — ошибка (ok == false).
func gateInt(c *gin.Context, name string, fallback int) (int, bool) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return fallback, true
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}

	return value, true
}

// gate — CI-гейт: отдаёт 200, если скан проходит пороги, и 409, если нет.
// Так `curl --fail` завершается ненулевым кодом и ломает пайплайн.
func (h *handler) gate(c *gin.Context) {
	lang := requestLang(c)

	minScore, ok := gateInt(c, "min_score", 0)
	if !ok || minScore < 0 || minScore > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "detail": texts.T(lang, texts.ErrGateMinScore)})
		return
	}

	maxCritical, ok := gateInt(c, "max_critical", 0)
	if !ok || maxCritical < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "detail": texts.T(lang, texts.ErrGateMaxCritical)})
		return
	}

	scan, err := h.store.Get(c.Request.Context(), c.Param("id"))
	if errors.Is(err, store.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "detail": texts.T(lang, texts.ErrNotFound)})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "store_error", "detail": texts.T(lang, texts.ErrStore)})
		return
	}

	// Чужой скан неотличим от несуществующего.
	if !isOwner(scan, clientID(c)) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "detail": texts.T(lang, texts.ErrNotFound)})
		return
	}

	switch scan.Status {
	case store.StatusQueued, store.StatusRunning:
		c.JSON(http.StatusConflict, gin.H{
			"id":     scan.ID,
			"url":    scan.URL,
			"status": scan.Status,
			"error":  "not_ready",
			"detail": texts.T(lang, texts.ErrGateNotReady),
		})
		return
	case store.StatusFailed:
		detail := scan.Error
		if detail == "" {
			detail = texts.T(lang, texts.ErrGateFailed)
		}
		c.JSON(http.StatusConflict, gin.H{
			"id":     scan.ID,
			"url":    scan.URL,
			"status": scan.Status,
			"error":  "failed",
			"detail": detail,
		})
		return
	}

	score := 0
	if scan.Score != nil {
		score = *scan.Score
	}

	reasons := make([]string, 0, 2)
	if score < minScore {
		reasons = append(reasons, texts.Tf(lang, texts.ErrGateScore, score, minScore))
	}
	if scan.Stats.Critical > maxCritical {
		reasons = append(reasons, texts.Tf(lang, texts.ErrGateCritical, scan.Stats.Critical, maxCritical))
	}

	passed := len(reasons) == 0
	status := http.StatusOK
	if !passed {
		status = http.StatusConflict
	}

	c.JSON(status, gateResponse{
		ID:       scan.ID,
		URL:      scan.URL,
		Status:   scan.Status,
		Passed:   passed,
		Score:    scan.Score,
		Critical: scan.Stats.Critical,
		Warning:  scan.Stats.Warning,
		Limits: gateLimits{
			MinScore:    minScore,
			MaxCritical: maxCritical,
		},
		Reasons: reasons,
	})
}
