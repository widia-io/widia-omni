package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type ReferralHandler struct {
	refSvc *service.ReferralService
}

func NewReferralHandler(refSvc *service.ReferralService) *ReferralHandler {
	return &ReferralHandler{refSvc: refSvc}
}

func (h *ReferralHandler) referralGate(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	limits, ok := middleware.GetEntitlements(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "entitlements not loaded")
		return uuid.Nil, false
	}
	if !limits.ReferralEnabled {
		writeError(w, http.StatusForbidden, "referral not available on your plan")
		return uuid.Nil, false
	}
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return uuid.Nil, false
	}
	return wsID, true
}

func (h *ReferralHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	wsID, ok := h.referralGate(w, r)
	if !ok {
		return
	}

	me, err := h.refSvc.GetMe(r.Context(), wsID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load referral data")
		return
	}
	writeJSON(w, http.StatusOK, me)
}

func (h *ReferralHandler) Regenerate(w http.ResponseWriter, r *http.Request) {
	wsID, ok := h.referralGate(w, r)
	if !ok {
		return
	}

	role, _ := middleware.GetRole(r.Context())
	if !role.CanManage() {
		writeError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	code, err := h.refSvc.RegenerateCode(r.Context(), wsID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to regenerate code")
		return
	}

	writeJSON(w, http.StatusOK, code)
}

func (h *ReferralHandler) ListAttributions(w http.ResponseWriter, r *http.Request) {
	wsID, ok := h.referralGate(w, r)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	items, err := h.refSvc.ListAttributions(r.Context(), wsID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list attributions")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *ReferralHandler) ListCredits(w http.ResponseWriter, r *http.Request) {
	wsID, ok := h.referralGate(w, r)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	items, err := h.refSvc.ListCredits(r.Context(), wsID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list credits")
		return
	}
	writeJSON(w, http.StatusOK, items)
}
