package middleware

import (
	"github.com/gorilla/mux"
	"github.com/turtacn/QuantaID/internal/domain/policy"
	"github.com/turtacn/QuantaID/internal/services/authorization"
	"net/http"
	"strings"
	"time"
)

// AuthorizationMiddleware enforces authorization for HTTP routes.
type AuthorizationMiddleware struct {
	authzSvc     *authorization.Service
	action       policy.Action
	resourceType string
}

// NewAuthorizationMiddleware creates a new authorization middleware instance.
func NewAuthorizationMiddleware(authzSvc *authorization.Service, action policy.Action, resourceType string) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{
		authzSvc:     authzSvc,
		action:       action,
		resourceType: resourceType,
	}
}

// Execute is the middleware handler function.
func (m *AuthorizationMiddleware) Execute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDContextKey).(string)
		if !ok {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		groups, _ := r.Context().Value(GroupsContextKey).([]string)

		vars := mux.Vars(r)
		resourceID := vars["id"]

		evalCtx := policy.EvaluationContext{
			Subject: policy.Subject{
				UserID: userID,
				Groups: groups,
			},
			Resource: policy.Resource{
				Type: m.resourceType,
				ID:   resourceID,
			},
			Action: m.action,
			Environment: policy.Environment{
				IP:   getClientIP(r),
				Time: time.Now().UTC(),
			},
		}

		decision, err := m.authzSvc.Authorize(r.Context(), evalCtx)
		if err != nil || decision != policy.DecisionAllow {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}
