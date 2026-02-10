package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/widia-io/widia-omni/internal/observability"
	"github.com/widia-io/widia-omni/internal/service"
)

func APIKeyAuth(apiKeySvc *service.APIKeyService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawKey := extractAPIKey(r)
			if rawKey == "" {
				http.Error(w, `{"error":"missing API key"}`, http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(rawKey, "wsk_") {
				http.Error(w, `{"error":"invalid API key format"}`, http.StatusUnauthorized)
				return
			}

			keyHash := service.HashAPIKey(rawKey)

			ak, limits, err := apiKeySvc.ValidateKey(r.Context(), keyHash)
			if err != nil {
				http.Error(w, `{"error":"invalid API key"}`, http.StatusUnauthorized)
				return
			}

			if ak.ExpiresAt != nil && ak.ExpiresAt.Before(time.Now().UTC()) {
				http.Error(w, `{"error":"API key expired"}`, http.StatusUnauthorized)
				return
			}

			if !limits.APIAccess {
				http.Error(w, `{"error":"API access not available on your plan"}`, http.StatusForbidden)
				return
			}

			ctx := SetWorkspaceID(r.Context(), ak.WorkspaceID)
			ctx = SetEntitlements(ctx, limits)
			ctx = SetAPIKeyID(ctx, ak.ID)
			ctx = observability.WithWorkspaceID(ctx, ak.WorkspaceID.String())

			go apiKeySvc.TouchLastUsed(r.Context(), ak.ID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractAPIKey(r *http.Request) string {
	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer wsk_") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}
