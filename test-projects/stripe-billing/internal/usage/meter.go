package usage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kabirnarang39/stripe-billing/internal/cache"
	"github.com/kabirnarang39/stripe-billing/internal/httpx"
	"github.com/kabirnarang39/stripe-billing/internal/limits"
)

// Middleware returns a quota-checking middleware that:
//  1. Extracts workspace_id from the request context (set by JWT middleware).
//  2. Loads plan limits (Redis → Postgres fallback).
//  3. If the plan has no API call limit (Free plan), passes through.
//  4. Increments the Redis usage counter (fail-open: 0 on error).
//  5. If the count exceeds the limit, returns 429 Too Many Requests.
//  6. Notifies the Flusher for eventual Postgres sync.
func Middleware(cacheClient *cache.Client, limitsCache *limits.Cache, flusher *Flusher) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			workspaceID, ok := httpx.WorkspaceIDFromContext(ctx)
			if !ok || workspaceID == "" {
				next.ServeHTTP(w, r)
				return
			}

			pl, err := limitsCache.Get(ctx, workspaceID)
			if err != nil {
				// If we can't load limits, fail open: allow the request.
				next.ServeHTTP(w, r)
				return
			}

			// Free plan: no quota.
			if pl.APICallLimit == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Increment the Redis counter (fail-open).
			now := time.Now().UTC()
			period := now.Format("2006-01")
			usageKey := cache.KeyUsage(workspaceID, period)
			count := cacheClient.IncrFailOpen(ctx, usageKey)

			// Set expiry on first increment (best-effort; idempotent on subsequent calls).
			if count == 1 {
				nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
				ttl := nextMonth.Sub(now) + 48*time.Hour // buffer for month-boundary skew
				_ = cacheClient.Expire(ctx, usageKey, ttl)
			}

			limit := *pl.APICallLimit
			if count > int64(limit) {
				resetAt := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
				body, _ := json.Marshal(map[string]interface{}{
					"error":    "quota_exceeded",
					"message":  fmt.Sprintf("API quota of %d calls exceeded for the current billing period", limit),
					"limit":    limit,
					"current":  count,
					"plan":     pl.Plan,
					"reset_at": resetAt.Format(time.RFC3339),
				})
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(time.Until(resetAt).Seconds())))
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write(body) //nolint:errcheck
				return
			}

			// Notify flusher for eventual Postgres sync.
			periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
			periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
			flusher.Notify(flushItem{
				WorkspaceID:        workspaceID,
				BillingPeriodStart: periodStart,
				BillingPeriodEnd:   periodEnd,
			})

			next.ServeHTTP(w, r)
		})
	}
}

