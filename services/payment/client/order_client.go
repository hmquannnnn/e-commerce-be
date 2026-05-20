package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ErrOrderNotFound được trả về khi order-service báo 404.
var ErrOrderNotFound = errors.New("order not found")

// ErrOrderNotPayable trả về khi order-service từ chối transition (đơn đã bị
// cancel / đã chuyển trạng thái khác PENDING). Caller nên log nhưng không retry.
var ErrOrderNotPayable = errors.New("order not in a payable state")

// OrderInfo là thông tin tối thiểu về order cần cho payment validation.
type OrderInfo struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TotalPrice float64
	Status     string // "PENDING", "PAID", "DELIVERING", "DELIVERED", "CANCELLED"
}

// OrderClient định nghĩa interface giao tiếp với order-service.
type OrderClient interface {
	GetOrder(ctx context.Context, orderID uuid.UUID) (*OrderInfo, error)
	// MarkOrderPaid yêu cầu order-service chuyển order PENDING → PAID.
	// Returns ErrOrderNotPayable nếu order không còn ở PENDING (đã cancel, etc).
	MarkOrderPaid(ctx context.Context, orderID uuid.UUID) error
	TouchPaymentDeadline(ctx context.Context, orderID uuid.UUID) error
}

type orderClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewOrderClient tạo HTTP client gọi sang order-service.
func NewOrderClient(baseURL string) OrderClient {
	return &orderClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type orderInternalResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		ID         uuid.UUID `json:"id"`
		UserID     uuid.UUID `json:"user_id"`
		TotalPrice float64   `json:"total_price"`
		Status     string    `json:"status"`
	} `json:"data"`
}

func (c *orderClient) GetOrder(ctx context.Context, orderID uuid.UUID) (*OrderInfo, error) {
	url := fmt.Sprintf("%s/internal/orders/%s", c.baseURL, orderID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call order-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrOrderNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("order-service returned status %d", resp.StatusCode)
	}

	var result orderInternalResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode order response: %w", err)
	}

	return &OrderInfo{
		ID:         result.Data.ID,
		UserID:     result.Data.UserID,
		TotalPrice: result.Data.TotalPrice,
		Status:     result.Data.Status,
	}, nil
}

func (c *orderClient) MarkOrderPaid(ctx context.Context, orderID uuid.UUID) error {
	url := fmt.Sprintf("%s/internal/orders/%s/mark-paid", c.baseURL, orderID)
	return c.postOrderAction(ctx, url)
}

func (c *orderClient) TouchPaymentDeadline(ctx context.Context, orderID uuid.UUID) error {
	url := fmt.Sprintf("%s/internal/orders/%s/touch-payment-deadline", c.baseURL, orderID)
	return c.postOrderAction(ctx, url)
}

func (c *orderClient) postOrderAction(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call order-service: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return ErrOrderNotFound
	case http.StatusConflict:
		return ErrOrderNotPayable
	default:
		return fmt.Errorf("order-service returned status %d", resp.StatusCode)
	}
}
