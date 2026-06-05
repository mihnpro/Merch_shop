package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	productpb "github.com/mihnpro/Merch_shop/services/products/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/products/internal/app/query"
	"github.com/mihnpro/Merch_shop/services/products/internal/app/service"
	"github.com/mihnpro/Merch_shop/services/products/internal/infrastructure/account"
	"github.com/mihnpro/Merch_shop/services/products/internal/infrastructure/persistence"
	psqlrepo "github.com/mihnpro/Merch_shop/services/products/internal/infrastructure/repository/psql"
	grpctransport "github.com/mihnpro/Merch_shop/services/products/internal/infrastructure/transport/grpc"
	resttransport "github.com/mihnpro/Merch_shop/services/products/internal/infrastructure/transport/rest"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Info("no .env file loaded", zap.Error(err))
	}

	ctx := context.Background()

	db, err := persistence.NewPostgres(ctx, persistence.PostgresConfig{
		URL:      os.Getenv("DATABASE_URL"),
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  os.Getenv("POSTGRES_SSLMODE"),
	})
	if err != nil {
		logger.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer db.Close()

	productRepo := psqlrepo.NewProductRepository(db)
	categoryRepo := psqlrepo.NewCategoryRepository(db)

	writeSvc := service.NewCatalogService(productRepo, categoryRepo)
	readSvc := query.NewCatalogReadService(productRepo, categoryRepo)
	validator := account.NewValidator(os.Getenv("JWT_ACCESS_SECRET"))

	grpcSrv := grpctransport.NewGRPCServer(writeSvc, readSvc)
	grpcPort := getEnv("GRPC_PORT", "50051")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Fatal("failed to listen", zap.String("port", grpcPort), zap.Error(err))
	}
	s := grpc.NewServer()
	productpb.RegisterProductServiceServer(s, grpcSrv)
	reflection.Register(s)

	httpPort := getEnv("HTTP_PORT", "8082")
	httpSrv := &http.Server{
		Addr: fmt.Sprintf(":%s", httpPort),
		Handler: resttransport.NewRouter(
			resttransport.NewServer(writeSvc, readSvc, logger),
			resttransport.NewMiddleware(validator),
		),
	}

	logger.Info("servers starting",
		zap.String("grpc_port", grpcPort),
		zap.String("http_port", httpPort),
	)

	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Fatal("failed to serve grpc", zap.Error(err))
		}
	}()
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to serve http", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down servers")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
	s.GracefulStop()
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
