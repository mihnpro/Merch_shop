package routes

import (
	"net/http"

	"github.com/mihnpro/Merch_shop/services/gateway/internal/infrastructure/rest/handlers"
	"github.com/mihnpro/Merch_shop/services/gateway/internal/infrastructure/rest/middleware"
)

func New(h *handlers.Server, mw *middleware.Middleware) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/auth/register", h.Register)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", h.Logout)

	mux.Handle("GET /api/v1/me", mw.Auth(http.HandlerFunc(h.Me)))

	return mw.CORS(mux)
}
