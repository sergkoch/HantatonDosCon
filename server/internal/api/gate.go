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
	ID       string     `json:"id"`
	URL      string     `json:"url"`
	Status   string     `json:"status"`
	Passed   bool       `json:"passed"`
	Score    *int       `json:"score"`
	Critical int        `json:"critical"`
	Warning  int        `json:"warning"`
	Limits   gateLimits `json:"limits"`
	Reasons  []string   `json:"reasons"`
}

// parseGateLimit читает целочисленный query-параметр.
// kind = "int" — не число, "range" — вне диапазона.
func parseGateLimit(raw string, fallback, min, max int) (int, string) {
	if strings.TrimSpace(raw) == "" {
		return fallback, ""
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, "int"
	}
	if value < min || (max >= 0 && value > max) {
		return 0, "range"
	}

	return value, ""
}

// gate проверяет скан по порогам и годится для CI: успех — 200, провал — 409.
func (h *handler) gate(c *gin.Context) {
	lang := requestLang(c)

	minScore, problem := parseGateLimit(c.Query("min_score"), 0, 0, 100)
	if problem != "" {
		key := texts.ErrParamRange
		if problem == "int" {
			key = texts.ErrParam
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "detail": texts.Tf(lang, key, "min_score")})
		return
	}

	maxCritical, problem := parseGateLimit(c.Query("max_critical"), 0, 0, -1)
	if problem != "" {
		key := texts.ErrParamRange
		if problem == "int" {
			key = texts.ErrParam
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_input", "detail": texts.Tf(lang, key, "max_critical")})
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
	if !isOwner(scan, clientID(c)) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "detail": texts.T(lang, texts.ErrNotFound)})
		return
	}

	if scan.Status == store.StatusQueued || scan.Status == store.StatusRunning {
		c.JSON(http.StatusConflict, gin.H{"error": "not_ready", "detail": texts.T(lang, texts.GateNotReady)})
		return
	}

	limits := gateLimits{MinScore: minScore, MaxCritical: maxCritical}
	reasons := make([]string, 0, 2)
	passed := scan.Status == store.StatusDone

	if !passed {
		reason := strings.TrimSpace(scan.Error)
		if reason == "" {
			reason = texts.T(lang, texts.GateNotReady)
		}
		reasons = append(reasons, reason)
	} else {
		if scan.Score != nil && *scan.Score < minScore {
			reasons = append(reasons, texts.Tf(lang, texts.GateScoreBelow, *scan.Score, minScore))
		}
		if scan.Stats.Critical > maxCritical {
			reasons = append(reasons, texts.Tf(lang, texts.GateCriticalAbove, scan.Stats.Critical, maxCritical))
		}
		passed = len(reasons) == 0
	}

	code := http.StatusOK
	if !passed {
		code = http.StatusConflict
	}

	c.JSON(code, gateResponse{
		ID:       scan.ID,
		URL:      scan.URL,
		Status:   string(scan.Status),
		Passed:   passed,
		Score:    scan.Score,
		Critical: scan.Stats.Critical,
		Warning:  scan.Stats.Warning,
		Limits:   limits,
		Reasons:  reasons,
	})
}
