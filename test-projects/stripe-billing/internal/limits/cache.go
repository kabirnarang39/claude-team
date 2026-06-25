package limits

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/kabirnarang39/stripe-billing/internal/billing"
	"github.com/kabirnarang39/stripe-billing/internal/cache"
)

// PlanLimits holds the resolved limits for a workspace's current plan.
type PlanLimits struct {
	Plan         string    `json:"plan"`
	UserLimit    int       `json:"user_limit"`
	APICallLimit *int      `json:"api_call_limit"` // nil for Free (no quota)
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
}

// Cache is a read-through cache for plan limits backed by Redis with a Postgres fallback.
type Cache struct {
	redis       *cache.Client
	store       StoreIface
	apiQuotaPro int
}

// NewCache creates a new limits Cache.
func NewCache(redis *cache.Client, store StoreIface, apiQuotaPro int) *Cache {
	return &Cache{
		redis:       redis,
		store:       store,
		apiQuotaPro: apiQuotaPro,
	}
}

// Get returns the plan limits for the given workspace.
// It tries Redis first (key: plan_limits:{workspaceID}), then falls back to Postgres.
// On a cache miss the result is stored in Redis with a 1-hour TTL.
func (c *Cache) Get(ctx context.Context, workspaceID string) (*PlanLimits, error) {
	key := cache.KeyPlanLimits(workspaceID)

	// Try cache.
	val, err := c.redis.Get(ctx, key)
	if err == nil {
		var pl PlanLimits
		if jsonErr := json.Unmarshal([]byte(val), &pl); jsonErr == nil {
			return &pl, nil
		}
	} else if err != redis.Nil {
		// Redis error — fall through to Postgres.
	}

	// Fallback: load from Postgres.
	sub, err := c.store.GetByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("limits: get plan limits from store: %w", err)
	}

	planLimits := billing.GetPlanLimits(sub.Plan, c.apiQuotaPro)

	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

	pl := &PlanLimits{
		Plan:         planLimits.Plan,
		UserLimit:    planLimits.UserLimit,
		APICallLimit: planLimits.APICallLimit,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
	}

	// Re-cache for 1 hour.
	if data, jsonErr := json.Marshal(pl); jsonErr == nil {
		_ = c.redis.Set(ctx, key, string(data), time.Hour)
	}

	return pl, nil
}

// Invalidate removes the plan_limits cache entry for the given workspace.
func (c *Cache) Invalidate(ctx context.Context, workspaceID string) error {
	return c.redis.Del(ctx, cache.KeyPlanLimits(workspaceID))
}
