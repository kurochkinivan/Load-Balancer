package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/kurochkinivan/load_balancer/internal/app"
	"github.com/kurochkinivan/load_balancer/internal/config"
	"github.com/kurochkinivan/load_balancer/internal/entity"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

// TODO:
// тесты, документирование кода, dockerfile + docker-compose, backends в виде []url (подождать второго задания)
func main() {
	cfg := config.MustLoadConfig()
	backends := MustMapBackends(cfg.Backends)

	log := setUpLogger(cfg.Env)

	log.Info("config is loaded and logger is set up",
		slog.String("env", cfg.Env),
		slog.String("proxy_host", cfg.Proxy.Host),
		slog.String("proxy_port", cfg.Proxy.Port),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	application := app.New(ctx, log, cfg, backends)
	go application.MustStart()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	sig := <-stop

	log.Info("stopping application", slog.String("signal", sig.String()))
	application.Stop()
	log.Info("application is stopped")
}

func setUpLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	default:
		panic("unknown env variable")
	}

	return log
}

func MustMapBackends(urls []string) []*entity.Backend {
	n := len(urls)
	parsedURLs := make([]*entity.Backend, n)

	for idx, unparsedURL := range urls {
		parsedURL, err := url.Parse(unparsedURL)
		if err != nil {
			err = fmt.Errorf("failed to parse url %q: %w", unparsedURL, err)
			panic(err)
		}
		parsedURLs[idx] = entity.NewBackend(parsedURL)
	}

	return parsedURLs
}
