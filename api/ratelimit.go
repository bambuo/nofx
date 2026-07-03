package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// rateLimiter is a simple in-memory sliding window rate limiter
type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	window   time.Duration
	maxReqs  int
}

// newRateLimiter creates a new rate limiter with the given window and max requests
func newRateLimiter(window time.Duration, maxReqs int) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		window:   window,
		maxReqs:  maxReqs,
	}
}

// allow checks if a request from the given key is allowed
func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Remove expired entries
	times := rl.requests[key]
	var valid []time.Time
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.maxReqs {
		rl.requests[key] = valid
		return false
	}

	rl.requests[key] = append(valid, now)
	return true
}

// cleanup periodically removes stale entries to prevent memory leaks
func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window * 2)
	for key, times := range rl.requests {
		var valid []time.Time
		for _, t := range times {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.requests, key)
		} else {
			rl.requests[key] = valid
		}
	}
}

// Global rate limiters for different endpoint categories
var (
	// Login rate limiter: 5 attempts per minute per IP
	loginLimiter = newRateLimiter(1*time.Minute, 5)
	// Registration rate limiter: 3 attempts per 10 minutes per IP
	registerLimiter = newRateLimiter(10*time.Minute, 3)
	// OTP rate limiter: 10 attempts per minute per IP
	otpLimiter = newRateLimiter(1*time.Minute, 10)
	// General API rate limiter: 60 requests per minute per IP
	apiLimiter = newRateLimiter(1*time.Minute, 60)
)

func init() {
	// Start cleanup goroutines for each limiter
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			loginLimiter.cleanup()
			registerLimiter.cleanup()
			otpLimiter.cleanup()
			apiLimiter.cleanup()
		}
	}()
}

// Client IP extraction helper
func clientIP(c *gin.Context) string {
	// Try X-Forwarded-For header first (for proxied requests)
	if fwd := c.GetHeader("X-Forwarded-For"); fwd != "" {
		return fwd
	}
	// Try X-Real-IP header
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}
	return c.ClientIP()
}

// rateLimitMiddleware returns a middleware that rate limits based on client IP
func rateLimitMiddleware(limiter *rateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := clientIP(c)
		if !limiter.allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
