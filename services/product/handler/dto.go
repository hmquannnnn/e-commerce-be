package handler

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/product-service/model"
)

// ─── Category DTOs ───────────────────────────────────────────────────────────

type CreateCategoryRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Description *string `json:"description,omitempty"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty"`
}

type CategoryResponse struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func ToCategoryResponse(c *model.Category) *CategoryResponse {
	return &CategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
	}
}

// ─── Product DTOs ─────────────────────────────────────────────────────────────

type CreateProductRequest struct {
	ProductID   *uuid.UUID       `json:"product_id,omitempty"`
	Name        string           `json:"name" binding:"required,min=1,max=50"`
	Description *string          `json:"description,omitempty"`
	Price       float64          `json:"price" binding:"required,min=0"`
	Specs       json.RawMessage  `json:"specs,omitempty"`
	CategoryID  *int             `json:"category_id,omitempty"`
	Images      []AddImageRequest `json:"images,omitempty"`
}

type UpdateProductRequest struct {
	Name        *string         `json:"name,omitempty" binding:"omitempty,min=1,max=50"`
	Description *string         `json:"description,omitempty"`
	Price       *float64        `json:"price,omitempty" binding:"omitempty,min=0"`
	Specs       json.RawMessage `json:"specs,omitempty"`
	CategoryID  *int            `json:"category_id,omitempty"`
}

type ListProductsQuery struct {
	CategoryID *int   `form:"category_id"`
	Search     string `form:"search"`
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=20"`
}

type AddImageRequest struct {
	URL          string `json:"url" binding:"required,url"`
	DisplayOrder int    `json:"display_order"`
	IsPrimary    bool   `json:"is_primary"`
}

type ProductImageResponse struct {
	ID           uuid.UUID `json:"id"`
	ProductID    uuid.UUID `json:"product_id"`
	URL          string    `json:"url"`
	DisplayOrder int       `json:"display_order"`
	IsPrimary    bool      `json:"is_primary"`
	CreatedAt    string    `json:"created_at"`
}

type ProductResponse struct {
	ID          uuid.UUID             `json:"id"`
	Name        string                `json:"name"`
	Description *string               `json:"description,omitempty"`
	Price       float64               `json:"price"`
	Specs       json.RawMessage       `json:"specs,omitempty"`
	CategoryID  *int                  `json:"category_id,omitempty"`
	Images      []ProductImageResponse `json:"images"`
	CreatedAt   string                `json:"created_at"`
	UpdatedAt   string                `json:"updated_at"`
}

type ProductListItemResponse struct {
	ID              uuid.UUID       `json:"id"`
	Name            string          `json:"name"`
	Description     *string         `json:"description,omitempty"`
	Price           float64         `json:"price"`
	Specs           json.RawMessage `json:"specs,omitempty"`
	CategoryID      *int            `json:"category_id,omitempty"`
	PrimaryImageURL *string         `json:"primary_image_url,omitempty"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}

func ToProductResponse(p *model.ProductWithImages) *ProductResponse {
	images := make([]ProductImageResponse, 0, len(p.Images))
	for _, img := range p.Images {
		images = append(images, ProductImageResponse{
			ID:           img.ID,
			ProductID:    img.ProductID,
			URL:          img.URL,
			DisplayOrder: img.DisplayOrder,
			IsPrimary:    img.IsPrimary,
			CreatedAt:    img.CreatedAt.Format(time.RFC3339),
		})
	}
	return &ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Specs:       p.Specs,
		CategoryID:  p.CategoryID,
		Images:      images,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
}

func ToProductListItemResponse(p *model.Product) *ProductListItemResponse {
	return &ProductListItemResponse{
		ID:              p.ID,
		Name:            p.Name,
		Description:     p.Description,
		Price:           p.Price,
		Specs:           p.Specs,
		CategoryID:      p.CategoryID,
		PrimaryImageURL: p.PrimaryImageURL,
		CreatedAt:       p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       p.UpdatedAt.Format(time.RFC3339),
	}
}

// ─── Inventory DTOs ───────────────────────────────────────────────────────────

type UpdateStockRequest struct {
	StockQuantity int `json:"stock_quantity" binding:"min=0"`
}

type ReserveStockRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
}

type ReleaseStockRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
}

type InventoryResponse struct {
	ID                uuid.UUID `json:"id"`
	ProductID         uuid.UUID `json:"product_id"`
	StockQuantity     int       `json:"stock_quantity"`
	ReservedQuantity  int       `json:"reserved_quantity"`
	AvailableQuantity int       `json:"available_quantity"`
	CreatedAt         string    `json:"created_at"`
	UpdatedAt         string    `json:"updated_at"`
}

func ToInventoryResponse(inv *model.Inventory) *InventoryResponse {
	return &InventoryResponse{
		ID:                inv.ID,
		ProductID:         inv.ProductID,
		StockQuantity:     inv.StockQuantity,
		ReservedQuantity:  inv.ReservedQuantity,
		AvailableQuantity: inv.AvailableQuantity(),
		CreatedAt:         inv.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         inv.UpdatedAt.Format(time.RFC3339),
	}
}
