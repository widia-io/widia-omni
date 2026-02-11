package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type WorkspaceHandler struct {
	wsSvc *service.WorkspaceService
}

func NewWorkspaceHandler(wsSvc *service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{wsSvc: wsSvc}
}

func (h *WorkspaceHandler) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.wsSvc.ListWorkspaces(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list workspaces")
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func (h *WorkspaceHandler) SwitchWorkspace(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		WorkspaceID uuid.UUID `json:"workspace_id"`
	}
	if err := parseBody(r, &req); err != nil || req.WorkspaceID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.wsSvc.SwitchWorkspace(r.Context(), userID, req.WorkspaceID); err != nil {
		if errors.Is(err, service.ErrWorkspaceNotFound) {
			writeError(w, http.StatusNotFound, "workspace not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to switch workspace")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *WorkspaceHandler) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	ws, err := h.wsSvc.GetWorkspace(r.Context(), wsID, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "workspace not found")
		return
	}

	writeJSON(w, http.StatusOK, ws)
}

func (h *WorkspaceHandler) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	role, _ := middleware.GetRole(r.Context())
	if !role.CanManage() {
		writeError(w, http.StatusForbidden, "insufficient permissions")
		return
	}

	var req service.UpdateWorkspaceRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ws, err := h.wsSvc.UpdateWorkspace(r.Context(), wsID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update workspace")
		return
	}

	writeJSON(w, http.StatusOK, ws)
}

func (h *WorkspaceHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	usage, err := h.wsSvc.GetUsage(r.Context(), wsID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get usage")
		return
	}

	writeJSON(w, http.StatusOK, usage)
}

func (h *WorkspaceHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	members, err := h.wsSvc.ListMembers(r.Context(), wsID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list members")
		return
	}
	writeJSON(w, http.StatusOK, members)
}

func (h *WorkspaceHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	actorUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	role, _ := middleware.GetRole(r.Context())

	targetUserID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	err = h.wsSvc.RemoveMember(r.Context(), wsID, actorUserID, role, targetUserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInsufficientPerms):
			writeError(w, http.StatusForbidden, "insufficient permissions")
		case errors.Is(err, service.ErrSelfRemovalForbidden):
			writeError(w, http.StatusBadRequest, "cannot remove yourself")
		case errors.Is(err, service.ErrOwnerRemovalForbidden):
			writeError(w, http.StatusBadRequest, "cannot remove owner")
		case errors.Is(err, service.ErrMemberNotFound):
			writeError(w, http.StatusNotFound, "member not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to remove member")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkspaceHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	actorUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	role, _ := middleware.GetRole(r.Context())

	var req service.CreateWorkspaceInviteRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	invite, err := h.wsSvc.CreateInvite(r.Context(), wsID, actorUserID, role, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInsufficientPerms):
			writeError(w, http.StatusForbidden, "insufficient permissions")
		case errors.Is(err, service.ErrFamilyNotEnabled):
			writeError(w, http.StatusForbidden, "family plan is not enabled")
		case errors.Is(err, service.ErrMembersLimitReached):
			writeError(w, http.StatusForbidden, "members limit reached")
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusCreated, invite)
}

func (h *WorkspaceHandler) ListInvites(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	invites, err := h.wsSvc.ListInvites(r.Context(), wsID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list invites")
		return
	}
	writeJSON(w, http.StatusOK, invites)
}

func (h *WorkspaceHandler) RevokeInvite(w http.ResponseWriter, r *http.Request) {
	wsID, ok := middleware.GetWorkspaceID(r.Context())
	if !ok {
		writeError(w, http.StatusForbidden, "workspace not found")
		return
	}

	actorUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	role, _ := middleware.GetRole(r.Context())

	inviteID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invite id")
		return
	}

	err = h.wsSvc.RevokeInvite(r.Context(), wsID, inviteID, actorUserID, role)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInsufficientPerms):
			writeError(w, http.StatusForbidden, "insufficient permissions")
		case errors.Is(err, service.ErrInviteNotFound):
			writeError(w, http.StatusNotFound, "invite not found")
		default:
			writeError(w, http.StatusInternalServerError, "failed to revoke invite")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkspaceHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Token string `json:"token"`
	}
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	wsID, err := h.wsSvc.AcceptInvite(r.Context(), userID, req.Token)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInviteNotFound):
			writeError(w, http.StatusNotFound, "invite not found")
		case errors.Is(err, service.ErrInviteExpired):
			writeError(w, http.StatusBadRequest, "invite expired")
		case errors.Is(err, service.ErrInviteRevoked):
			writeError(w, http.StatusBadRequest, "invite revoked")
		case errors.Is(err, service.ErrInviteAccepted):
			writeError(w, http.StatusBadRequest, "invite already accepted")
		case errors.Is(err, service.ErrInviteEmailMismatch):
			writeError(w, http.StatusForbidden, "invite email mismatch")
		case errors.Is(err, service.ErrFamilyNotEnabled):
			writeError(w, http.StatusForbidden, "family plan is not enabled")
		case errors.Is(err, service.ErrMembersLimitReached):
			writeError(w, http.StatusForbidden, "members limit reached")
		default:
			writeError(w, http.StatusInternalServerError, "failed to accept invite")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":       "accepted",
		"workspace_id": wsID.String(),
	})
}
