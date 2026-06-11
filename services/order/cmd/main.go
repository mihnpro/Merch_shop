package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	orderpb "github.com/mihnpro/Merch_shop/services/order/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/order/internal/app/query"
	"github.com/mihnpro/Merch_shop/services/order/internal/app/service"
	"github.com/mihnpro/Merch_shop/services/order/internal/infrastructure/account"
	"github.com/mihnpro/Merch_shop/services/order/internal/infrastructure/clients"
	kafkainfra "github.com/mihnpro/Merch_shop/services/order/internal/infrastructure/kafka"
	"github.com/mihnpro/Merch_shop/services/order/internal/infrastructure/persistence"
	psqlrepo "github.com/mihnpro/Merch_shop/services/order/internal/infrastructure/repository/psql"
	grpctransport "github.com/mihnpro/Merch_shop/services/order/internal/infrastructure/transport/grpc"
	resttransport "github.com/mihnpro/Merch_shop/services/order/internal/infrastructure/transport/rest"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Info("no .env file loaded", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
		logger.Fatal("postgres connect", zap.Error(err))
	}
	defer db.Close()

	cartClient, err := clients.NewCartClient(getEnv("CART_SERVICE_ADDR", "cart:50055"))
	if err != nil {
		logger.Fatal("cart client", zap.Error(err))
	}

	userClient, err := clients.NewUserClient(getEnv("USER_SERVICE_ADDR", "user:50051"))
	if err != nil {
		logger.Fatal("user client", zap.Error(err))
	}

	inventoryClient, err := clients.NewInventoryClient(getEnv("INVENTORY_SERVICE_ADDR", "inventory:50051"))
	if err != nil {
		logger.Fatal("inventory client", zap.Error(err))
	}

	outboxRepo := psqlrepo.NewOutboxRepository(db)
	orderRepo := psqlrepo.NewOrderRepository(db, outboxRepo)
	orderSvc := service.NewOrderService(orderRepo, cartClient, userClient, inventoryClient)
	queryStr := query.NewOrderQueryService(orderRepo)

	// Kafka outbox worker
	brokers := strings.Split(getEnv("KAFKA_BROKERS", "kafka:9092"), ",")
	outboxWorker := kafkainfra.NewOutboxWorker(outboxRepo, brokers, "order.created", logger)
	go outboxWorker.Run(ctx)

	// gRPC server (for Inventory callback)
	grpcSrv := grpctransport.NewGRPCServer(orderSvc)
	grpcPort := getEnv("GRPC_PORT", "50056")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Fatal("grpc listen", zap.String("port", grpcPort), zap.Error(err))
	}
	s := grpc.NewServer()
	orderpb.RegisterOrderServiceServer(s, grpcSrv)
	reflection.Register(s)

	// HTTP server
	validator := account.NewValidator(os.Getenv("JWT_ACCESS_SECRET"))
	httpPort := getEnv("HTTP_PORT", "8086")
	httpSrv := &http.Server{
		Addr: fmt.Sprintf(":%s", httpPort),
		Handler: resttransport.NewRouter(
			resttransport.NewServer(orderSvc, queryStr, logger),
			resttransport.NewMiddleware(validator),
		),
	}

	logger.Info("servers starting",
		zap.String("grpc_port", grpcPort),
		zap.String("http_port", httpPort),
	)

	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Fatal("grpc serve", zap.Error(err))
		}
	}()
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http serve", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutCtx)
	s.GracefulStop()
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
