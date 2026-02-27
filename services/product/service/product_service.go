package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/product-service/model"
	"github.com/hmquannnnn/e-commerce/product-service/repository"
)

type ProductService interface {
	CreateProduct(ctx context.Context, params CreateProductParams) (*model.Product, error)
	GetProduct(ctx context.Context, id uuid.UUID) (*model.ProductWithImages, error)
	ListProducts(ctx context.Context, filter model.ListProductsFilter) ([]*model.Product, int64, error)
	UpdateProduct(ctx context.Context, id uuid.UUID, params UpdateProductParams) (*model.ProductWithImages, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) error
	AddProductImage(ctx context.Context, productID uuid.UUID, url string, displayOrder int, isPrimary bool) (*model.ProductImage, error)
	DeleteProductImage(ctx context.Context, imageID uuid.UUID) error
}

type CreateProductParams struct {
	Name        string
	Description *string
	Price       float64
	Specs       []byte
	CategoryID  *int
}

type UpdateProductParams struct {
	Name        *string
	Description *string
	Price       *float64
	Specs       []byte
	CategoryID  *int
}

type productService struct {
	productRepo   repository.ProductRepository
	inventoryRepo repository.InventoryRepository
}

func NewProductService(
	productRepo repository.ProductRepository,
	inventoryRepo repository.InventoryRepository,
) ProductService {
	return &productService{
		productRepo:   productRepo,
		inventoryRepo: inventoryRepo,
	}
}

func (s *productService) CreateProduct(ctx context.Context, params CreateProductParams) (*model.Product, error) {
	if params.Name == "" {
		return nil, ErrInvalidInput
	}
	if params.Price < 0 {
		return nil, ErrInvalidInput
	}

	product, err := s.productRepo.Create(ctx, &model.CreateProductParams{
		Name:        params.Name,
		Description: params.Description,
		Price:       params.Price,
		Specs:       params.Specs,
		CategoryID:  params.CategoryID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Initialize inventory with 0 stock
	if _, err := s.inventoryRepo.Create(ctx, product.ID, 0); err != nil {
		return nil, fmt.Errorf("failed to initialize inventory: %w", err)
	}

	return product, nil
}

func (s *productService) GetProduct(ctx context.Context, id uuid.UUID) (*model.ProductWithImages, error) {
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return product, nil
}

func (s *productService) ListProducts(ctx context.Context, filter model.ListProductsFilter) ([]*model.Product, int64, error) {
	products, total, err := s.productRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	return products, total, nil
}

func (s *productService) UpdateProduct(ctx context.Context, id uuid.UUID, params UpdateProductParams) (*model.ProductWithImages, error) {
	if params.Price != nil && *params.Price < 0 {
		return nil, ErrInvalidInput
	}

	if err := s.productRepo.Update(ctx, id, &model.UpdateProductParams{
		Name:        params.Name,
		Description: params.Description,
		Price:       params.Price,
		Specs:       params.Specs,
		CategoryID:  params.CategoryID,
	}); err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return s.GetProduct(ctx, id)
}

func (s *productService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	if err := s.productRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return ErrProductNotFound
		}
		return fmt.Errorf("failed to delete product: %w", err)
	}
	return nil
}

func (s *productService) AddProductImage(ctx context.Context, productID uuid.UUID, url string, displayOrder int, isPrimary bool) (*model.ProductImage, error) {
	img, err := s.productRepo.AddImage(ctx, productID, url, displayOrder, isPrimary)
	if err != nil {
		return nil, fmt.Errorf("failed to add product image: %w", err)
	}
	return img, nil
}

func (s *productService) DeleteProductImage(ctx context.Context, imageID uuid.UUID) error {
	if err := s.productRepo.DeleteImage(ctx, imageID); err != nil {
		return fmt.Errorf("failed to delete product image: %w", err)
	}
	return nil
}
