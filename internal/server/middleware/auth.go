package middleware

import (
	"context"
	"encoding/json"
	"github.com/turtacn/QuantaID/internal/services/authorization"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

// AuthMiddleware is a middleware component that protects HTTP routes by requiring
// a valid JSON Web Token (JWT) in the Authorization header.
type AuthMiddleware struct {
	logger       utils.Logger
	crypto       *utils.CryptoManager
	authzService *authorization.ApplicationService
}

// ContextKey is a custom type for context keys to prevent collisions between packages.
type ContextKey string

// UserIDContextKey is the key used to store the authenticated user's ID in the request context.
const UserIDContextKey ContextKey = "userID"

// NewAuthMiddleware creates a new instance of the authentication middleware.
//
// Parameters:
//   - authzService: The authorization service, which might be used for further checks (currently unused).
//   - crypto: The cryptographic utility for validating JWTs.
//   - logger: The logger for middleware-specific messages.
//
// Returns:
//   A new AuthMiddleware instance.
func NewAuthMiddleware(authzService *authorization.ApplicationService, crypto *utils.CryptoManager, logger utils.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		logger:       logger,
		crypto:       crypto,
		authzService: authzService,
	}
}

// Execute is the main middleware handler function. It inspects the request for a
// 'Bearer' token, validates it, and if successful, injects the user's ID into
// the request's context before passing control to the next handler in the chain.
func (m *AuthMiddleware) Execute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.writeError(w, r, types.ErrUnauthorized.WithDetails(map[string]string{"reason": "missing authorization header"}))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			m.writeError(w, r, types.ErrUnauthorized.WithDetails(map[string]string{"reason": "malformed authorization header"}))
			return
		}

		tokenString := parts[1]
		claims, err := m.crypto.ValidateJWT(tokenString)
		if err != nil {
			m.writeError(w, r, types.ErrInvalidToken.WithCause(err))
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			m.writeError(w, r, types.ErrInvalidToken.WithDetails(map[string]string{"reason": "missing or invalid subject claim"}))
			return
		}

		ctx := context.WithValue(r.Context(), UserIDContextKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// writeError is a helper function to log an authentication error and write a
// standardized JSON error response to the client.
func (m *AuthMiddleware) writeError(w http.ResponseWriter, r *http.Request, err *types.Error) {
	m.logger.Warn(r.Context(), "Authentication failed", zap.String("path", r.URL.Path), zap.String("error", err.Message))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HttpStatus)
	// Using a simple struct to ensure correct JSON formatting for the error message.
	errorResponse := map[string]string{"error": err.Message}
	json.NewEncoder(w).Encode(errorResponse)
}
