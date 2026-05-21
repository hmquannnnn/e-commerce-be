package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

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

func cartLinesToInventoryItems(items []*model.CartItem) []client.InventoryItem {
	result := make([]client.InventoryItem, 0, len(items))
	for _, item := range items {
		result = append(result, client.InventoryItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}
	return result
}

func orderItemsToInventoryItems(items []model.OrderItem) []client.InventoryItem {
	result := make([]client.InventoryItem, 0, len(items))
	for _, item := range items {
		result = append(result, client.InventoryItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}
	return result
}

type OrderService interface {
	CreateOrder(ctx context.Context, params model.CreateOrderParams, lines []model.CheckoutLine) (*model.OrderWithItems, error)
	GetOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) (*model.OrderWithItems, error)
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*model.OrderWithItems, error)
	ListUserOrders(ctx context.Context, userID uuid.UUID, page, limit int) ([]*model.Order, int64, error)
	CancelOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) error
	AdminListOrders(ctx context.Context, status *model.OrderStatus, search string, page, limit int) ([]*model.OrderWithCustomer, int64, error)
	AdminGetOrder(ctx context.Context, orderID uuid.UUID) (*model.OrderWithItemsAndCustomer, error)
	AdminUpdateStatus(ctx context.Context, orderID uuid.UUID, status model.OrderStatus) error
	TouchPaymentDeadline(ctx context.Context, orderID uuid.UUID) error
	// MarkOrderPaid transitions a PENDING order to PAID. Idempotent: returns nil
	// if the order is already PAID. Any other state returns ErrInvalidStatusTransition.
	// Used by payment-service via gRPC after PayOS webhook.
	MarkOrderPaid(ctx context.Context, orderID uuid.UUID, amount int64) error
}

type orderService struct {
	orderRepo     repository.OrderRepository
	cartRepo      repository.CartRepository
	productClient client.ProductClient
	userClient    client.UserClient
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	productClient client.ProductClient,
	userClient client.UserClient,
) OrderService {
	return &orderService{
		orderRepo:     orderRepo,
		cartRepo:      cartRepo,
		productClient: productClient,
		userClient:    userClient,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, params model.CreateOrderParams, lines []model.CheckoutLine) (*model.OrderWithItems, error) {
	params.ShippingPhone = strings.TrimSpace(params.ShippingPhone)
	params.ShippingAddress = strings.TrimSpace(params.ShippingAddress)
	if !model.IsValidPaymentMethod(params.PaymentMethod) ||
		params.UserID == uuid.Nil ||
		params.ShippingPhone == "" ||
		params.ShippingAddress == "" ||
		len(params.ShippingPhone) > 20 ||
		len(params.ShippingAddress) > 500 {
		return nil, ErrInvalidInput
	}

	merged, err := mergeCheckoutLines(lines)
	if err != nil {
		return nil, err
	}

	cartItems, err := s.cartRepo.GetByUserID(ctx, params.UserID)
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

	reserved := cartLinesToInventoryItems(orderLines)
	if err := s.productClient.ReserveInventory(ctx, reserved); err != nil {
		if errors.Is(err, client.ErrInsufficientStock) {
			return nil, ErrInsufficientStock
		}
		if errors.Is(err, client.ErrProductNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to reserve inventory: %w", err)
	}

	order, err := s.orderRepo.Create(ctx, &params, orderLines, params.UserID, merged)
	if err != nil {
		_ = s.productClient.ReleaseInventory(ctx, reserved)
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

func (s *orderService) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*model.OrderWithItems, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
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

	_ = s.productClient.ReleaseInventory(ctx, orderItemsToInventoryItems(order.Items))

	if err := s.orderRepo.UpdateStatus(ctx, orderID, model.OrderStatusCancelled); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}
	return nil
}

func (s *orderService) AdminListOrders(ctx context.Context, status *model.OrderStatus, search string, page, limit int) ([]*model.OrderWithCustomer, int64, error) {
	matchedUserIDs := []uuid.UUID{}
	if strings.TrimSpace(search) != "" {
		users, err := s.userClient.SearchUsers(ctx, search, 100)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to search users: %w", err)
		}
		for _, user := range users {
			matchedUserIDs = append(matchedUserIDs, user.ID)
		}
	}

	orders, total, err := s.orderRepo.List(ctx, model.ListOrdersFilter{
		Status:  status,
		Search:  search,
		UserIDs: matchedUserIDs,
		Page:    page,
		Limit:   limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}

	usersByID, err := s.loadCustomerSummaries(ctx, orders)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*model.OrderWithCustomer, 0, len(orders))
	for _, order := range orders {
		item := &model.OrderWithCustomer{Order: *order}
		if customer, ok := usersByID[order.UserID]; ok {
			item.Customer = customer
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (s *orderService) AdminGetOrder(ctx context.Context, orderID uuid.UUID) (*model.OrderWithItemsAndCustomer, error) {
	order, err := s.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	usersByID, err := s.loadCustomerSummaries(ctx, []*model.Order{&order.Order})
	if err != nil {
		return nil, err
	}

	result := &model.OrderWithItemsAndCustomer{OrderWithItems: *order}
	if customer, ok := usersByID[order.UserID]; ok {
		result.Customer = customer
	}
	return result, nil
}

func (s *orderService) loadCustomerSummaries(ctx context.Context, orders []*model.Order) (map[uuid.UUID]*model.CustomerSummary, error) {
	userIDs := make([]uuid.UUID, 0, len(orders))
	seen := make(map[uuid.UUID]struct{}, len(orders))
	for _, order := range orders {
		if _, ok := seen[order.UserID]; ok {
			continue
		}
		seen[order.UserID] = struct{}{}
		userIDs = append(userIDs, order.UserID)
	}

	users, err := s.userClient.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load customers: %w", err)
	}

	result := make(map[uuid.UUID]*model.CustomerSummary, len(users))
	for id, user := range users {
		result[id] = &model.CustomerSummary{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		}
	}
	return result, nil
}

// AdminUpdateStatus enforces the order state machine for admin transitions.
// Allowed:
//   - (PENDING|PAID)   → DELIVERING  (admin xác nhận đơn để giao)
//   - DELIVERING       → DELIVERED   (admin xác nhận đã giao xong)
//   - (PENDING|PAID)   → CANCELLED   (admin huỷ thay user; release inventory)
//
// All other transitions return ErrInvalidStatusTransition. DELIVERED và
// CANCELLED là terminal states.
func (s *orderService) AdminUpdateStatus(ctx context.Context, orderID uuid.UUID, target model.OrderStatus) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return ErrOrderNotFound
		}
		return fmt.Errorf("failed to get order: %w", err)
	}

	if !isValidAdminTransition(order.Status, target) {
		return ErrInvalidStatusTransition
	}

	if target == model.OrderStatusCancelled {
		_ = s.productClient.ReleaseInventory(ctx, orderItemsToInventoryItems(order.Items))
	}

	if err := s.orderRepo.UpdateStatus(ctx, orderID, target); err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return ErrOrderNotFound
		}
		return fmt.Errorf("failed to update order status: %w", err)
	}
	return nil
}

func isValidAdminTransition(from, to model.OrderStatus) bool {
	switch to {
	case model.OrderStatusDelivering:
		return from == model.OrderStatusPending || from == model.OrderStatusPaid
	case model.OrderStatusDelivered:
		return from == model.OrderStatusDelivering
	case model.OrderStatusCancelled:
		return from == model.OrderStatusPending || from == model.OrderStatusPaid
	}
	return false
}

func (s *orderService) TouchPaymentDeadline(ctx context.Context, orderID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return ErrOrderNotFound
		}
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order.Status != model.OrderStatusPending || order.PaymentMethod == model.PaymentMethodCash {
		return ErrInvalidStatusTransition
	}

	if err := s.orderRepo.TouchPaymentDeadline(ctx, orderID); err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return ErrInvalidStatusTransition
		}
		return fmt.Errorf("failed to touch payment deadline: %w", err)
	}
	return nil
}

// MarkOrderPaid is the internal-only entrypoint used by payment-service after
// a successful provider callback. Only PENDING → PAID is allowed; if the order
// has already been moved past PENDING (e.g. user cancelled in the meantime,
// admin shipped a CASH order, etc.) the call returns ErrInvalidStatusTransition
// so the caller can log/alert without overwriting business state. Already-PAID
// is treated as a no-op for webhook-retry safety.
func (s *orderService) MarkOrderPaid(ctx context.Context, orderID uuid.UUID, amount int64) error {
	if amount <= 0 {
		return ErrInvalidInput
	}

	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return ErrOrderNotFound
		}
		return fmt.Errorf("failed to get order: %w", err)
	}

	expectedAmount := int64(math.Round(order.TotalPrice))
	if diff := amount - expectedAmount; diff < -1 || diff > 1 {
		return ErrAmountMismatch
	}

	if order.Status == model.OrderStatusPaid {
		return nil
	}
	if order.Status != model.OrderStatusPending || order.PaymentMethod == model.PaymentMethodCash {
		return ErrInvalidStatusTransition
	}

	if err := s.orderRepo.UpdateStatus(ctx, orderID, model.OrderStatusPaid); err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return ErrOrderNotFound
		}
		return fmt.Errorf("failed to mark order paid: %w", err)
	}
	return nil
}

// TotalPages is a helper used by handlers.
func TotalPages(total int64, limit int) int {
	return int(math.Ceil(float64(total) / float64(limit)))
}
