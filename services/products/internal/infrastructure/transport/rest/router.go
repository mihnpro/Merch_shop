package rest

import "net/http"

func NewRouter(s *Server, mw *Middleware) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /api/v1/products", mw.Auth(http.HandlerFunc(s.ListProducts)))
	mux.Handle("GET /api/v1/products/{product_id}", mw.Auth(http.HandlerFunc(s.GetProduct)))
	mux.Handle("GET /api/v1/categories", mw.Auth(http.HandlerFunc(s.ListCategories)))


	mux.Handle("POST /api/v1/admin/products", mw.Admin(http.HandlerFunc(s.CreateProduct)))
	mux.Handle("PUT /api/v1/admin/products/{product_id}", mw.Admin(http.HandlerFunc(s.UpdateProduct)))
	mux.Handle("DELETE /api/v1/admin/products/{product_id}", mw.Admin(http.HandlerFunc(s.DeactivateProduct)))
	mux.Handle("POST /api/v1/admin/categories", mw.Admin(http.HandlerFunc(s.CreateCategory)))
	mux.Handle("PUT /api/v1/admin/categories/{category_id}", mw.Admin(http.HandlerFunc(s.UpdateCategory)))

	return mux
}
