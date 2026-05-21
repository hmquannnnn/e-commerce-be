package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/pkg/common/utils"
	"github.com/hmquannnnn/e-commerce/product-service/model"
)

var (
	ErrInventoryNotFound = errors.New("inventory not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type InventoryRepository interface {
	Create(ctx context.Context, productID uuid.UUID, initialStock int) (*model.Inventory, error)
	GetByProductID(ctx context.Context, productID uuid.UUID) (*model.Inventory, error)
	UpdateStock(ctx context.Context, productID uuid.UUID, params *model.UpdateInventoryParams) error
	ReserveStock(ctx context.Context, params *model.ReserveStockParams) error
	ReserveStocks(ctx context.Context, params []model.ReserveStockParams) error
	ReleaseStock(ctx context.Context, params *model.ReleaseStockParams) error
	ReleaseStocks(ctx context.Context, params []model.ReleaseStockParams) error
}

type inventoryRepository struct {
	db *sql.DB
}

func NewInventoryRepository(db *sql.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) Create(ctx context.Context, productID uuid.UUID, initialStock int) (*model.Inventory, error) {
	query := `
		INSERT INTO inventory (product_id, stock_quantity)
		VALUES ($1, $2)
		RETURNING id, product_id, stock_quantity, reserved_quantity, created_at, updated_at
	`
	inv := &model.Inventory{}
	err := r.db.QueryRowContext(ctx, query, productID, initialStock).Scan(
		&inv.ID, &inv.ProductID, &inv.StockQuantity, &inv.ReservedQuantity,
		&inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create inventory: %w", err)
	}
	return inv, nil
}

func (r *inventoryRepository) GetByProductID(ctx context.Context, productID uuid.UUID) (*model.Inventory, error) {
	query := `
		SELECT id, product_id, stock_quantity, reserved_quantity, created_at, updated_at
		FROM inventory WHERE product_id = $1
	`
	inv := &model.Inventory{}
	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&inv.ID, &inv.ProductID, &inv.StockQuantity, &inv.ReservedQuantity,
		&inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInventoryNotFound
		}
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}
	return inv, nil
}

func (r *inventoryRepository) UpdateStock(ctx context.Context, productID uuid.UUID, params *model.UpdateInventoryParams) error {
	updates := make(map[string]interface{})
	if params.StockQuantity != nil {
		updates["stock_quantity"] = *params.StockQuantity
	}

	rowsAffected, err := utils.ExecPatch(ctx, r.db, utils.PatchUpdate{
		Table:     "inventory",
		Updates:   updates,
		Where:     "product_id = $1",
		WhereArgs: []interface{}{productID},
	})
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}
	if rowsAffected == 0 {
		return ErrInventoryNotFound
	}
	return nil
}

func (r *inventoryRepository) ReserveStock(ctx context.Context, params *model.ReserveStockParams) error {
	return r.reserveStock(ctx, r.db, params)
}

type sqlExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func (r *inventoryRepository) reserveStock(ctx context.Context, exec sqlExecutor, params *model.ReserveStockParams) error {
	query := `
		UPDATE inventory
		SET reserved_quantity = reserved_quantity + $1, updated_at = NOW()
		WHERE product_id = $2
		  AND (stock_quantity - reserved_quantity) >= $1
	`
	result, err := exec.ExecContext(ctx, query, params.Quantity, params.ProductID)
	if err != nil {
		return fmt.Errorf("failed to reserve stock: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		inv, err := r.getByProductID(ctx, exec, params.ProductID)
		if err != nil {
			return ErrInventoryNotFound
		}
		if inv.AvailableQuantity() < params.Quantity {
			return ErrInsufficientStock
		}
		return ErrInventoryNotFound
	}
	return nil
}

func (r *inventoryRepository) ReserveStocks(ctx context.Context, params []model.ReserveStockParams) error {
	if len(params) == 0 {
		return nil
	}

	ordered := append([]model.ReserveStockParams(nil), params...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].ProductID.String() < ordered[j].ProductID.String()
	})

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin inventory transaction: %w", err)
	}
	defer tx.Rollback()

	for i := range ordered {
		if err := r.reserveStock(ctx, tx, &ordered[i]); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit inventory transaction: %w", err)
	}
	return nil
}

func (r *inventoryRepository) ReleaseStock(ctx context.Context, params *model.ReleaseStockParams) error {
	return r.releaseStock(ctx, r.db, params)
}

func (r *inventoryRepository) releaseStock(ctx context.Context, exec sqlExecutor, params *model.ReleaseStockParams) error {
	query := `
		UPDATE inventory
		SET reserved_quantity = GREATEST(0, reserved_quantity - $1), updated_at = NOW()
		WHERE product_id = $2
	`
	result, err := exec.ExecContext(ctx, query, params.Quantity, params.ProductID)
	if err != nil {
		return fmt.Errorf("failed to release stock: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrInventoryNotFound
	}
	return nil
}

func (r *inventoryRepository) ReleaseStocks(ctx context.Context, params []model.ReleaseStockParams) error {
	if len(params) == 0 {
		return nil
	}

	ordered := append([]model.ReleaseStockParams(nil), params...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].ProductID.String() < ordered[j].ProductID.String()
	})

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin inventory transaction: %w", err)
	}
	defer tx.Rollback()

	for i := range ordered {
		if err := r.releaseStock(ctx, tx, &ordered[i]); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit inventory transaction: %w", err)
	}
	return nil
}

func (r *inventoryRepository) getByProductID(ctx context.Context, exec sqlExecutor, productID uuid.UUID) (*model.Inventory, error) {
	query := `
		SELECT id, product_id, stock_quantity, reserved_quantity, created_at, updated_at
		FROM inventory WHERE product_id = $1
	`
	inv := &model.Inventory{}
	err := exec.QueryRowContext(ctx, query, productID).Scan(
		&inv.ID, &inv.ProductID, &inv.StockQuantity, &inv.ReservedQuantity,
		&inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInventoryNotFound
		}
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}
	return inv, nil
}
