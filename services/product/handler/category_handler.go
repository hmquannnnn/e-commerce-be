package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hmquannnnn/e-commerce/product-service/service"
)

type CategoryHandler struct {
	categoryService service.CategoryService
}

func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	cat, err := h.categoryService.CreateCategory(c.Request.Context(), service.CreateCategoryParams{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, service.ErrCategoryAlreadyExists) {
			respondError(c, http.StatusConflict, "CATEGORY_EXISTS", "Category with this name already exists")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create category")
		return
	}

	respondCreated(c, "Category created successfully", ToCategoryResponse(cat))
}

func (h *CategoryHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid category ID")
		return
	}

	cat, err := h.categoryService.GetCategory(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			respondError(c, http.StatusNotFound, "CATEGORY_NOT_FOUND", "Category not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get category")
		return
	}

	respondOK(c, "Category retrieved successfully", ToCategoryResponse(cat))
}

func (h *CategoryHandler) List(c *gin.Context) {
	cats, err := h.categoryService.ListCategories(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list categories")
		return
	}

	resp := make([]*CategoryResponse, 0, len(cats))
	for _, cat := range cats {
		resp = append(resp, ToCategoryResponse(cat))
	}

	respondOK(c, "Categories retrieved successfully", resp)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid category ID")
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	cat, err := h.categoryService.UpdateCategory(c.Request.Context(), id, service.UpdateCategoryParams{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			respondError(c, http.StatusNotFound, "CATEGORY_NOT_FOUND", "Category not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update category")
		return
	}

	respondOK(c, "Category updated successfully", ToCategoryResponse(cat))
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid category ID")
		return
	}

	if err := h.categoryService.DeleteCategory(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			respondError(c, http.StatusNotFound, "CATEGORY_NOT_FOUND", "Category not found")
			return
		}
		if errors.Is(err, service.ErrCategoryHasProducts) {
			respondError(c, http.StatusConflict, "CATEGORY_HAS_PRODUCTS", "Cannot delete category that still has products")
			return
		}
		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete category")
		return
	}

	respondOK(c, "Category deleted successfully", nil)
}
