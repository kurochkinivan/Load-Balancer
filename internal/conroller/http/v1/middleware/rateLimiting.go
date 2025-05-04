package middleware

import (
	"context"
	"log/slog"
	"net"
	"net/http"

	"github.com/kurochkinivan/load_balancer/internal/entity"
)

type ClientProvider interface {
	Client(ctx context.Context, ipAdress string) (*entity.Client, bool)
}

type ClientCreator interface {
	CreateClient(ctx context.Context, client *entity.Client) error
}

func RateLimitingMiddleware(
	logger *slog.Logger,
	clientProvider ClientProvider,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipAddress, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "failed to split host and port", http.StatusInternalServerError)
			return
		}

		client, ok := clientProvider.Client(r.Context(), ipAddress)
		if !ok {
			http.Error(w, "rate limit not configured for this client", http.StatusForbidden)
			return
		}

		if !client.Allow() {
			logger.Info("rate limit exceeded", slog.String("ip_address", ipAddress))

			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
