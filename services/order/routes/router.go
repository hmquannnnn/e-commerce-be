package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hmquannnnn/e-commerce/order-service/handler"
	"github.com/hmquannnnn/e-commerce/order-service/service"
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

	// Internal routes (không cần auth — chỉ dùng cho service-to-service call)
	internal := router.Group("/internal")
	{
		internal.GET("/orders/:id", r.orderHandler.InternalGetOrder)
		// Replaces the old PATCH /internal/orders/:id/status — payment-service only
		// ever needs to flip PENDING → PAID, and the narrow endpoint makes the
		// state-machine guard explicit.
		internal.POST("/orders/:id/mark-paid", r.orderHandler.InternalMarkPaid)
		internal.POST("/orders/:id/touch-payment-deadline", r.orderHandler.InternalTouchPaymentDeadline)
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
