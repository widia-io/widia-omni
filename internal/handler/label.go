package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type LabelHandler struct {
	labelSvc *service.LabelService
}

func NewLabelHandler(labelSvc *service.LabelService) *LabelHandler {
	return &LabelHandler{labelSvc: labelSvc}
}

func (h *LabelHandler) List(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	labels, err := h.labelSvc.List(r.Context(), wsID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list labels")
		return
	}
	writeJSON(w, http.StatusOK, labels)
}

func (h *LabelHandler) Create(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	var req service.CreateLabelRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	label, err := h.labelSvc.Create(r.Context(), wsID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create label")
		return
	}
	writeJSON(w, http.StatusCreated, label)
}

func (h *LabelHandler) Update(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid label id")
		return
	}

	var req service.UpdateLabelRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	label, err := h.labelSvc.Update(r.Context(), wsID, id, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update label")
		return
	}
	writeJSON(w, http.StatusOK, label)
}

func (h *LabelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid label id")
		return
	}

	if err := h.labelSvc.Delete(r.Context(), wsID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete label")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
