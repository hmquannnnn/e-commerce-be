package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/order-service/client"
	"github.com/hmquannnnn/e-commerce/order-service/model"
	"github.com/hmquannnnn/e-commerce/order-service/repository"
)

type CartService interface {
	GetCart(ctx context.Context, userID uuid.UUID) ([]*model.CartItem, error)
	AddItem(ctx context.Context, userID, productID uuid.UUID, quantity int) (*model.CartItem, error)
	UpdateItemQuantity(ctx context.Context, userID, productID uuid.UUID, quantity int) (*model.CartItem, error)
	RemoveItem(ctx context.Context, userID, productID uuid.UUID) error
	ClearCart(ctx context.Context, userID uuid.UUID) error
}

type cartService struct {
	cartRepo      repository.CartRepository
	productClient client.ProductClient
}

func NewCartService(cartRepo repository.CartRepository, productClient client.ProductClient) CartService {
	return &cartService{
		cartRepo:      cartRepo,
		productClient: productClient,
	}
}

func (s *cartService) GetCart(ctx context.Context, userID uuid.UUID) ([]*model.CartItem, error) {
	items, err := s.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}
	return items, nil
}

func (s *cartService) AddItem(ctx context.Context, userID, productID uuid.UUID, quantity int) (*model.CartItem, error) {
	if quantity <= 0 {
		return nil, ErrInvalidInput
	}

	product, err := s.productClient.GetProduct(ctx, productID)
	if err != nil {
		if errors.Is(err, client.ErrProductNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product info: %w", err)
	}

	item, err := s.cartRepo.Upsert(ctx, &model.AddCartItemParams{
		UserID:      userID,
		ProductID:   productID,
		ProductName: product.Name,
		UnitPrice:   product.Price,
		ImageURL:    product.ImageURL,
		Quantity:    quantity,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add cart item: %w", err)
	}
	return item, nil
}

func (s *cartService) UpdateItemQuantity(ctx context.Context, userID, productID uuid.UUID, quantity int) (*model.CartItem, error) {
	if quantity <= 0 {
		return nil, ErrInvalidInput
	}

	item, err := s.cartRepo.UpdateQuantity(ctx, &model.UpdateCartItemParams{
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
	})
	if err != nil {
		if errors.Is(err, repository.ErrCartItemNotFound) {
			return nil, ErrCartItemNotFound
		}
		return nil, fmt.Errorf("failed to update cart item: %w", err)
	}
	return item, nil
}

func (s *cartService) RemoveItem(ctx context.Context, userID, productID uuid.UUID) error {
	if err := s.cartRepo.DeleteItem(ctx, userID, productID); err != nil {
		if errors.Is(err, repository.ErrCartItemNotFound) {
			return ErrCartItemNotFound
		}
		return fmt.Errorf("failed to remove cart item: %w", err)
	}
	return nil
}

func (s *cartService) ClearCart(ctx context.Context, userID uuid.UUID) error {
	if err := s.cartRepo.ClearByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}
	return nil
}
