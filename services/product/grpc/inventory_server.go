package grpcserver

import (
	"context"
	"errors"

	"github.com/google/uuid"
	inventorypb "github.com/hmquannnnn/e-commerce/pkg/proto/inventory"
	"github.com/hmquannnnn/e-commerce/product-service/model"
	"github.com/hmquannnnn/e-commerce/product-service/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InventoryServer struct {
	inventorypb.UnimplementedInventoryServiceServer
	inventoryService service.InventoryService
}

func NewInventoryServer(inventoryService service.InventoryService) *InventoryServer {
	return &InventoryServer{inventoryService: inventoryService}
}

func (s *InventoryServer) ReserveStock(ctx context.Context, req *inventorypb.ReserveRequest) (*inventorypb.ReserveResponse, error) {
	items, err := parseReserveItems(req.GetItems())
	if err != nil {
		return nil, err
	}

	if err := s.inventoryService.ReserveStockBatch(ctx, items); err != nil {
		return nil, mapInventoryError(err)
	}

	return &inventorypb.ReserveResponse{
		Success: true,
		Message: "Stock reserved successfully",
	}, nil
}

func (s *InventoryServer) ReleaseStock(ctx context.Context, req *inventorypb.ReleaseRequest) (*inventorypb.ReleaseResponse, error) {
	items, err := parseReleaseItems(req.GetItems())
	if err != nil {
		return nil, err
	}

	if err := s.inventoryService.ReleaseStockBatch(ctx, items); err != nil {
		return nil, mapInventoryError(err)
	}

	return &inventorypb.ReleaseResponse{
		Success: true,
		Message: "Stock released successfully",
	}, nil
}

func parseReserveItems(items []*inventorypb.OrderItem) ([]model.ReserveStockParams, error) {
	if len(items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "items are required")
	}

	result := make([]model.ReserveStockParams, 0, len(items))
	for _, item := range items {
		productID, err := uuid.Parse(item.GetProductId())
		if err != nil || productID == uuid.Nil {
			return nil, status.Error(codes.InvalidArgument, "invalid product_id")
		}
		if item.GetQuantity() <= 0 {
			return nil, status.Error(codes.InvalidArgument, "quantity must be greater than zero")
		}
		result = append(result, model.ReserveStockParams{
			ProductID: productID,
			Quantity:  int(item.GetQuantity()),
		})
	}
	return result, nil
}

func parseReleaseItems(items []*inventorypb.OrderItem) ([]model.ReleaseStockParams, error) {
	if len(items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "items are required")
	}

	result := make([]model.ReleaseStockParams, 0, len(items))
	for _, item := range items {
		productID, err := uuid.Parse(item.GetProductId())
		if err != nil || productID == uuid.Nil {
			return nil, status.Error(codes.InvalidArgument, "invalid product_id")
		}
		if item.GetQuantity() <= 0 {
			return nil, status.Error(codes.InvalidArgument, "quantity must be greater than zero")
		}
		result = append(result, model.ReleaseStockParams{
			ProductID: productID,
			Quantity:  int(item.GetQuantity()),
		})
	}
	return result, nil
}

func mapInventoryError(err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, "invalid inventory request")
	case errors.Is(err, service.ErrInventoryNotFound):
		return status.Error(codes.NotFound, "inventory not found")
	case errors.Is(err, service.ErrInsufficientStock):
		return status.Error(codes.FailedPrecondition, "insufficient stock")
	default:
		return status.Errorf(codes.Internal, "inventory service error: %v", err)
	}
}
