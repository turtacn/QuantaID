package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/turtacn/QuantaID/internal/domain/policy"
	"github.com/turtacn/QuantaID/internal/services/authorization"
	"github.com/turtacn/QuantaID/internal/server/middleware"
)

func TestAuthorizationMiddleware(t *testing.T) {
	rules := []authorization.Rule{
		{
			Name:     "allow-admins-to-read-dashboard",
			Effect:   policy.DecisionAllow,
			Actions:  []string{"dashboard.read"},
			Subjects: []string{"group:admins"},
		},
		{
			Name:        "allow-access-from-whitelist",
			Effect:      policy.DecisionAllow,
			Actions:     []string{"api.read"},
			Subjects:    []string{"user:whitelisted-user"},
			IPWhitelist: []string{"10.0.0.1/32"},
		},
	}
	evaluator := authorization.NewDefaultEvaluator(rules)
	authzService := authorization.NewService(evaluator)
	authzMiddleware := middleware.NewAuthorizationMiddleware(authzService, "dashboard.read", "dashboard")
	apiMiddleware := middleware.NewAuthorizationMiddleware(authzService, "api.read", "api")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router := mux.NewRouter()
	router.Handle("/dashboard", authzMiddleware.Execute(handler))
	router.Handle("/api", apiMiddleware.Execute(handler))

	testCases := []struct {
		Name           string
		Path           string
		UserID         string
		Groups         []string
		Headers        map[string]string
		ExpectedStatus int
	}{
		{
			Name:           "admin-user-can-access-dashboard",
			Path:           "/dashboard",
			UserID:         "admin-user",
			Groups:         []string{"admins"},
			ExpectedStatus: http.StatusOK,
		},
		{
			Name:           "non-admin-user-cannot-access-dashboard",
			Path:           "/dashboard",
			UserID:         "regular-user",
			Groups:         []string{"users"},
			ExpectedStatus: http.StatusForbidden,
		},
		{
			Name:   "whitelisted-user-can-access-api-with-x-forwarded-for",
			Path:   "/api",
			UserID: "whitelisted-user",
			Headers: map[string]string{
				"X-Forwarded-For": "10.0.0.1",
			},
			ExpectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tc.Path, nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, tc.UserID)
			ctx = context.WithValue(ctx, middleware.GroupsContextKey, tc.Groups)
			req = req.WithContext(ctx)

			for key, value := range tc.Headers {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, tc.ExpectedStatus, rr.Code)
		})
	}
}
