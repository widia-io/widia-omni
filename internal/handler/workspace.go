package handler

import (
	"net/http"

	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type WorkspaceHandler struct {
	wsSvc *service.WorkspaceService
}

func NewWorkspaceHandler(wsSvc *service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{wsSvc: wsSvc}
}

func (h *WorkspaceHandler) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	ws, err := h.wsSvc.GetWorkspace(r.Context(), wsID, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "workspace not found")
		return
	}

	writeJSON(w, http.StatusOK, ws)
}

func (h *WorkspaceHandler) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	role, _ := middleware.GetRole(r.Context())
	if !role.CanManage() {
		writeError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	var req service.UpdateWorkspaceRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ws, err := h.wsSvc.UpdateWorkspace(r.Context(), wsID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update workspace")
		return
	}

	writeJSON(w, http.StatusOK, ws)
}

func (h *WorkspaceHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	usage, err := h.wsSvc.GetUsage(r.Context(), wsID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get usage")
		return
	}

	writeJSON(w, http.StatusOK, usage)
}
