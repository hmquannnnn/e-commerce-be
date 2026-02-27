package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/product-service/service"
)

type InventoryHandler struct {
	inventoryService service.InventoryService
}

func NewInventoryHandler(inventoryService service.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService}
}

func (h *InventoryHandler) GetByProductID(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	inv, err := h.inventoryService.GetInventory(c.Request.Context(), productID)
	if err != nil {
		if errors.Is(err, service.ErrInventoryNotFound) {
			respondError(c, http.StatusNotFound, "INVENTORY_NOT_FOUND", "Inventory not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get inventory")
		return
	}

	respondOK(c, "Inventory retrieved successfully", ToInventoryResponse(inv))
}

func (h *InventoryHandler) UpdateStock(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	var req UpdateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	inv, err := h.inventoryService.UpdateStock(c.Request.Context(), productID, req.StockQuantity)
	if err != nil {
		if errors.Is(err, service.ErrInventoryNotFound) {
			respondError(c, http.StatusNotFound, "INVENTORY_NOT_FOUND", "Inventory not found")
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			respondError(c, http.StatusBadRequest, "INVALID_INPUT", "Stock quantity must be non-negative")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update stock")
		return
	}

	respondOK(c, "Stock updated successfully", ToInventoryResponse(inv))
}

func (h *InventoryHandler) ReserveStock(c *gin.Context) {
	var req ReserveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.inventoryService.ReserveStock(c.Request.Context(), req.ProductID, req.Quantity); err != nil {
		if errors.Is(err, service.ErrInsufficientStock) {
			respondError(c, http.StatusConflict, "INSUFFICIENT_STOCK", "Insufficient stock available")
			return
		}
		if errors.Is(err, service.ErrInventoryNotFound) {
			respondError(c, http.StatusNotFound, "INVENTORY_NOT_FOUND", "Inventory not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to reserve stock")
		return
	}

	respondOK(c, "Stock reserved successfully", nil)
}

func (h *InventoryHandler) ReleaseStock(c *gin.Context) {
	var req ReleaseStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.inventoryService.ReleaseStock(c.Request.Context(), req.ProductID, req.Quantity); err != nil {
		if errors.Is(err, service.ErrInventoryNotFound) {
			respondError(c, http.StatusNotFound, "INVENTORY_NOT_FOUND", "Inventory not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to release stock")
		return
	}

	respondOK(c, "Stock released successfully", nil)
}
