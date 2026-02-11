package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type BillingHandler struct {
	billingSvc *service.BillingService
}

func NewBillingHandler(billingSvc *service.BillingService) *BillingHandler {
	return &BillingHandler{billingSvc: billingSvc}
}

func (h *BillingHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.billingSvc.ListPlans(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list plans")
		return
	}
	writeJSON(w, http.StatusOK, plans)
}

func (h *BillingHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	sub, err := h.billingSvc.GetSubscription(r.Context(), wsID)
	if err != nil {
		writeError(w, http.StatusNotFound, "subscription not found")
		return
	}
	writeJSON(w, http.StatusOK, sub)
}

func (h *BillingHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	var req struct {
		Tier     string `json:"tier"`
		Interval string `json:"interval"`
	}
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Tier = strings.ToLower(strings.TrimSpace(req.Tier))
	req.Interval = strings.ToLower(strings.TrimSpace(req.Interval))
	if req.Tier != "pro" && req.Tier != "premium" {
		writeError(w, http.StatusBadRequest, "invalid tier")
		return
	}
	if req.Interval != "monthly" && req.Interval != "yearly" {
		writeError(w, http.StatusBadRequest, "invalid interval")
		return
	}

	url, err := h.billingSvc.CreateCheckoutSession(r.Context(), wsID, req.Tier, req.Interval)
	if err != nil {
		if errors.Is(err, service.ErrBillingPlanUnavailable) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create checkout session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

func (h *BillingHandler) CreatePortal(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	url, err := h.billingSvc.CreatePortalSession(r.Context(), wsID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create portal session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}
