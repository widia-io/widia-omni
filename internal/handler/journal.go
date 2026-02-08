package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type JournalHandler struct {
	journalSvc *service.JournalService
}

func NewJournalHandler(journalSvc *service.JournalService) *JournalHandler {
	return &JournalHandler{journalSvc: journalSvc}
}

func (h *JournalHandler) List(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	entries, err := h.journalSvc.List(r.Context(), wsID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list journal entries")
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func (h *JournalHandler) Get(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	date, err := time.Parse("2006-01-02", chi.URLParam(r, "date"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
		return
	}

	entry, err := h.journalSvc.GetByDate(r.Context(), wsID, date)
	if err != nil {
		writeError(w, http.StatusNotFound, "journal entry not found")
		return
	}
	writeJSON(w, http.StatusOK, entry)
}

func (h *JournalHandler) Upsert(w http.ResponseWriter, r *http.Request) {
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
	if !limits.JournalEnabled {
		writeError(w, http.StatusForbidden, "journal not available on your plan")
		return
	}

	date, err := time.Parse("2006-01-02", chi.URLParam(r, "date"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
		return
	}

	var req service.UpsertJournalRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	entry, err := h.journalSvc.Upsert(r.Context(), wsID, date, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save journal entry")
		return
	}
	writeJSON(w, http.StatusOK, entry)
}

func (h *JournalHandler) Delete(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	date, err := time.Parse("2006-01-02", chi.URLParam(r, "date"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid date format, use YYYY-MM-DD")
		return
	}

	if err := h.journalSvc.Delete(r.Context(), wsID, date); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete journal entry")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
