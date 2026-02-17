package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/widia-io/widia-omni/internal/service"
)

type AdminHandler struct {
	adminSvc *service.AdminService
}

func NewAdminHandler(adminSvc *service.AdminService) *AdminHandler {
	return &AdminHandler{adminSvc: adminSvc}
}

func (h *AdminHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.adminSvc.GetMetrics(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get metrics")
		return
	}
	writeJSON(w, http.StatusOK, metrics)
}

func (h *AdminHandler) GetOnboardingFunnel(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	funnel, err := h.adminSvc.GetOnboardingFunnel(r.Context(), days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get onboarding funnel")
		return
	}
	writeJSON(w, http.StatusOK, funnel)
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	users, total, err := h.adminSvc.ListUsers(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"users": users,
		"total": total,
	})
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.adminSvc.GetUser(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *AdminHandler) GetWorkspaceUsage(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid workspace id")
		return
	}

	usage, err := h.adminSvc.GetWorkspaceUsage(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "workspace not found")
		return
	}
	writeJSON(w, http.StatusOK, usage)
}

func (h *AdminHandler) OverrideEntitlement(w http.ResponseWriter, r *http.Request) {
	var req service.OverrideEntitlementRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.WorkspaceID == uuid.Nil || req.Tier == "" || len(req.Limits) == 0 {
		writeError(w, http.StatusBadRequest, "workspace_id, tier, and limits are required")
		return
	}

	if err := h.adminSvc.OverrideEntitlement(r.Context(), req); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to override entitlement")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
