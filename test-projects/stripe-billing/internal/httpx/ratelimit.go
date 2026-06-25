package httpx

import (
	"fmt"
	"net/http"
	"time"

	"github.com/kabirnarang39/stripe-billing/internal/cache"
)

const (
	rateLimitMax    = 20
	rateLimitWindow = 60 // seconds
)

// RateLimitMiddleware limits authenticated users to 20 POST requests per 60-second
// window per user. The limit applies per user (from JWT) and per minute bucket.
//
// Key format: ratelimit:{userID}:{unix_minute_bucket}
// where unix_minute_bucket = unix_timestamp / 60.
func RateLimitMiddleware(cacheClient *cache.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only rate-limit POST requests.
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			userID, ok := UserIDFromContext(r.Context())
			if !ok || userID == "" {
				// No user in context; skip rate limiting (JWT middleware should have rejected first).
				next.ServeHTTP(w, r)
				return
			}

			now := time.Now().Unix()
			minuteBucket := fmt.Sprintf("%d", now/rateLimitWindow)
			key := cache.KeyRateLimit(userID, minuteBucket)

			count := cacheClient.IncrFailOpen(r.Context(), key)
			if count == 1 {
				// Set expiry on the first request in this bucket.
				_ = cacheClient.Expire(r.Context(), key, 2*time.Minute) // generous TTL for safety
			}

			if count > rateLimitMax {
				RateLimitError(w, rateLimitWindow)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
