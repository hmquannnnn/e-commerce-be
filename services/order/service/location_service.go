package service

import (
	"context"
	"strings"

	"github.com/hmquannnnn/e-commerce/order-service/model"
	"github.com/hmquannnnn/e-commerce/order-service/repository"
)

type LocationService interface {
	ListProvinces(ctx context.Context) ([]model.LocationUnit, error)
	ListDistricts(ctx context.Context, provinceCode string) ([]model.LocationUnit, error)
	ListWards(ctx context.Context, districtCode string) ([]model.LocationUnit, error)
}

type locationService struct {
	locationRepo repository.LocationRepository
}

func NewLocationService(locationRepo repository.LocationRepository) LocationService {
	return &locationService{locationRepo: locationRepo}
}

func (s *locationService) ListProvinces(ctx context.Context) ([]model.LocationUnit, error) {
	return s.locationRepo.ListProvinces(ctx)
}

func (s *locationService) ListDistricts(ctx context.Context, provinceCode string) ([]model.LocationUnit, error) {
	provinceCode = strings.TrimSpace(provinceCode)
	if provinceCode == "" {
		return nil, ErrInvalidInput
	}
	return s.locationRepo.ListDistricts(ctx, provinceCode)
}

func (s *locationService) ListWards(ctx context.Context, districtCode string) ([]model.LocationUnit, error) {
	districtCode = strings.TrimSpace(districtCode)
	if districtCode == "" {
		return nil, ErrInvalidInput
	}
	return s.locationRepo.ListWards(ctx, districtCode)
}
