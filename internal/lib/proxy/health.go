package proxy

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"syscall"
	"time"

	"github.com/kurochkinivan/load_balancer/internal/entity"
)

// StartHealthChecks starts periodical health checks for all backends.
//
// It will create a separate goroutine for each backend and check its health every given delay.
//
// The function will stop checking health when the context is canceled.
//
// The function uses a counting semaphore strategy to limit the number of concurrent health checks.
// The number of concurrent checks is limited by the number of workers.
func (p *ReverseProxy) StartHealthChecks(ctx context.Context, delay time.Duration, workers int) {
	ticker := time.NewTicker(delay)
	tokens := make(chan struct{}, workers)

	for {
		select {
		case <-ticker.C:
			p.checkAllBackends(tokens)
		case <-ctx.Done():
			p.log.Info("health checks stopped due to context cancellation")
			ticker.Stop()
			return
		}
	}
}

// checkAllBackends checks the health of all backends and updates their availability.
//
// It will start a separate goroutine for each backend and limit the number of concurrent checks
// using a counting semaphore strategy.
func (p *ReverseProxy) checkAllBackends(tokens chan struct{}) {
	p.log.Info("starting health check for all backends")

	for _, backend := range p.backends {
		tokens <- struct{}{}

		go func(backend *entity.Backend) {
			defer func() { <-tokens }()

			log := p.log.With(
				slog.String("backend", backend.URL.Host),
			)

			healthURL := backend.URL.String() + "/health"
			resp, err := http.Get(healthURL)
			if err != nil {
				if errors.Is(err, syscall.ECONNREFUSED) {
					log.Warn("backend is unhealthy", slog.String("error", ErrBackendRefusedConnection.Error()))
				} else {
					log.Warn("error while checking backend health", slog.String("error", err.Error()))
				}

				backend.SetAvailable(false)
				return
			}

			if resp.StatusCode == http.StatusOK {
				log.Info("backend is healthy")
				backend.SetAvailable(true)
			} else {
				log.Warn("backend is unhealthy", slog.String("status_code", resp.Status))
				backend.SetAvailable(false)
			}
		}(backend)
	}
}

