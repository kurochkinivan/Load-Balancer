package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// LogMiddleware wraps an http.Handler and logs incoming requests.
func LogMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path
		clientIP := r.RemoteAddr

		log := logger.With(
			slog.String("client", clientIP),
			slog.String("path", path),
		)

		log.Info("handling request")

		next.ServeHTTP(w, r)

		log.Info("request completed",
			slog.Duration("duration", time.Since(start)),
		)
	})
}
