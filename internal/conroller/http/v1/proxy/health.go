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
// This function performs an initial health check of all backend servers immediately upon being called,
// and then continues checking at the specified interval. Each health check is performed concurrently,
// up to a limit defined by the 'workers' parameter.
//
// The health checks will continue until the provided context is canceled or timed out.
//
// Parameters:
//   - ctx: Context used to control the lifecycle of the health checks.
//   - interval: Time duration between each round of health checks.
//   - workers: Maximum number of concurrent health checks allowed.
func (p *ReverseProxy) StartHealthChecks(ctx context.Context, interval time.Duration, workers int) {
	ticker := time.NewTicker(interval)
	tokens := make(chan struct{}, workers)

	p.log.Info("starting initial health check")
	p.healthCheckAllBackends(tokens)
	p.log.Info("initial health check is completed")

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

// healthCheckAllBackends performs a health check on all backend servers concurrently.
//
// This function spawns a goroutine for each backend to check its /health endpoint.
// It uses a buffered channel as a semaphore to limit the number of concurrent checks.
//
// Parameters:
//   - tokens: A buffered channel used as a counting semaphore to limit concurrency.
func (p *ReverseProxy) healthCheckAllBackends(tokens chan struct{}) {
	p.log.Info("starting health check for all backends")

	for _, backend := range p.backends {
		tokens <- struct{}{} // Acquire token

		go func(backend *entity.Backend) {
			defer func() { <-tokens }() // Release token when done

			log := p.log.With(
				slog.String("backend", backend.URL.Host),
			)

			healthURL := backend.URL.String() + "/health"

			resp, err := http.Get(healthURL)
			if err != nil {
				if errors.Is(err, syscall.ECONNREFUSED) {
					log.Warn("backend is unhealthy", slog.String("error", "backend refused connection"))
				} else {
					log.Warn("error while checking backend health", slog.String("error", err.Error()))
				}

				backend.SetAvailable(false)
				return
			}

			if resp.StatusCode == http.StatusOK {
				log.Debug("backend is healthy")
				backend.SetAvailable(true)
			} else {
				log.Warn("backend is unhealthy", slog.String("status_code", resp.Status))
				backend.SetAvailable(false)
			}
		}(backend)
	}
}
