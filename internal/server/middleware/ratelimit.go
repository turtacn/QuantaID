package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/turtacn/QuantaID/internal/services/platform"
	"github.com/turtacn/QuantaID/pkg/types"
	"go.uber.org/zap"
)

// Lua script for distributed rate limiting (Fixed Window with Expiry)
// Optimized to be atomic.
const redisRateLimitScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])

local current = redis.call("INCR", key)
if current == 1 then
    redis.call("EXPIRE", key, window)
end

if current > limit then
    return 0 -- Rejected
else
    return 1 -- Allowed
end
`

// RateLimitMiddleware implements distributed rate limiting using Redis.
type RateLimitMiddleware struct {
	redisClient   *redis.Client
	apiKeyService *platform.APIKeyService // Optional service for policy lookup
	defaultLimit  int
	defaultWindow int
	logger        *zap.Logger
}

// NewRateLimitMiddleware creates a new RateLimitMiddleware.
func NewRateLimitMiddleware(redisClient *redis.Client, apiKeyService *platform.APIKeyService, defaultLimit, defaultWindow int, logger *zap.Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		redisClient:   redisClient,
		apiKeyService: apiKeyService,
		defaultLimit:  defaultLimit,
		defaultWindow: defaultWindow,
		logger:        logger,
	}
}

// Execute is the middleware handler.
func (m *RateLimitMiddleware) Execute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Determine Identifier (AppID > IP)
		var identifier string
		var prefix string
		var appID string

		if val, ok := r.Context().Value(types.ContextKeyAppID).(string); ok && val != "" {
			appID = val
			identifier = appID
			prefix = "ratelimit:app:"
		} else {
			// Fallback to IP
			ip := r.RemoteAddr
			// Handle X-Forwarded-For if behind proxy
			if idx := strings.LastIndex(ip, ":"); idx != -1 {
				ip = ip[:idx]
			}
			identifier = ip
			prefix = "ratelimit:ip:"
		}

		key := prefix + identifier

		// 2. Determine Limit & Window
		limit := m.defaultLimit
		window := m.defaultWindow

		// Fetch custom policy if AppID is present and service is available
		if appID != "" && m.apiKeyService != nil {
			// Note: This is a synchronous DB call. In high traffic, this should be cached.
			policy, err := m.apiKeyService.GetRateLimitPolicy(r.Context(), appID)
			if err == nil && policy != nil {
				limit = policy.Limit
				window = policy.Window
			} else if err != nil && m.logger != nil {
				m.logger.Warn("Failed to fetch rate limit policy", zap.String("app_id", appID), zap.Error(err))
			}
		}

		// 3. Execute Lua Script
		ctx := r.Context()
		allowed, err := m.redisClient.Eval(ctx, redisRateLimitScript, []string{key}, limit, window).Result()

		if err != nil {
			// Fail open on Redis error
			if m.logger != nil {
				m.logger.Error("Rate limit Redis error", zap.Error(err))
			}
			next.ServeHTTP(w, r)
			return
		}

		if allowed.(int64) == 0 {
			// Rejected
			w.Header().Set("Retry-After", strconv.Itoa(window)) // Rough estimate
			err := types.ErrTooManyRequests
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(fmt.Sprintf(`{"code":"%s","message":"%s"}`, err.Code, err.Message)))
			return
		}

		next.ServeHTTP(w, r)
	})
}
