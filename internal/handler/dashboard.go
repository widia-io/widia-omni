package handler

import (
	"net/http"

	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type DashboardHandler struct {
	dashSvc *service.DashboardService
}

func NewDashboardHandler(dashSvc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashSvc: dashSvc}
}

func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "user not found")
		return
	}

	data, err := h.dashSvc.GetDashboard(r.Context(), wsID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get dashboard")
		return
	}
	writeJSON(w, http.StatusOK, data)
}
