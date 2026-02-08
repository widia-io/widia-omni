package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type NotificationHandler struct {
	notifSvc *service.NotificationService
}

func NewNotificationHandler(notifSvc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifSvc: notifSvc}
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "user not found")
		return
	}

	unreadOnly := r.URL.Query().Get("unread") == "true"
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	notifs, err := h.notifSvc.List(r.Context(), wsID, userID, unreadOnly, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list notifications")
		return
	}
	writeJSON(w, http.StatusOK, notifs)
}

func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "user not found")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	if err := h.notifSvc.MarkRead(r.Context(), wsID, userID, id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to mark notification read")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "read"})
}

func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "user not found")
		return
	}

	if err := h.notifSvc.MarkAllRead(r.Context(), wsID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to mark all notifications read")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "all read"})
}

func (h *NotificationHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "user not found")
		return
	}

	count, err := h.notifSvc.UnreadCount(r.Context(), wsID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get unread count")
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"count": count})
}
