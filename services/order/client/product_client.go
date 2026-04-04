package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type ProductInfo struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Price    float64   `json:"price"`
	ImageURL *string   `json:"image_url,omitempty"`
}

type InventoryItem struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

type ProductClient interface {
	GetProduct(ctx context.Context, productID uuid.UUID) (*ProductInfo, error)
	ReserveInventory(ctx context.Context, item InventoryItem) error
	ReleaseInventory(ctx context.Context, item InventoryItem) error
}

type productClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewProductClient(baseURL string) ProductClient {
	return &productClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type productDetailResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID     uuid.UUID `json:"id"`
		Name   string    `json:"name"`
		Price  float64   `json:"price"`
		Images []struct {
			URL       string `json:"url"`
			IsPrimary bool   `json:"is_primary"`
		} `json:"images"`
	} `json:"data"`
	Message string `json:"message"`
}

func (c *productClient) GetProduct(ctx context.Context, productID uuid.UUID) (*ProductInfo, error) {
	url := fmt.Sprintf("%s/api/products/%s", c.baseURL, productID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call product-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrProductNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product-service returned status %d", resp.StatusCode)
	}

	var result productDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode product response: %w", err)
	}

	info := &ProductInfo{
		ID:    result.Data.ID,
		Name:  result.Data.Name,
		Price: result.Data.Price,
	}

	for _, img := range result.Data.Images {
		if img.IsPrimary {
			url := img.URL
			info.ImageURL = &url
			break
		}
	}
	if info.ImageURL == nil && len(result.Data.Images) > 0 {
		url := result.Data.Images[0].URL
		info.ImageURL = &url
	}

	return info, nil
}

func (c *productClient) ReserveInventory(ctx context.Context, item InventoryItem) error {
	return c.callInventory(ctx, "/api/internal/inventory/reserve", item)
}

func (c *productClient) ReleaseInventory(ctx context.Context, item InventoryItem) error {
	return c.callInventory(ctx, "/api/internal/inventory/release", item)
}

func (c *productClient) callInventory(ctx context.Context, path string, item InventoryItem) error {
	body, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call product-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return ErrInsufficientStock
	}
	if resp.StatusCode == http.StatusNotFound {
		return ErrProductNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("product-service inventory call returned status %d", resp.StatusCode)
	}

	return nil
}
