package middleware

import (
	"net/http"
	"time"

	"github.com/turtacn/QuantaID/pkg/utils"
	"go.uber.org/zap"
)

// LoggingMiddleware logs incoming HTTP requests.
type LoggingMiddleware struct {
	logger utils.Logger
}

// NewLoggingMiddleware creates a new logging middleware instance.
func NewLoggingMiddleware(logger utils.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: logger}
}

// Execute is the middleware handler function.
func (m *LoggingMiddleware) Execute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := &loggingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(lrw, r)

		m.logger.Info(r.Context(), "HTTP Request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", lrw.statusCode),
			zap.Duration("duration", time.Since(start)),
			zap.String("ip", getClientIP(r)),
		)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
