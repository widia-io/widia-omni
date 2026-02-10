package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/widia-io/widia-omni/internal/middleware"
	"github.com/widia-io/widia-omni/internal/service"
)

type APIKeyHandler struct {
	apiKeySvc *service.APIKeyService
}

func NewAPIKeyHandler(apiKeySvc *service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{apiKeySvc: apiKeySvc}
}

func (h *APIKeyHandler) apiKeyGate(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
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
	if !limits.APIAccess {
		writeError(w, http.StatusForbidden, "API access not available on your plan")
		return uuid.Nil, false
	}
	return wsID, true
}

func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	wsID, ok := h.apiKeyGate(w, r)
	if !ok {
		return
	}

	keys, err := h.apiKeySvc.List(r.Context(), wsID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list API keys")
		return
	}
	writeJSON(w, http.StatusOK, keys)
}

func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	wsID, ok := h.apiKeyGate(w, r)
	if !ok {
		return
	}

	role, ok := middleware.GetRole(r.Context())
	if !ok || !role.CanManage() {
		writeError(w, http.StatusForbidden, "only owners and admins can create API keys")
		return
	}

	userID, _ := middleware.GetUserID(r.Context())

	var req service.CreateAPIKeyRequest
	if err := parseBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	key, err := h.apiKeySvc.Create(r.Context(), wsID, userID, req)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, key)
}

func (h *APIKeyHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	wsID, ok := h.apiKeyGate(w, r)
	if !ok {
		return
	}

	role, ok := middleware.GetRole(r.Context())
	if !ok || !role.CanManage() {
		writeError(w, http.StatusForbidden, "only owners and admins can revoke API keys")
		return
	}

	keyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid key ID")
		return
	}

	if err := h.apiKeySvc.Revoke(r.Context(), wsID, keyID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
