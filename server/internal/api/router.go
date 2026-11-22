package api

import (
	"log/slog"
	"net/http"
	"time"

	"accessible-content/server/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Options struct {
	Store        store.Store
	Log          *slog.Logger
	AIConfigured bool
}

type handler struct {
	store        store.Store
	log          *slog.Logger
	aiConfigured bool
}

func NewRouter(opts Options) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	log := opts.Log
	if log == nil {
		log = slog.Default()
	}

	router := gin.New()
	router.Use(gin.Recovery(), requestID(), cors(), securityHeaders(), requestLogger(log))

	h := &handler{store: opts.Store, log: log, aiConfigured: opts.AIConfigured}
	api := router.Group("/api")
	api.GET("/health", h.health)
	api.POST("/scans", h.createScan)
	api.GET("/scans/:id", h.getScan)

	return router
}

func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Request-ID", uuid.NewString())
		c.Next()
	}
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

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
