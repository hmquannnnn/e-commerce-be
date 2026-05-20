package handler

import (
	"time"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/order-service/model"
)

// ─── Cart DTOs ────────────────────────────────────────────────────────────────

type AddCartItemRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

type CartItemResponse struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	UnitPrice   float64   `json:"unit_price"`
	ImageURL    *string   `json:"image_url,omitempty"`
	Quantity    int       `json:"quantity"`
	Subtotal    float64   `json:"subtotal"`
	UpdatedAt   string    `json:"updated_at"`
}

type CartResponse struct {
	Items         []CartItemResponse `json:"items"`
	TotalQuantity int                `json:"total_quantity"`
	TotalPrice    float64            `json:"total_price"`
}

func ToCartItemResponse(item *model.CartItem) CartItemResponse {
	return CartItemResponse{
		ProductID:   item.ProductID,
		ProductName: item.ProductName,
		UnitPrice:   item.UnitPrice,
		ImageURL:    item.ImageURL,
		Quantity:    item.Quantity,
		Subtotal:    item.UnitPrice * float64(item.Quantity),
		UpdatedAt:   item.UpdatedAt.Format(time.RFC3339),
	}
}

func ToCartResponse(items []*model.CartItem) CartResponse {
	resp := CartResponse{
		Items: make([]CartItemResponse, 0, len(items)),
	}
	for _, item := range items {
		resp.Items = append(resp.Items, ToCartItemResponse(item))
		resp.TotalQuantity += item.Quantity
		resp.TotalPrice += item.UnitPrice * float64(item.Quantity)
	}
	return resp
}

// ─── Order DTOs ───────────────────────────────────────────────────────────────

type CreateOrderItemRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
}

type CreateOrderRequest struct {
	PaymentMethod string                   `json:"payment_method" binding:"required,oneof=QR_CODE CASH"`
	Items         []CreateOrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

// UpdateOrderStatusRequest is used by the admin endpoint. It only accepts the
// transitions the admin UI is allowed to drive. PAID is intentionally excluded
// from the admin path — that transition is owned by payment-service via the
// internal /mark-paid endpoint.
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=DELIVERING DELIVERED CANCELLED"`
}

type ListOrdersQuery struct {
	Status string `form:"status"`
	Search string `form:"search"`
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=20"`
}

type OrderItemResponse struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	UnitPrice   float64   `json:"unit_price"`
	ImageURL    *string   `json:"image_url,omitempty"`
	Quantity    int       `json:"quantity"`
	Subtotal    float64   `json:"subtotal"`
}

type CustomerSummaryResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
}

type OrderResponse struct {
	ID            uuid.UUID                `json:"id"`
	UserID        uuid.UUID                `json:"user_id"`
	Customer      *CustomerSummaryResponse `json:"customer,omitempty"`
	TotalPrice    float64                  `json:"total_price"`
	Status        model.OrderStatus        `json:"status"`
	PaymentMethod model.PaymentMethod      `json:"payment_method"`
	Items         []OrderItemResponse      `json:"items"`
	CreatedAt     string                   `json:"created_at"`
	UpdatedAt     string                   `json:"updated_at"`
}

type OrderListItemResponse struct {
	ID            uuid.UUID                `json:"id"`
	UserID        uuid.UUID                `json:"user_id"`
	Customer      *CustomerSummaryResponse `json:"customer,omitempty"`
	TotalPrice    float64                  `json:"total_price"`
	Status        model.OrderStatus        `json:"status"`
	PaymentMethod model.PaymentMethod      `json:"payment_method"`
	CreatedAt     string                   `json:"created_at"`
	UpdatedAt     string                   `json:"updated_at"`
}

func ToOrderResponse(o *model.OrderWithItems) *OrderResponse {
	items := make([]OrderItemResponse, 0, len(o.Items))
	for _, item := range o.Items {
		items = append(items, OrderItemResponse{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			UnitPrice:   item.UnitPrice,
			ImageURL:    item.ImageURL,
			Quantity:    item.Quantity,
			Subtotal:    item.UnitPrice * float64(item.Quantity),
		})
	}
	return &OrderResponse{
		ID:            o.ID,
		UserID:        o.UserID,
		TotalPrice:    o.TotalPrice,
		Status:        o.Status,
		PaymentMethod: o.PaymentMethod,
		Items:         items,
		CreatedAt:     o.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     o.UpdatedAt.Format(time.RFC3339),
	}
}

func ToAdminOrderResponse(o *model.OrderWithItemsAndCustomer) *OrderResponse {
	resp := ToOrderResponse(&o.OrderWithItems)
	resp.Customer = ToCustomerSummaryResponse(o.Customer)
	return resp
}

func ToOrderListItemResponse(o *model.Order) *OrderListItemResponse {
	return &OrderListItemResponse{
		ID:            o.ID,
		UserID:        o.UserID,
		TotalPrice:    o.TotalPrice,
		Status:        o.Status,
		PaymentMethod: o.PaymentMethod,
		CreatedAt:     o.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     o.UpdatedAt.Format(time.RFC3339),
	}
}

func ToAdminOrderListItemResponse(o *model.OrderWithCustomer) *OrderListItemResponse {
	resp := ToOrderListItemResponse(&o.Order)
	resp.Customer = ToCustomerSummaryResponse(o.Customer)
	return resp
}

func ToCustomerSummaryResponse(customer *model.CustomerSummary) *CustomerSummaryResponse {
	if customer == nil {
		return nil
	}
	return &CustomerSummaryResponse{
		ID:    customer.ID,
		Email: customer.Email,
		Name:  customer.Name,
	}
}
