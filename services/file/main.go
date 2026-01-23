package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/hmquannnnn/e-commerce/file-service/config"
	"github.com/hmquannnnn/e-commerce/file-service/routes"
	"github.com/hmquannnnn/e-commerce/file-service/service"
	commonstorage "github.com/hmquannnnn/e-commerce/pkg/storage"
	commonminio "github.com/hmquannnnn/e-commerce/pkg/storage/minio"
	"github.com/joho/godotenv"
)

func runServer() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting %s in %s mode...", cfg.App.Name, cfg.App.Environment)

	// Initialize MinIO client
	minioConfig := commonminio.NewConfigFromEnv()
	log.Println("minioConfig", minioConfig)
	minioClient, err := commonminio.NewClient(minioConfig)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO: %v", err)
	}

	// Ensure bucket exists
	bucketName := cfg.Storage.BucketName
	if err := minioClient.EnsureBucket(context.Background(), bucketName); err != nil {
		log.Fatalf("Failed to ensure bucket exists: %v", err)
	}
	log.Println("✓ MinIO connected successfully")

	// Initialize storage manager
	storageManager := commonstorage.NewManager(minioClient)
	log.Println("✓ Storage manager initialized")

	// Initialize services
	fileService := service.NewFileService(
		storageManager,
		bucketName,
	)
	log.Println("✓ Services initialized")

	// Initialize router
	router := routes.NewRouter(fileService, cfg.JWT.SecretKey)
	handler := router.SetupRoutes()
	log.Println("✓ Routes configured")

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	log.Printf("🚀 Server starting on port %s", cfg.App.Port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func main() {
	runServer()
}
