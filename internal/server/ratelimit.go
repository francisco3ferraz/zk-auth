package server

import (
	"net/http"
	"sync"
	"time"

	"github.com/francisco3ferraz/zk-auth/internal/config"
	"github.com/francisco3ferraz/zk-auth/internal/errors"
)

// RateLimiter implements a sliding window rate limiter per IP address
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	maxReqs  int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter with the given configuration
func NewRateLimiter(cfg *config.SecurityConfig) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		maxReqs:  cfg.RateLimitReqs,
		window:   cfg.RateLimitWindow,
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// Allow checks if the given IP is allowed to make a request
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get existing requests for this IP
	requests := rl.requests[ip]

	// Filter out expired requests
	var validRequests []time.Time
	for _, t := range requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}

	// Check if limit exceeded
	if len(validRequests) >= rl.maxReqs {
		rl.requests[ip] = validRequests
		return false
	}

	// Add new request
	validRequests = append(validRequests, now)
	rl.requests[ip] = validRequests

	return true
}

// cleanupLoop periodically removes expired entries
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	for ip, requests := range rl.requests {
		var validRequests []time.Time
		for _, t := range requests {
			if t.After(windowStart) {
				validRequests = append(validRequests, t)
			}
		}

		if len(validRequests) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = validRequests
		}
	}
}

// RateLimitMiddleware creates a middleware that limits requests per IP
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			if !limiter.Allow(ip) {
				errors.NewTooManyRequestsError().WriteResponse(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
