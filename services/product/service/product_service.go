package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/pkg/storage/minio"
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

type ImageInput struct {
	URL          string
	DisplayOrder int
	IsPrimary    bool
}

type CreateProductParams struct {
	ProductID   *uuid.UUID
	Name        string
	Description *string
	Price       float64
	Specs       []byte
	CategoryID  *int
	Images      []ImageInput
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
	minioClient   *minio.Client
	minioBucket   string
}

func NewProductService(
	productRepo repository.ProductRepository,
	inventoryRepo repository.InventoryRepository,
	minioClient *minio.Client,
	minioBucket string,
) ProductService {
	return &productService{
		productRepo:   productRepo,
		inventoryRepo: inventoryRepo,
		minioClient:   minioClient,
		minioBucket:   minioBucket,
	}
}

func (s *productService) CreateProduct(ctx context.Context, params CreateProductParams) (*model.Product, error) {
	if params.Name == "" {
		return nil, ErrInvalidInput
	}
	if params.Price < 0 {
		return nil, ErrInvalidInput
	}

	// Use provided ID or generate a new one
	productID := uuid.New()
	if params.ProductID != nil {
		productID = *params.ProductID
	}

	product, err := s.productRepo.Create(ctx, &model.CreateProductParams{
		ID:          productID,
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

	// Bulk insert images
	for i, img := range params.Images {
		if _, err := s.productRepo.AddImage(ctx, product.ID, img.URL, img.DisplayOrder, img.IsPrimary); err != nil {
			slog.Warn("failed to add product image",
				"product_id", product.ID,
				"image_index", i,
				"error", err,
			)
		}
	}

	// Create MinIO folder for product images (best-effort)
	if s.minioClient != nil {
		folderPath := "products/" + product.ID.String()
		if err := s.minioClient.CreateFolder(ctx, s.minioBucket, folderPath); err != nil {
			slog.Warn("failed to create MinIO folder for product",
				"product_id", product.ID,
				"folder", folderPath,
				"error", err,
			)
		}
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
	slog.Info("products", "products", products)
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
