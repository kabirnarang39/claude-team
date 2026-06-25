package billing

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kabirnarang39/stripe-billing/internal/store"
	"github.com/stripe/stripe-go/v76"
)

// StoreIface is the store dependency interface for the billing package.
// Defining it here avoids import cycles between billing and store.
type StoreIface interface {
	GetByWorkspace(ctx context.Context, workspaceID string) (*store.Subscription, error)
	UpsertFromWebhook(ctx context.Context, tx pgx.Tx, sub *store.Subscription) error
	CreateFree(ctx context.Context, workspaceID string) (*store.Subscription, error)
	SetCancelAtPeriodEnd(ctx context.Context, workspaceID string, cancelAt time.Time) error
	SetPending(ctx context.Context, workspaceID, customerID, subscriptionID string) error
	InsertFromWebhook(ctx context.Context, tx pgx.Tx, inv *store.Invoice) error
	ListByWorkspace(ctx context.Context, workspaceID string, page, perPage int) ([]*store.Invoice, int, error)
	InsertEvent(ctx context.Context, tx pgx.Tx, eventID, eventType string) (bool, error)
	UpsertUsage(ctx context.Context, workspaceID string, periodStart, periodEnd time.Time, delta int64) error
	GetUsage(ctx context.Context, workspaceID string, periodStart time.Time) (*store.APIUsage, error)
	WithTx(ctx context.Context, fn func(pgx.Tx) error) error
}

// StripeIface is the Stripe client dependency interface for the billing package.
type StripeIface interface {
	EnsureCustomer(ctx context.Context, workspaceID, email, paymentMethodID string) (string, error)
	CreateProSubscription(ctx context.Context, customerID, priceID, pmID string) (*stripe.Subscription, error)
	CancelAtPeriodEnd(ctx context.Context, subID string) (*stripe.Subscription, error)
	CancelNow(ctx context.Context, subID string) (*stripe.Subscription, error)
	NewPortalSession(ctx context.Context, customerID, returnURL string) (*stripe.BillingPortalSession, error)
}

// CacheIface is the cache dependency interface for the billing package.
type CacheIface interface {
	Del(ctx context.Context, keys ...string) error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error)
}
