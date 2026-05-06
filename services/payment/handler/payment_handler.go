package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/payment-service/model"
	"github.com/hmquannnnn/e-commerce/payment-service/service"
)

type PaymentHandler struct {
	paymentService service.PaymentService
}

func NewPaymentHandler(paymentService service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	p, err := h.paymentService.CreatePayment(c.Request.Context(), service.CreatePaymentInput{
		OrderID:       req.OrderID,
		UserID:        userID,
		Provider:      model.Provider(req.Provider),
		PaymentMethod: model.PaymentMethod(req.PaymentMethod),
		Amount:        req.Amount,
		Currency:      req.Currency,
		ReturnURL:     req.ReturnURL,
		CancelURL:     req.CancelURL,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			respondError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid payment input")
			return
		}
		if errors.Is(err, service.ErrUnsupported) {
			respondError(c, http.StatusBadRequest, "UNSUPPORTED_PROVIDER", "Unsupported payment provider")
			return
		}
		if errors.Is(err, service.ErrOrderNotFound) {
			respondError(c, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			respondError(c, http.StatusForbidden, "FORBIDDEN", "Access denied")
			return
		}
		if errors.Is(err, service.ErrOrderNotPayable) {
			respondError(c, http.StatusConflict, "ORDER_NOT_PAYABLE", "Order is not in a payable state")
			return
		}
		if errors.Is(err, service.ErrAmountMismatch) {
			respondError(c, http.StatusBadRequest, "AMOUNT_MISMATCH", "Amount does not match order total")
			return
		}
		slog.Error("create payment failed", "error", err, "order_id", req.OrderID, "provider", req.Provider)
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create payment")
		return
	}

	respondCreated(c, "Payment created successfully", ToPaymentResponse(p))
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid payment ID")
		return
	}
	p, err := h.paymentService.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, service.ErrPaymentNotFound) {
			respondError(c, http.StatusNotFound, "PAYMENT_NOT_FOUND", "Payment not found")
			return
		}
		slog.Error("get payment failed", "error", err, "payment_id", c.Param("id"))
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get payment")
		return
	}
	respondOK(c, "Payment retrieved successfully", ToPaymentResponse(p))
}

func (h *PaymentHandler) GetPaymentByOrderID(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	orderID := c.Param("orderId")
	p, err := h.paymentService.GetByOrderID(c.Request.Context(), orderID, userID)
	if err != nil {
		if errors.Is(err, service.ErrPaymentNotFound) {
			respondError(c, http.StatusNotFound, "PAYMENT_NOT_FOUND", "Payment not found")
			return
		}
		slog.Error("get payment by order failed", "error", err, "order_id", orderID)
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get payment")
		return
	}
	respondOK(c, "Payment retrieved successfully", ToPaymentResponse(p))
}

func (h *PaymentHandler) CancelPayment(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid payment ID")
		return
	}
	if err := h.paymentService.CancelPayment(c.Request.Context(), id, userID); err != nil {
		if errors.Is(err, service.ErrPaymentNotFound) {
			respondError(c, http.StatusNotFound, "PAYMENT_NOT_FOUND", "Payment not found")
			return
		}
		slog.Error("cancel payment failed", "error", err, "payment_id", c.Param("id"))
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cancel payment")
		return
	}
	respondOK(c, "Payment cancelled successfully", nil)
}

func (h *PaymentHandler) PayOSWebhook(c *gin.Context) {
	if err := h.paymentService.HandlePayOSWebhook(c.Request); err != nil {
		slog.Error("payos webhook failed", "error", err)
		respondError(c, http.StatusBadRequest, "INVALID_WEBHOOK", err.Error())
		return
	}
	respondOK(c, "Webhook processed", gin.H{"received": true})
}

// PayOSWebhookProbe answers GET for URL reachability checks (no body).
func (h *PaymentHandler) PayOSWebhookProbe(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "payos webhook endpoint ready"})
}
