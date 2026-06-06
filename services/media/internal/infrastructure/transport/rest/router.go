package rest

import "net/http"

func NewRouter(s *Server, mw *Middleware) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/admin/media/photos", mw.AdminOnly(s.UploadPhoto))

	return mux
}
