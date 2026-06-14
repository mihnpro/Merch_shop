package rest

import "net/http"

func NewRouter(s *Server, mw *Middleware) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /api/v1/stock", mw.Auth(http.HandlerFunc(s.GetStock)))
	mux.Handle("GET /api/v1/admin/inventory", mw.Admin(http.HandlerFunc(s.ListStock)))
	mux.Handle("POST /api/v1/admin/inventory/adjust", mw.Admin(http.HandlerFunc(s.AdjustStock)))

	return mux
}
