package middleware

import (
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// responseWriter is a wrapper around http.ResponseWriter that allows for capturing
// the status code of the response, which is not normally available to middleware.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// newResponseWriter creates a new responseWriter, wrapping the original http.ResponseWriter.
// It initializes the status code to http.StatusOK.
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

// WriteHeader captures the status code before calling the original WriteHeader method.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware is a middleware component that logs structured information
// about each incoming HTTP request and its response.
type LoggingMiddleware struct {
	logger utils.Logger
}

// NewLoggingMiddleware creates a new instance of the logging middleware.
//
// Parameters:
//   - logger: The logger to be used for logging request information.
//
// Returns:
//   A new LoggingMiddleware instance.
func NewLoggingMiddleware(logger utils.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Execute is the main middleware handler function. It wraps the next handler
// to log details such as method, path, status code, and duration for each request.
func (m *LoggingMiddleware) Execute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		userID, _ := r.Context().Value(UserIDContextKey).(string)

		m.logger.Info(r.Context(), "HTTP Request Handled",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("user_agent", r.UserAgent()),
			zap.Int("status_code", rw.statusCode),
			zap.Duration("duration", duration),
			zap.String("user_id", userID),
		)
	})
}
