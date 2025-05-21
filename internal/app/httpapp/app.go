package httpapp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/kurochkinivan/load_balancer/internal/config"
	v1 "github.com/kurochkinivan/load_balancer/internal/conroller/http/v1/api"
	"github.com/kurochkinivan/load_balancer/internal/conroller/http/v1/middleware"
	"github.com/kurochkinivan/load_balancer/internal/conroller/http/v1/proxy"
	"github.com/kurochkinivan/load_balancer/internal/entity"
	roundrobin "github.com/kurochkinivan/load_balancer/internal/lib/roundRobin"
)

type App struct {
	log                *slog.Logger
	server             *http.Server
	reverseProxy       *proxy.ReverseProxy
	tokenRifiller      TokenRefiller
	healtCheckInterval time.Duration
	workers            int
}

const (
	bytesLimit = 1 << 20 // 1024*1024
)

type TokenRefiller interface {
	StartTokenRefiller(ctx context.Context)
}

func New(
	log *slog.Logger,
	cfg *config.Config,
	backends []*entity.Backend,
	tokenRifiller TokenRefiller,
	clientsUseCase v1.ClientsUseCase,
	clientProvider middleware.ClientProvider,
	clientCreator middleware.ClientCreator,
) *App {
	r := httprouter.New()

	balancer := roundrobin.New(backends)
	reverseProxy := proxy.New(log, backends, balancer)
	r.NotFound = reverseProxy

	clientsHandler := v1.NewClientsHandler(clientsUseCase, bytesLimit)
	clientsHandler.Register(r)

	// Base handler
	baseHandler := func(w http.ResponseWriter, req *http.Request) error {
		r.ServeHTTP(w, req)
		return nil
	}

	// Middleware chain
	handler := middleware.RateLimitingMiddleware(log, clientProvider, clientCreator, baseHandler)
	handler = middleware.LogMiddleware(log, handler)
	finalHandler := middleware.ErrorMiddleware(handler)

	server := &http.Server{
		Addr:         net.JoinHostPort(cfg.Proxy.Host, cfg.Proxy.Port),
		Handler:      finalHandler,
		ReadTimeout:  cfg.Proxy.ReadTimeout,
		WriteTimeout: cfg.Proxy.WriteTimeout,
		IdleTimeout:  cfg.Proxy.IdleTimeout,
	}

	return &App{
		log:                log,
		server:             server,
		reverseProxy:       reverseProxy,
		tokenRifiller:      tokenRifiller,
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
	go a.tokenRifiller.StartTokenRefiller(ctx)

	a.log.Info("listening to the server...", slog.String("addr", a.server.Addr))
	err := a.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to listen and serve: %w", err)
	}
	return nil
}

func (a *App) Stop(ctx context.Context) {
	a.log.Info("stopping http server")

	if err := a.server.Shutdown(ctx); err != nil {
		a.log.Error("failed to shutdown the server", slog.String("err", err.Error()))
	}
}
