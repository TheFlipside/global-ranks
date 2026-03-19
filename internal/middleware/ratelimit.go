package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// maxLimiters caps the number of per-key rate limiters to prevent
// memory exhaustion from distributed attacks with many unique keys.
const maxLimiters = 100000

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter holds per-key rate limiters with automatic cleanup of stale entries.
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*limiterEntry
	rps      rate.Limit
	burst    int
}

// NewRateLimiter creates a rate limiter that allows rps requests per second
// with the given burst size.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*limiterEntry),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
	go rl.cleanup()
	return rl
}

// Allow checks whether the given key is within its rate limit.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	entry, ok := rl.limiters[key]
	if !ok {
		if len(rl.limiters) >= maxLimiters {
			rl.mu.Unlock()
			return false
		}
		entry = &limiterEntry{
			limiter: rate.NewLimiter(rl.rps, rl.burst),
		}
		rl.limiters[key] = entry
	}
	entry.lastSeen = time.Now()
	rl.mu.Unlock()

	return entry.limiter.Allow()
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		for key, entry := range rl.limiters {
			if entry.lastSeen.Before(cutoff) {
				delete(rl.limiters, key)
			}
		}
		rl.mu.Unlock()
	}
}

// PerIPRateLimit applies a general per-IP rate limit to all requests.
// trustedProxies defines CIDRs whose X-Forwarded-For header is trusted.
func PerIPRateLimit(rl *RateLimiter, trustedProxies []*net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractIP(r, trustedProxies)
			if !rl.Allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "5")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate limit exceeded"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// extractIP returns the client IP for rate limiting.
// If the request comes from a trusted proxy, the rightmost untrusted IP
// in X-Forwarded-For is used. Otherwise, RemoteAddr is returned directly.
func extractIP(r *http.Request, trusted []*net.IPNet) string {
	remoteHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteHost = r.RemoteAddr
	}

	if len(trusted) == 0 {
		return remoteHost
	}

	remoteIP := net.ParseIP(remoteHost)
	if remoteIP == nil || !isTrusted(remoteIP, trusted) {
		return remoteHost
	}

	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded == "" {
		return remoteHost
	}

	// Walk the chain from right to left, skipping trusted proxy IPs.
	// The first non-trusted IP is the real client.
	parts := strings.Split(forwarded, ",")
	for i := len(parts) - 1; i >= 0; i-- {
		candidate := strings.TrimSpace(parts[i])
		ip := net.ParseIP(candidate)
		if ip == nil {
			// Skip non-IP values (hostnames, garbage) to prevent
			// injection of arbitrary strings as rate-limit keys.
			continue
		}
		if !isTrusted(ip, trusted) {
			return ip.String()
		}
	}

	return remoteHost
}

func isTrusted(ip net.IP, nets []*net.IPNet) bool {
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
