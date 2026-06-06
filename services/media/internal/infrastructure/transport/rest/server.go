package rest

import (
	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/media/internal/app/service"
)

type Server struct {
	uploader service.Uploader
	log      *zap.Logger
}

func NewServer(uploader service.Uploader, log *zap.Logger) *Server {
	return &Server{uploader: uploader, log: log}
}
