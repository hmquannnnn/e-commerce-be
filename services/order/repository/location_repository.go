package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hmquannnnn/e-commerce/order-service/model"
)

type LocationRepository interface {
	ListProvinces(ctx context.Context) ([]model.LocationUnit, error)
	ListDistricts(ctx context.Context, provinceCode string) ([]model.LocationUnit, error)
	ListWards(ctx context.Context, districtCode string) ([]model.LocationUnit, error)
}

type locationRepository struct {
	db *sql.DB
}

func NewLocationRepository(db *sql.DB) LocationRepository {
	return &locationRepository{db: db}
}

func (r *locationRepository) ListProvinces(ctx context.Context) ([]model.LocationUnit, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT code, name, full_name, COALESCE(code_name, '')
		FROM vn_provinces
		ORDER BY code
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list provinces: %w", err)
	}
	defer rows.Close()
	return scanLocationUnits(rows)
}

func (r *locationRepository) ListDistricts(ctx context.Context, provinceCode string) ([]model.LocationUnit, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT code, name, full_name, COALESCE(code_name, '')
		FROM vn_districts
		WHERE province_code = $1
		ORDER BY code
	`, provinceCode)
	if err != nil {
		return nil, fmt.Errorf("failed to list districts: %w", err)
	}
	defer rows.Close()
	return scanLocationUnits(rows)
}

func (r *locationRepository) ListWards(ctx context.Context, districtCode string) ([]model.LocationUnit, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT code, name, full_name, COALESCE(code_name, '')
		FROM vn_wards
		WHERE district_code = $1
		ORDER BY code
	`, districtCode)
	if err != nil {
		return nil, fmt.Errorf("failed to list wards: %w", err)
	}
	defer rows.Close()
	return scanLocationUnits(rows)
}

func scanLocationUnits(rows *sql.Rows) ([]model.LocationUnit, error) {
	items := []model.LocationUnit{}
	for rows.Next() {
		item := model.LocationUnit{}
		if err := rows.Scan(&item.Code, &item.Name, &item.FullName, &item.CodeName); err != nil {
			return nil, fmt.Errorf("failed to scan location unit: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
