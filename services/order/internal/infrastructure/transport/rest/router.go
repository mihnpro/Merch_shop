package rest

import "net/http"

func NewRouter(s *Server, mw *Middleware) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /api/v1/orders", mw.Auth(http.HandlerFunc(s.CreateOrder)))
	mux.Handle("GET /api/v1/orders", mw.Auth(http.HandlerFunc(s.GetMyOrders)))
	mux.Handle("GET /api/v1/orders/{id}", mw.Auth(http.HandlerFunc(s.GetOrder)))
	mux.Handle("PUT /api/v1/admin/orders/{id}/status", mw.Admin(http.HandlerFunc(s.AdminUpdateOrderStatus)))
	mux.Handle("GET /api/v1/admin/orders", mw.Admin(http.HandlerFunc(s.AdminListOrders)))
	mux.Handle("GET /api/v1/admin/analytics", mw.Admin(http.HandlerFunc(s.AdminAnalytics)))

	return mux
}
