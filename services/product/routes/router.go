package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hmquannnnn/e-commerce/product-service/handler"
	"github.com/hmquannnnn/e-commerce/product-service/service"
)

type Router struct {
	categoryHandler  *handler.CategoryHandler
	productHandler   *handler.ProductHandler
	inventoryHandler *handler.InventoryHandler
}

func NewRouter(
	categoryService service.CategoryService,
	productService service.ProductService,
	inventoryService service.InventoryService,
) *Router {
	return &Router{
		categoryHandler:  handler.NewCategoryHandler(categoryService),
		productHandler:   handler.NewProductHandler(productService),
		inventoryHandler: handler.NewInventoryHandler(inventoryService),
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	router := gin.Default()

	router.Use(RequestIDMiddleware())
	router.Use(CORSMiddleware())

	router.GET("/health", r.healthCheck)
	router.GET("/api/health", r.healthCheck)

	v1 := router.Group("/api")
	{
		// Categories
		categories := v1.Group("/categories")
		{
			// Public
			categories.GET("", r.categoryHandler.List)
			categories.GET("/:id", r.categoryHandler.GetByID)

			// Admin only
			adminCategories := categories.Group("")
			adminCategories.Use(RequireAuth(), RequireAdmin())
			{
				adminCategories.POST("", r.categoryHandler.Create)
				adminCategories.PATCH("/:id", r.categoryHandler.Update)
				adminCategories.DELETE("/:id", r.categoryHandler.Delete)
			}
		}

		// Products
		products := v1.Group("/products")
		{
			// Public
			products.GET("", r.productHandler.List)
			products.GET("/:id", r.productHandler.GetByID)

			// Admin only
			adminProducts := products.Group("")
			adminProducts.Use(RequireAuth(), RequireAdmin())
			{
				adminProducts.POST("", r.productHandler.Create)
				adminProducts.PATCH("/:id", r.productHandler.Update)
				adminProducts.DELETE("/:id", r.productHandler.Delete)
				adminProducts.POST("/:id/images", r.productHandler.AddImage)
				adminProducts.DELETE("/:id/images/:image_id", r.productHandler.DeleteImage)
			}
		}

		// Inventory - admin only
		inventory := v1.Group("/inventory")
		inventory.Use(RequireAuth(), RequireAdmin())
		{
			inventory.GET("/:product_id", r.inventoryHandler.GetByProductID)
			inventory.PATCH("/:product_id/stock", r.inventoryHandler.UpdateStock)
		}

		// Internal stock management (service-to-service)
		internal := v1.Group("/internal/inventory")
		internal.Use(RequireAuth())
		{
			internal.POST("/reserve", r.inventoryHandler.ReserveStock)
			internal.POST("/release", r.inventoryHandler.ReleaseStock)
		}
	}

	return router
}

func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Service is healthy",
		"data": gin.H{
			"service": "product-service",
			"status":  "ok",
		},
	})
}
