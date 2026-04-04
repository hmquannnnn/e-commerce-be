package model

import (
	"time"

	"github.com/google/uuid"
)

type CartItem struct {
	ID          int64     `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	UnitPrice   float64   `json:"unit_price"`
	ImageURL    *string   `json:"image_url,omitempty"`
	Quantity    int       `json:"quantity"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AddCartItemParams struct {
	UserID      uuid.UUID
	ProductID   uuid.UUID
	ProductName string
	UnitPrice   float64
	ImageURL    *string
	Quantity    int
}

type UpdateCartItemParams struct {
	UserID    uuid.UUID
	ProductID uuid.UUID
	Quantity  int
}
