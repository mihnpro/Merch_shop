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

	inventorypb "github.com/mihnpro/Merch_shop/services/invetory/api/server/AccountInternal"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/app/query"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/app/service"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/infrastructure/account"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/infrastructure/clients"
	kafkainfra "github.com/mihnpro/Merch_shop/services/invetory/internal/infrastructure/kafka"
	"github.com/mihnpro/Merch_shop/services/invetory/internal/infrastructure/persistence"
	psqlrepo "github.com/mihnpro/Merch_shop/services/invetory/internal/infrastructure/repository/psql"
	grpctransport "github.com/mihnpro/Merch_shop/services/invetory/internal/infrastructure/transport/grpc"
	resttransport "github.com/mihnpro/Merch_shop/services/invetory/internal/infrastructure/transport/rest"
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
		logger.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer db.Close()

	stockRepo := psqlrepo.NewStockRepository(db)
	reservationRepo := psqlrepo.NewReservationRepository(db)

	orderClient, err := clients.NewOrderClient(getEnv("ORDER_SERVICE_ADDR", "order:50056"))
	if err != nil {
		logger.Fatal("order client", zap.Error(err))
	}

	writeSvc := service.NewInventoryService(stockRepo, reservationRepo, orderClient, logger)
	readSvc := query.NewInventoryReadService(stockRepo)
	validator := account.NewValidator(os.Getenv("JWT_ACCESS_SECRET"))

	brokers := strings.Split(getEnv("KAFKA_BROKERS", "kafka:9092"), ",")
	consumer := kafkainfra.NewConsumer(brokers, writeSvc, logger)
	go consumer.Run(ctx)

	grpcSrv := grpctransport.NewGRPCServer(writeSvc, readSvc)
	grpcPort := getEnv("GRPC_PORT", "50051")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Fatal("failed to listen", zap.String("port", grpcPort), zap.Error(err))
	}
	s := grpc.NewServer()
	inventorypb.RegisterInventoryServiceServer(s, grpcSrv)
	reflection.Register(s)

	httpPort := getEnv("HTTP_PORT", "8081")
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

	<-ctx.Done()

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
