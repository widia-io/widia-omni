package handler

import (
	"net/http"

	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type OnboardingHandler struct {
	onboardingSvc *service.OnboardingService
}

func NewOnboardingHandler(onboardingSvc *service.OnboardingService) *OnboardingHandler {
	return &OnboardingHandler{onboardingSvc: onboardingSvc}
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
	writeJSON(w, http.StatusOK, status)
}

func (h *OnboardingHandler) SetupAreas(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	var req struct {
		Areas []service.SetupAreaItem `json:"areas"`
	}
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.onboardingSvc.SetupAreas(r.Context(), wsID, req.Areas); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to setup areas")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "areas created"})
}

func (h *OnboardingHandler) SetupGoals(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
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
		writeError(w, http.StatusInternalServerError, "failed to setup goals")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "goals created"})
}

func (h *OnboardingHandler) SetupHabits(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
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
	writeJSON(w, http.StatusCreated, map[string]string{"status": "habits created"})
}

func (h *OnboardingHandler) Complete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.onboardingSvc.Complete(r.Context(), userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to complete onboarding")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "onboarding completed"})
}
