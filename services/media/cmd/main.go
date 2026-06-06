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

	mediapb "github.com/mihnpro/Merch_shop/services/media/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/media/internal/app/service"
	"github.com/mihnpro/Merch_shop/services/media/internal/infrastructure/account"
	miniostore "github.com/mihnpro/Merch_shop/services/media/internal/infrastructure/storage/minio"
	grpctransport "github.com/mihnpro/Merch_shop/services/media/internal/infrastructure/transport/grpc"
	resttransport "github.com/mihnpro/Merch_shop/services/media/internal/infrastructure/transport/rest"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Info("no .env file loaded", zap.Error(err))
	}

	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	if accessSecret == "" {
		logger.Fatal("JWT_ACCESS_SECRET is required")
	}
	verifier := account.NewVerifier(accessSecret)

	store, err := miniostore.New(miniostore.Config{
		Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		SecretKey: os.Getenv("MINIO_SECRET_KEY"),
		UseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
		Bucket:    getEnv("MINIO_BUCKET", "merch-products"),
	})
	if err != nil {
		logger.Fatal("failed to init minio client", zap.Error(err))
	}

	ctx := context.Background()
	if err := store.EnsureBucket(ctx); err != nil {
		logger.Fatal("failed to ensure bucket", zap.Error(err))
	}

	uploader := service.NewUploader(store)

	grpcPort := getEnv("GRPC_PORT", "50051")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Fatal("failed to listen", zap.String("port", grpcPort), zap.Error(err))
	}

	s := grpc.NewServer(
		grpc.StreamInterceptor(grpctransport.AdminOnlyStreamInterceptor(verifier)),
	)
	mediapb.RegisterMediaServiceServer(s, grpctransport.NewGRPCServer(uploader, logger))
	reflection.Register(s)

	httpPort := getEnv("HTTP_PORT", "8083")
	httpSrv := &http.Server{
		Addr: fmt.Sprintf(":%s", httpPort),
		Handler: resttransport.NewRouter(
			resttransport.NewServer(uploader, logger),
			resttransport.NewMiddleware(verifier),
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
