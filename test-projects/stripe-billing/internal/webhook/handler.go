package webhook

import (
	"io"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v76/webhook"
)

// Handler is an http.Handler that processes Stripe webhook events.
type Handler struct {
	store    StoreIface
	cache    CacheIface
	dispatch *Dispatcher
	secret   string
}

// NewHandler creates a new webhook Handler.
func NewHandler(store StoreIface, cache CacheIface, secret string) *Handler {
	return &Handler{
		store:    store,
		cache:    cache,
		dispatch: NewDispatcher(store, cache),
		secret:   secret,
	}
}

// ServeHTTP handles POST /webhooks/stripe.
//
// Request processing:
//  1. Limit request body to 65 536 bytes to guard against large payloads.
//  2. Read the raw body (needed for signature verification).
//  3. Verify the Stripe-Signature header and construct the event.
//  4. Dispatch to the appropriate event handler.
//  5. Return 200 OK on success.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 65536)

	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("webhook: read body: %v", err)
		http.Error(w, "request body too large or unreadable", http.StatusBadRequest)
		return
	}

	event, err := webhook.ConstructEvent(rawBody, r.Header.Get("Stripe-Signature"), h.secret)
	if err != nil {
		log.Printf("webhook: construct event: %v", err)
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	if err := h.dispatch.Dispatch(r.Context(), event); err != nil {
		log.Printf("webhook: dispatch event %s type=%s: %v", event.ID, event.Type, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
