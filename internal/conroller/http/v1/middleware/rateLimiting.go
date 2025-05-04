package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/kurochkinivan/load_balancer/internal/entity"
)

type ClientProvider interface {
	Client(ctx context.Context, ipAdress string) (*entity.Client, bool)
}

func RateLimitingMiddleware(logger *slog.Logger, clientProvider ClientProvider, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipAddress := r.RemoteAddr

		client, ok := clientProvider.Client(r.Context(), ipAddress)
		if !ok {
			logger.Warn("Rate limit not configured for this client")
			
		}

		if !client.Allow() {
			logger.Info("rate limit exceeded", slog.String("ip_address", ipAddress))

			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
