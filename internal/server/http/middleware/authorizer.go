package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/turtacn/QuantaID/internal/policy/engine"
	"github.com/turtacn/QuantaID/internal/server/http/handlers"
	"github.com/turtacn/QuantaID/pkg/types"
)

type contextKey string
const userIDKey contextKey = "user_id"

// RequirePermission is a middleware that checks if the user has the required permission.
func RequirePermission(evaluator engine.Evaluator, requiredPermission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(userIDKey).(string)
			if !ok || userID == "" {
				handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusUnauthorized, Message: "Unauthorized"}, http.StatusUnauthorized)
				return
			}

			parts := strings.Split(requiredPermission, ":")
			if len(parts) != 2 {
				handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusInternalServerError, Message: "Invalid permission format"}, http.StatusInternalServerError)
				return
			}
			resource := parts[0]
			action := parts[1]

			req := engine.EvaluationRequest{
				SubjectID: userID,
				Action:    action,
				Resource:  resource,
				Context:   make(map[string]interface{}), // Context can be enriched here
			}

			allowed, err := evaluator.Evaluate(r.Context(), req)
			if err != nil {
				handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusInternalServerError, Message: "Policy evaluation error"}, http.StatusInternalServerError)
				return
			}

			if !allowed {
				handlers.WriteJSONError(w, &types.Error{HttpStatus: http.StatusForbidden, Message: "Permission denied"}, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// WithUserID adds the user ID to the request context.
// This is a helper function that would be used by the authentication middleware.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
