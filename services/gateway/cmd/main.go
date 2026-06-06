package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/mihnpro/Merch_shop/services/gateway/internal/infrastructure/config"
	"github.com/mihnpro/Merch_shop/services/gateway/internal/infrastructure/account"
	"github.com/mihnpro/Merch_shop/services/gateway/internal/infrastructure/rest"
	"github.com/mihnpro/Merch_shop/services/gateway/internal/infrastructure/rest/middleware"
	"github.com/mihnpro/Merch_shop/services/gateway/internal/infrastructure/rest/proxy"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	verifier := account.NewVerifier(cfg.JWTAccessSecret)
	mw := middleware.NewMiddleware(logger, verifier, cfg.CORSAllowedOrigins)

	p, err := proxy.New(cfg.Routes, mw)
	if err != nil {
		logger.Fatal("failed to build proxy", zap.Error(err))
	}

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:      rest.New(mw, p),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	go func() {
		logger.Info("gateway starting", zap.String("port", cfg.HTTPPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down gateway")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("forced shutdown", zap.Error(err))
	}
}
