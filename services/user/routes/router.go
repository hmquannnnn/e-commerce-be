package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hmquannnnn/e-commerce/user-service/handler"
	"github.com/hmquannnnn/e-commerce/user-service/internal/util"
	"github.com/hmquannnnn/e-commerce/user-service/service"
)

type Router struct {
	authHandler    *handler.AuthHandler
	userHandler    *handler.UserHandler
	authMiddleware *AuthMiddleware
}

func NewRouter(
	authService service.AuthService,
	userService service.UserService,
	jwtManager *util.JWTManager,
) *Router {
	return &Router{
		authHandler:    handler.NewAuthHandler(authService, userService, jwtManager),
		userHandler:    handler.NewUserHandler(userService),
		authMiddleware: NewAuthMiddleware(jwtManager),
	}
}

func (router *Router) SetupRoutes() *gin.Engine {
	r := gin.Default()

	r.Use(RequestIDMiddleware())

	r.GET("/health", router.healthCheck)
	r.GET("/api/health", router.healthCheck)

	v1 := r.Group("/api")
	{
		auth := v1.Group("/auth")
		{
			// Public routes
			auth.POST("/register", router.authHandler.Register)
			auth.POST("/login", router.authHandler.Login)
			auth.POST("/refresh", router.authHandler.RefreshToken)

			// Protected routes
			protected := auth.Group("")
			protected.Use(router.authMiddleware.RequireAuth())
			{
				protected.POST("/logout", router.authHandler.Logout)
				protected.GET("/me", router.authHandler.GetMe)
			}
		}

		users := v1.Group("/users")
		users.Use(router.authMiddleware.RequireAuth())
		{
			users.PATCH("/profile", router.userHandler.UpdateProfile)
		}
	}

	internal := r.Group("/internal")
	{
		internal.GET("/users", router.userHandler.InternalListUsers)
		internal.GET("/users/search", router.userHandler.InternalSearchUsers)
	}

	return r
}

func (router *Router) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Service is healthy",
		"data": gin.H{
			"service": "user-service",
			"status":  "ok",
		},
	})
}
