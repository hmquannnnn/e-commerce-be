package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/hmquannnnn/e-commerce/user-service/model"
)

var (
	ErrCityNotFound     = errors.New("city not found")
	ErrDistrictNotFound = errors.New("district not found")
	ErrWardNotFound     = errors.New("ward not found")
)

// LocationRepository defines the interface for location data access
type LocationRepository interface {
	// City operations
	GetAllCities(ctx context.Context) ([]*model.City, error)

	// District operations
	GetDistrictsByCityID(ctx context.Context, cityID int) ([]*model.District, error)

	// Ward operations
	GetWardsByDistrictID(ctx context.Context, districtID int) ([]*model.Ward, error)
}

type locationRepository struct {
	db *sql.DB
}

// NewLocationRepository creates a new location repository
func NewLocationRepository(db *sql.DB) LocationRepository {
	return &locationRepository{db: db}
}

func (r *locationRepository) GetAllCities(ctx context.Context) ([]*model.City, error) {
	query := `
		SELECT id, name, code, created_at, updated_at
		FROM cities
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get cities: %w", err)
	}
	defer rows.Close()

	cities := make([]*model.City, 0)
	for rows.Next() {
		city := &model.City{}
		err := rows.Scan(&city.ID, &city.Name, &city.Code, &city.CreatedAt, &city.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan city: %w", err)
		}
		cities = append(cities, city)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cities: %w", err)
	}

	return cities, nil
}

func (r *locationRepository) GetDistrictsByCityID(ctx context.Context, cityID int) ([]*model.District, error) {
	query := `
		SELECT id, city_id, name, code, created_at, updated_at
		FROM districts
		WHERE city_id = $1
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, cityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get districts: %w", err)
	}
	defer rows.Close()

	districts := make([]*model.District, 0)
	for rows.Next() {
		district := &model.District{}
		err := rows.Scan(
			&district.ID, &district.CityID, &district.Name, &district.Code,
			&district.CreatedAt, &district.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan district: %w", err)
		}
		districts = append(districts, district)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating districts: %w", err)
	}

	return districts, nil
}

func (r *locationRepository) GetWardsByDistrictID(ctx context.Context, districtID int) ([]*model.Ward, error) {
	query := `
		SELECT id, district_id, name, code, created_at, updated_at
		FROM wards
		WHERE district_id = $1
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, districtID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wards: %w", err)
	}
	defer rows.Close()

	wards := make([]*model.Ward, 0)
	for rows.Next() {
		ward := &model.Ward{}
		err := rows.Scan(
			&ward.ID, &ward.DistrictID, &ward.Name, &ward.Code,
			&ward.CreatedAt, &ward.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ward: %w", err)
		}
		wards = append(wards, ward)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating wards: %w", err)
	}

	return wards, nil
}
