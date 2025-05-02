package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/kurochkinivan/load_balancer/internal/config"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

// TODO: 
// тесты, документирование кода
func main() {
	cfg := config.MustLoadConfig()
	fmt.Println(cfg)

	log := setUpLogger(cfg.Env)
	log.Info("logger is set up")

	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			server := cfg.Backends[0]
			targetURL, _ := url.Parse(server)

			log.Info("Перенаправление запроса", slog.String("redirect_url", targetURL.String()))

			pr.SetURL(targetURL)
			pr.Out.Host = targetURL.Host
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			fmt.Println("failed to get resp!")
		},
	}

	log.Info("Прокси сервер запущен на :8080")
	http.Handle("/", proxy)
	http.ListenAndServe(":8080", nil)
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
