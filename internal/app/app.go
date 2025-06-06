package app

import (
	"context"
	"log/slog"

	"github.com/kurochkinivan/load_balancer/internal/app/httpapp"
	"github.com/kurochkinivan/load_balancer/internal/app/pgapp"
	"github.com/kurochkinivan/load_balancer/internal/config"
	"github.com/kurochkinivan/load_balancer/internal/entity"
	"github.com/kurochkinivan/load_balancer/internal/usecase"
	"github.com/kurochkinivan/load_balancer/internal/usecase/storage/cache"
	"github.com/kurochkinivan/load_balancer/internal/usecase/storage/pg"
)

type App struct {
	log           *slog.Logger
	HTTPApp       *httpapp.App
	PostgreSQLApp *pgapp.App
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config, backends []*entity.Backend) *App {
	pgApp := pgapp.New(ctx, log, cfg.PostgreSQL)

	clientsStorage := pg.New(pgApp.Pool)
	clientsCache := cache.NewClientsCache(log, cfg.Cache.MaxElements)

	clientsUseCase := usecase.New(log, clientsStorage, clientsCache)

	httpApp := httpapp.New(log, cfg, backends, clientsCache, clientsUseCase, clientsUseCase, clientsUseCase)

	return &App{
		log:           log,
		PostgreSQLApp: pgApp,
		HTTPApp:       httpApp,
	}
}

func (a *App) Run(ctx context.Context, cfg config.PostgreSQLConnection) {
	go a.PostgreSQLApp.MustRun(ctx, cfg.Attempts, cfg.Delay)
	go a.HTTPApp.MustStart(ctx)
}

func (a *App) Stop(ctx context.Context) {
	a.HTTPApp.Stop(ctx)
	a.PostgreSQLApp.Stop()
}
