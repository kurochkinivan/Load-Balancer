// ReverseProxy is a custom http.Handler that can be used to proxy requests to multiple backend servers.
//
// It uses a LoadBalanceAlgorithm to determine which backend server to use for each incoming request.
// It also handles errors that may occur when communicating with the backend servers.
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
}

type ReverseProxy struct {
	log      *slog.Logger
	backends []*entity.Backend
	balancer LoadBalanceAlgorithm
}

// New creates a new ReverseProxy instance.
func New(log *slog.Logger, backends []*entity.Backend, balancer LoadBalanceAlgorithm) *ReverseProxy {
	return &ReverseProxy{
		log:      log,
		backends: backends,
		balancer: balancer,
	}
}

// ServeHTTP implements the http.Handler interface.
// It proxies the incoming request to one of the available backend servers.
func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	next, ok := p.balancer.Next()
	if !ok {
		http.Error(w, ErrNoBackendsAvailable.Error(), http.StatusServiceUnavailable)
		return
	}

	backend := p.backends[next]
	proxy := p.createProxy(backend)

	proxy.ServeHTTP(w, r)
}

// createProxy creates a new httputil.ReverseProxy instance that proxies requests to the given backend server.
func (p *ReverseProxy) createProxy(backend *entity.Backend) *httputil.ReverseProxy {
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
			p.handleError(w, r, err, backend, log)
		},
	}
}

// handleError is called when an error occurs while proxying a request.
// It logs the error and sets the backend server as unavailable if the error is syscall.ECONNREFUSED.
// If the error is different from syscall.ECONNREFUSED, it returns 502 BadGateway error
func (p *ReverseProxy) handleError(w http.ResponseWriter, r *http.Request, err error, backend *entity.Backend, log *slog.Logger) {
	if errors.Is(err, syscall.ECONNREFUSED) {
		log.Warn(ErrBackendRefusedConnection.Error())
		backend.SetAvailable(false)
		p.ServeHTTP(w, r)
		return
	}

	log.Error("proxy error", slog.String("error", err.Error()))
	http.Error(w, "Bad Gateway", http.StatusBadGateway)
}
