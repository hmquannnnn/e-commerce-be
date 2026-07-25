package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hmquannnnn/e-commerce/payment-service/handler"
	"github.com/hmquannnnn/e-commerce/payment-service/service"
	"github.com/hmquannnnn/e-commerce/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Router struct {
	paymentHandler *handler.PaymentHandler
}

func NewRouter(paymentService service.PaymentService) *Router {
	return &Router{
		paymentHandler: handler.NewPaymentHandler(paymentService),
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	router := gin.Default()
	router.Use(RequestIDMiddleware())

	router.Use(metrics.Middleware())
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/health", r.healthCheck)
	router.GET("/api/health", r.healthCheck)

	api := router.Group("/api")
	{
		api.POST("/payments/webhook/payos", r.paymentHandler.PayOSWebhook)
		api.GET("/payments/webhook/payos", r.paymentHandler.PayOSWebhookProbe)

		secured := api.Group("")
		secured.Use(RequireAuth())
		{
			payments := secured.Group("/payments")
			{
				payments.POST("", r.paymentHandler.CreatePayment)
				payments.GET("/:id", r.paymentHandler.GetPayment)
				payments.GET("/order/:orderId", r.paymentHandler.GetPaymentByOrderID)
				payments.POST("/:id/cancel", r.paymentHandler.CancelPayment)
			}
		}
	}
	return router
}

func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Service is healthy",
		"data": gin.H{
			"service": "payment-service",
			"status":  "ok",
		},
	})
}
