package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hmquannnnn/e-commerce/file-service/handler"
	"github.com/hmquannnnn/e-commerce/file-service/service"
)

type Router struct {
	fileHandler *handler.FileHandler
}

func NewRouter(
	fileService service.FileService,
	jwtSecret string,
) *Router {
	return &Router{
		fileHandler: handler.NewFileHandler(fileService),
	}
}

func (router *Router) SetupRoutes() *gin.Engine {
	r := gin.Default()

	r.Use(RequestIDMiddleware())

	r.GET("/health", router.healthCheck)
	r.GET("/api/health", router.healthCheck)

	v1 := r.Group("/api")
	{
		// Note: Authentication is handled by API Gateway
		// This service assumes requests are already authenticated
		files := v1.Group("/files")
		{
			files.POST("/presigned-url", router.fileHandler.GetPresignedUploadURL)
		}
	}

	return r
}

func (router *Router) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Service is healthy",
		"data": gin.H{
			"service": "file-service",
			"status":  "ok",
		},
	})
}
