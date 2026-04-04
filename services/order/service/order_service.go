package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/order-service/client"
	"github.com/hmquannnnn/e-commerce/order-service/model"
	"github.com/hmquannnnn/e-commerce/order-service/repository"
)

func mergeCheckoutLines(lines []model.CheckoutLine) (map[uuid.UUID]int, error) {
	if len(lines) == 0 {
		return nil, ErrInvalidOrderLines
	}
	merged := make(map[uuid.UUID]int)
	for _, l := range lines {
		if l.Quantity < 1 {
			return nil, ErrInvalidOrderLines
		}
		merged[l.ProductID] += l.Quantity
	}
	return merged, nil
}

type OrderService interface {
	CreateOrder(ctx context.Context, userID uuid.UUID, paymentMethod model.PaymentMethod, lines []model.CheckoutLine) (*model.OrderWithItems, error)
	GetOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) (*model.OrderWithItems, error)
	ListUserOrders(ctx context.Context, userID uuid.UUID, page, limit int) ([]*model.Order, int64, error)
	CancelOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) error
	AdminListOrders(ctx context.Context, status *model.OrderStatus, page, limit int) ([]*model.Order, int64, error)
	AdminUpdateStatus(ctx context.Context, orderID uuid.UUID, status model.OrderStatus) error
}

type orderService struct {
	orderRepo     repository.OrderRepository
	cartRepo      repository.CartRepository
	productClient client.ProductClient
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	productClient client.ProductClient,
) OrderService {
	return &orderService{
		orderRepo:     orderRepo,
		cartRepo:      cartRepo,
		productClient: productClient,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, userID uuid.UUID, paymentMethod model.PaymentMethod, lines []model.CheckoutLine) (*model.OrderWithItems, error) {
	if paymentMethod != model.PaymentMethodVNPAY && paymentMethod != model.PaymentMethodCash {
		return nil, ErrInvalidInput
	}

	merged, err := mergeCheckoutLines(lines)
	if err != nil {
		return nil, err
	}

	cartItems, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}
	if len(cartItems) == 0 {
		return nil, ErrCartEmpty
	}

	cartByProduct := make(map[uuid.UUID]*model.CartItem, len(cartItems))
	for _, ci := range cartItems {
		cartByProduct[ci.ProductID] = ci
	}

	orderLines := make([]*model.CartItem, 0, len(merged))
	for productID, qty := range merged {
		ci, ok := cartByProduct[productID]
		if !ok || qty > ci.Quantity {
			return nil, ErrInvalidOrderLines
		}
		line := *ci
		line.Quantity = qty
		orderLines = append(orderLines, &line)
	}

	reserved := make([]client.InventoryItem, 0, len(orderLines))
	for _, item := range orderLines {
		if err := s.productClient.ReserveInventory(ctx, client.InventoryItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}); err != nil {
			for _, r := range reserved {
				_ = s.productClient.ReleaseInventory(ctx, r)
			}
			if errors.Is(err, client.ErrInsufficientStock) {
				return nil, ErrInsufficientStock
			}
			if errors.Is(err, client.ErrProductNotFound) {
				return nil, ErrProductNotFound
			}
			return nil, fmt.Errorf("failed to reserve inventory: %w", err)
		}
		reserved = append(reserved, client.InventoryItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	order, err := s.orderRepo.Create(ctx, &model.CreateOrderParams{
		UserID:        userID,
		PaymentMethod: paymentMethod,
	}, orderLines, userID, merged)
	if err != nil {
		for _, r := range reserved {
			_ = s.productClient.ReleaseInventory(ctx, r)
		}
		if errors.Is(err, repository.ErrCartQuantityMismatch) {
			return nil, ErrCartChanged
		}
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return order, nil
}

func (s *orderService) GetOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) (*model.OrderWithItems, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order.UserID != userID {
		return nil, ErrForbidden
	}

	return order, nil
}

func (s *orderService) ListUserOrders(ctx context.Context, userID uuid.UUID, page, limit int) ([]*model.Order, int64, error) {
	orders, total, err := s.orderRepo.List(ctx, model.ListOrdersFilter{
		UserID: &userID,
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}
	return orders, total, nil
}

func (s *orderService) CancelOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return ErrOrderNotFound
		}
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order.UserID != userID {
		return ErrForbidden
	}

	if order.Status != model.OrderStatusPending && order.Status != model.OrderStatusPaid {
		return ErrOrderNotCancellable
	}

	// Release inventory
	for _, item := range order.Items {
		_ = s.productClient.ReleaseInventory(ctx, client.InventoryItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	if err := s.orderRepo.UpdateStatus(ctx, orderID, model.OrderStatusCancelled); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}
	return nil
}

func (s *orderService) AdminListOrders(ctx context.Context, status *model.OrderStatus, page, limit int) ([]*model.Order, int64, error) {
	orders, total, err := s.orderRepo.List(ctx, model.ListOrdersFilter{
		Status: status,
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}
	return orders, total, nil
}

func (s *orderService) AdminUpdateStatus(ctx context.Context, orderID uuid.UUID, status model.OrderStatus) error {
	if err := s.orderRepo.UpdateStatus(ctx, orderID, status); err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return ErrOrderNotFound
		}
		return fmt.Errorf("failed to update order status: %w", err)
	}
	return nil
}

// TotalPages is a helper used by handlers.
func TotalPages(total int64, limit int) int {
	return int(math.Ceil(float64(total) / float64(limit)))
}
