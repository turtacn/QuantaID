package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/turtacn/QuantaID/internal/domain/identity"
	"github.com/turtacn/QuantaID/pkg/utils"
)

// AuthMiddleware validates JWT tokens and adds user info to the context.
type AuthMiddleware struct {
	cryptoManager         *utils.CryptoManager
	logger                utils.Logger
	identityDomainService identity.IService
}

// NewAuthMiddleware creates a new authentication middleware.
func NewAuthMiddleware(cryptoManager *utils.CryptoManager, logger utils.Logger, identityDomainService identity.IService) *AuthMiddleware {
	return &AuthMiddleware{
		cryptoManager:         cryptoManager,
		logger:                logger,
		identityDomainService: identityDomainService,
	}
}

// Execute is the middleware handler function.
func (m *AuthMiddleware) Execute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		claims, err := m.cryptoManager.ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDContextKey, claims["sub"].(string))

		user, err := m.identityDomainService.GetUser(ctx, claims["sub"].(string))
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var groupNames []string
		for _, group := range user.Groups {
			groupNames = append(groupNames, group.Name)
		}
		ctx = context.WithValue(ctx, GroupsContextKey, groupNames)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
