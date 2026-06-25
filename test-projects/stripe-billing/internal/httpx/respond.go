package httpx

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// JSON writes v as a JSON response with the given HTTP status code.
func JSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Response headers already written; nothing to do.
		return
	}
}

// Error writes a structured JSON error response.
func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, map[string]string{
		"error":   code,
		"message": message,
	})
}

// LimitExceededError writes a 403 response indicating the workspace has hit a
// non-API-quota limit (e.g. user count limit).
func LimitExceededError(w http.ResponseWriter, limit, current int, plan string) {
	JSON(w, http.StatusForbidden, map[string]interface{}{
		"error":   "limit_exceeded",
		"message": "workspace limit exceeded for the current plan",
		"limit":   limit,
		"current": current,
		"plan":    plan,
	})
}

// QuotaExceededError writes a 429 response indicating the workspace has exhausted
// its API call quota for the current billing period.
func QuotaExceededError(w http.ResponseWriter, limit, current int, plan string, resetAt time.Time) {
	retryAfter := int(time.Until(resetAt).Seconds())
	if retryAfter < 0 {
		retryAfter = 0
	}
	w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	JSON(w, http.StatusTooManyRequests, map[string]interface{}{
		"error":    "quota_exceeded",
		"message":  "API quota exceeded for the current billing period",
		"limit":    limit,
		"current":  current,
		"plan":     plan,
		"reset_at": resetAt.Format(time.RFC3339),
	})
}

// RateLimitError writes a 429 response indicating the caller has exceeded the
// per-user request rate limit.
func RateLimitError(w http.ResponseWriter, retryAfter int) {
	w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
	JSON(w, http.StatusTooManyRequests, map[string]interface{}{
		"error":       "rate_limit_exceeded",
		"message":     "too many requests; please slow down",
		"retry_after": retryAfter,
	})
}
