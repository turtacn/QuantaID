package middleware

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/turtacn/QuantaID/internal/audit"
	"github.com/turtacn/QuantaID/pkg/audit/events"
)

// AuditorMiddleware logs HTTP requests.
func AuditorMiddleware(logger *audit.AuditLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// We only want to audit state-changing methods
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodDelete {
				next.ServeHTTP(w, r)
				return
			}

			var bodyBytes []byte
			if r.Body != nil {
				bodyBytes, _ = ioutil.ReadAll(r.Body)
			}
			r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore the body

			// Capture the response status
			lrw := NewResponseWriter(w)
			next.ServeHTTP(lrw, r)

			// Get the user ID from the context
			userID, _ := r.Context().Value(UserIDContextKey).(string)

			// Mask sensitive data in the request body
			maskedBody := maskSensitiveData(bodyBytes)

			// Log the audit event
			logger.Record(r.Context(), &events.AuditEvent{
				EventType: events.EventDataModified,
				Actor:     events.Actor{ID: userID, Type: "user"},
				Target:    events.Target{Type: "http-request"},
				Result:    getResult(lrw.StatusCode()),
				Metadata: map[string]interface{}{
					"path":   r.URL.Path,
					"method": r.Method,
					"status": lrw.StatusCode(),
					"body":   string(maskedBody),
				},
				IPAddress: r.RemoteAddr,
				UserAgent: r.UserAgent(),
			})
		})
	}
}

// maskSensitiveData masks sensitive fields in a JSON byte slice.
func maskSensitiveData(body []byte) []byte {
	if len(body) == 0 {
		return nil
	}
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return body // Not a JSON body, so we can't mask it
	}

	for key, value := range data {
		if _, ok := value.(string); ok {
			lowerKey := strings.ToLower(key)
			if strings.Contains(lowerKey, "password") || strings.Contains(lowerKey, "secret") || strings.Contains(lowerKey, "token") {
				data[key] = "***"
			}
		}
	}

	maskedBody, _ := json.Marshal(data)
	return maskedBody
}

// getResult converts an HTTP status code to an audit result.
func getResult(statusCode int) events.Result {
	if statusCode >= 200 && statusCode < 300 {
		return events.ResultSuccess
	}
	return events.ResultFailure
}

// responseWriter is a wrapper around http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// NewResponseWriter creates a new responseWriter.
func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

// WriteHeader captures the status code before writing the header.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// StatusCode returns the captured status code.
func (rw *responseWriter) StatusCode() int {
	return rw.statusCode
}
