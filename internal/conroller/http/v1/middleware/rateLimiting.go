package middleware

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"

	httperror "github.com/kurochkinivan/load_balancer/internal/conroller/http/v1/errors"
	"github.com/kurochkinivan/load_balancer/internal/entity"
	"github.com/kurochkinivan/load_balancer/internal/lib/sl"
)

// ClientProvider is an interface that defines the method to retrieve a client based on IP address.
type ClientProvider interface {
	Client(ctx context.Context, ipAdress string) (*entity.Client, bool)
}

// ClientProvider is an interface that defines the method to create a client
type ClientCreator interface {
	CreateClient(ctx context.Context, client *entity.Client) error
}

// RateLimitingMiddleware is an HTTP middleware that applies rate limiting based on client IP address.
// It uses a ClientProvider to retrieve client information and check if the client is allowed to proceed.
//
// ClientCreator can be nil. If it is not nil, it will be used to create a default client if the client is not found in the database.
func RateLimitingMiddleware(
	log *slog.Logger,
	clientProvider ClientProvider,
	clientCreator ClientCreator,
	next AppHandler,
) AppHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		// Bypass rate limiting for client-related API paths.
		if strings.HasPrefix(r.URL.String(), "/v1/api/clients") {
			log.Debug("skipping rate limiting", slog.String("path", r.URL.Path))
			return next(w, r)
		}

		// Extract IP address from the request's remote address.
		ipAddress, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Error("failed to split host and port", slog.String("remote_addr", r.RemoteAddr))
			return httperror.BadRequest(err, "failed to split host and port")
		}

		// Retrieve the client information using the ClientProvider.
		client, ok := clientProvider.Client(r.Context(), ipAddress)
		if !ok {
			log.Warn("unknown client", slog.String("ip_address", ipAddress))
			if clientCreator != nil {
				err = clientCreator.CreateClient(r.Context(), &entity.Client{IPAddress: ipAddress, Capacity: 1000, RatePerSecond: 100})
				if err != nil {
					log.Error("failed to create client", sl.Error(err))
				}
			} else {
				return httperror.ErrUnknownClient
			}
		}

		// Check if the client is allowed to proceed based on rate limiting.
		if !client.Allow() {
			log.Info("rate limit exceeded", slog.String("ip_address", ipAddress))
			return httperror.ErrRateLimitExceeded
		}

		return next(w, r)
	}
}
