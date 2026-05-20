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

// GenerateID returns a new random UUID for use as a product ID before form submission.
func (h *ProductHandler) GenerateID(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"product_id": uuid.New().String()},
	})
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	images := make([]service.ImageInput, 0, len(req.Images))
	for _, img := range req.Images {
		images = append(images, service.ImageInput{
			URL:          img.URL,
			DisplayOrder: img.DisplayOrder,
			IsPrimary:    img.IsPrimary,
		})
	}

	product, err := h.productService.CreateProduct(c.Request.Context(), service.CreateProductParams{
		ProductID:   req.ProductID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Specs:       req.Specs,
		CategoryID:  req.CategoryID,
		Images:      images,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			respondError(c, http.StatusBadRequest, "INVALID_INPUT", "Invalid product data")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create product")
		return
	}

	// Fetch the product with images to include in response
	productWithImages, err := h.productService.GetProduct(c.Request.Context(), product.ID)
	if err != nil {
		// Fallback: return basic product without images
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
		return
	}

	respondCreated(c, "Product created successfully", ToProductResponse(productWithImages))
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
	if query.MinPrice != nil && query.MaxPrice != nil && *query.MinPrice > *query.MaxPrice {
		respondError(c, http.StatusBadRequest, "INVALID_PRICE_RANGE", "min_price must be less than or equal to max_price")
		return
	}

	filter := model.ListProductsFilter{
		CategoryID: query.CategoryID,
		Search:     query.Search,
		MinPrice:   query.MinPrice,
		MaxPrice:   query.MaxPrice,
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
