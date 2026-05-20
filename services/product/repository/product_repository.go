package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/pkg/common/utils"
	"github.com/hmquannnnn/e-commerce/pkg/storage"
	"github.com/hmquannnnn/e-commerce/product-service/model"
)

var ErrProductNotFound = errors.New("product not found")

type ProductRepository interface {
	Create(ctx context.Context, params *model.CreateProductParams) (*model.Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.ProductWithImages, error)
	List(ctx context.Context, filter model.ListProductsFilter) ([]*model.Product, int64, error)
	Update(ctx context.Context, id uuid.UUID, params *model.UpdateProductParams) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddImage(ctx context.Context, productID uuid.UUID, url string, displayOrder int, isPrimary bool) (*model.ProductImage, error)
	DeleteImage(ctx context.Context, imageID uuid.UUID) error
}

type productRepository struct {
	db         *sql.DB
	storageMgr *storage.Manager
}

func NewProductRepository(db *sql.DB, storageMgr *storage.Manager) ProductRepository {
	return &productRepository{db: db, storageMgr: storageMgr}
}

func (r *productRepository) Create(ctx context.Context, params *model.CreateProductParams) (*model.Product, error) {
	query := `
		INSERT INTO products (id, name, description, price, specs, category_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, description, price, specs, category_id, created_at, updated_at
	`
	p := &model.Product{}
	var specsBytes []byte
	err := r.db.QueryRowContext(ctx, query,
		params.ID, params.Name, params.Description, params.Price, params.Specs, params.CategoryID,
	).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &specsBytes, &p.CategoryID,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}
	if specsBytes != nil {
		p.Specs = json.RawMessage(specsBytes)
	}
	return p, nil
}

func (r *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ProductWithImages, error) {
	query := `
		SELECT id, name, description, price, specs, category_id, created_at, updated_at
		FROM products WHERE id = $1
	`
	p := &model.ProductWithImages{}
	var specsBytes []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &specsBytes, &p.CategoryID,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	if specsBytes != nil {
		p.Specs = json.RawMessage(specsBytes)
	}

	images, err := r.getImages(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Images = images

	return p, nil
}

func (r *productRepository) getImages(ctx context.Context, productID uuid.UUID) ([]model.ProductImage, error) {
	query := `
		SELECT id, product_id, url, display_order, is_primary, created_at
		FROM product_images WHERE product_id = $1
		ORDER BY display_order ASC, created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product images: %w", err)
	}
	defer rows.Close()

	var images []model.ProductImage
	for rows.Next() {
		img := model.ProductImage{}
		if err := rows.Scan(&img.ID, &img.ProductID, &img.URL, &img.DisplayOrder, &img.IsPrimary, &img.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}
		images = append(images, img)
	}
	return images, rows.Err()
}

func (r *productRepository) List(ctx context.Context, filter model.ListProductsFilter) ([]*model.Product, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}

	conditions := []string{}
	args := []interface{}{}
	argIdx := 1

	if filter.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIdx))
		args = append(args, *filter.CategoryID)
		argIdx++
	}
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}
	if filter.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price >= $%d", argIdx))
		args = append(args, *filter.MinPrice)
		argIdx++
	}
	if filter.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price <= $%d", argIdx))
		args = append(args, *filter.MaxPrice)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM products %s`, where)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	offset := (filter.Page - 1) * filter.Limit
	dataArgs := append(args, filter.Limit, offset)
	dataQuery := fmt.Sprintf(`
		SELECT p.id, p.name, p.description, p.price, p.specs, p.category_id, p.created_at, p.updated_at,
		       (SELECT url FROM product_images WHERE product_id = p.id AND is_primary = TRUE LIMIT 1) AS primary_image_url
		FROM products p %s
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)

	rows, err := r.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		p := &model.Product{}
		var specsBytes []byte
		var rawImageURL *string
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &specsBytes, &p.CategoryID, &p.CreatedAt, &p.UpdatedAt, &rawImageURL); err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		if specsBytes != nil {
			p.Specs = json.RawMessage(specsBytes)
		}
		if rawImageURL != nil && r.storageMgr != nil {
			// url := r.storageMgr.GetPublicURL(*rawImageURL)
			p.PrimaryImageURL = rawImageURL
		}
		products = append(products, p)
	}
	return products, total, rows.Err()
}

func (r *productRepository) Update(ctx context.Context, id uuid.UUID, params *model.UpdateProductParams) error {
	updates := make(map[string]interface{})
	if params.Name != nil {
		updates["name"] = *params.Name
	}
	if params.Description != nil {
		updates["description"] = *params.Description
	}
	if params.Price != nil {
		updates["price"] = *params.Price
	}
	if params.Specs != nil {
		updates["specs"] = []byte(params.Specs)
	}
	if params.CategoryID != nil {
		updates["category_id"] = *params.CategoryID
	}

	rowsAffected, err := utils.ExecPatch(ctx, r.db, utils.PatchUpdate{
		Table:     "products",
		Updates:   updates,
		Where:     "id = $1",
		WhereArgs: []interface{}{id},
	})
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	if rowsAffected == 0 {
		return ErrProductNotFound
	}
	return nil
}

func (r *productRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrProductNotFound
	}
	return nil
}

func (r *productRepository) AddImage(ctx context.Context, productID uuid.UUID, url string, displayOrder int, isPrimary bool) (*model.ProductImage, error) {
	if isPrimary {
		_, err := r.db.ExecContext(ctx,
			`UPDATE product_images SET is_primary = FALSE WHERE product_id = $1`, productID)
		if err != nil {
			return nil, fmt.Errorf("failed to unset primary images: %w", err)
		}
	}

	query := `
		INSERT INTO product_images (product_id, url, display_order, is_primary)
		VALUES ($1, $2, $3, $4)
		RETURNING id, product_id, url, display_order, is_primary, created_at
	`
	img := &model.ProductImage{}
	err := r.db.QueryRowContext(ctx, query, productID, url, displayOrder, isPrimary).Scan(
		&img.ID, &img.ProductID, &img.URL, &img.DisplayOrder, &img.IsPrimary, &img.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add product image: %w", err)
	}
	return img, nil
}

func (r *productRepository) DeleteImage(ctx context.Context, imageID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM product_images WHERE id = $1`, imageID)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("image not found")
	}
	return nil
}
