package app

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"

	"github.com/kurochkinivan/load_balancer/internal/config"
	"github.com/kurochkinivan/load_balancer/internal/entity"
	"github.com/kurochkinivan/load_balancer/internal/lib/proxy"
	roundrobin "github.com/kurochkinivan/load_balancer/internal/lib/roundRobin"
)

type App struct {
	log    *slog.Logger
	ctx    context.Context
	server *http.Server
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config, backends []*entity.Backend) *App {
	balancer := roundrobin.New(backends)

	reverseProxy := proxy.New(log, backends, balancer)
	go reverseProxy.StartHealthChecks(ctx, cfg.Proxy.HealthCheck.Delay, cfg.Proxy.HealthCheck.WorkersCount)

	server := &http.Server{
		Addr:         net.JoinHostPort(cfg.Proxy.Host, cfg.Proxy.Port),
		Handler:      reverseProxy,
		ReadTimeout:  cfg.Proxy.ReadTimeout,
		WriteTimeout: cfg.Proxy.WriteTimeout,
		IdleTimeout:  cfg.Proxy.IdleTimeout,
	}

	return &App{
		log:    log,
		ctx:    ctx,
		server: server,
	}
}

func (a *App) MustStart() {
	if err := a.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

func (a *App) Start() error {
	a.log.Info("listening to the server...", slog.String("addr", a.server.Addr))

	return a.server.ListenAndServe()
}

func (a *App) Stop() {
	if err := a.server.Shutdown(a.ctx); err != nil {
		a.log.Error("failed to shutdown the server", slog.String("err", err.Error()))
	}
}
