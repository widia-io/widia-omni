package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/widia-io/widia-omni/internal/domain"
	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type GoalHandler struct {
	goalSvc *service.GoalService
}

func NewGoalHandler(goalSvc *service.GoalService) *GoalHandler {
	return &GoalHandler{goalSvc: goalSvc}
}

func (h *GoalHandler) List(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	var f service.GoalFilters
	if v := r.URL.Query().Get("area_id"); v != "" {
		id, err := uuid.Parse(v)
		if err == nil {
			f.AreaID = &id
		}
	}
	if v := r.URL.Query().Get("period"); v != "" {
		p := domain.GoalPeriod(v)
		f.Period = &p
	}
	if v := r.URL.Query().Get("status"); v != "" {
		s := domain.GoalStatus(v)
		f.Status = &s
	}

	goals, err := h.goalSvc.List(r.Context(), wsID, f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list goals")
		return
	}
	writeJSON(w, http.StatusOK, goals)
}

func (h *GoalHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid goal id")
		return
	}

	goal, err := h.goalSvc.GetByID(r.Context(), wsID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "goal not found")
		return
	}
	writeJSON(w, http.StatusOK, goal)
}

func (h *GoalHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req service.CreateGoalRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	goal, err := h.goalSvc.Create(r.Context(), wsID, limits, req)
	if err != nil {
		if err.Error() == "goal limit reached" {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create goal")
		return
	}
	writeJSON(w, http.StatusCreated, goal)
}

func (h *GoalHandler) Update(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid goal id")
		return
	}

	var req service.UpdateGoalRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	goal, err := h.goalSvc.Update(r.Context(), wsID, id, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update goal")
		return
	}
	writeJSON(w, http.StatusOK, goal)
}

func (h *GoalHandler) Delete(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid goal id")
		return
	}

	if err := h.goalSvc.Delete(r.Context(), wsID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete goal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *GoalHandler) UpdateProgress(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid goal id")
		return
	}

	var req struct {
		CurrentValue float64 `json:"current_value"`
	}
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	goal, err := h.goalSvc.UpdateProgress(r.Context(), wsID, id, req.CurrentValue)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update progress")
		return
	}
	writeJSON(w, http.StatusOK, goal)
}
