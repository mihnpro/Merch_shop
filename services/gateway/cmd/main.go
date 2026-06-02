package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/gateway/internal/app/port"
	"github.com/mihnpro/Merch_shop/services/gateway/internal/app/service"
	"github.com/mihnpro/Merch_shop/services/gateway/internal/config"
	"github.com/mihnpro/Merch_shop/services/gateway/internal/infrastructure/account"
	grpcclient "github.com/mihnpro/Merch_shop/services/gateway/internal/infrastructure/client/grpc"
	"github.com/mihnpro/Merch_shop/services/gateway/internal/infrastructure/rest/handlers"
	"github.com/mihnpro/Merch_shop/services/gateway/internal/infrastructure/rest/middleware"
	"github.com/mihnpro/Merch_shop/services/gateway/internal/infrastructure/rest/routes"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	clients, err := grpcclient.New(cfg, logger)
	if err != nil {
		logger.Fatal("failed to init grpc clients", zap.Error(err))
	}
	defer clients.Close()

	adapters := port.Adapters{
		User: grpcclient.NewUserAdapter(clients.User),
	}

	verifier := account.NewVerifier(cfg.JWTAccessSecret, cfg.JWTRefreshSecret)
	mw := middleware.NewMiddleware(logger, verifier)

	authSvc := service.NewAuthQueryService(adapters.User)

	server := handlers.NewServer(authSvc, logger)
	router := routes.New(server, mw)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	go func() {
		logger.Info("HTTP server starting", zap.String("port", cfg.HTTPPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down HTTP server")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("forced shutdown", zap.Error(err))
	}
}
