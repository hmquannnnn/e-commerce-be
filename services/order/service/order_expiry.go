package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/hmquannnnn/e-commerce/order-service/client"
	"github.com/hmquannnnn/e-commerce/order-service/model"
	"github.com/hmquannnnn/e-commerce/order-service/repository"
)

const (
	orderExpiryTTL      = 15 * time.Minute
	orderExpiryInterval = 1 * time.Minute
)

type OrderExpiryWorker struct {
	orderRepo     repository.OrderRepository
	productClient client.ProductClient
}

func NewOrderExpiryWorker(
	orderRepo repository.OrderRepository,
	productClient client.ProductClient,
) *OrderExpiryWorker {
	return &OrderExpiryWorker{
		orderRepo:     orderRepo,
		productClient: productClient,
	}
}

// Start runs the expiry loop until ctx is cancelled.
func (w *OrderExpiryWorker) Start(ctx context.Context) {
	slog.Info("order expiry worker started", "ttl", orderExpiryTTL, "interval", orderExpiryInterval)
	ticker := time.NewTicker(orderExpiryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("order expiry worker stopped")
			return
		case <-ticker.C:
			w.cancelExpiredOrders(ctx)
		}
	}
}

func (w *OrderExpiryWorker) cancelExpiredOrders(ctx context.Context) {
	cutoff := time.Now().Add(-orderExpiryTTL)
	ids, err := w.orderRepo.ListExpiredPendingOrders(ctx, cutoff)
	if err != nil {
		slog.Error("expiry: failed to list expired orders", "error", err)
		return
	}
	if len(ids) == 0 {
		return
	}

	slog.Info("expiry: cancelling expired orders", "count", len(ids))
	for _, orderID := range ids {
		order, err := w.orderRepo.GetByID(ctx, orderID)
		if err != nil {
			slog.Error("expiry: failed to get order", "error", err, "order_id", orderID)
			continue
		}

		_ = w.productClient.ReleaseInventory(ctx, orderItemsToInventoryItems(order.Items))

		if err := w.orderRepo.UpdateStatus(ctx, orderID, model.OrderStatusCancelled); err != nil {
			slog.Error("expiry: failed to cancel order", "error", err, "order_id", orderID)
			continue
		}
		slog.Info("expiry: cancelled expired order", "order_id", orderID)
	}
}
