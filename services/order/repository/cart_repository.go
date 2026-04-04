package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/order-service/model"
)

var ErrCartItemNotFound = errors.New("cart item not found")

type CartRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.CartItem, error)
	Upsert(ctx context.Context, params *model.AddCartItemParams) (*model.CartItem, error)
	UpdateQuantity(ctx context.Context, params *model.UpdateCartItemParams) (*model.CartItem, error)
	DeleteItem(ctx context.Context, userID, productID uuid.UUID) error
	ClearByUserID(ctx context.Context, userID uuid.UUID) error
}

type cartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*model.CartItem, error) {
	query := `
		SELECT id, user_id, product_id, product_name, unit_price, image_url, quantity, updated_at
		FROM cart_items
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}
	defer rows.Close()

	var items []*model.CartItem
	for rows.Next() {
		item := &model.CartItem{}
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.ProductID,
			&item.ProductName, &item.UnitPrice, &item.ImageURL,
			&item.Quantity, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *cartRepository) Upsert(ctx context.Context, params *model.AddCartItemParams) (*model.CartItem, error) {
	query := `
		INSERT INTO cart_items (user_id, product_id, product_name, unit_price, image_url, quantity, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (user_id, product_id) DO UPDATE
		SET quantity    = cart_items.quantity + EXCLUDED.quantity,
		    product_name = EXCLUDED.product_name,
		    unit_price   = EXCLUDED.unit_price,
		    image_url    = EXCLUDED.image_url,
		    updated_at   = NOW()
		RETURNING id, user_id, product_id, product_name, unit_price, image_url, quantity, updated_at
	`
	item := &model.CartItem{}
	err := r.db.QueryRowContext(ctx, query,
		params.UserID, params.ProductID, params.ProductName,
		params.UnitPrice, params.ImageURL, params.Quantity,
	).Scan(
		&item.ID, &item.UserID, &item.ProductID,
		&item.ProductName, &item.UnitPrice, &item.ImageURL,
		&item.Quantity, &item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert cart item: %w", err)
	}
	return item, nil
}

func (r *cartRepository) UpdateQuantity(ctx context.Context, params *model.UpdateCartItemParams) (*model.CartItem, error) {
	query := `
		UPDATE cart_items
		SET quantity = $1, updated_at = NOW()
		WHERE user_id = $2 AND product_id = $3
		RETURNING id, user_id, product_id, product_name, unit_price, image_url, quantity, updated_at
	`
	item := &model.CartItem{}
	err := r.db.QueryRowContext(ctx, query,
		params.Quantity, params.UserID, params.ProductID,
	).Scan(
		&item.ID, &item.UserID, &item.ProductID,
		&item.ProductName, &item.UnitPrice, &item.ImageURL,
		&item.Quantity, &item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCartItemNotFound
		}
		return nil, fmt.Errorf("failed to update cart item: %w", err)
	}
	return item, nil
}

func (r *cartRepository) DeleteItem(ctx context.Context, userID, productID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2`,
		userID, productID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete cart item: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrCartItemNotFound
	}
	return nil
}

func (r *cartRepository) ClearByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM cart_items WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}
	return nil
}
