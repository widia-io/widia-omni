package handler

import (
	"net/http"

	"github.com/widia-io/widia-omni/internal/service"
)

type StripeWebhookHandler struct {
	billingSvc *service.BillingService
}

func NewStripeWebhookHandler(billingSvc *service.BillingService) *StripeWebhookHandler {
	return &StripeWebhookHandler{billingSvc: billingSvc}
}

func (h *StripeWebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	sig := r.Header.Get("Stripe-Signature")
	if sig == "" {
		writeError(w, http.StatusBadRequest, "missing stripe signature")
		return
	}

	if err := h.billingSvc.HandleWebhook(r.Context(), r.Body, sig); err != nil {
		writeError(w, http.StatusBadRequest, "webhook processing failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
