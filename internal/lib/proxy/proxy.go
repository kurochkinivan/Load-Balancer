package proxy

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"syscall"

	"github.com/kurochkinivan/load_balancer/internal/entity"
)

type LoadBalanceAlgorithm interface {
	Next() (int32, bool)
	Reset()
}

type ReverseProxy struct {
	log      *slog.Logger
	backends []*entity.Backend
	balancer LoadBalanceAlgorithm
}

func New(log *slog.Logger, backends []*entity.Backend, balancer LoadBalanceAlgorithm) *ReverseProxy {
	return &ReverseProxy{
		log:      log,
		backends: backends,
		balancer: balancer,
	}
}

func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	next, ok := p.balancer.Next()
	if !ok {
		http.Error(w, ErrNoServicesAvailable.Error(), http.StatusServiceUnavailable)
		return
	}

	backend := p.backends[next]
	proxy := p.NewProxy(backend)

	proxy.ServeHTTP(w, r)
}

func (p *ReverseProxy) NewProxy(backend *entity.Backend) *httputil.ReverseProxy {
	log := p.log.With(
		slog.String("backend", backend.URL.Host),
	)

	return &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			log.Info("proxying request", slog.String("path", pr.Out.URL.Path))

			pr.SetURL(backend.URL)
			pr.Out.Host = backend.URL.Host
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			if errors.Is(err, syscall.ECONNREFUSED) {
				log.Warn("backend refused connection")

				backend.SetAvailable(false)

				p.ServeHTTP(w, r)
				return
			}

			log.Error("proxy error", slog.String("error", err.Error()))
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		},
	}
}
