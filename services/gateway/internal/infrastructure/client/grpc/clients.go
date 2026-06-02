package grpcclient

import (
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/mihnpro/Merch_shop/services/gateway/internal/config"
	userpb "github.com/mihnpro/Merch_shop/services/user_customer/api/server/AccountInternal"
)

type Clients struct {
	User  userpb.UserServiceClient
	conns []*grpc.ClientConn
	log   *zap.Logger
}

func New(cfg *config.Config, log *zap.Logger) (*Clients, error) {
	c := &Clients{log: log}

	userConn, err := dial(cfg.UserServiceAddr)
	if err != nil {
		return nil, fmt.Errorf("dial user-service: %w", err)
	}
	c.conns = append(c.conns, userConn)
	c.User = userpb.NewUserServiceClient(userConn)

	log.Info("gRPC clients initialised",
		zap.String("user_service", cfg.UserServiceAddr),
	)
	return c, nil
}

func (c *Clients) Close() {
	for _, conn := range c.conns {
		if err := conn.Close(); err != nil {
			c.log.Warn("close grpc conn", zap.Error(err))
		}
	}
}

func dial(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}
