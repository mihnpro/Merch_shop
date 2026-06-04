package rest

import (
	"net/http"

	"github.com/mihnpro/Merch_shop/services/gateway/internal/infrastructure/rest/middleware"
	"github.com/mihnpro/Merch_shop/services/gateway/internal/infrastructure/rest/proxy"
)

func New(mw *middleware.Middleware, p *proxy.Proxy) http.Handler {
	return mw.CORS(mw.Auth(p))
}
