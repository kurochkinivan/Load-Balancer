package pgapp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kurochkinivan/load_balancer/internal/config"
	pgclient "github.com/kurochkinivan/pgClient"
)

type App struct {
	Pool     *pgxpool.Pool
	log      *slog.Logger
	host     string
	port     string
	username string
	password string
	db       string
}

func New(ctx context.Context, log *slog.Logger, cfg config.PostgreSQLConfig) *App {
	pool, err := pgclient.NewClient(ctx, cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DB)
	if err != nil {
		panic("pgapp.New: " + err.Error())
	}

	return &App{
		log:      log,
		Pool:     pool,
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		db:       cfg.DB,
	}
}

func (a *App) MustRun(ctx context.Context, maxAttempts int, delay time.Duration) {
	if err := a.Run(ctx, maxAttempts, delay); err != nil {
		panic(err)
	}
}

func (a *App) Run(ctx context.Context, maxAttempts int, delay time.Duration) error {
	const op = "pgapp.Run"
	log := a.log.With(
		slog.String("op", op),
		slog.String("host", a.host),
		slog.String("port", a.port),
		slog.String("username", a.username),
		slog.String("db", a.db),
	)

	err := pgclient.PingWithAttempts(ctx, log, a.Pool, maxAttempts, delay)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("connection to postgresql database is established")

	return nil
}

func (a *App) Stop() {
	const op = "pgapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("aborting postgresql connection...",
			slog.String("host", a.host),
			slog.String("port", a.port),
		)

	a.Pool.Close()
}
