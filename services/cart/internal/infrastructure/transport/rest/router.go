package rest

import "net/http"

func NewRouter(s *Server, mw *Middleware) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /api/v1/cart", mw.Auth(http.HandlerFunc(s.GetCart)))
	mux.Handle("POST /api/v1/cart/items", mw.Auth(http.HandlerFunc(s.AddItem)))
	mux.Handle("PATCH /api/v1/cart/items/{id}", mw.Auth(http.HandlerFunc(s.UpdateItem)))
	mux.Handle("DELETE /api/v1/cart/items/{id}", mw.Auth(http.HandlerFunc(s.RemoveItem)))
	mux.Handle("DELETE /api/v1/cart", mw.Auth(http.HandlerFunc(s.ClearCart)))

	return mux
}
