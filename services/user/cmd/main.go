package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	userpb "github.com/mihnpro/Merch_shop/services/user_customer/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/query"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/app/service"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/infrastructure/account"
	"github.com/mihnpro/Merch_shop/services/user_customer/internal/infrastructure/persistence"
	psqlrepo "github.com/mihnpro/Merch_shop/services/user_customer/internal/infrastructure/repository/psql"
	redisrepo "github.com/mihnpro/Merch_shop/services/user_customer/internal/infrastructure/repository/redis"
	grpctransport "github.com/mihnpro/Merch_shop/services/user_customer/internal/infrastructure/transport/grpc"
	resttransport "github.com/mihnpro/Merch_shop/services/user_customer/internal/infrastructure/transport/rest"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
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

	redisClient, err := persistence.NewRedis(ctx, persistence.RedisConfig{
		URL:      os.Getenv("REDIS_URL"),
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       atoiOr(os.Getenv("REDIS_DB"), 0),
	})
	if err != nil {
		logger.Fatal("failed to connect to redis", zap.Error(err))
	}
	defer redisClient.Close()

	userRepo := psqlrepo.NewUserRepository(db)
	tokenStore := redisrepo.NewTokenStore(redisClient, logger)
	acct := account.NewAccount(
		os.Getenv("JWT_ACCESS_SECRET"),
		os.Getenv("JWT_REFRESH_SECRET"),
		accessTokenTTL,
		refreshTokenTTL,
	)

	readSvc := query.NewAuthReadService(userRepo)
	querySvc := service.NewAuthService(userRepo, readSvc, acct, tokenStore, refreshTokenTTL)

	grpcSrv := grpctransport.NewGRPCServer(querySvc)

	grpcPort := getEnv("GRPC_PORT", "50051")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Fatal("failed to listen", zap.String("port", grpcPort), zap.Error(err))
	}

	s := grpc.NewServer()
	userpb.RegisterUserServiceServer(s, grpcSrv)
	reflection.Register(s)

	httpPort := getEnv("HTTP_PORT", "8081")
	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%s", httpPort),
		Handler: resttransport.NewRouter(
			resttransport.NewServer(querySvc, logger),
			resttransport.NewMiddleware(acct),
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

func atoiOr(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return n
}
