package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/product-service/model"
	"github.com/hmquannnnn/e-commerce/product-service/repository"
)

type InventoryService interface {
	GetInventory(ctx context.Context, productID uuid.UUID) (*model.Inventory, error)
	UpdateStock(ctx context.Context, productID uuid.UUID, stockQuantity int) (*model.Inventory, error)
	ReserveStock(ctx context.Context, productID uuid.UUID, quantity int) error
	ReserveStockBatch(ctx context.Context, items []model.ReserveStockParams) error
	ReleaseStock(ctx context.Context, productID uuid.UUID, quantity int) error
	ReleaseStockBatch(ctx context.Context, items []model.ReleaseStockParams) error
}

type inventoryService struct {
	repo repository.InventoryRepository
}

func NewInventoryService(repo repository.InventoryRepository) InventoryService {
	return &inventoryService{repo: repo}
}

func (s *inventoryService) GetInventory(ctx context.Context, productID uuid.UUID) (*model.Inventory, error) {
	inv, err := s.repo.GetByProductID(ctx, productID)
	if err != nil {
		if errors.Is(err, repository.ErrInventoryNotFound) {
			return nil, ErrInventoryNotFound
		}
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}
	return inv, nil
}

func (s *inventoryService) UpdateStock(ctx context.Context, productID uuid.UUID, stockQuantity int) (*model.Inventory, error) {
	if stockQuantity < 0 {
		return nil, ErrInvalidInput
	}

	if err := s.repo.UpdateStock(ctx, productID, &model.UpdateInventoryParams{
		StockQuantity: &stockQuantity,
	}); err != nil {
		if errors.Is(err, repository.ErrInventoryNotFound) {
			return nil, ErrInventoryNotFound
		}
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	return s.GetInventory(ctx, productID)
}

func (s *inventoryService) ReserveStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidInput
	}

	if err := s.repo.ReserveStock(ctx, &model.ReserveStockParams{
		ProductID: productID,
		Quantity:  quantity,
	}); err != nil {
		if errors.Is(err, repository.ErrInsufficientStock) {
			return ErrInsufficientStock
		}
		if errors.Is(err, repository.ErrInventoryNotFound) {
			return ErrInventoryNotFound
		}
		return fmt.Errorf("failed to reserve stock: %w", err)
	}
	return nil
}

func (s *inventoryService) ReserveStockBatch(ctx context.Context, items []model.ReserveStockParams) error {
	if len(items) == 0 {
		return ErrInvalidInput
	}
	for _, item := range items {
		if item.ProductID == uuid.Nil || item.Quantity <= 0 {
			return ErrInvalidInput
		}
	}

	if err := s.repo.ReserveStocks(ctx, items); err != nil {
		if errors.Is(err, repository.ErrInsufficientStock) {
			return ErrInsufficientStock
		}
		if errors.Is(err, repository.ErrInventoryNotFound) {
			return ErrInventoryNotFound
		}
		return fmt.Errorf("failed to reserve stock batch: %w", err)
	}
	return nil
}

func (s *inventoryService) ReleaseStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidInput
	}

	if err := s.repo.ReleaseStock(ctx, &model.ReleaseStockParams{
		ProductID: productID,
		Quantity:  quantity,
	}); err != nil {
		if errors.Is(err, repository.ErrInventoryNotFound) {
			return ErrInventoryNotFound
		}
		return fmt.Errorf("failed to release stock: %w", err)
	}
	return nil
}

func (s *inventoryService) ReleaseStockBatch(ctx context.Context, items []model.ReleaseStockParams) error {
	if len(items) == 0 {
		return ErrInvalidInput
	}
	for _, item := range items {
		if item.ProductID == uuid.Nil || item.Quantity <= 0 {
			return ErrInvalidInput
		}
	}

	if err := s.repo.ReleaseStocks(ctx, items); err != nil {
		if errors.Is(err, repository.ErrInventoryNotFound) {
			return ErrInventoryNotFound
		}
		return fmt.Errorf("failed to release stock batch: %w", err)
	}
	return nil
}
