package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/humfurie/pulpulitiko/api/pkg/cache"
)

type RateLimiter struct {
	cache      *cache.RedisCache
	maxReqs    int64
	windowSecs int64
}

func NewRateLimiter(redisCache *cache.RedisCache, maxRequests int64, windowSeconds int64) *RateLimiter {
	return &RateLimiter{
		cache:      redisCache,
		maxReqs:    maxRequests,
		windowSecs: windowSeconds,
	}
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		key := cache.RateLimitKey(ip)

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		count, err := rl.cache.Increment(ctx, key)
		if err != nil {
			// If Redis fails, allow the request
			next.ServeHTTP(w, r)
			return
		}

		// Set expiry on first request
		if count == 1 {
			ttl := time.Duration(rl.windowSecs) * time.Second
			_ = rl.cache.Set(ctx, key, count, ttl)
		}

		if count > rl.maxReqs {
			w.Header().Set("Retry-After", "60")
			http.Error(w, `{"success":false,"error":{"code":"RATE_LIMITED","message":"too many requests, please try again later"}}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header (common in reverse proxy setups)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}

	// Check for X-Real-IP header (used by nginx)
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
