package handler

import (
	"net/http"

	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type ExportHandler struct {
	exportSvc *service.ExportService
}

func NewExportHandler(exportSvc *service.ExportService) *ExportHandler {
	return &ExportHandler{exportSvc: exportSvc}
}

func (h *ExportHandler) Export(w http.ResponseWriter, r *http.Request) {
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
	limits, ok := middleware.GetEntitlements(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "entitlements not loaded")
		return
	}
	if !limits.ExportEnabled {
		writeError(w, http.StatusForbidden, "export not available on your plan")
		return
	}

	data, err := h.exportSvc.Export(r.Context(), wsID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to export data")
		return
	}
	writeJSON(w, http.StatusOK, data)
}
