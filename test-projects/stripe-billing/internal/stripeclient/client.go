package stripeclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/client"
)

// Client wraps stripe's client.API with domain-specific helper methods.
type Client struct {
	sc *client.API
}

// New creates a new Client. It refuses non-test keys to prevent accidental
// live charges.
func New(secretKey string) (*Client, error) {
	if !strings.HasPrefix(secretKey, "sk_test_") {
		return nil, fmt.Errorf("stripeclient: refusing non-test key — key must start with sk_test_")
	}
	sc := &client.API{}
	sc.Init(secretKey, nil)
	return &Client{sc: sc}, nil
}

// EnsureCustomer looks up a Stripe customer by workspace_id metadata. If none
// exists it creates one and attaches the payment method.
func (c *Client) EnsureCustomer(ctx context.Context, workspaceID, email, paymentMethodID string) (string, error) {
	// Search for an existing customer by metadata.
	searchParams := &stripe.CustomerSearchParams{
		SearchParams: stripe.SearchParams{
			Query: fmt.Sprintf("metadata['workspace_id']:'%s'", workspaceID),
		},
	}
	iter := c.sc.Customers.Search(searchParams)
	for iter.Next() {
		return iter.Customer().ID, nil
	}
	if err := iter.Err(); err != nil {
		return "", fmt.Errorf("stripeclient: search customer: %w", err)
	}

	// No existing customer — create one.
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Metadata: map[string]string{
			"workspace_id": workspaceID,
		},
	}
	if paymentMethodID != "" {
		params.PaymentMethod = stripe.String(paymentMethodID)
		params.InvoiceSettings = &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethodID),
		}
	}
	cust, err := c.sc.Customers.New(params)
	if err != nil {
		return "", fmt.Errorf("stripeclient: create customer: %w", err)
	}
	return cust.ID, nil
}

// CreateProSubscription creates a Stripe subscription for the given customer,
// price, and payment method. It sets the subscription metadata with workspace_id
// (extracted from customer metadata lookup is not needed here — caller provides it
// via the payment flow; the workspace_id should be attached to the customer).
func (c *Client) CreateProSubscription(ctx context.Context, customerID, priceID, pmID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{Price: stripe.String(priceID)},
		},
		DefaultPaymentMethod: stripe.String(pmID),
		PaymentBehavior:      stripe.String("default_incomplete"),
		Expand:               []*string{stripe.String("latest_invoice.payment_intent")},
	}
	sub, err := c.sc.Subscriptions.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripeclient: create subscription: %w", err)
	}
	return sub, nil
}

// CancelAtPeriodEnd sets cancel_at_period_end=true on the Stripe subscription.
func (c *Client) CancelAtPeriodEnd(ctx context.Context, subID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}
	sub, err := c.sc.Subscriptions.Update(subID, params)
	if err != nil {
		return nil, fmt.Errorf("stripeclient: cancel at period end: %w", err)
	}
	return sub, nil
}

// CancelNow immediately cancels the Stripe subscription.
func (c *Client) CancelNow(ctx context.Context, subID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionCancelParams{}
	sub, err := c.sc.Subscriptions.Cancel(subID, params)
	if err != nil {
		return nil, fmt.Errorf("stripeclient: cancel now: %w", err)
	}
	return sub, nil
}

// NewSetupIntent creates a Stripe SetupIntent for adding a payment method.
func (c *Client) NewSetupIntent(ctx context.Context, customerID string) (*stripe.SetupIntent, error) {
	params := &stripe.SetupIntentParams{
		Customer:           stripe.String(customerID),
		PaymentMethodTypes: []*string{stripe.String("card")},
	}
	si, err := c.sc.SetupIntents.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripeclient: create setup intent: %w", err)
	}
	return si, nil
}

// NewPortalSession creates a Stripe Billing Portal session for the customer.
func (c *Client) NewPortalSession(ctx context.Context, customerID, returnURL string) (*stripe.BillingPortalSession, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}
	session, err := c.sc.BillingPortalSessions.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripeclient: create portal session: %w", err)
	}
	return session, nil
}
