package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/order-service/service"
)

type CartHandler struct {
	cartService service.CartService
}

func NewCartHandler(cartService service.CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

func (h *CartHandler) GetCart(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	items, err := h.cartService.GetCart(c.Request.Context(), userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get cart")
		return
	}

	respondOK(c, "Cart retrieved successfully", ToCartResponse(items))
}

func (h *CartHandler) AddItem(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req AddCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	item, err := h.cartService.AddItem(c.Request.Context(), userID, req.ProductID, req.Quantity)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			respondError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found")
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			respondError(c, http.StatusBadRequest, "INVALID_INPUT", "Quantity must be greater than 0")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to add item to cart")
		return
	}

	respondCreated(c, "Item added to cart", ToCartItemResponse(item))
}

func (h *CartHandler) UpdateItem(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	var req UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	item, err := h.cartService.UpdateItemQuantity(c.Request.Context(), userID, productID, req.Quantity)
	if err != nil {
		if errors.Is(err, service.ErrCartItemNotFound) {
			respondError(c, http.StatusNotFound, "CART_ITEM_NOT_FOUND", "Cart item not found")
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			respondError(c, http.StatusBadRequest, "INVALID_INPUT", "Quantity must be greater than 0")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update cart item")
		return
	}

	respondOK(c, "Cart item updated", ToCartItemResponse(item))
}

func (h *CartHandler) DeleteItem(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	if err := h.cartService.RemoveItem(c.Request.Context(), userID, productID); err != nil {
		if errors.Is(err, service.ErrCartItemNotFound) {
			respondError(c, http.StatusNotFound, "CART_ITEM_NOT_FOUND", "Cart item not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove cart item")
		return
	}

	respondOK(c, "Cart item removed", nil)
}
