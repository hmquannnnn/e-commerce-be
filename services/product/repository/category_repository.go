package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/hmquannnnn/e-commerce/pkg/common/utils"
	"github.com/hmquannnnn/e-commerce/product-service/model"
)

var (
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryAlreadyExists = errors.New("category already exists")
)

type CategoryRepository interface {
	Create(ctx context.Context, params *model.CreateCategoryParams) (*model.Category, error)
	GetByID(ctx context.Context, id int) (*model.Category, error)
	GetByName(ctx context.Context, name string) (*model.Category, error)
	List(ctx context.Context) ([]*model.Category, error)
	Update(ctx context.Context, id int, params *model.UpdateCategoryParams) error
	Delete(ctx context.Context, id int) error
}

type categoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, params *model.CreateCategoryParams) (*model.Category, error) {
	query := `
		INSERT INTO categories (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at
	`
	cat := &model.Category{}
	err := r.db.QueryRowContext(ctx, query, params.Name, params.Description).Scan(
		&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt, &cat.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrCategoryAlreadyExists
		}
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	return cat, nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id int) (*model.Category, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM categories WHERE id = $1`
	cat := &model.Category{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt, &cat.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return cat, nil
}

func (r *categoryRepository) GetByName(ctx context.Context, name string) (*model.Category, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM categories WHERE name = $1`
	cat := &model.Category{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt, &cat.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return cat, nil
}

func (r *categoryRepository) List(ctx context.Context) ([]*model.Category, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM categories ORDER BY name ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	var categories []*model.Category
	for rows.Next() {
		cat := &model.Category{}
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt, &cat.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, cat)
	}
	return categories, rows.Err()
}

func (r *categoryRepository) Update(ctx context.Context, id int, params *model.UpdateCategoryParams) error {
	updates := make(map[string]interface{})
	if params.Name != nil {
		updates["name"] = *params.Name
	}
	if params.Description != nil {
		updates["description"] = *params.Description
	}

	rowsAffected, err := utils.ExecPatch(ctx, r.db, utils.PatchUpdate{
		Table:     "categories",
		Updates:   updates,
		Where:     "id = $1",
		WhereArgs: []interface{}{id},
	})
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}
	if rowsAffected == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (r *categoryRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return len(errStr) > 0 && (contains(errStr, "duplicate key") || contains(errStr, "unique constraint"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
