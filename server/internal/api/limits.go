package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	maxRequestBodyBytes = 1 << 20 // 1 MiB
	rateWindow          = time.Minute
	defaultRateLimit    = 20
)

type rateLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	hits   []time.Time
}

func newRateLimiter(perMinute int) *rateLimiter {
	if perMinute <= 0 {
		perMinute = defaultRateLimit
	}
	return &rateLimiter{limit: perMinute, window: rateWindow}
}

func (l *rateLimiter) allow(now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := now.Add(-l.window)
	kept := l.hits[:0]
	for _, hit := range l.hits {
		if hit.After(cutoff) {
			kept = append(kept, hit)
		}
	}
	l.hits = kept

	if len(l.hits) >= l.limit {
		return false
	}

	l.hits = append(l.hits, now)
	return true
}

func rateLimit(limiter *rateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.allow(time.Now()) {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":  "rate_limited",
				"detail": "too many scans, try again in a minute",
			})
			return
		}
		c.Next()
	}
}

func limitBody(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
