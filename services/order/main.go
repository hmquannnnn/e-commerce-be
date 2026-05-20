package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"github.com/hmquannnnn/e-commerce/order-service/client"
	"github.com/hmquannnnn/e-commerce/order-service/config"
	appdb "github.com/hmquannnnn/e-commerce/order-service/internal/db"
	"github.com/hmquannnnn/e-commerce/order-service/repository"
	"github.com/hmquannnnn/e-commerce/order-service/routes"
	"github.com/hmquannnnn/e-commerce/order-service/service"
)

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

	productClient := client.NewProductClient(cfg.ProductSvc.URL)
	userClient := client.NewUserClient(cfg.UserSvc.URL)

	cartRepo := repository.NewCartRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	cartSvc := service.NewCartService(cartRepo, productClient)
	orderSvc := service.NewOrderService(orderRepo, cartRepo, productClient, userClient)

	if cfg.App.OrderExpiryEnabled {
		expiryWorker := service.NewOrderExpiryWorker(orderRepo, productClient)
		go expiryWorker.Start(context.Background())
	} else {
		slog.Warn("order expiry worker disabled by config")
	}

	router := routes.NewRouter(cartSvc, orderSvc)
	engine := router.SetupRoutes()

	addr := ":" + cfg.App.Port
	slog.Info("order-service starting", "port", cfg.App.Port, "env", cfg.App.Environment)
	if err := engine.Run(addr); err != nil {
		slog.Error("failed to run server", "error", err)
		os.Exit(1)
	}
}
