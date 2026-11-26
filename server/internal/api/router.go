package api

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"accessible-content/server/internal/ai"
	"accessible-content/server/internal/crawler"
	"accessible-content/server/internal/pdf"
	"accessible-content/server/internal/safehttp"
	"accessible-content/server/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Options struct {
	Store        store.Store
	Log          *slog.Logger
	AIConfigured bool
	AIVision     bool
	Crawler      crawler.Crawler
	AI           ai.Client
	Policy       *safehttp.Policy
	HTTPClient   *http.Client
	RateLimit    int
	PDF          pdf.Analyzer
}

type handler struct {
	store        store.Store
	log          *slog.Logger
	aiConfigured bool
	aiVision     bool
	crawler      crawler.Crawler
	aiClient     ai.Client
	pdfAnalyzer  pdf.Analyzer
	httpClient   *http.Client
	policy       *safehttp.Policy
	pdfClient    *http.Client
}

func NewRouter(opts Options) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	log := opts.Log
	if log == nil {
		log = slog.Default()
	}

	router := gin.New()
	router.Use(gin.Recovery(), requestID(), cors(), securityHeaders(), requestLogger(log))
	_ = router.SetTrustedProxies(nil)

	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}

	pdfClient := &http.Client{Timeout: 30 * time.Second}
	if opts.Policy != nil {
		pdfClient = safehttp.NewClient(opts.Policy, 30*time.Second)
	}

	h := &handler{
		store:        opts.Store,
		log:          log,
		aiConfigured: opts.AIConfigured,
		aiVision:     opts.AIVision,
		crawler:      opts.Crawler,
		aiClient:     opts.AI,
		httpClient:   httpClient,
		policy:       opts.Policy,
		pdfAnalyzer:  opts.PDF,
		pdfClient:    pdfClient,
	}
	limiter := newRateLimiter(opts.RateLimit)

	api := router.Group("/api")
	api.GET("/health", h.health)
	api.POST("/scans", limitBody(maxRequestBodyBytes), rateLimit(limiter), h.createScan)
	api.GET("/scans", h.listScans)
	api.GET("/scans/:id", h.getScan)
	api.GET("/scans/:id/pdf", h.downloadPDF)

	return router
}

func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Request-ID", uuid.NewString())
		c.Next()
	}
}

func cors() gin.HandlerFunc {
	origin := os.Getenv("CORS_ORIGIN")
	if origin == "" {
		origin = "*" // development default
	}

	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Client-ID")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Next()
	}
}

func requestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()

		log.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", time.Since(started).String(),
		)
	}
}
