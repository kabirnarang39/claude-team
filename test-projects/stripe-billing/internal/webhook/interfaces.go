package webhook

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kabirnarang39/stripe-billing/internal/store"
)

// StoreIface is the store dependency interface for the webhook package.
type StoreIface interface {
	WithTx(ctx context.Context, fn func(pgx.Tx) error) error
	InsertEvent(ctx context.Context, tx pgx.Tx, eventID, eventType string) (bool, error)
	UpsertFromWebhook(ctx context.Context, tx pgx.Tx, sub *store.Subscription) error
	InsertFromWebhook(ctx context.Context, tx pgx.Tx, inv *store.Invoice) error
	UpsertUsage(ctx context.Context, workspaceID string, periodStart, periodEnd time.Time, delta int64) error
	GetByWorkspace(ctx context.Context, workspaceID string) (*store.Subscription, error)
}

// CacheIface is the cache dependency interface for the webhook package.
type CacheIface interface {
	Del(ctx context.Context, keys ...string) error
}
