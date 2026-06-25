package store

import "time"

// Subscription mirrors the subscriptions table.
type Subscription struct {
	ID                   string
	WorkspaceID          string
	Plan                 string
	Status               string // free, pending, active, past_due, cancel_at_period_end, cancelled
	StripeCustomerID     *string
	StripeSubscriptionID *string
	UserLimit            int
	APICallLimit         *int
	BillingPeriodStart   *time.Time
	BillingPeriodEnd     *time.Time
	CancelAtPeriodEnd    bool
	CancelAt             *time.Time
	PastDueSince         *time.Time
	CancelledAt          *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// Invoice mirrors the invoices table.
type Invoice struct {
	ID              string
	WorkspaceID     string
	StripeInvoiceID string
	AmountPaid      int64 // cents
	Currency        string
	Status          string
	PeriodStart     time.Time
	PeriodEnd       time.Time
	InvoicePDFURL   *string
	CreatedAt       time.Time
}

// APIUsage mirrors the api_usage table.
type APIUsage struct {
	ID                 int64
	WorkspaceID        string
	BillingPeriodStart time.Time
	BillingPeriodEnd   time.Time
	CallCount          int64
	UpdatedAt          time.Time
}

// WebhookEvent mirrors the webhook_events table (idempotency log).
type WebhookEvent struct {
	ID        string
	EventID   string
	EventType string
	CreatedAt time.Time
}
