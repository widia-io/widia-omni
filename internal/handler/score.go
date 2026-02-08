package handler

import (
	"net/http"
	"strconv"

	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type ScoreHandler struct {
	scoreSvc *service.ScoreService
}

func NewScoreHandler(scoreSvc *service.ScoreService) *ScoreHandler {
	return &ScoreHandler{scoreSvc: scoreSvc}
}

func (h *ScoreHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
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

	weeks, _ := strconv.Atoi(r.URL.Query().Get("weeks"))
	if weeks <= 0 {
		weeks = 4
	}
	if limits.ScoreHistoryWeeks > 0 && weeks > limits.ScoreHistoryWeeks {
		weeks = limits.ScoreHistoryWeeks
	}

	history, err := h.scoreSvc.GetHistory(r.Context(), wsID, weeks)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get score history")
		return
	}
	writeJSON(w, http.StatusOK, history)
}

func (h *ScoreHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	score, err := h.scoreSvc.GetCurrent(r.Context(), wsID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"score": nil, "message": "no scores computed yet"})
		return
	}
	writeJSON(w, http.StatusOK, score)
}
