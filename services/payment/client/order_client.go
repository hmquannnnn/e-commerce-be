package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	orderpb "github.com/hmquannnnn/e-commerce/pkg/proto/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var ErrOrderNotFound = errors.New("order not found")

// ErrOrderNotPayable is returned when order-service refuses a payment action
// because the order is no longer PENDING/payable.
var ErrOrderNotPayable = errors.New("order not in a payable state")

var ErrOrderAmountMismatch = errors.New("order amount mismatch")

type OrderInfo struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	TotalPrice    int64
	Status        string
	PaymentMethod string
}

type OrderClient interface {
	GetOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (*OrderInfo, error)
	MarkOrderPaid(ctx context.Context, orderID uuid.UUID, amount int64) error
	TouchPaymentDeadline(ctx context.Context, orderID uuid.UUID) error
	Close() error
}

type orderClient struct {
	conn   *grpc.ClientConn
	client orderpb.OrderServiceClient
}

func NewOrderClient(address string) (OrderClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order-service grpc: %w", err)
	}
	return &orderClient{
		conn:   conn,
		client: orderpb.NewOrderServiceClient(conn),
	}, nil
}

func (c *orderClient) GetOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (*OrderInfo, error) {
	resp, err := c.client.GetOrder(ctx, &orderpb.GetOrderRequest{
		OrderId: orderID.String(),
		UserId:  userID.String(),
	})
	if err != nil {
		return nil, mapOrderRPCError(err)
	}

	parsedOrderID, err := uuid.Parse(resp.GetOrderId())
	if err != nil {
		return nil, fmt.Errorf("order-service returned invalid order_id: %w", err)
	}
	parsedUserID, err := uuid.Parse(resp.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("order-service returned invalid user_id: %w", err)
	}

	return &OrderInfo{
		ID:            parsedOrderID,
		UserID:        parsedUserID,
		TotalPrice:    resp.GetTotalPrice(),
		Status:        resp.GetStatus(),
		PaymentMethod: resp.GetPaymentMethod(),
	}, nil
}

func (c *orderClient) MarkOrderPaid(ctx context.Context, orderID uuid.UUID, amount int64) error {
	resp, err := c.client.MarkOrderPaid(ctx, &orderpb.MarkPaidRequest{
		OrderId: orderID.String(),
		Amount:  amount,
	})
	if err != nil {
		return mapOrderRPCError(err)
	}
	if !resp.GetSuccess() {
		return fmt.Errorf("order-service mark paid failed: %s", resp.GetMessage())
	}
	return nil
}

func (c *orderClient) TouchPaymentDeadline(ctx context.Context, orderID uuid.UUID) error {
	resp, err := c.client.TouchPaymentDeadline(ctx, &orderpb.TouchRequest{
		OrderId: orderID.String(),
	})
	if err != nil {
		return mapOrderRPCError(err)
	}
	if !resp.GetSuccess() {
		return fmt.Errorf("order-service touch payment deadline failed")
	}
	return nil
}

func (c *orderClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func mapOrderRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("order-service grpc call failed: %w", err)
	}

	switch st.Code() {
	case codes.NotFound:
		return ErrOrderNotFound
	case codes.FailedPrecondition:
		if st.Message() == "order amount mismatch" {
			return ErrOrderAmountMismatch
		}
		return ErrOrderNotPayable
	case codes.PermissionDenied:
		return ErrOrderNotFound
	default:
		return fmt.Errorf("order-service grpc call failed: %w", err)
	}
}
