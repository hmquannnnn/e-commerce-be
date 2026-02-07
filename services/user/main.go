package main

import (
	"log"
	"net/http"
	"time"

	"github.com/hmquannnnn/e-commerce/user-service/config"
	"github.com/hmquannnnn/e-commerce/user-service/internal/db"
	"github.com/hmquannnnn/e-commerce/user-service/internal/util"
	"github.com/hmquannnnn/e-commerce/user-service/repository"
	"github.com/hmquannnnn/e-commerce/user-service/routes"
	"github.com/hmquannnnn/e-commerce/user-service/service"
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

	// Initialize database connection
	database, err := db.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Println("✓ Database connected successfully")

	// Note: Redis cache removed for simplicity. You can add it later when learning about caching.
	// See internal/cache/redis.go for example implementation

	// Initialize JWT manager
	jwtManager := util.NewJWTManager(
		cfg.JWT.SecretKey,
		cfg.JWT.AccessTokenDuration,
		cfg.JWT.RefreshTokenDuration,
	)
	log.Println("✓ JWT manager initialized")

	// Initialize repositories
	userRepo := repository.NewUserRepository(database)
	refreshTokenRepo := repository.NewRefreshTokenRepository(database)
	log.Println("✓ Repositories initialized")

	// Initialize services
	authService := service.NewAuthService(
		userRepo,
		refreshTokenRepo,
		jwtManager,
	)
	userService := service.NewUserService(
		userRepo,
	)
	log.Println("✓ Services initialized")

	// Initialize router
	router := routes.NewRouter(authService, userService, jwtManager)
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
