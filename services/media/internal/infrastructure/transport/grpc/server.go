package grpc

import (
	"go.uber.org/zap"

	mediapb "github.com/mihnpro/Merch_shop/services/media/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/media/internal/app/service"
)

type GRPCServer struct {
	uploader service.Uploader
	log      *zap.Logger
	mediapb.UnimplementedMediaServiceServer
}

func NewGRPCServer(uploader service.Uploader, log *zap.Logger) *GRPCServer {
	return &GRPCServer{uploader: uploader, log: log}
}
