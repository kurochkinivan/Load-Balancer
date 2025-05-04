package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/kurochkinivan/load_balancer/internal/config"
	"github.com/kurochkinivan/load_balancer/internal/entity"
	"github.com/kurochkinivan/load_balancer/internal/lib/middleware"
	"github.com/kurochkinivan/load_balancer/internal/lib/proxy"
	roundrobin "github.com/kurochkinivan/load_balancer/internal/lib/roundRobin"
)

type App struct {
	log                *slog.Logger
	server             *http.Server
	reverseProxy       *proxy.ReverseProxy
	healtCheckInterval time.Duration
	workers            int
}

func New(log *slog.Logger, cfg *config.Config, backends []*entity.Backend) *App {
	balancer := roundrobin.New(backends)

	reverseProxy := proxy.New(log, backends, balancer)

	server := &http.Server{
		Addr:         net.JoinHostPort(cfg.Proxy.Host, cfg.Proxy.Port),
		Handler:      middleware.LogMiddleware(log, reverseProxy),
		ReadTimeout:  cfg.Proxy.ReadTimeout,
		WriteTimeout: cfg.Proxy.WriteTimeout,
		IdleTimeout:  cfg.Proxy.IdleTimeout,
	}

	return &App{
		log:                log,
		server:             server,
		reverseProxy:       reverseProxy,
		healtCheckInterval: cfg.Proxy.HealthCheck.Interval,
		workers:            cfg.Proxy.HealthCheck.WorkersCount,
	}
}

func (a *App) MustStart(ctx context.Context) {
	if err := a.Start(ctx); err != nil {
		panic(err)
	}
}

func (a *App) Start(ctx context.Context) error {
	go a.reverseProxy.StartHealthChecks(ctx, a.healtCheckInterval, a.workers)

	a.log.Info("listening to the server...", slog.String("addr", a.server.Addr))
	err := a.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to listen and serve: %w", err)
	}
	return nil
}

func (a *App) Stop(ctx context.Context) {
	if err := a.server.Shutdown(ctx); err != nil {
		a.log.Error("failed to shutdown the server", slog.String("err", err.Error()))
	}
}
