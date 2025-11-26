package middleware

import (
	"fmt"
	"net/http"

	"github.com/turtacn/QuantaID/internal/storage/redis"
)

// IPBlacklistMiddleware checks if the request's IP is in the blacklist.
func IPBlacklistMiddleware(redisClient redis.RedisClientInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr // In a real app, consider X-Forwarded-For
			key := fmt.Sprintf("security:blacklist:ip:%s", ip)

			exists, err := redisClient.Exists(r.Context(), key)
			if err != nil {
				// Log the error but allow the request to proceed to avoid blocking
				// legitimate users due to a Redis issue.
				next.ServeHTTP(w, r)
				return
			}

			if exists > 0 {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
