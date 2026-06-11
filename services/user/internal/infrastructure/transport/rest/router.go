package rest

import "net/http"

func NewRouter(s *Server, mw *Middleware) http.Handler {
	mux := http.NewServeMux()

	// Auth (public)
	mux.HandleFunc("POST /api/v1/auth/register", s.Register)
	mux.HandleFunc("POST /api/v1/auth/login", s.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", s.Refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", s.Logout)

	// Self (authenticated user)
	mux.Handle("GET /api/v1/me", mw.Auth(http.HandlerFunc(s.Me)))
	mux.Handle("GET /api/v1/me/balance", mw.Auth(http.HandlerFunc(s.GetMyBalance)))
	mux.Handle("GET /api/v1/me/transactions", mw.Auth(http.HandlerFunc(s.GetMyTransactions)))

	// Admin
	mux.Handle("GET /api/v1/admin/users", mw.Admin(http.HandlerFunc(s.AdminListUsers)))
	mux.Handle("GET /api/v1/admin/users/{id}", mw.Admin(http.HandlerFunc(s.AdminGetUser)))
	mux.Handle("POST /api/v1/admin/users/{id}/grant-points", mw.Admin(http.HandlerFunc(s.AdminGrantPoints)))
	mux.Handle("POST /api/v1/admin/users/{id}/reset-password", mw.Admin(http.HandlerFunc(s.AdminResetPassword)))
	mux.Handle("PUT /api/v1/admin/users/{id}/status", mw.Admin(http.HandlerFunc(s.AdminBlockUser)))
	mux.Handle("PUT /api/v1/admin/users/{id}/role", mw.Admin(http.HandlerFunc(s.AdminChangeRole)))
	mux.Handle("GET /api/v1/admin/users/{id}/transactions", mw.Admin(http.HandlerFunc(s.AdminGetUserTransactions)))

	return mux
}
