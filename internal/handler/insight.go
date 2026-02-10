package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/widia-io/widia-omni/internal/domain"
	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type InsightHandler struct {
	insightSvc *service.InsightService
}

func NewInsightHandler(insightSvc *service.InsightService) *InsightHandler {
	return &InsightHandler{insightSvc: insightSvc}
}

func (h *InsightHandler) insightGate(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return uuid.Nil, false
	}
	limits, ok := middleware.GetEntitlements(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "entitlements not loaded")
		return uuid.Nil, false
	}
	if !limits.AIInsights {
		writeError(w, http.StatusForbidden, "AI insights not available on your plan")
		return uuid.Nil, false
	}
	return wsID, true
}

func (h *InsightHandler) List(w http.ResponseWriter, r *http.Request) {
	wsID, ok := h.insightGate(w, r)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	insights, err := h.insightSvc.List(r.Context(), wsID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list insights")
		return
	}
	writeJSON(w, http.StatusOK, insights)
}

func (h *InsightHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	wsID, ok := h.insightGate(w, r)
	if !ok {
		return
	}

	insight, err := h.insightSvc.GetLatest(r.Context(), wsID)
	if err != nil {
		writeError(w, http.StatusNotFound, "no insights found")
		return
	}
	writeJSON(w, http.StatusOK, insight)
}

func (h *InsightHandler) Generate(w http.ResponseWriter, r *http.Request) {
	wsID, ok := h.insightGate(w, r)
	if !ok {
		return
	}

	canGenerate, err := h.insightSvc.CanGenerate(r.Context(), wsID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check rate limit")
		return
	}
	if !canGenerate {
		writeError(w, http.StatusTooManyRequests, "insight generation limited to once per day")
		return
	}

	insight, err := h.insightSvc.Generate(r.Context(), wsID, domain.InsightOnDemand)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate insight")
		return
	}
	writeJSON(w, http.StatusCreated, insight)
}
