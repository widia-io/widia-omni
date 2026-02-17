package handler

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/widia-io/widia-omni/internal/domain"
	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type OnboardingHandler struct {
	onboardingSvc *service.OnboardingService
	projectSvc    *service.ProjectService
	taskSvc       *service.TaskService
	auditSvc      *service.AuditService
}

func NewOnboardingHandler(
	onboardingSvc *service.OnboardingService,
	projectSvc *service.ProjectService,
	taskSvc *service.TaskService,
	auditSvc *service.AuditService,
) *OnboardingHandler {
	return &OnboardingHandler{
		onboardingSvc: onboardingSvc,
		projectSvc:    projectSvc,
		taskSvc:       taskSvc,
		auditSvc:      auditSvc,
	}
}

func (h *OnboardingHandler) GetAreaTemplates(w http.ResponseWriter, r *http.Request) {
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en-US"
	}
	writeJSON(w, http.StatusOK, service.GetAreaTemplates(locale))
}

func (h *OnboardingHandler) GetGoalSuggestions(w http.ResponseWriter, r *http.Request) {
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en-US"
	}
	areaSlug := r.URL.Query().Get("area_slug")
	if areaSlug == "" {
		writeError(w, http.StatusBadRequest, "area_slug is required")
		return
	}
	writeJSON(w, http.StatusOK, service.GetGoalSuggestions(locale, areaSlug))
}

func (h *OnboardingHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	status, err := h.onboardingSvc.GetStatus(r.Context(), userID, wsID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get onboarding status")
		return
	}

	if !status.Completed &&
		!status.Steps.Areas &&
		!status.Steps.Goals &&
		!status.Steps.Habits &&
		!status.Steps.Project &&
		!status.Steps.FirstTask {
		h.logAction(r, wsID, userID, "onboarding_started", nil)
	}

	writeJSON(w, http.StatusOK, status)
}

func (h *OnboardingHandler) SetupAreas(w http.ResponseWriter, r *http.Request) {
	userID, wsID, ok := h.getTenantIDs(w, r)
	if !ok {
		return
	}

	var req struct {
		Areas []service.SetupAreaItem `json:"areas"`
	}
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	areas, err := h.onboardingSvc.SetupAreas(r.Context(), wsID, req.Areas)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to setup areas")
		return
	}

	h.logAction(r, wsID, userID, "onboarding_areas_completed", map[string]any{
		"count": len(areas),
	})
	writeJSON(w, http.StatusCreated, areas)
}

func (h *OnboardingHandler) SetupGoals(w http.ResponseWriter, r *http.Request) {
	userID, wsID, ok := h.getTenantIDs(w, r)
	if !ok {
		return
	}

	var req struct {
		Goals []service.SetupGoalItem `json:"goals"`
	}
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.onboardingSvc.SetupGoals(r.Context(), wsID, req.Goals); err != nil {
		if errors.Is(err, service.ErrOnboardingGoalAreaRequired) || errors.Is(err, service.ErrOnboardingAreaNotFound) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to setup goals")
		return
	}

	h.logAction(r, wsID, userID, "onboarding_goals_completed", map[string]any{
		"count": len(req.Goals),
	})
	writeJSON(w, http.StatusCreated, map[string]string{"status": "goals created"})
}

func (h *OnboardingHandler) SetupHabits(w http.ResponseWriter, r *http.Request) {
	userID, wsID, ok := h.getTenantIDs(w, r)
	if !ok {
		return
	}

	var req struct {
		Habits []service.SetupHabitItem `json:"habits"`
	}
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.onboardingSvc.SetupHabits(r.Context(), wsID, req.Habits); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to setup habits")
		return
	}

	if len(req.Habits) > 0 {
		h.logAction(r, wsID, userID, "onboarding_habits_completed", map[string]any{
			"count": len(req.Habits),
		})
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "habits created"})
}

func (h *OnboardingHandler) SkipHabits(w http.ResponseWriter, r *http.Request) {
	userID, wsID, ok := h.getTenantIDs(w, r)
	if !ok {
		return
	}

	h.logAction(r, wsID, userID, "onboarding_habits_skipped", nil)
	writeJSON(w, http.StatusOK, map[string]string{"status": "habits skipped"})
}

func (h *OnboardingHandler) SetupProject(w http.ResponseWriter, r *http.Request) {
	userID, wsID, ok := h.getTenantIDs(w, r)
	if !ok {
		return
	}
	limits, ok := middleware.GetEntitlements(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "entitlements not loaded")
		return
	}

	var req service.CreateProjectRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	req.Title = strings.TrimSpace(req.Title)

	project, err := h.projectSvc.Create(r.Context(), wsID, limits, req)
	if err != nil {
		if err.Error() == "project limit reached" {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create onboarding project")
		return
	}

	h.logAction(r, wsID, userID, "onboarding_project_completed", map[string]any{
		"project_id": project.ID,
	})
	writeJSON(w, http.StatusCreated, project)
}

func (h *OnboardingHandler) SetupFirstTask(w http.ResponseWriter, r *http.Request) {
	userID, wsID, ok := h.getTenantIDs(w, r)
	if !ok {
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
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if req.ProjectID == nil {
		writeError(w, http.StatusBadRequest, "project_id is required")
		return
	}
	if req.Priority == "" {
		req.Priority = domain.PriorityMedium
	}

	task, err := h.taskSvc.Create(r.Context(), wsID, limits, req)
	if err != nil {
		if err.Error() == "daily task limit reached" {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create first task")
		return
	}

	h.logAction(r, wsID, userID, "onboarding_first_task_completed", map[string]any{
		"task_id": task.ID,
	})
	writeJSON(w, http.StatusCreated, task)
}

func (h *OnboardingHandler) Complete(w http.ResponseWriter, r *http.Request) {
	userID, wsID, ok := h.getTenantIDs(w, r)
	if !ok {
		return
	}

	if err := h.onboardingSvc.Complete(r.Context(), userID, wsID); err != nil {
		var incompleteErr *service.OnboardingIncompleteError
		if errors.As(err, &incompleteErr) {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
				"error":         "onboarding_incomplete",
				"missing_steps": incompleteErr.MissingSteps,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to complete onboarding")
		return
	}

	h.logAction(r, wsID, userID, "onboarding_completed", nil)
	writeJSON(w, http.StatusOK, map[string]string{"status": "onboarding completed"})
}

func (h *OnboardingHandler) getTenantIDs(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return uuid.Nil, uuid.Nil, false
	}
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return uuid.Nil, uuid.Nil, false
	}
	return userID, wsID, true
}

func (h *OnboardingHandler) logAction(r *http.Request, wsID, userID uuid.UUID, action string, metadata map[string]any) {
	if h.auditSvc == nil {
		return
	}
	ip, ua := requestClientMeta(r)
	_ = h.auditSvc.LogActionOnce(r.Context(), wsID, userID, action, metadata, ip, ua)
}

func requestClientMeta(r *http.Request) (*string, *string) {
	var ip *string
	for _, raw := range []string{
		strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]),
		strings.TrimSpace(r.Header.Get("X-Real-IP")),
		strings.TrimSpace(r.RemoteAddr),
	} {
		if raw == "" {
			continue
		}
		host := raw
		if strings.Contains(raw, ":") {
			if h, _, err := net.SplitHostPort(raw); err == nil {
				host = h
			}
		}
		if parsed := net.ParseIP(host); parsed != nil {
			s := parsed.String()
			ip = &s
			break
		}
	}

	var ua *string
	if trimmed := strings.TrimSpace(r.UserAgent()); trimmed != "" {
		ua = &trimmed
	}
	return ip, ua
}
