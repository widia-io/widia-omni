package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type HabitHandler struct {
	habitSvc *service.HabitService
}

func NewHabitHandler(habitSvc *service.HabitService) *HabitHandler {
	return &HabitHandler{habitSvc: habitSvc}
}

func (h *HabitHandler) List(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	habits, err := h.habitSvc.List(r.Context(), wsID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list habits")
		return
	}
	writeJSON(w, http.StatusOK, habits)
}

func (h *HabitHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req service.CreateHabitRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	habit, err := h.habitSvc.Create(r.Context(), wsID, limits, req)
	if err != nil {
		if err.Error() == "habit limit reached" {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create habit")
		return
	}
	writeJSON(w, http.StatusCreated, habit)
}

func (h *HabitHandler) Update(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid habit id")
		return
	}

	var req service.UpdateHabitRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	habit, err := h.habitSvc.Update(r.Context(), wsID, id, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update habit")
		return
	}
	writeJSON(w, http.StatusOK, habit)
}

func (h *HabitHandler) Delete(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid habit id")
		return
	}

	if err := h.habitSvc.Delete(r.Context(), wsID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete habit")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *HabitHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	habitID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid habit id")
		return
	}

	var req service.CheckInRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	entry, err := h.habitSvc.CheckIn(r.Context(), wsID, habitID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check in")
		return
	}
	writeJSON(w, http.StatusCreated, entry)
}

func (h *HabitHandler) DeleteCheckIn(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	habitID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid habit id")
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		writeError(w, http.StatusBadRequest, "date required")
		return
	}

	if err := h.habitSvc.DeleteCheckIn(r.Context(), wsID, habitID, date); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete check-in")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *HabitHandler) ListEntries(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		writeError(w, http.StatusBadRequest, "from and to dates required")
		return
	}

	entries, err := h.habitSvc.ListEntries(r.Context(), wsID, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list entries")
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func (h *HabitHandler) GetStreaks(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	streaks, err := h.habitSvc.GetStreaks(r.Context(), wsID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get streaks")
		return
	}
	writeJSON(w, http.StatusOK, streaks)
}
