package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"os"

	"github.com/joho/godotenv"

	"github.com/hmquannnnn/e-commerce/order-service/client"
	"github.com/hmquannnnn/e-commerce/order-service/config"
	grpcserver "github.com/hmquannnnn/e-commerce/order-service/grpc"
	appdb "github.com/hmquannnnn/e-commerce/order-service/internal/db"
	"github.com/hmquannnnn/e-commerce/order-service/repository"
	"github.com/hmquannnnn/e-commerce/order-service/routes"
	"github.com/hmquannnnn/e-commerce/order-service/service"
	orderpb "github.com/hmquannnnn/e-commerce/pkg/proto/order"
	"google.golang.org/grpc"
)

func startGRPCServer(port string, orderService service.OrderService) *grpc.Server {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("failed to listen on grpc port", "error", err, "port", port)
		os.Exit(1)
	}

	server := grpc.NewServer()
	orderpb.RegisterOrderServiceServer(server, grpcserver.NewOrderServer(orderService))

	go func() {
		slog.Info("order-service grpc starting", "port", port)
		if err := server.Serve(listener); err != nil {
			slog.Error("failed to run grpc server", "error", err)
			os.Exit(1)
		}
	}()

	return server
}

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

	productClient, err := client.NewProductClient(cfg.ProductSvc.URL, cfg.ProductSvc.GRPCAddress)
	if err != nil {
		slog.Error("failed to create product-service client", "error", err)
		os.Exit(1)
	}
	defer productClient.Close()
	userClient := client.NewUserClient(cfg.UserSvc.URL)

	cartRepo := repository.NewCartRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	locationRepo := repository.NewLocationRepository(db)

	cartSvc := service.NewCartService(cartRepo, productClient)
	orderSvc := service.NewOrderService(orderRepo, cartRepo, productClient, userClient)
	locationSvc := service.NewLocationService(locationRepo)

	grpcServer := startGRPCServer(cfg.App.GRPCPort, orderSvc)
	defer grpcServer.GracefulStop()

	if cfg.App.OrderExpiryEnabled {
		expiryWorker := service.NewOrderExpiryWorker(orderRepo, productClient)
		go expiryWorker.Start(context.Background())
	} else {
		slog.Warn("order expiry worker disabled by config")
	}

	router := routes.NewRouter(cartSvc, orderSvc, locationSvc)
	engine := router.SetupRoutes()

	addr := ":" + cfg.App.Port
	slog.Info("order-service starting", "port", cfg.App.Port, "env", cfg.App.Environment)
	if err := engine.Run(addr); err != nil {
		slog.Error("failed to run server", "error", err)
		os.Exit(1)
	}
}
