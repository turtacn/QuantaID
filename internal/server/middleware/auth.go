package middleware

import (
	"context"
	"encoding/json"
	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

// AuthMiddleware is a middleware component that protects HTTP routes by requiring
// a valid JSON Web Token (JWT) in the Authorization header.
type AuthMiddleware struct {
	logger         utils.Logger
	crypto         *utils.CryptoManager
	identityDomain identity.IService
}

// ContextKey is a custom type for context keys to prevent collisions between packages.
type ContextKey string

// UserIDContextKey is the key used to store the authenticated user's ID in the request context.
const UserIDContextKey ContextKey = "userID"

// GroupsContextKey is the key used to store the authenticated user's groups in the request context.
const GroupsContextKey ContextKey = "groups"

// NewAuthMiddleware creates a new instance of the authentication middleware.
func NewAuthMiddleware(crypto *utils.CryptoManager, logger utils.Logger, identityDomain identity.IService) *AuthMiddleware {
	return &AuthMiddleware{
		logger:         logger,
		crypto:         crypto,
		identityDomain: identityDomain,
	}
}

// Execute is the main middleware handler function.
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

		userGroups, err := m.identityDomain.GetUserGroups(r.Context(), userID)
		if err != nil {
			m.logger.Warn(r.Context(), "Could not fetch user groups for auth middleware", zap.Error(err), zap.String("userID", userID))
		}

		groupIDs := make([]string, len(userGroups))
		for i, g := range userGroups {
			groupIDs[i] = g.ID
		}

		ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
		ctx = context.WithValue(ctx, GroupsContextKey, groupIDs)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// writeError is a helper function to log an authentication error and write a
// standardized JSON error response to the client.
func (m *AuthMiddleware) writeError(w http.ResponseWriter, r *http.Request, err *types.Error) {
	m.logger.Warn(r.Context(), "Authentication failed", zap.String("path", r.URL.Path), zap.String("error", err.Message))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HttpStatus)
	errorResponse := map[string]string{"error": err.Message}
	json.NewEncoder(w).Encode(errorResponse)
}
