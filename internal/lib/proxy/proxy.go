package proxy

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type LoadBalanceAlgorithm interface {
	Next() int32
	Reset()
}

type ReverseProxy struct {
	log      *slog.Logger
	backends []string
	balancer LoadBalanceAlgorithm
}

func New(log *slog.Logger, backends []string, balancer LoadBalanceAlgorithm) *ReverseProxy {
	return &ReverseProxy{
		log:      log,
		backends: backends,
		balancer: balancer,
	}
}

// TODO: точно ли все урлы верны? ПЕРЕДЕЛАТЬ backend в []*url.Url
func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	proxy := httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			next := p.balancer.Next()
			server := p.backends[next]
			targetURL, _ := url.Parse(server)

			url := pr.Out.URL.String()

			pr.SetURL(targetURL)
			pr.Out.Host = targetURL.Host

			p.log.Info("proxying request",
				slog.String("host", pr.Out.Host),
				slog.String("target", url),
			)
		},
	}

	proxy.ServeHTTP(w, r)
}
