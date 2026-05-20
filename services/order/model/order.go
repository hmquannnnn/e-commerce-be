package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string
type PaymentMethod string

// Order state machine:
//
//	PENDING ──(QR_CODE paid)──► PAID ──(admin)──► DELIVERING ──(admin)──► DELIVERED
//	  │                              │                   │
//	  │ (admin/user cancel,          │ (admin/user       │  (terminal — không
//	  │  hoặc expiry 15' với         │  cancel)          │   thể cancel)
//	  │  QR_CODE)                    ▼                   ▼
//	  ▼                          CANCELLED          ─────
//	CANCELLED
//
// CASH bỏ qua bước PAID: PENDING ──(admin xác nhận)──► DELIVERING.
const (
	OrderStatusPending    OrderStatus = "PENDING"
	OrderStatusPaid       OrderStatus = "PAID"
	OrderStatusDelivering OrderStatus = "DELIVERING"
	OrderStatusDelivered  OrderStatus = "DELIVERED"
	OrderStatusCancelled  OrderStatus = "CANCELLED"
)

const (
	PaymentMethodQRCode PaymentMethod = "QR_CODE"
	PaymentMethodCash   PaymentMethod = "CASH"
)

func IsValidPaymentMethod(m PaymentMethod) bool {
	switch m {
	case PaymentMethodQRCode, PaymentMethodCash:
		return true
	}
	return false
}

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

type CustomerSummary struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
}

type OrderWithCustomer struct {
	Order
	Customer *CustomerSummary `json:"customer,omitempty"`
}

type OrderWithItemsAndCustomer struct {
	OrderWithItems
	Customer *CustomerSummary `json:"customer,omitempty"`
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
	UserID  *uuid.UUID
	UserIDs []uuid.UUID
	Status  *OrderStatus
	Search  string
	Page    int
	Limit   int
}
