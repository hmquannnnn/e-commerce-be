package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID              uuid.UUID       `json:"id"`
	Name            string          `json:"name"`
	Description     *string         `json:"description,omitempty"`
	Price           float64         `json:"price"`
	Specs           json.RawMessage `json:"specs,omitempty"`
	CategoryID      *int            `json:"category_id,omitempty"`
	PrimaryImageURL *string         `json:"primary_image_url,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type ProductImage struct {
	ID           uuid.UUID `json:"id"`
	ProductID    uuid.UUID `json:"product_id"`
	URL          string    `json:"url"`
	DisplayOrder int       `json:"display_order"`
	IsPrimary    bool      `json:"is_primary"`
	CreatedAt    time.Time `json:"created_at"`
}

type ProductWithImages struct {
	Product
	Images []ProductImage `json:"images"`
}

type CreateProductParams struct {
	ID          uuid.UUID
	Name        string
	Description *string
	Price       float64
	Specs       json.RawMessage
	CategoryID  *int
}

type UpdateProductParams struct {
	Name        *string
	Description *string
	Price       *float64
	Specs       json.RawMessage
	CategoryID  *int
}

type ListProductsFilter struct {
	CategoryID *int
	Search     string
	MinPrice   *float64
	MaxPrice   *float64
	Page       int
	Limit      int
}
