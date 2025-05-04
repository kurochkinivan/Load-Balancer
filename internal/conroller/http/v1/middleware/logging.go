package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// LogMiddleware wraps an http.Handler and logs incoming requests.
// It logs the client IP, request path, duration of the request, and the status code of the response.
//
// The logger is expected to be a *slog.Logger.
func LogMiddleware(logger *slog.Logger, next errorHandler) errorHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		start := time.Now()
		path := r.URL.Path
		clientIP := r.RemoteAddr

		log := logger.With(
			slog.String("client", clientIP),
			slog.String("path", path),
		)
		log.Info("handling request")

		rw := &responseWriterWrapper{ResponseWriter: w}

		err := next(rw, r)

		log.Info("request completed",
			slog.Duration("duration", time.Since(start)),
			slog.Int("status", rw.status),
		)

		return err
	}
}

// responseWriterWrapper is a wrapper around http.ResponseWriter that records the status code of the response.
type responseWriterWrapper struct {
	http.ResponseWriter
	status int
}

// WriteHeader writes the header with the given status code.
func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
