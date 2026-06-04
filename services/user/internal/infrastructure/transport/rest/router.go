package rest

import "net/http"

func NewRouter(s *Server, mw *Middleware) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/auth/register", s.Register)
	mux.HandleFunc("POST /api/v1/auth/login", s.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", s.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", s.Logout)

	mux.Handle("GET /api/v1/me", mw.Auth(http.HandlerFunc(s.Me)))

	return mux
}
