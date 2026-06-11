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
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	cartpb "github.com/mihnpro/Merch_shop/services/cart/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/cart/internal/app/query"
	"github.com/mihnpro/Merch_shop/services/cart/internal/app/service"
	"github.com/mihnpro/Merch_shop/services/cart/internal/infrastructure/account"
	"github.com/mihnpro/Merch_shop/services/cart/internal/infrastructure/clients"
	"github.com/mihnpro/Merch_shop/services/cart/internal/infrastructure/persistence"
	"github.com/mihnpro/Merch_shop/services/cart/internal/domain/repository"
	psqlrepo "github.com/mihnpro/Merch_shop/services/cart/internal/infrastructure/repository/psql"
	redisrepo "github.com/mihnpro/Merch_shop/services/cart/internal/infrastructure/repository/redis"
	grpctransport "github.com/mihnpro/Merch_shop/services/cart/internal/infrastructure/transport/grpc"
	resttransport "github.com/mihnpro/Merch_shop/services/cart/internal/infrastructure/transport/rest"
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

	var rc *redis.Client
	rc, err = persistence.NewRedis(ctx, persistence.RedisConfig{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: os.Getenv("REDIS_PASSWORD"),
	})
	if err != nil {
		logger.Warn("redis unavailable, cache disabled", zap.Error(err))
		rc = nil
	}
	if rc != nil {
		defer rc.Close()
	}

	productClient, err := clients.NewProductClient(getEnv("PRODUCT_SERVICE_ADDR", "product:50052"))
	if err != nil {
		logger.Fatal("failed to connect to product service", zap.Error(err))
	}

	inventoryClient, err := clients.NewInventoryClient(getEnv("INVENTORY_SERVICE_ADDR", "inventory:50054"))
	if err != nil {
		logger.Fatal("failed to connect to inventory service", zap.Error(err))
	}

	cartRepo := psqlrepo.NewCartRepository(db)
	validator := account.NewValidator(os.Getenv("JWT_ACCESS_SECRET"))

	var cacheRepo repository.CacheRepository
	if rc != nil {
		cacheRepo = redisrepo.NewCacheRepository(rc, logger)
	}

	writeSvc := service.NewCartService(cartRepo, productClient, inventoryClient, cacheRepo, logger)
	readSvc := query.NewCartQueryService(cartRepo, cacheRepo, logger)

	grpcSrv := grpctransport.NewGRPCServer(writeSvc, readSvc)
	grpcPort := getEnv("GRPC_PORT", "50055")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Fatal("failed to listen", zap.String("port", grpcPort), zap.Error(err))
	}
	s := grpc.NewServer()
	cartpb.RegisterCartServiceServer(s, grpcSrv)
	reflection.Register(s)

	httpPort := getEnv("HTTP_PORT", "8085")
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
