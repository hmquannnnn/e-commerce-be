package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string
type PaymentMethod string

const (
	OrderStatusPending    OrderStatus = "PENDING"
	OrderStatusPaid       OrderStatus = "PAID"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusCancelled  OrderStatus = "CANCELLED"
)

const (
	PaymentMethodVNPAY PaymentMethod = "VNPAY"
	PaymentMethodCash  PaymentMethod = "CASH"
)

type Order struct {
	ID            uuid.UUID     `json:"id"`
	UserID        uuid.UUID     `json:"user_id"`
	TotalPrice    float64       `json:"total_price"`
	Status        OrderStatus   `json:"status"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type OrderItem struct {
	ID          int64     `json:"id"`
	OrderID     uuid.UUID `json:"order_id"`
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	UnitPrice   float64   `json:"unit_price"`
	ImageURL    *string   `json:"image_url,omitempty"`
	Quantity    int       `json:"quantity"`
}

type OrderWithItems struct {
	Order
	Items []OrderItem `json:"items"`
}

type CreateOrderParams struct {
	UserID        uuid.UUID
	PaymentMethod PaymentMethod
}

// CheckoutLine is one product line the client wants to buy from the server-side cart.
type CheckoutLine struct {
	ProductID uuid.UUID
	Quantity  int
}

type ListOrdersFilter struct {
	UserID *uuid.UUID
	Status *OrderStatus
	Page   int
	Limit  int
}
