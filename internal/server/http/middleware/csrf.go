package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/turtacn/QuantaID/internal/server/http/handlers"
	"github.com/turtacn/QuantaID/pkg/types"
)

const (
	csrfCookieName = "_csrf"
	csrfHeaderName = "X-CSRF-Token"
	csrfFormName   = "_csrf"
	csrfTokenBytes = 32
)

type csrfContextKey string

const csrfTokenContextKey = csrfContextKey("csrf_token")

// CSRFMiddleware protects against CSRF attacks using the double-submit cookie pattern.
// It applies only to UI-related routes, skipping API endpoints.
func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF check for API endpoints, which should use token-based auth
		if strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		// For idempotent methods, generate and set the token.
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			token, err := generateCSRFToken()
			if err != nil {
				appErr := types.ErrInternal.WithCause(err)
				handlers.WriteJSONError(w, appErr, http.StatusInternalServerError)
				return
			}

			// Set the token in a cookie
			http.SetCookie(w, &http.Cookie{
				Name:     csrfCookieName,
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				Secure:   r.TLS != nil,
				SameSite: http.SameSiteLaxMode,
			})

			// Store the token in the request context to be available in templates
			ctx := context.WithValue(r.Context(), csrfTokenContextKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// For non-idempotent methods, verify the token.
		cookieToken, err := r.Cookie(csrfCookieName)
		if err != nil {
			appErr := types.ErrForbidden.WithCause(err)
			handlers.WriteJSONError(w, appErr, http.StatusForbidden)
			return
		}

		// Extract token from form value first, then header
		formToken := r.PostFormValue(csrfFormName)
		if formToken == "" {
			formToken = r.Header.Get(csrfHeaderName)
		}

		if cookieToken.Value == "" || formToken == "" || cookieToken.Value != formToken {
			appErr := types.ErrForbidden
			handlers.WriteJSONError(w, appErr, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetCSRFToken retrieves the CSRF token from the request context.
// This is used by the UI renderer to embed the token in forms.
func GetCSRFToken(r *http.Request) string {
	if token, ok := r.Context().Value(csrfTokenContextKey).(string); ok {
		return token
	}
	return ""
}

func generateCSRFToken() (string, error) {
	b := make([]byte, csrfTokenBytes)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
