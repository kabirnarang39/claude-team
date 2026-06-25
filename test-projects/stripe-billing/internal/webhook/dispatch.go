package webhook

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v76"
)

// Dispatcher routes Stripe events to their specific handlers.
type Dispatcher struct {
	store StoreIface
	cache CacheIface
}

// NewDispatcher creates a new Dispatcher.
func NewDispatcher(store StoreIface, cache CacheIface) *Dispatcher {
	return &Dispatcher{store: store, cache: cache}
}

// Dispatch routes the given Stripe event to the appropriate handler.
// Unknown event types return nil (the caller responds 200 to acknowledge receipt).
func (d *Dispatcher) Dispatch(ctx context.Context, event stripe.Event) error {
	switch event.Type {
	case "invoice.payment_succeeded":
		return d.handlePaymentSucceeded(ctx, event)
	case "invoice.payment_failed":
		return d.handlePaymentFailed(ctx, event)
	case "customer.subscription.deleted":
		return d.handleSubscriptionDeleted(ctx, event)
	default:
		// Acknowledge but do not process unknown events.
		return nil
	}
}

// workspaceIDFromInvoice extracts the workspace_id from invoice subscription metadata.
func workspaceIDFromInvoice(inv *stripe.Invoice) (string, bool) {
	if inv.SubscriptionDetails != nil {
		if id, ok := inv.SubscriptionDetails.Metadata["workspace_id"]; ok && id != "" {
			return id, true
		}
	}
	if inv.Subscription != nil {
		if id, ok := inv.Subscription.Metadata["workspace_id"]; ok && id != "" {
			return id, true
		}
	}
	return "", false
}

// workspaceIDFromSub extracts the workspace_id from a Stripe subscription's metadata.
func workspaceIDFromSub(sub *stripe.Subscription) (string, bool) {
	if id, ok := sub.Metadata["workspace_id"]; ok && id != "" {
		return id, true
	}
	return "", false
}

// planLimitsForEvent returns safe defaults for a Pro subscription activated by webhook.
func planLimitsForEvent() (userLimit int, apiCallLimit *int) {
	q := 10000 // conservative default; real limits come from config via billing.GetPlanLimits
	return 50, &q
}

// missingWorkspaceID logs a warning and returns fmt.Errorf with a sentinel context.
func missingWorkspaceID(eventID, eventType string) error {
	return fmt.Errorf("webhook: event %s (%s): missing workspace_id in metadata — skipping", eventID, eventType)
}
