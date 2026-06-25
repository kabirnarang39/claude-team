package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kabirnarang39/stripe-billing/internal/config"
	"github.com/kabirnarang39/stripe-billing/internal/store"
)

// ErrEnterpriseRequiresContact is returned when a user tries to self-serve
// subscribe to the Enterprise plan. Maps to HTTP 402.
var ErrEnterpriseRequiresContact = errors.New("enterprise_requires_contact")

// Service orchestrates subscription lifecycle by coordinating the store,
// Stripe client, and cache.
type Service struct {
	store  StoreIface
	stripe StripeIface
	cache  CacheIface
	cfg    *config.Config
}

// NewService constructs a Service.
func NewService(store StoreIface, stripe StripeIface, cache CacheIface, cfg *config.Config) *Service {
	return &Service{
		store:  store,
		stripe: stripe,
		cache:  cache,
		cfg:    cfg,
	}
}

// Subscribe initiates a Pro subscription. It calls Stripe, writes a pending
// status to the store (API-layer write), and returns the pending subscription.
// The webhook handler is the sole writer of confirmed statuses (active/cancelled).
//
// Returns ErrEnterpriseRequiresContact if plan == "enterprise".
func (s *Service) Subscribe(ctx context.Context, workspaceID, email, paymentMethodID, plan string) (*store.Subscription, error) {
	if plan == PlanEnterprise {
		return nil, ErrEnterpriseRequiresContact
	}
	if plan != PlanPro {
		return nil, fmt.Errorf("billing: unsupported plan %q; only %q is self-serve subscribable", plan, PlanPro)
	}

	// Ensure a Stripe customer exists (idempotent).
	customerID, err := s.stripe.EnsureCustomer(ctx, workspaceID, email, paymentMethodID)
	if err != nil {
		return nil, fmt.Errorf("billing: ensure customer: %w", err)
	}

	// Create the Stripe subscription.
	stripeSub, err := s.stripe.CreateProSubscription(ctx, customerID, s.cfg.ProPriceID, paymentMethodID)
	if err != nil {
		return nil, fmt.Errorf("billing: create stripe subscription: %w", err)
	}

	// Write pending status — API-layer write only. Confirmed status comes via webhook.
	if err := s.store.SetPending(ctx, workspaceID, customerID, stripeSub.ID); err != nil {
		return nil, fmt.Errorf("billing: set pending: %w", err)
	}

	// Invalidate plan limits cache.
	_ = s.cache.Del(ctx, "plan_limits:"+workspaceID)

	return s.store.GetByWorkspace(ctx, workspaceID)
}

// Downgrade sets cancel_at_period_end=true on Stripe and mirrors the flag to
// the store. The subscription stays active until the period end; the webhook
// will confirm cancellation.
func (s *Service) Downgrade(ctx context.Context, workspaceID string) (cancelAt time.Time, err error) {
	sub, err := s.store.GetByWorkspace(ctx, workspaceID)
	if err != nil {
		return time.Time{}, fmt.Errorf("billing: downgrade: get subscription: %w", err)
	}
	if sub.StripeSubscriptionID == nil {
		return time.Time{}, fmt.Errorf("billing: downgrade: workspace %s has no Stripe subscription", workspaceID)
	}

	stripeSub, err := s.stripe.CancelAtPeriodEnd(ctx, *sub.StripeSubscriptionID)
	if err != nil {
		return time.Time{}, fmt.Errorf("billing: downgrade: stripe cancel at period end: %w", err)
	}

	cancelAt = time.Unix(stripeSub.CancelAt, 0).UTC()
	if err := s.store.SetCancelAtPeriodEnd(ctx, workspaceID, cancelAt); err != nil {
		return time.Time{}, fmt.Errorf("billing: downgrade: store update: %w", err)
	}

	return cancelAt, nil
}

// Cancel cancels the subscription. If immediate is true it cancels right away;
// otherwise it sets cancel_at_period_end (same as Downgrade).
func (s *Service) Cancel(ctx context.Context, workspaceID string, immediate bool) error {
	sub, err := s.store.GetByWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("billing: cancel: get subscription: %w", err)
	}
	if sub.StripeSubscriptionID == nil {
		return fmt.Errorf("billing: cancel: workspace %s has no Stripe subscription", workspaceID)
	}

	if immediate {
		_, err = s.stripe.CancelNow(ctx, *sub.StripeSubscriptionID)
		if err != nil {
			return fmt.Errorf("billing: cancel now: %w", err)
		}
		// Webhook customer.subscription.deleted will write confirmed cancelled status.
	} else {
		stripeSub, err := s.stripe.CancelAtPeriodEnd(ctx, *sub.StripeSubscriptionID)
		if err != nil {
			return fmt.Errorf("billing: cancel at period end: %w", err)
		}
		cancelAt := time.Unix(stripeSub.CancelAt, 0).UTC()
		if err := s.store.SetCancelAtPeriodEnd(ctx, workspaceID, cancelAt); err != nil {
			return fmt.Errorf("billing: cancel: store update: %w", err)
		}
	}
	return nil
}

// GetSubscription returns the current subscription for a workspace.
func (s *Service) GetSubscription(ctx context.Context, workspaceID string) (*store.Subscription, error) {
	return s.store.GetByWorkspace(ctx, workspaceID)
}

// CreateFree creates a Free subscription for a newly registered workspace.
// No Stripe objects are created.
func (s *Service) CreateFree(ctx context.Context, workspaceID string) (*store.Subscription, error) {
	return s.store.CreateFree(ctx, workspaceID)
}
