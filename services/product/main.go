package main

import (
	"log"
	"net"
	"net/http"
	"time"

	inventorypb "github.com/hmquannnnn/e-commerce/pkg/proto/inventory"
	"github.com/hmquannnnn/e-commerce/pkg/storage"
	"github.com/hmquannnnn/e-commerce/pkg/storage/minio"
	"github.com/hmquannnnn/e-commerce/product-service/config"
	grpcserver "github.com/hmquannnnn/e-commerce/product-service/grpc"
	"github.com/hmquannnnn/e-commerce/product-service/internal/db"
	"github.com/hmquannnnn/e-commerce/product-service/repository"
	"github.com/hmquannnnn/e-commerce/product-service/routes"
	"github.com/hmquannnnn/e-commerce/product-service/service"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func startGRPCServer(port string, inventoryService service.InventoryService) *grpc.Server {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", port, err)
	}

	server := grpc.NewServer()
	inventorypb.RegisterInventoryServiceServer(server, grpcserver.NewInventoryServer(inventoryService))

	go func() {
		log.Printf("gRPC server starting on port %s", port)
		if err := server.Serve(listener); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	return server
}

func runServer() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting %s in %s mode...", cfg.App.Name, cfg.App.Environment)

	database, err := db.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Println("✓ Database connected successfully")

	minioClient, err := minio.NewClient(minio.Config{
		Endpoint:        cfg.MinIO.Endpoint,
		AccessKey:       cfg.MinIO.AccessKey,
		SecretAccessKey: cfg.MinIO.SecretAccessKey,
		UseSSL:          cfg.MinIO.UseSSL,
	})
	if err != nil {
		log.Printf("Warning: failed to connect to MinIO: %v — folder auto-creation disabled", err)
		minioClient = nil
	} else {
		log.Println("✓ MinIO connected successfully")
	}

	var storageMgr *storage.Manager
	if minioClient != nil {
		storageMgr = storage.NewManager(minioClient)
	}

	categoryRepo := repository.NewCategoryRepository(database)
	productRepo := repository.NewProductRepository(database, storageMgr)
	inventoryRepo := repository.NewInventoryRepository(database)
	log.Println("✓ Repositories initialized")

	categoryService := service.NewCategoryService(categoryRepo, productRepo)
	inventoryService := service.NewInventoryService(inventoryRepo)
	productService := service.NewProductService(productRepo, inventoryRepo, minioClient, cfg.MinIO.BucketName)
	log.Println("✓ Services initialized")

	grpcServer := startGRPCServer(cfg.App.GRPCPort, inventoryService)
	defer grpcServer.GracefulStop()

	router := routes.NewRouter(categoryService, productService, inventoryService)
	handler := router.SetupRoutes()
	log.Println("✓ Routes configured")

	server := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀 Server starting on port %s", cfg.App.Port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func main() {
	runServer()
}
