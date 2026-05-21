package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hmquannnnn/e-commerce/order-service/service"
)

type LocationHandler struct {
	locationService service.LocationService
}

func NewLocationHandler(locationService service.LocationService) *LocationHandler {
	return &LocationHandler{locationService: locationService}
}

func (h *LocationHandler) ListProvinces(c *gin.Context) {
	items, err := h.locationService.ListProvinces(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list provinces")
		return
	}
	respondOK(c, "Provinces retrieved successfully", items)
}

func (h *LocationHandler) ListDistricts(c *gin.Context) {
	items, err := h.locationService.ListDistricts(c.Request.Context(), c.Query("province_code"))
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			respondError(c, http.StatusBadRequest, "INVALID_PROVINCE_CODE", "province_code is required")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list districts")
		return
	}
	respondOK(c, "Districts retrieved successfully", items)
}

func (h *LocationHandler) ListWards(c *gin.Context) {
	items, err := h.locationService.ListWards(c.Request.Context(), c.Query("district_code"))
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			respondError(c, http.StatusBadRequest, "INVALID_DISTRICT_CODE", "district_code is required")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list wards")
		return
	}
	respondOK(c, "Wards retrieved successfully", items)
}
