package main

import (
	"log"
	"net/http"
	"time"

	"github.com/hmquannnnn/e-commerce/pkg/storage"
	"github.com/hmquannnnn/e-commerce/pkg/storage/minio"
	"github.com/hmquannnnn/e-commerce/product-service/config"
	"github.com/hmquannnnn/e-commerce/product-service/internal/db"
	"github.com/hmquannnnn/e-commerce/product-service/repository"
	"github.com/hmquannnnn/e-commerce/product-service/routes"
	"github.com/hmquannnnn/e-commerce/product-service/service"
	"github.com/joho/godotenv"
)

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

	categoryService := service.NewCategoryService(categoryRepo)
	inventoryService := service.NewInventoryService(inventoryRepo)
	productService := service.NewProductService(productRepo, inventoryRepo, minioClient, cfg.MinIO.BucketName)
	log.Println("✓ Services initialized")

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
