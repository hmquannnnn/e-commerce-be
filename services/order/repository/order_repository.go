package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/order-service/model"
)

var (
	ErrOrderNotFound           = errors.New("order not found")
	ErrCartQuantityMismatch    = errors.New("cart quantity mismatch during checkout")
)

type OrderRepository interface {
	Create(ctx context.Context, params *model.CreateOrderParams, items []*model.CartItem, userID uuid.UUID, cartDeductions map[uuid.UUID]int) (*model.OrderWithItems, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.OrderWithItems, error)
	List(ctx context.Context, filter model.ListOrdersFilter) ([]*model.Order, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) error
	ListExpiredPendingOrders(ctx context.Context, cutoff time.Time) ([]uuid.UUID, error)
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, params *model.CreateOrderParams, items []*model.CartItem, userID uuid.UUID, cartDeductions map[uuid.UUID]int) (*model.OrderWithItems, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var totalPrice float64
	for _, item := range items {
		totalPrice += item.UnitPrice * float64(item.Quantity)
	}

	order := &model.Order{}
	err = tx.QueryRowContext(ctx, `
		INSERT INTO orders (user_id, total_price, status, payment_method)
		VALUES ($1, $2, 'PENDING', $3)
		RETURNING id, user_id, total_price, status, payment_method, created_at, updated_at
	`, params.UserID, totalPrice, params.PaymentMethod).Scan(
		&order.ID, &order.UserID, &order.TotalPrice,
		&order.Status, &order.PaymentMethod,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	orderItems := make([]model.OrderItem, 0, len(items))
	for _, cartItem := range items {
		item := model.OrderItem{}
		err = tx.QueryRowContext(ctx, `
			INSERT INTO order_items (order_id, product_id, product_name, unit_price, image_url, quantity)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, order_id, product_id, product_name, unit_price, image_url, quantity
		`, order.ID, cartItem.ProductID, cartItem.ProductName,
			cartItem.UnitPrice, cartItem.ImageURL, cartItem.Quantity,
		).Scan(
			&item.ID, &item.OrderID, &item.ProductID,
			&item.ProductName, &item.UnitPrice, &item.ImageURL, &item.Quantity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert order item: %w", err)
		}
		orderItems = append(orderItems, item)
	}

	for productID, deductQty := range cartDeductions {
		var remaining int
		err = tx.QueryRowContext(ctx, `
			SELECT quantity FROM cart_items
			WHERE user_id = $1 AND product_id = $2
			FOR UPDATE
		`, userID, productID).Scan(&remaining)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("%w", ErrCartQuantityMismatch)
			}
			return nil, fmt.Errorf("failed to lock cart row: %w", err)
		}
		if remaining < deductQty {
			return nil, fmt.Errorf("%w", ErrCartQuantityMismatch)
		}
		if remaining == deductQty {
			_, err = tx.ExecContext(ctx, `
				DELETE FROM cart_items WHERE user_id = $1 AND product_id = $2
			`, userID, productID)
		} else {
			_, err = tx.ExecContext(ctx, `
				UPDATE cart_items
				SET quantity = quantity - $1, updated_at = NOW()
				WHERE user_id = $2 AND product_id = $3
			`, deductQty, userID, productID)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to update cart after order: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &model.OrderWithItems{Order: *order, Items: orderItems}, nil
}

func (r *orderRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.OrderWithItems, error) {
	order := &model.Order{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, total_price, status, payment_method, created_at, updated_at
		FROM orders WHERE id = $1
	`, id).Scan(
		&order.ID, &order.UserID, &order.TotalPrice,
		&order.Status, &order.PaymentMethod,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, order_id, product_id, product_name, unit_price, image_url, quantity
		FROM order_items WHERE order_id = $1
	`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	var items []model.OrderItem
	for rows.Next() {
		item := model.OrderItem{}
		if err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID,
			&item.ProductName, &item.UnitPrice, &item.ImageURL, &item.Quantity,
		); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &model.OrderWithItems{Order: *order, Items: items}, nil
}

func (r *orderRepository) List(ctx context.Context, filter model.ListOrdersFilter) ([]*model.Order, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}

	conditions := []string{}
	args := []interface{}{}
	argIdx := 1

	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, *filter.UserID)
		argIdx++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	if err := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM orders %s", where),
		args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	offset := (filter.Page - 1) * filter.Limit
	dataArgs := append(args, filter.Limit, offset)
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, user_id, total_price, status, payment_method, created_at, updated_at
		FROM orders %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1), dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		o := &model.Order{}
		if err := rows.Scan(
			&o.ID, &o.UserID, &o.TotalPrice,
			&o.Status, &o.PaymentMethod,
			&o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, o)
	}
	return orders, total, rows.Err()
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrOrderNotFound
	}
	return nil
}

func (r *orderRepository) ListExpiredPendingOrders(ctx context.Context, cutoff time.Time) ([]uuid.UUID, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id FROM orders
		WHERE status = 'PENDING'
		  AND payment_method != 'CASH'
		  AND created_at < $1
	`, cutoff)
	if err != nil {
		return nil, fmt.Errorf("failed to list expired orders: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan expired order id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
