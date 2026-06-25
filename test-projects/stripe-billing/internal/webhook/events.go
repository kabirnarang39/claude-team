package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kabirnarang39/stripe-billing/internal/cache"
	"github.com/kabirnarang39/stripe-billing/internal/store"
	"github.com/stripe/stripe-go/v76"
)

// handlePaymentSucceeded handles invoice.payment_succeeded.
//
// Steps:
//  1. Unmarshal the Stripe Invoice from the event payload.
//  2. Extract workspace_id from subscription metadata.
//  3. Open a database transaction.
//  4. Insert idempotency record — if the event was already processed, return nil.
//  5. Upsert the subscription as active with the new billing period.
//  6. Insert the invoice record.
//  7. Upsert (reset) the API usage counter for the new billing period (delta=0).
//  8. Commit.
//  9. Invalidate the plan_limits cache key.
func (d *Dispatcher) handlePaymentSucceeded(ctx context.Context, event stripe.Event) error {
	var inv stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		return fmt.Errorf("webhook: payment_succeeded: unmarshal invoice: %w", err)
	}

	workspaceID, ok := workspaceIDFromInvoice(&inv)
	if !ok {
		log.Printf("webhook: payment_succeeded event %s: missing workspace_id in metadata — skipping", event.ID)
		return nil
	}

	periodStart := time.Unix(inv.PeriodStart, 0).UTC()
	periodEnd := time.Unix(inv.PeriodEnd, 0).UTC()

	// Determine plan limits for Pro.
	userLimit, apiCallLimit := planLimitsForEvent()

	var invoicePDFURL *string
	if inv.InvoicePDF != "" {
		u := inv.InvoicePDF
		invoicePDFURL = &u
	}

	// Customer ID from the invoice.
	var customerID *string
	if inv.Customer != nil && inv.Customer.ID != "" {
		id := inv.Customer.ID
		customerID = &id
	}

	// Subscription ID.
	var stripeSubID *string
	if inv.Subscription != nil && inv.Subscription.ID != "" {
		id := inv.Subscription.ID
		stripeSubID = &id
	}

	txErr := d.store.WithTx(ctx, func(tx pgx.Tx) error {
		inserted, err := d.store.InsertEvent(ctx, tx, event.ID, string(event.Type))
		if err != nil {
			return fmt.Errorf("insert event: %w", err)
		}
		if !inserted {
			return nil // duplicate delivery
		}

		sub := &store.Subscription{
			WorkspaceID:          workspaceID,
			Plan:                 "pro",
			Status:               "active",
			StripeCustomerID:     customerID,
			StripeSubscriptionID: stripeSubID,
			UserLimit:            userLimit,
			APICallLimit:         apiCallLimit,
			BillingPeriodStart:   &periodStart,
			BillingPeriodEnd:     &periodEnd,
		}
		if err := d.store.UpsertFromWebhook(ctx, tx, sub); err != nil {
			return fmt.Errorf("upsert subscription: %w", err)
		}

		currency := string(inv.Currency)
		storeInv := &store.Invoice{
			WorkspaceID:     workspaceID,
			StripeInvoiceID: inv.ID,
			AmountPaid:      inv.AmountPaid,
			Currency:        currency,
			Status:          string(inv.Status),
			PeriodStart:     periodStart,
			PeriodEnd:       periodEnd,
			InvoicePDFURL:   invoicePDFURL,
		}
		if err := d.store.InsertFromWebhook(ctx, tx, storeInv); err != nil {
			return fmt.Errorf("insert invoice: %w", err)
		}

		// delta=0 creates the usage row at 0 for the new period (upsert semantics).
		if err := d.store.UpsertUsage(ctx, workspaceID, periodStart, periodEnd, 0); err != nil {
			return fmt.Errorf("upsert usage: %w", err)
		}

		return nil
	})
	if txErr != nil {
		return fmt.Errorf("webhook: payment_succeeded tx: %w", txErr)
	}

	// Invalidate plan_limits cache outside the transaction.
	if err := d.cache.Del(ctx, cache.KeyPlanLimits(workspaceID)); err != nil {
		log.Printf("webhook: payment_succeeded: invalidate cache: %v", err)
	}

	return nil
}

// handlePaymentFailed handles invoice.payment_failed.
//
// Transitions subscription status to past_due and sets past_due_since=NOW().
func (d *Dispatcher) handlePaymentFailed(ctx context.Context, event stripe.Event) error {
	var inv stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		return fmt.Errorf("webhook: payment_failed: unmarshal invoice: %w", err)
	}

	workspaceID, ok := workspaceIDFromInvoice(&inv)
	if !ok {
		log.Printf("webhook: payment_failed event %s: missing workspace_id — skipping", event.ID)
		return nil
	}

	now := time.Now().UTC()

	var customerID *string
	if inv.Customer != nil && inv.Customer.ID != "" {
		id := inv.Customer.ID
		customerID = &id
	}
	var stripeSubID *string
	if inv.Subscription != nil && inv.Subscription.ID != "" {
		id := inv.Subscription.ID
		stripeSubID = &id
	}

	txErr := d.store.WithTx(ctx, func(tx pgx.Tx) error {
		inserted, err := d.store.InsertEvent(ctx, tx, event.ID, string(event.Type))
		if err != nil {
			return fmt.Errorf("insert event: %w", err)
		}
		if !inserted {
			return nil // duplicate delivery
		}

		sub := &store.Subscription{
			WorkspaceID:          workspaceID,
			Plan:                 "pro",
			Status:               "past_due",
			StripeCustomerID:     customerID,
			StripeSubscriptionID: stripeSubID,
			PastDueSince:         &now,
		}
		if err := d.store.UpsertFromWebhook(ctx, tx, sub); err != nil {
			return fmt.Errorf("upsert subscription: %w", err)
		}
		return nil
	})
	if txErr != nil {
		return fmt.Errorf("webhook: payment_failed tx: %w", txErr)
	}

	if err := d.cache.Del(ctx, cache.KeyPlanLimits(workspaceID)); err != nil {
		log.Printf("webhook: payment_failed: invalidate cache: %v", err)
	}

	return nil
}

// handleSubscriptionDeleted handles customer.subscription.deleted.
//
// Transitions subscription to cancelled and resets limits to free tier.
func (d *Dispatcher) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	var stripeSub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
		return fmt.Errorf("webhook: subscription_deleted: unmarshal subscription: %w", err)
	}

	workspaceID, ok := workspaceIDFromSub(&stripeSub)
	if !ok {
		log.Printf("webhook: subscription_deleted event %s: missing workspace_id — skipping", event.ID)
		return nil
	}

	now := time.Now().UTC()
	freePlanUserLimit := 5

	var customerID *string
	if stripeSub.Customer != nil && stripeSub.Customer.ID != "" {
		id := stripeSub.Customer.ID
		customerID = &id
	}
	subID := stripeSub.ID

	txErr := d.store.WithTx(ctx, func(tx pgx.Tx) error {
		inserted, err := d.store.InsertEvent(ctx, tx, event.ID, string(event.Type))
		if err != nil {
			return fmt.Errorf("insert event: %w", err)
		}
		if !inserted {
			return nil // duplicate delivery
		}

		sub := &store.Subscription{
			WorkspaceID:          workspaceID,
			Plan:                 "free",
			Status:               "cancelled",
			StripeCustomerID:     customerID,
			StripeSubscriptionID: &subID,
			UserLimit:            freePlanUserLimit,
			APICallLimit:         nil, // free plan has no API call quota
			CancelledAt:          &now,
		}
		if err := d.store.UpsertFromWebhook(ctx, tx, sub); err != nil {
			return fmt.Errorf("upsert subscription: %w", err)
		}
		return nil
	})
	if txErr != nil {
		return fmt.Errorf("webhook: subscription_deleted tx: %w", txErr)
	}

	// Invalidate plan limits and usage cache.
	now2 := time.Now().UTC()
	period := now2.Format("2006-01")
	if err := d.cache.Del(ctx,
		cache.KeyPlanLimits(workspaceID),
		cache.KeyUsage(workspaceID, period),
	); err != nil {
		log.Printf("webhook: subscription_deleted: invalidate cache: %v", err)
	}

	return nil
}
