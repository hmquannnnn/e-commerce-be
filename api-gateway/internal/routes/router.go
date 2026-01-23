package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hmquannnnn/e-commerce/api-gateway/internal/middleware"
	"github.com/hmquannnnn/e-commerce/api-gateway/internal/proxy"
)

// SetupRouter configures all routes for the API Gateway
func SetupRouter(userServiceURL string, fileServiceURL string, jwtSecret string) *gin.Engine {
	r := gin.Default()

	// Global middlewares
	r.Use(middleware.CORSMiddleware())

	// Health check endpoint
	r.GET("/health", healthCheck)
	r.GET("/api/health", healthCheck)

	userServiceReverseProxy := proxy.NewReverseProxy(userServiceURL)
	fileServiceReverseProxy := proxy.NewReverseProxy(fileServiceURL)

	// API v1 group
	v1 := r.Group("/api")
	{
		// Public auth routes - No authentication required
		// Routes: POST /api/auth/register, POST /api/auth/login, POST /api/auth/refresh
		authPublic := v1.Group("/auth")
		{
			authPublic.POST("/register", userServiceReverseProxy)
			authPublic.POST("/login", userServiceReverseProxy)
			authPublic.POST("/refresh", userServiceReverseProxy)
		}

		// Protected routes - Authentication required
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		{
			// Protected auth routes: logout, me
			// Routes: POST /api/auth/logout, GET /api/auth/me
			authProtected := protected.Group("/auth")
			{
				authProtected.POST("/logout", userServiceReverseProxy)
				authProtected.GET("/me", userServiceReverseProxy)
			}

			// User management routes
			// Route: PATCH /api/users/profile
			users := protected.Group("/users")
			{
				users.Any("/*path", userServiceReverseProxy)
			}

			// File management routes
			// Route: POST /api/files/presigned-url
			files := protected.Group("/files")
			{
				files.Any("/*path", fileServiceReverseProxy)
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
