package middleware

import (
	"context"
	"net/http"

	"github.com/turtacn/QuantaID/internal/services/platform"
	"github.com/turtacn/QuantaID/pkg/types"
)

// APIKeyAuthMiddleware handles API key authentication.
type APIKeyAuthMiddleware struct {
	service *platform.APIKeyService
}

// NewAPIKeyAuthMiddleware creates a new APIKeyAuthMiddleware.
func NewAPIKeyAuthMiddleware(service *platform.APIKeyService) *APIKeyAuthMiddleware {
	return &APIKeyAuthMiddleware{service: service}
}

// Execute is the middleware handler.
func (m *APIKeyAuthMiddleware) Execute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract API Key from Header
		apiKeyHeader := r.Header.Get("X-API-Key")
		if apiKeyHeader == "" {
			// If no key, maybe proceed as anonymous?
			// Or if this middleware is applied to protected routes, return 401.
			// Assuming protected routes.
			writeUnauthorized(w, "Missing API Key")
			return
		}

		// 2. Validate Key
		apiKey, err := m.service.ValidateKey(r.Context(), apiKeyHeader)
		if err != nil {
			writeUnauthorized(w, "Invalid API Key")
			return
		}

		// 3. Inject AppID into Context
		ctx := context.WithValue(r.Context(), types.ContextKeyAppID, apiKey.AppID)
		// Also inject full key info if needed
		ctx = context.WithValue(ctx, types.ContextKeyAPIKey, apiKey)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"code":"unauthorized", "message":"` + msg + `"}`))
}
