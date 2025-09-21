package middleware

import (
	"context"
	"github.com/turtacn/QuantaID/internal/services/authorization"
	"github.com/turtacn/QuantaID/pkg/types"
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

// AuthMiddleware protects routes by validating JWTs.
type AuthMiddleware struct {
	logger       utils.Logger
	crypto       *utils.CryptoManager
	authzService *authorization.ApplicationService
}

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const UserIDContextKey ContextKey = "userID"

// NewAuthMiddleware creates a new authentication middleware.
func NewAuthMiddleware(authzService *authorization.ApplicationService, crypto *utils.CryptoManager, logger utils.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		logger:       logger,
		crypto:       crypto,
		authzService: authzService,
	}
}

// Execute is the middleware handler function.
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

func (m *AuthMiddleware) writeError(w http.ResponseWriter, r *http.Request, err *types.Error) {
	m.logger.Warn(r.Context(), "Authentication failed", zap.String("path", r.URL.Path), zap.String("error", err.Message))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HttpStatus)
	w.Write([]byte(`{"error":"` + err.Message + `"}`))
}

//Personal.AI order the ending
