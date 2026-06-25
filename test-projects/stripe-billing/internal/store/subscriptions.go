package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// GetByWorkspace returns the subscription for the given workspace ID.
// Returns pgx.ErrNoRows (wrapped) if none exists.
func (s *Store) GetByWorkspace(ctx context.Context, workspaceID string) (*Subscription, error) {
	const q = `
		SELECT id, workspace_id, plan, status,
		       stripe_subscription_id, stripe_customer_id,
		       user_limit, api_call_limit,
		       current_period_start, current_period_end,
		       cancel_at_period_end, cancel_at, past_due_since, cancelled_at,
		       created_at, updated_at
		FROM subscriptions
		WHERE workspace_id = $1
		LIMIT 1`

	sub := &Subscription{}
	err := s.pool.QueryRow(ctx, q, workspaceID).Scan(
		&sub.ID,
		&sub.WorkspaceID,
		&sub.Plan,
		&sub.Status,
		&sub.StripeSubscriptionID,
		&sub.StripeCustomerID,
		&sub.UserLimit,
		&sub.APICallLimit,
		&sub.BillingPeriodStart,
		&sub.BillingPeriodEnd,
		&sub.CancelAtPeriodEnd,
		&sub.CancelAt,
		&sub.PastDueSince,
		&sub.CancelledAt,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("store: subscription not found for workspace %s: %w", workspaceID, pgx.ErrNoRows)
		}
		return nil, fmt.Errorf("store: get subscription: %w", err)
	}
	return sub, nil
}

// UpsertFromWebhook inserts or updates a subscription record using the Stripe
// subscription ID as the conflict target. It must be called inside a transaction
// because webhook processing requires atomic event-dedup + data update.
// This is the ONLY writer of confirmed statuses (active, past_due, cancelled).
func (s *Store) UpsertFromWebhook(ctx context.Context, tx pgx.Tx, sub *Subscription) error {
	const q = `
		INSERT INTO subscriptions (
			workspace_id, plan, status,
			stripe_subscription_id, stripe_customer_id,
			user_limit, api_call_limit,
			current_period_start, current_period_end,
			cancel_at_period_end, cancel_at, past_due_since, cancelled_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW())
		ON CONFLICT (workspace_id) DO UPDATE SET
			plan                  = EXCLUDED.plan,
			status                = EXCLUDED.status,
			stripe_subscription_id= EXCLUDED.stripe_subscription_id,
			stripe_customer_id    = EXCLUDED.stripe_customer_id,
			user_limit            = EXCLUDED.user_limit,
			api_call_limit        = EXCLUDED.api_call_limit,
			current_period_start  = EXCLUDED.current_period_start,
			current_period_end    = EXCLUDED.current_period_end,
			cancel_at_period_end  = EXCLUDED.cancel_at_period_end,
			cancel_at             = EXCLUDED.cancel_at,
			past_due_since        = EXCLUDED.past_due_since,
			cancelled_at          = EXCLUDED.cancelled_at,
			updated_at            = NOW()`

	_, err := tx.Exec(ctx, q,
		sub.WorkspaceID,
		sub.Plan,
		sub.Status,
		sub.StripeSubscriptionID,
		sub.StripeCustomerID,
		sub.UserLimit,
		sub.APICallLimit,
		sub.BillingPeriodStart,
		sub.BillingPeriodEnd,
		sub.CancelAtPeriodEnd,
		sub.CancelAt,
		sub.PastDueSince,
		sub.CancelledAt,
	)
	if err != nil {
		return fmt.Errorf("store: upsert subscription: %w", err)
	}
	return nil
}

// CreateFree inserts a free-tier subscription for a newly created workspace.
// No Stripe IDs are set at this point.
func (s *Store) CreateFree(ctx context.Context, workspaceID string) (*Subscription, error) {
	const q = `
		INSERT INTO subscriptions (workspace_id, plan, status, user_limit)
		VALUES ($1, 'free', 'active', 5)
		RETURNING id, workspace_id, plan, status,
		          stripe_subscription_id, stripe_customer_id,
		          user_limit, api_call_limit,
		          current_period_start, current_period_end,
		          cancel_at_period_end, cancel_at, past_due_since, cancelled_at,
		          created_at, updated_at`

	sub := &Subscription{}
	err := s.pool.QueryRow(ctx, q, workspaceID).Scan(
		&sub.ID,
		&sub.WorkspaceID,
		&sub.Plan,
		&sub.Status,
		&sub.StripeSubscriptionID,
		&sub.StripeCustomerID,
		&sub.UserLimit,
		&sub.APICallLimit,
		&sub.BillingPeriodStart,
		&sub.BillingPeriodEnd,
		&sub.CancelAtPeriodEnd,
		&sub.CancelAt,
		&sub.PastDueSince,
		&sub.CancelledAt,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("store: create free subscription: %w", err)
	}
	return sub, nil
}

// SetPending transitions the subscription to pending status after the API layer
// creates a Stripe subscription but before the webhook confirms payment.
func (s *Store) SetPending(ctx context.Context, workspaceID, customerID, subscriptionID string) error {
	const q = `
		UPDATE subscriptions
		SET status                  = 'pending',
		    stripe_customer_id      = $2,
		    stripe_subscription_id  = $3,
		    updated_at              = NOW()
		WHERE workspace_id = $1`

	tag, err := s.pool.Exec(ctx, q, workspaceID, customerID, subscriptionID)
	if err != nil {
		return fmt.Errorf("store: set pending: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("store: set pending: no subscription found for workspace %s", workspaceID)
	}
	return nil
}

// SetCancelAtPeriodEnd mirrors the cancel_at_period_end flag and cancel_at
// timestamp after an API-initiated cancellation request (not a webhook event).
func (s *Store) SetCancelAtPeriodEnd(ctx context.Context, workspaceID string, cancelAt time.Time) error {
	const q = `
		UPDATE subscriptions
		SET cancel_at_period_end = TRUE,
		    cancel_at            = $2,
		    updated_at           = NOW()
		WHERE workspace_id = $1`

	tag, err := s.pool.Exec(ctx, q, workspaceID, cancelAt)
	if err != nil {
		return fmt.Errorf("store: set cancel_at_period_end: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("store: set cancel_at_period_end: no subscription found for workspace %s", workspaceID)
	}
	return nil
}
