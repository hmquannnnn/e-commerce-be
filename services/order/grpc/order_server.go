package grpcserver

import (
	"context"
	"errors"
	"math"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/order-service/model"
	"github.com/hmquannnnn/e-commerce/order-service/service"
	orderpb "github.com/hmquannnnn/e-commerce/pkg/proto/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderServer struct {
	orderpb.UnimplementedOrderServiceServer
	orderService service.OrderService
}

func NewOrderServer(orderService service.OrderService) *OrderServer {
	return &OrderServer{orderService: orderService}
}

func (s *OrderServer) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.OrderResponse, error) {
	orderID, err := uuid.Parse(req.GetOrderId())
	if err != nil || orderID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order_id")
	}

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil || userID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	order, err := s.orderService.GetOrder(ctx, userID, orderID)
	if err != nil {
		return nil, mapOrderError(err)
	}
	if order.Status != model.OrderStatusPending || order.PaymentMethod == model.PaymentMethodCash {
		return nil, status.Error(codes.FailedPrecondition, "order is not payable")
	}

	return &orderpb.OrderResponse{
		OrderId:       order.ID.String(),
		UserId:        order.UserID.String(),
		TotalPrice:    int64(math.Round(order.TotalPrice)),
		Status:        string(order.Status),
		PaymentMethod: string(order.PaymentMethod),
	}, nil
}

func (s *OrderServer) MarkOrderPaid(ctx context.Context, req *orderpb.MarkPaidRequest) (*orderpb.MarkPaidResponse, error) {
	orderID, err := uuid.Parse(req.GetOrderId())
	if err != nil || orderID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order_id")
	}
	if req.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be greater than zero")
	}

	if err := s.orderService.MarkOrderPaid(ctx, orderID, req.GetAmount()); err != nil {
		return nil, mapOrderError(err)
	}

	return &orderpb.MarkPaidResponse{
		Success: true,
		Message: "Order marked paid",
	}, nil
}

func (s *OrderServer) TouchPaymentDeadline(ctx context.Context, req *orderpb.TouchRequest) (*orderpb.TouchResponse, error) {
	orderID, err := uuid.Parse(req.GetOrderId())
	if err != nil || orderID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order_id")
	}

	if err := s.orderService.TouchPaymentDeadline(ctx, orderID); err != nil {
		return nil, mapOrderError(err)
	}

	return &orderpb.TouchResponse{Success: true}, nil
}

func mapOrderError(err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, "invalid order request")
	case errors.Is(err, service.ErrOrderNotFound):
		return status.Error(codes.NotFound, "order not found")
	case errors.Is(err, service.ErrForbidden):
		return status.Error(codes.PermissionDenied, "order does not belong to user")
	case errors.Is(err, service.ErrAmountMismatch):
		return status.Error(codes.FailedPrecondition, "order amount mismatch")
	case errors.Is(err, service.ErrInvalidStatusTransition):
		return status.Error(codes.FailedPrecondition, "order is not payable")
	default:
		return status.Errorf(codes.Internal, "order service error: %v", err)
	}
}
