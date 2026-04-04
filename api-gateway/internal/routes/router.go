package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hmquannnnn/e-commerce/api-gateway/internal/middleware"
	"github.com/hmquannnnn/e-commerce/api-gateway/internal/proxy"
)

// SetupRouter configures all routes for the API Gateway
func SetupRouter(userServiceURL string, fileServiceURL string, productServiceURL string, orderServiceURL string, jwtSecret string) *gin.Engine {
	r := gin.Default()

	// Global middlewares
	r.Use(middleware.CORSMiddleware())

	// Health check endpoint
	r.GET("/health", healthCheck)
	r.GET("/api/health", healthCheck)

	userServiceReverseProxy := proxy.NewReverseProxy(userServiceURL)
	fileServiceReverseProxy := proxy.NewReverseProxy(fileServiceURL)
	productServiceReverseProxy := proxy.NewReverseProxy(productServiceURL)
	orderServiceReverseProxy := proxy.NewReverseProxy(orderServiceURL)

	// API v1 group
	v1 := r.Group("/api")
	{
		// Public auth routes
		authPublic := v1.Group("/auth")
		{
			authPublic.POST("/register", userServiceReverseProxy)
			authPublic.POST("/login", userServiceReverseProxy)
			authPublic.POST("/refresh", userServiceReverseProxy)
		}

		// Public product routes (no auth required for browsing)
		// new-id must be before /:id to avoid param conflict
		v1.GET("/products/new-id", productServiceReverseProxy)
		v1.GET("/products", productServiceReverseProxy)
		v1.GET("/products/:id", productServiceReverseProxy)
		v1.GET("/categories", productServiceReverseProxy)
		v1.GET("/categories/:id", productServiceReverseProxy)

		// Protected routes - Authentication required
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		{
			// Auth protected
			authProtected := protected.Group("/auth")
			{
				authProtected.POST("/logout", userServiceReverseProxy)
				authProtected.GET("/me", userServiceReverseProxy)
			}

			// User management
			users := protected.Group("/users")
			{
				users.Any("/*path", userServiceReverseProxy)
			}

			// File management
			files := protected.Group("/files")
			{
				files.Any("/*path", fileServiceReverseProxy)
			}

			// Product admin routes (explicit methods to avoid GET conflict)
			products := protected.Group("/products")
			{
				products.POST("", productServiceReverseProxy)
				products.PATCH("/:id", productServiceReverseProxy)
				products.DELETE("/:id", productServiceReverseProxy)
				products.POST("/:id/images", productServiceReverseProxy)
				products.DELETE("/:id/images/:image_id", productServiceReverseProxy)
			}

			// Category admin routes
			categories := protected.Group("/categories")
			{
				categories.POST("", productServiceReverseProxy)
				categories.PATCH("/:id", productServiceReverseProxy)
				categories.DELETE("/:id", productServiceReverseProxy)
			}

			// Inventory admin routes
			inventory := protected.Group("/inventory")
			{
				inventory.GET("/:product_id", productServiceReverseProxy)
				inventory.PATCH("/:product_id/stock", productServiceReverseProxy)
			}

			// Internal inventory (service-to-service)
			internal := protected.Group("/internal/inventory")
			{
				internal.POST("/reserve", productServiceReverseProxy)
				internal.POST("/release", productServiceReverseProxy)
			}

			// Cart routes
			cart := protected.Group("/cart")
			{
				cart.Any("", orderServiceReverseProxy)
				cart.Any("/*path", orderServiceReverseProxy)
			}

			// Order routes
			orders := protected.Group("/orders")
			{
				orders.Any("", orderServiceReverseProxy)
				orders.Any("/*path", orderServiceReverseProxy)
			}

			// Admin order routes
			adminOrders := protected.Group("/admin/orders")
			{
				adminOrders.Any("", orderServiceReverseProxy)
				adminOrders.Any("/*path", orderServiceReverseProxy)
			}
		}
	}

	return r
}

// healthCheck returns the health status of the API Gateway
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "API Gateway is healthy",
		"data": gin.H{
			"service": "api-gateway",
			"status":  "ok",
		},
	})
}
