package handler

import (
	"errors"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/product-service/model"
	"github.com/hmquannnnn/e-commerce/product-service/service"
)

type ProductHandler struct {
	productService service.ProductService
}

func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	product, err := h.productService.CreateProduct(c.Request.Context(), service.CreateProductParams{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Specs:       req.Specs,
		CategoryID:  req.CategoryID,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			respondError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid product data")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create product")
		return
	}

	// Return product without images since it's brand new
	resp := &ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Specs:       product.Specs,
		CategoryID:  product.CategoryID,
		Images:      []ProductImageResponse{},
		CreatedAt:   product.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	respondCreated(c, "Product created successfully", resp)
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	product, err := h.productService.GetProduct(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			respondError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get product")
		return
	}

	respondOK(c, "Product retrieved successfully", ToProductResponse(product))
}

func (h *ProductHandler) List(c *gin.Context) {
	var query ListProductsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_QUERY", "Invalid query parameters")
		return
	}

	filter := model.ListProductsFilter{
		CategoryID: query.CategoryID,
		Search:     query.Search,
		Page:       query.Page,
		Limit:      query.Limit,
	}

	products, total, err := h.productService.ListProducts(c.Request.Context(), filter)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list products")
		return
	}

	items := make([]*ProductListItemResponse, 0, len(products))
	for _, p := range products {
		items = append(items, ToProductListItemResponse(p))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Products retrieved successfully",
		"data": gin.H{
			"items":       items,
			"total":       total,
			"page":        filter.Page,
			"limit":       filter.Limit,
			"total_pages": int(math.Ceil(float64(total) / float64(filter.Limit))),
		},
	})
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	product, err := h.productService.UpdateProduct(c.Request.Context(), id, service.UpdateProductParams{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Specs:       req.Specs,
		CategoryID:  req.CategoryID,
	})
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			respondError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found")
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			respondError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid product data")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update product")
		return
	}

	respondOK(c, "Product updated successfully", ToProductResponse(product))
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	if err := h.productService.DeleteProduct(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			respondError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "Product not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete product")
		return
	}

	respondOK(c, "Product deleted successfully", nil)
}

func (h *ProductHandler) AddImage(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID")
		return
	}

	var req AddImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	img, err := h.productService.AddProductImage(c.Request.Context(), productID, req.URL, req.DisplayOrder, req.IsPrimary)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to add image")
		return
	}

	respondCreated(c, "Image added successfully", ProductImageResponse{
		ID:           img.ID,
		ProductID:    img.ProductID,
		URL:          img.URL,
		DisplayOrder: img.DisplayOrder,
		IsPrimary:    img.IsPrimary,
		CreatedAt:    img.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *ProductHandler) DeleteImage(c *gin.Context) {
	imageID, err := uuid.Parse(c.Param("image_id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid image ID")
		return
	}

	if err := h.productService.DeleteProductImage(c.Request.Context(), imageID); err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete image")
		return
	}

	respondOK(c, "Image deleted successfully", nil)
}
