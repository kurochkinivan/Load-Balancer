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

// StartHealthChecks starts checking health of all backends with given interval and workers number.
//
// It uses counting semaphore strategy to limit the number of concurrent workers.
// It will stop checking health if context is canceled.
func (p *ReverseProxy) StartHealthChecks(ctx context.Context, interval time.Duration, workers int) {
	ticker := time.NewTicker(interval)
	tokens := make(chan struct{}, workers)

	p.log.Info("starting initial health check")
	p.healthCheckAllBackends(tokens)

	for {
		select {
		case <-ticker.C:
			p.healthCheckAllBackends(tokens)
		case <-ctx.Done():
			p.log.Info("health checks stopped due to context cancellation")
			ticker.Stop()
			return
		}
	}
}

// healthCheckAllBackends checks health of all backends and sets availability accordingly.
//
// Tokens is a buffered channel that is used to limit the number of concurrent workers.
func (p *ReverseProxy) healthCheckAllBackends(tokens chan struct{}) {
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

