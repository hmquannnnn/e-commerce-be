package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/order-service/model"
	"github.com/hmquannnnn/e-commerce/order-service/service"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	lines := make([]model.CheckoutLine, 0, len(req.Items))
	for _, it := range req.Items {
		lines = append(lines, model.CheckoutLine{ProductID: it.ProductID, Quantity: it.Quantity})
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), userID, model.PaymentMethod(req.PaymentMethod), lines)
	if err != nil {
		if errors.Is(err, service.ErrCartEmpty) {
			respondError(c, http.StatusBadRequest, "CART_EMPTY", "Cart is empty")
			return
		}
		if errors.Is(err, service.ErrInvalidOrderLines) {
			respondError(c, http.StatusBadRequest, "INVALID_ORDER_ITEMS", "Items must exist in cart and quantities must not exceed cart")
			return
		}
		if errors.Is(err, service.ErrCartChanged) {
			respondError(c, http.StatusConflict, "CART_CHANGED", "Cart changed during checkout; please refresh and try again")
			return
		}
		if errors.Is(err, service.ErrInsufficientStock) {
			respondError(c, http.StatusConflict, "INSUFFICIENT_STOCK", "Insufficient stock for one or more items")
			return
		}
		if errors.Is(err, service.ErrProductNotFound) {
			respondError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "One or more products not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create order")
		return
	}

	respondCreated(c, "Order created successfully", ToOrderResponse(order))
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid order ID")
		return
	}

	order, err := h.orderService.GetOrder(c.Request.Context(), userID, orderID)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			respondError(c, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			respondError(c, http.StatusForbidden, "FORBIDDEN", "Access denied")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get order")
		return
	}

	respondOK(c, "Order retrieved successfully", ToOrderResponse(order))
}

func (h *OrderHandler) ListUserOrders(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var query ListOrdersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_QUERY", "Invalid query parameters")
		return
	}

	orders, total, err := h.orderService.ListUserOrders(c.Request.Context(), userID, query.Page, query.Limit)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list orders")
		return
	}

	items := make([]*OrderListItemResponse, 0, len(orders))
	for _, o := range orders {
		items = append(items, ToOrderListItemResponse(o))
	}

	respondOK(c, "Orders retrieved successfully", gin.H{
		"items":       items,
		"total":       total,
		"page":        query.Page,
		"limit":       query.Limit,
		"total_pages": service.TotalPages(total, query.Limit),
	})
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid order ID")
		return
	}

	if err := h.orderService.CancelOrder(c.Request.Context(), userID, orderID); err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			respondError(c, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			respondError(c, http.StatusForbidden, "FORBIDDEN", "Access denied")
			return
		}
		if errors.Is(err, service.ErrOrderNotCancellable) {
			respondError(c, http.StatusConflict, "ORDER_NOT_CANCELLABLE", "Order cannot be cancelled in current status")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cancel order")
		return
	}

	respondOK(c, "Order cancelled successfully", nil)
}

// ─── Admin Handlers ───────────────────────────────────────────────────────────

func (h *OrderHandler) AdminListOrders(c *gin.Context) {
	var query ListOrdersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_QUERY", "Invalid query parameters")
		return
	}

	var status *model.OrderStatus
	if query.Status != "" {
		s := model.OrderStatus(query.Status)
		status = &s
	}

	orders, total, err := h.orderService.AdminListOrders(c.Request.Context(), status, query.Page, query.Limit)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list orders")
		return
	}

	items := make([]*OrderListItemResponse, 0, len(orders))
	for _, o := range orders {
		items = append(items, ToOrderListItemResponse(o))
	}

	respondOK(c, "Orders retrieved successfully", gin.H{
		"items":       items,
		"total":       total,
		"page":        query.Page,
		"limit":       query.Limit,
		"total_pages": service.TotalPages(total, query.Limit),
	})
}

func (h *OrderHandler) AdminUpdateStatus(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid order ID")
		return
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.orderService.AdminUpdateStatus(c.Request.Context(), orderID, model.OrderStatus(req.Status)); err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			respondError(c, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update order status")
		return
	}

	respondOK(c, "Order status updated", nil)
}
