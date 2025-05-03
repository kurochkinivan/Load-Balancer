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
