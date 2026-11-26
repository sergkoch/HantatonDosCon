package api

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	maxRequestBodyBytes = 1 << 20 // 1 MiB
	rateWindow          = time.Minute
	defaultRateLimit    = 20
	// globalLimitFactor задаёт общий потолок: даже если сканы приходят
	// с разных адресов, сервис не берёт больше limit*factor за минуту.
	globalLimitFactor = 5
	// maxTrackedIPs ограничивает память под счётчики по адресам.
	maxTrackedIPs = 4096
)

type rateLimiter struct {
	mu     sync.Mutex
	limit  int
	global int
	window time.Duration
	hits   []time.Time
	perIP  map[string][]time.Time
}

func newRateLimiter(perMinute int) *rateLimiter {
	if perMinute <= 0 {
		perMinute = defaultRateLimit
	}
	return &rateLimiter{
		limit:  perMinute,
		global: perMinute * globalLimitFactor,
		window: rateWindow,
		perIP:  make(map[string][]time.Time),
	}
}

// allow проверяет только общий лимит. Оставлен для тестов и простых вызовов.
func (l *rateLimiter) allow(now time.Time) bool {
	return l.allowFor(now, "")
}

// allowFor проверяет лимит для конкретного адреса и общий потолок.
func (l *rateLimiter) allowFor(now time.Time, ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := now.Add(-l.window)
	l.hits = pruneHits(l.hits, cutoff)

	if len(l.hits) >= l.global {
		return false
	}

	if ip != "" {
		ipHits := pruneHits(l.perIP[ip], cutoff)
		if len(ipHits) >= l.limit {
			l.perIP[ip] = ipHits
			return false
		}
		l.perIP[ip] = append(ipHits, now)
		l.dropStaleIPs(cutoff)
	}

	l.hits = append(l.hits, now)
	return true
}

// dropStaleIPs убирает адреса без активности, чтобы карта не росла бесконечно.
func (l *rateLimiter) dropStaleIPs(cutoff time.Time) {
	if len(l.perIP) <= maxTrackedIPs {
		return
	}
	for ip, hits := range l.perIP {
		if len(pruneHits(hits, cutoff)) == 0 {
			delete(l.perIP, ip)
		}
	}
}

func pruneHits(hits []time.Time, cutoff time.Time) []time.Time {
	kept := hits[:0]
	for _, hit := range hits {
		if hit.After(cutoff) {
			kept = append(kept, hit)
		}
	}
	return kept
}

func rateLimit(limiter *rateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.allowFor(time.Now(), requestIP(c)) {
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

// requestIP берёт последний адрес из X-Forwarded-For: его дописывает наш
// nginx, остальные элементы клиент может подделать.
func requestIP(c *gin.Context) string {
	if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if ip := strings.TrimSpace(parts[len(parts)-1]); ip != "" {
			return ip
		}
	}
	if host, _, err := net.SplitHostPort(c.Request.RemoteAddr); err == nil {
		return host
	}
	return c.Request.RemoteAddr
}

func limitBody(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
