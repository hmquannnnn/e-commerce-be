package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	inventorypb "github.com/hmquannnnn/e-commerce/pkg/proto/inventory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
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
	ReserveInventory(ctx context.Context, items []InventoryItem) error
	ReleaseInventory(ctx context.Context, items []InventoryItem) error
	Close() error
}

type productClient struct {
	baseURL         string
	httpClient      *http.Client
	inventoryConn   *grpc.ClientConn
	inventoryClient inventorypb.InventoryServiceClient
}

func NewProductClient(baseURL, inventoryAddress string) (ProductClient, error) {
	conn, err := grpc.Dial(inventoryAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to product-service inventory grpc: %w", err)
	}
	return &productClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		inventoryConn:   conn,
		inventoryClient: inventorypb.NewInventoryServiceClient(conn),
	}, nil
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

func (c *productClient) ReserveInventory(ctx context.Context, items []InventoryItem) error {
	if len(items) == 0 {
		return nil
	}

	resp, err := c.inventoryClient.ReserveStock(ctx, &inventorypb.ReserveRequest{
		Items: toProtoItems(items),
	})
	if err != nil {
		return mapInventoryRPCError(err)
	}
	if !resp.GetSuccess() {
		return fmt.Errorf("product-service inventory reserve failed: %s", resp.GetMessage())
	}
	return nil
}

func (c *productClient) ReleaseInventory(ctx context.Context, items []InventoryItem) error {
	if len(items) == 0 {
		return nil
	}

	resp, err := c.inventoryClient.ReleaseStock(ctx, &inventorypb.ReleaseRequest{
		Items: toProtoItems(items),
	})
	if err != nil {
		return mapInventoryRPCError(err)
	}
	if !resp.GetSuccess() {
		return fmt.Errorf("product-service inventory release failed: %s", resp.GetMessage())
	}
	return nil
}

func (c *productClient) Close() error {
	if c.inventoryConn == nil {
		return nil
	}
	return c.inventoryConn.Close()
}

func toProtoItems(items []InventoryItem) []*inventorypb.OrderItem {
	result := make([]*inventorypb.OrderItem, 0, len(items))
	for _, item := range items {
		result = append(result, &inventorypb.OrderItem{
			ProductId: item.ProductID.String(),
			Quantity:  int32(item.Quantity),
		})
	}
	return result
}

func mapInventoryRPCError(err error) error {
	switch status.Code(err) {
	case codes.NotFound:
		return ErrProductNotFound
	case codes.FailedPrecondition:
		return ErrInsufficientStock
	default:
		return fmt.Errorf("product-service inventory grpc call failed: %w", err)
	}
}
