package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type AreaHandler struct {
	areaSvc *service.AreaService
}

func NewAreaHandler(areaSvc *service.AreaService) *AreaHandler {
	return &AreaHandler{areaSvc: areaSvc}
}

func (h *AreaHandler) List(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	if r.URL.Query().Get("include") == "stats" {
		areas, err := h.areaSvc.ListWithStats(r.Context(), wsID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list areas")
			return
		}
		writeJSON(w, http.StatusOK, areas)
		return
	}

	areas, err := h.areaSvc.List(r.Context(), wsID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list areas")
		return
	}
	writeJSON(w, http.StatusOK, areas)
}

func (h *AreaHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid area id")
		return
	}

	area, err := h.areaSvc.GetByID(r.Context(), wsID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "area not found")
		return
	}
	writeJSON(w, http.StatusOK, area)
}

func (h *AreaHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid area id")
		return
	}

	summary, err := h.areaSvc.GetSummary(r.Context(), wsID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "area not found")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *AreaHandler) Create(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}
	limits, ok := middleware.GetEntitlements(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "entitlements not loaded")
		return
	}

	var req service.CreateAreaRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	area, err := h.areaSvc.Create(r.Context(), wsID, limits, req)
	if err != nil {
		if err.Error() == "area limit reached" {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create area")
		return
	}
	writeJSON(w, http.StatusCreated, area)
}

func (h *AreaHandler) Update(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid area id")
		return
	}

	var req service.UpdateAreaRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	area, err := h.areaSvc.Update(r.Context(), wsID, id, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update area")
		return
	}
	writeJSON(w, http.StatusOK, area)
}

func (h *AreaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid area id")
		return
	}

	if err := h.areaSvc.Delete(r.Context(), wsID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete area")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *AreaHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid area id")
		return
	}

	var req struct {
		SortOrder int `json:"sort_order"`
	}
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.areaSvc.Reorder(r.Context(), wsID, id, req.SortOrder); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reorder area")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "reordered"})
}
