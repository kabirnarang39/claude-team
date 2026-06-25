package usage

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/kabirnarang39/stripe-billing/internal/cache"
)

// flushItem carries the workspace and billing period information for a usage
// notification from the quota middleware.
type flushItem struct {
	WorkspaceID        string
	BillingPeriodStart time.Time
	BillingPeriodEnd   time.Time
}

// Flusher accumulates Redis-counted API calls and periodically (or on threshold)
// flushes the delta to Postgres so that invoices and dashboards stay accurate.
//
// Design notes:
//   - Redis is the primary counter (fast, fail-open).
//   - Postgres is the source of truth for billing; Flusher keeps it in sync.
//   - A threshold-based flush prevents unbounded Redis-only counts.
type Flusher struct {
	store     StoreIface
	cache     *cache.Client
	counters  map[string]int64 // key: "workspaceID:YYYY-MM"
	mu        sync.Mutex
	threshold int64
	ticker    *time.Ticker
	ch        chan flushItem
}

// NewFlusher creates a Flusher. Call Start to begin processing.
func NewFlusher(store StoreIface, cacheClient *cache.Client) *Flusher {
	return &Flusher{
		store:     store,
		cache:     cacheClient,
		counters:  make(map[string]int64),
		threshold: 1000,
		ticker:    time.NewTicker(30 * time.Second),
		ch:        make(chan flushItem, 4096),
	}
}

// Start begins the background goroutine that drains the notification channel
// and periodically flushes to Postgres. It returns when ctx is cancelled.
func (f *Flusher) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			f.flush(context.Background()) // final drain
			return
		case item := <-f.ch:
			f.mu.Lock()
			k := item.WorkspaceID + ":" + item.BillingPeriodStart.Format("2006-01")
			f.counters[k]++
			shouldFlush := f.counters[k]%f.threshold == 0
			f.mu.Unlock()
			if shouldFlush {
				f.flush(ctx)
			}
		case <-f.ticker.C:
			f.flush(ctx)
		}
	}
}

// Notify enqueues a flush notification from the middleware. Non-blocking: drops
// if the channel is full rather than blocking the request path.
func (f *Flusher) Notify(item flushItem) {
	select {
	case f.ch <- item:
	default:
		// Channel full; drop notification. The ticker will catch up.
	}
}

// flush reads current Redis counts and upserts deltas to Postgres.
func (f *Flusher) flush(ctx context.Context) {
	f.mu.Lock()
	counters := f.counters
	f.counters = make(map[string]int64)
	f.mu.Unlock()

	if len(counters) == 0 {
		return
	}

	for k, delta := range counters {
		workspaceID, period, ok := parseCounterKey(k)
		if !ok {
			log.Printf("flusher: malformed key %q — skipping", k)
			continue
		}

		periodStart := periodStartFromPeriod(period)
		periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

		// Read current Redis counter to sync; if unavailable use the accumulated delta.
		redisKey := cache.KeyUsage(workspaceID, period)
		if v := f.cache.GetFailOpen(ctx, redisKey); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				delta = n
			}
		}

		if err := f.store.UpsertUsage(ctx, workspaceID, periodStart, periodEnd, delta); err != nil {
			log.Printf("flusher: upsert usage workspace=%s period=%s delta=%d: %v",
				workspaceID, period, delta, err)
		}
	}
}

// parseCounterKey splits "workspaceID:YYYY-MM" into its parts.
func parseCounterKey(k string) (workspaceID, period string, ok bool) {
	// workspace IDs are UUIDs (36 chars) + ":" + "YYYY-MM" (7 chars)
	if len(k) < 9 {
		return "", "", false
	}
	idx := len(k) - 8 // ":" + 7
	if k[idx] != ':' {
		return "", "", false
	}
	return k[:idx], k[idx+1:], true
}

// periodStartFromPeriod parses "YYYY-MM" and returns the first instant of that month.
func periodStartFromPeriod(period string) time.Time {
	t, err := time.Parse("2006-01", period)
	if err != nil {
		return time.Time{}
	}
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

