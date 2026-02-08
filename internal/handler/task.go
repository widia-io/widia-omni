package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type TaskHandler struct {
	taskSvc *service.TaskService
}

func NewTaskHandler(taskSvc *service.TaskService) *TaskHandler {
	return &TaskHandler{taskSvc: taskSvc}
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	var f service.TaskFilters
	if v := r.URL.Query().Get("area_id"); v != "" {
		id, err := uuid.Parse(v)
		if err == nil {
			f.AreaID = &id
		}
	}
	if v := r.URL.Query().Get("goal_id"); v != "" {
		id, err := uuid.Parse(v)
		if err == nil {
			f.GoalID = &id
		}
	}
	if v := r.URL.Query().Get("is_completed"); v != "" {
		b := v == "true"
		f.IsCompleted = &b
	}
	if v := r.URL.Query().Get("due_from"); v != "" {
		f.DueFrom = &v
	}
	if v := r.URL.Query().Get("due_to"); v != "" {
		f.DueTo = &v
	}

	tasks, err := h.taskSvc.List(r.Context(), wsID, f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tasks")
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req service.CreateTaskRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	task, err := h.taskSvc.Create(r.Context(), wsID, limits, req)
	if err != nil {
		if err.Error() == "daily task limit reached" {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create task")
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req service.UpdateTaskRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	task, err := h.taskSvc.Update(r.Context(), wsID, id, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update task")
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	if err := h.taskSvc.Delete(r.Context(), wsID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete task")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *TaskHandler) Complete(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := h.taskSvc.Complete(r.Context(), wsID, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to complete task")
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) ToggleFocus(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	task, err := h.taskSvc.ToggleFocus(r.Context(), wsID, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to toggle focus")
		return
	}
	writeJSON(w, http.StatusOK, task)
}
