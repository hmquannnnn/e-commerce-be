package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hmquannnnn/e-commerce/order-service/handler"
	"github.com/hmquannnnn/e-commerce/order-service/service"
	"github.com/hmquannnnn/e-commerce/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Router struct {
	cartHandler     *handler.CartHandler
	orderHandler    *handler.OrderHandler
	locationHandler *handler.LocationHandler
}

func NewRouter(cartService service.CartService, orderService service.OrderService, locationService service.LocationService) *Router {
	return &Router{
		cartHandler:     handler.NewCartHandler(cartService),
		orderHandler:    handler.NewOrderHandler(orderService),
		locationHandler: handler.NewLocationHandler(locationService),
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	router := gin.Default()

	router.Use(RequestIDMiddleware())

	router.Use(metrics.Middleware())
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/health", r.healthCheck)
	router.GET("/api/health", r.healthCheck)

	locations := router.Group("/api/locations")
	{
		locations.GET("/provinces", r.locationHandler.ListProvinces)
		locations.GET("/cities", r.locationHandler.ListProvinces)
		locations.GET("/districts", r.locationHandler.ListDistricts)
		locations.GET("/wards", r.locationHandler.ListWards)
	}

	v1 := router.Group("/api")
	v1.Use(RequireAuth())
	{
		// Cart routes
		cart := v1.Group("/cart")
		{
			cart.GET("", r.cartHandler.GetCart)
			cart.POST("/items", r.cartHandler.AddItem)
			cart.PUT("/items/:product_id", r.cartHandler.UpdateItem)
			cart.DELETE("/items/:product_id", r.cartHandler.DeleteItem)
		}

		// Order routes
		orders := v1.Group("/orders")
		{
			orders.POST("", r.orderHandler.CreateOrder)
			orders.GET("", r.orderHandler.ListUserOrders)
			orders.GET("/:id", r.orderHandler.GetOrder)
			orders.PATCH("/:id/cancel", r.orderHandler.CancelOrder)
		}

		// Admin order routes
		adminOrders := v1.Group("/admin/orders")
		adminOrders.Use(RequireAdmin())
		{
			adminOrders.GET("", r.orderHandler.AdminListOrders)
			adminOrders.GET("/:id", r.orderHandler.AdminGetOrder)
			adminOrders.PATCH("/:id/status", r.orderHandler.AdminUpdateStatus)
		}
	}

	return router
}

func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Service is healthy",
		"data": gin.H{
			"service": "order-service",
			"status":  "ok",
		},
	})
}
