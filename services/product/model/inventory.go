package model

import (
	"time"

	"github.com/google/uuid"
)

type Inventory struct {
	ID               uuid.UUID `json:"id"`
	ProductID        uuid.UUID `json:"product_id"`
	StockQuantity    int       `json:"stock_quantity"`
	ReservedQuantity int       `json:"reserved_quantity"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (inv *Inventory) AvailableQuantity() int {
	return inv.StockQuantity - inv.ReservedQuantity
}

type UpdateInventoryParams struct {
	StockQuantity *int
}

type ReserveStockParams struct {
	ProductID uuid.UUID
	Quantity  int
}

type ReleaseStockParams struct {
	ProductID uuid.UUID
	Quantity  int
}
