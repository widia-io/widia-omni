package middleware

import (
	"net/http"
)

func RequireFeature(feature string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ent, ok := GetEntitlements(r.Context())
			if !ok {
				http.Error(w, `{"error":"entitlements not loaded"}`, http.StatusForbidden)
				return
			}

			allowed := false
			switch feature {
			case "finance":
				allowed = ent.FinanceEnabled
			case "export":
				allowed = ent.ExportEnabled
			case "ai_insights":
				allowed = ent.AIInsights
			case "api_access":
				allowed = ent.APIAccess
			case "journal":
				allowed = ent.JournalEnabled
			default:
				allowed = true
			}

			if !allowed {
				http.Error(w, `{"error":"feature not available on your plan"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireLimit(limitKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit checks are done at the service layer where we have access
			// to both counters and entitlements. This middleware is a placeholder
			// for route-level enforcement if needed.
			next.ServeHTTP(w, r)
		})
	}
}
