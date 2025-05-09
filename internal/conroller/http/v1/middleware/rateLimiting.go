package middleware

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"

	httperror "github.com/kurochkinivan/load_balancer/internal/conroller/http/v1/errors"
	"github.com/kurochkinivan/load_balancer/internal/entity"
)

type ClientProvider interface {
	Client(ctx context.Context, ipAdress string) (*entity.Client, bool)
}

func RateLimitingMiddleware(
	log *slog.Logger,
	clientProvider ClientProvider,
	next AppHandler,
) AppHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		if strings.HasPrefix(r.URL.String(), "/v1/api/clients") {
			log.Debug("skipping rate limiting", slog.String("path", r.URL.Path))
			return next(w, r)
		}

		ipAddress, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Error("failed to split host and port", slog.String("remote_addr", r.RemoteAddr))

			return httperror.BadRequest(err, "failed to split host and port")
		}

		client, ok := clientProvider.Client(r.Context(), ipAddress)
		if !ok {
			log.Warn("unknown client", slog.String("ip_address", ipAddress))

			return httperror.ErrUnknownClient
		}

		if !client.Allow() {
			log.Info("rate limit exceeded", slog.String("ip_address", ipAddress))

			return httperror.ErrRateLimitExceeded
		}

		return next(w, r)
	}
}
