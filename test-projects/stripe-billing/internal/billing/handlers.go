package billing

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kabirnarang39/stripe-billing/internal/cache"
	"github.com/kabirnarang39/stripe-billing/internal/config"
	"github.com/kabirnarang39/stripe-billing/internal/httpx"
)

// RegisterRoutes mounts all billing HTTP handlers on the given router.
// All routes expect JWT middleware to have run upstream so that workspace_id
// and user_id are available in the request context.
func RegisterRoutes(r chi.Router, svc *Service, cacheClient *cache.Client, storeSvc StoreIface, cfg *config.Config) {
	h := &handler{
		svc:         svc,
		cacheClient: cacheClient,
		store:       storeSvc,
		cfg:         cfg,
	}

	r.Post("/subscribe", h.subscribe)
	r.Post("/downgrade", h.downgrade)
	r.Post("/cancel", h.cancel)
	r.Get("/subscription", h.getSubscription)
	r.Get("/usage", h.getUsage)
	r.Get("/invoices", h.listInvoices)
	r.Post("/portal", h.portal)
}

type handler struct {
	svc         *Service
	cacheClient *cache.Client
	store       StoreIface
	cfg         *config.Config
}

// POST /subscribe
// Body: {"payment_method_id":"pm_xxx","plan":"pro","email":"user@example.com"}
// Header: Idempotency-Key
func (h *handler) subscribe(w http.ResponseWriter, r *http.Request) {
	workspaceID, ok := httpx.WorkspaceIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized", "missing workspace_id in context")
		return
	}
	userID, _ := httpx.UserIDFromContext(r.Context())

	var body struct {
		PaymentMethodID string `json:"payment_method_id"`
		Plan            string `json:"plan"`
		Email           string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}

	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey != "" && userID != "" {
		cacheKey := cache.KeyIdempotency(userID, idempotencyKey)
		if cached := h.cacheClient.GetFailOpen(r.Context(), cacheKey); cached != "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(cached)) //nolint:errcheck
			return
		}
	}

	sub, err := h.svc.Subscribe(r.Context(), workspaceID, body.Email, body.PaymentMethodID, body.Plan)
	if err != nil {
		if errors.Is(err, ErrEnterpriseRequiresContact) {
			httpx.Error(w, http.StatusPaymentRequired, "enterprise_requires_contact", "contact sales for enterprise plans")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "internal_error", "failed to create subscription")
		return
	}

	resp := map[string]interface{}{
		"status":                 "pending",
		"stripe_subscription_id": sub.StripeSubscriptionID,
	}
	respBytes, _ := json.Marshal(resp)

	// Cache the response for 24 hours under the idempotency key.
	if idempotencyKey != "" && userID != "" {
		cacheKey := cache.KeyIdempotency(userID, idempotencyKey)
		_ = h.cacheClient.Set(r.Context(), cacheKey, string(respBytes), 24*time.Hour)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(respBytes) //nolint:errcheck
}

// POST /downgrade
func (h *handler) downgrade(w http.ResponseWriter, r *http.Request) {
	workspaceID, ok := httpx.WorkspaceIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized", "missing workspace_id in context")
		return
	}

	cancelAt, err := h.svc.Downgrade(r.Context(), workspaceID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal_error", "failed to downgrade subscription")
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]interface{}{
		"cancel_at": cancelAt.Format(time.RFC3339),
		"status":    "active",
	})
}

// POST /cancel?immediate=true
func (h *handler) cancel(w http.ResponseWriter, r *http.Request) {
	workspaceID, ok := httpx.WorkspaceIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized", "missing workspace_id in context")
		return
	}

	immediate := r.URL.Query().Get("immediate") == "true"
	if err := h.svc.Cancel(r.Context(), workspaceID, immediate); err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal_error", "failed to cancel subscription")
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]string{"status": "cancelling"})
}

// GET /subscription
func (h *handler) getSubscription(w http.ResponseWriter, r *http.Request) {
	workspaceID, ok := httpx.WorkspaceIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized", "missing workspace_id in context")
		return
	}

	sub, err := h.svc.GetSubscription(r.Context(), workspaceID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal_error", "failed to get subscription")
		return
	}

	httpx.JSON(w, http.StatusOK, sub)
}

// GET /usage
func (h *handler) getUsage(w http.ResponseWriter, r *http.Request) {
	workspaceID, ok := httpx.WorkspaceIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized", "missing workspace_id in context")
		return
	}

	now := time.Now().UTC()
	period := now.Format("2006-01")

	// Try Redis first for the current count.
	usageKey := cache.KeyUsage(workspaceID, period)
	var apiCallsUsed int64
	if val := h.cacheClient.GetFailOpen(r.Context(), usageKey); val != "" {
		apiCallsUsed, _ = strconv.ParseInt(val, 10, 64)
	}

	sub, err := h.svc.GetSubscription(r.Context(), workspaceID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal_error", "failed to get subscription")
		return
	}

	limits := GetPlanLimits(sub.Plan, h.cfg.APIQuotaPro)
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

	resp := map[string]interface{}{
		"plan":           sub.Plan,
		"api_calls_used": apiCallsUsed,
		"api_call_limit": limits.APICallLimit,
		"period_start":   periodStart.Format(time.RFC3339),
		"period_end":     periodEnd.Format(time.RFC3339),
	}
	httpx.JSON(w, http.StatusOK, resp)
}

// GET /invoices?page=1&per_page=20
func (h *handler) listInvoices(w http.ResponseWriter, r *http.Request) {
	workspaceID, ok := httpx.WorkspaceIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized", "missing workspace_id in context")
		return
	}

	page := 1
	perPage := 20
	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := r.URL.Query().Get("per_page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			perPage = n
		}
	}

	invoices, total, err := h.store.ListByWorkspace(r.Context(), workspaceID, page, perPage)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal_error", "failed to list invoices")
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]interface{}{
		"invoices": invoices,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

// POST /portal
// Body: {"return_url":"https://app.example.com/settings/billing"}
func (h *handler) portal(w http.ResponseWriter, r *http.Request) {
	workspaceID, ok := httpx.WorkspaceIDFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "unauthorized", "missing workspace_id in context")
		return
	}

	var body struct {
		ReturnURL string `json:"return_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "bad_request", "invalid JSON body")
		return
	}

	sub, err := h.svc.GetSubscription(r.Context(), workspaceID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal_error", "failed to get subscription")
		return
	}
	if sub.StripeCustomerID == nil {
		httpx.Error(w, http.StatusBadRequest, "no_customer", "workspace has no Stripe customer")
		return
	}

	session, err := h.svc.stripe.NewPortalSession(r.Context(), *sub.StripeCustomerID, body.ReturnURL)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "internal_error", "failed to create portal session")
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]string{"url": session.URL})
}
