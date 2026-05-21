package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/hmquannnnn/e-commerce/product-service/model"
	"github.com/hmquannnnn/e-commerce/product-service/repository"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, params CreateCategoryParams) (*model.Category, error)
	GetCategory(ctx context.Context, id int) (*model.Category, error)
	ListCategories(ctx context.Context) ([]*model.Category, error)
	UpdateCategory(ctx context.Context, id int, params UpdateCategoryParams) (*model.Category, error)
	DeleteCategory(ctx context.Context, id int) error
}

type CreateCategoryParams struct {
	Name        string
	Description *string
}

type UpdateCategoryParams struct {
	Name        *string
	Description *string
}

type categoryService struct {
	repo        repository.CategoryRepository
	productRepo repository.ProductRepository
}

func NewCategoryService(repo repository.CategoryRepository, productRepo repository.ProductRepository) CategoryService {
	return &categoryService{repo: repo, productRepo: productRepo}
}

func (s *categoryService) CreateCategory(ctx context.Context, params CreateCategoryParams) (*model.Category, error) {
	if params.Name == "" {
		return nil, ErrInvalidInput
	}

	cat, err := s.repo.Create(ctx, &model.CreateCategoryParams{
		Name:        params.Name,
		Description: params.Description,
	})
	if err != nil {
		if errors.Is(err, repository.ErrCategoryAlreadyExists) {
			return nil, ErrCategoryAlreadyExists
		}
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	return cat, nil
}

func (s *categoryService) GetCategory(ctx context.Context, id int) (*model.Category, error) {
	cat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return cat, nil
}

func (s *categoryService) ListCategories(ctx context.Context) ([]*model.Category, error) {
	cats, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	return cats, nil
}

func (s *categoryService) UpdateCategory(ctx context.Context, id int, params UpdateCategoryParams) (*model.Category, error) {
	if err := s.repo.Update(ctx, id, &model.UpdateCategoryParams{
		Name:        params.Name,
		Description: params.Description,
	}); err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return s.GetCategory(ctx, id)
}

func (s *categoryService) DeleteCategory(ctx context.Context, id int) error {
	count, err := s.productRepo.CountByCategory(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check products in category: %w", err)
	}
	if count > 0 {
		return ErrCategoryHasProducts
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			return ErrCategoryNotFound
		}
		return fmt.Errorf("failed to delete category: %w", err)
	}
	return nil
}
