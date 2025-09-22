package middleware

import (
	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// responseWriter is a wrapper around http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs information about each HTTP request.
type LoggingMiddleware struct {
	logger utils.Logger
}

// NewLoggingMiddleware creates a new logging middleware.
func NewLoggingMiddleware(logger utils.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Execute is the middleware handler function.
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

//Personal.AI order the ending
