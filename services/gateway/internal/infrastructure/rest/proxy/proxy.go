package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strings"

	"github.com/mihnpro/Merch_shop/services/gateway/internal/infrastructure/config"
)


type Authorizer interface {
	Authorize(w http.ResponseWriter, r *http.Request) bool
}

type route struct {
	prefix string
	public bool
	proxy  *httputil.ReverseProxy
}

type Proxy struct {
	routes []route
	authz  Authorizer
}

func New(routes []config.Route, authz Authorizer) (*Proxy, error) {
	p := &Proxy{authz: authz}
	for _, r := range routes {
		target, err := url.Parse(r.Target)
		if err != nil {
			return nil, fmt.Errorf("invalid target %q: %w", r.Target, err)
		}
		p.routes = append(p.routes, route{
			prefix: r.Prefix,
			public: r.Public,
			proxy:  httputil.NewSingleHostReverseProxy(target),
		})
	}
	sort.Slice(p.routes, func(i, j int) bool {
		return len(p.routes[i].prefix) > len(p.routes[j].prefix)
	})
	return p, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, rt := range p.routes {
		if strings.HasPrefix(r.URL.Path, rt.prefix) {
			if rt.public || p.authz.Authorize(w, r) {
				rt.proxy.ServeHTTP(w, r)
			}
			return
		}
	}
	http.NotFound(w, r)
}
