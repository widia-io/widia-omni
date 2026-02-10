package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func RateLimit(rdb *redis.Client, requestsPerMinute int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rdb == nil {
				next.ServeHTTP(w, r)
				return
			}

			wsID, ok := GetWorkspaceID(r.Context())
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			ent, ok := GetEntitlements(r.Context())
			if ok && ent.APIRateLimitPerMinute > 0 {
				requestsPerMinute = ent.APIRateLimitPerMinute
			}

			prefix := "rl"
			if _, isAPI := GetAPIKeyID(r.Context()); isAPI {
				prefix = "rl:api"
			}
			key := fmt.Sprintf("%s:%s", prefix, wsID.String())
			window := time.Minute

			count, err := rdb.Incr(r.Context(), key).Result()
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {
				rdb.Expire(r.Context(), key, window)
			}

			ttl, _ := rdb.TTL(r.Context(), key).Result()

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(requestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(max(0, requestsPerMinute-int(count))))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))

			if int(count) > requestsPerMinute {
				w.Header().Set("Retry-After", strconv.FormatInt(int64(ttl.Seconds()), 10))
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
