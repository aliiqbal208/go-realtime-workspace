package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
	RedisClient       *redis.Client
	Logger            zerolog.Logger
}

// RateLimit middleware implements Redis-based rate limiting per IP
func RateLimit(config RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract IP address
			ip := getClientIP(r)
			key := fmt.Sprintf("rate_limit:%s", ip)

			ctx := context.Background()

			// Get current count
			count, err := config.RedisClient.Get(ctx, key).Int()
			if err != nil && err != redis.Nil {
				config.Logger.Error().Err(err).Str("ip", ip).Msg("Rate limit check failed")
				// On error, allow request to proceed
				next.ServeHTTP(w, r)
				return
			}

			// Check if rate limit exceeded
			if count >= config.RequestsPerMinute {
				// Get TTL to inform client when to retry
				ttl, _ := config.RedisClient.TTL(ctx, key).Result()

				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))
				w.Header().Set("Retry-After", strconv.Itoa(int(ttl.Seconds())))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)

				requestID := GetRequestID(r.Context())
				fmt.Fprintf(w, `{"error":"Rate limit exceeded","retry_after":%d,"request_id":"%s"}`, int(ttl.Seconds()), requestID)
				return
			}

			// Increment counter
			pipe := config.RedisClient.Pipeline()
			pipe.Incr(ctx, key)
			if count == 0 {
				// Set expiration only on first request
				pipe.Expire(ctx, key, time.Minute)
			}
			_, err = pipe.Exec(ctx)
			if err != nil {
				config.Logger.Error().Err(err).Str("ip", ip).Msg("Rate limit increment failed")
			}

			// Add rate limit headers
			remaining := config.RequestsPerMinute - count - 1
			if remaining < 0 {
				remaining = 0
			}
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the real client IP address
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (behind proxy)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := len(xff); idx > 0 {
			for i := 0; i < len(xff); i++ {
				if xff[i] == ',' {
					return xff[:i]
				}
			}
			return xff
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}
