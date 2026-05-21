package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"github.com/hmquannnnn/e-commerce/payment-service/client"
	"github.com/hmquannnnn/e-commerce/payment-service/config"
	appdb "github.com/hmquannnnn/e-commerce/payment-service/internal/db"
	"github.com/hmquannnnn/e-commerce/payment-service/provider"
	"github.com/hmquannnnn/e-commerce/payment-service/repository"
	"github.com/hmquannnnn/e-commerce/payment-service/routes"
	"github.com/hmquannnnn/e-commerce/payment-service/service"
)

// test jenkins pipeline
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := appdb.NewPostgresDB(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("connected to database", "host", cfg.Database.Host, "db", cfg.Database.DBName)

	repo := repository.NewPaymentRepository(db)
	payosProvider, err := provider.NewPayOSProvider(cfg.PayOS)
	if err != nil {
		slog.Error("failed to init PayOS provider", "error", err)
		os.Exit(1)
	}
	orderClient, err := client.NewOrderClient(cfg.OrderServiceGRPCAddress)
	if err != nil {
		slog.Error("failed to create order-service grpc client", "error", err)
		os.Exit(1)
	}
	defer orderClient.Close()
	paymentSvc := service.NewPaymentService(repo, payosProvider, orderClient)

	slog.Info("order-service grpc client configured", "address", cfg.OrderServiceGRPCAddress)

	router := routes.NewRouter(paymentSvc)
	engine := router.SetupRoutes()

	addr := ":" + cfg.App.Port
	slog.Info("payment-service starting", "port", cfg.App.Port, "env", cfg.App.Environment)
	if err := engine.Run(addr); err != nil {
		slog.Error("failed to run server", "error", err)
		os.Exit(1)
	}
}
