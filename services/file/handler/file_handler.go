package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hmquannnnn/e-commerce/file-service/service"
)

type FileHandler struct {
	fileService service.FileService
}

func NewFileHandler(fileService service.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

type GetPresignedURLRequest struct {
	FileType    string `json:"file_type" binding:"required,oneof=avatar image product document"` // avatar, image, product, document
	ContentType string `json:"content_type" binding:"required"`                                  // e.g., image/jpeg, image/png
}

func (h *FileHandler) GetPresignedUploadURL(c *gin.Context) {
	var req GetPresignedURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "INVALID_REQUEST",
			"message": "Invalid request body",
		})
		return
	}

	response, err := h.fileService.GetPresignedUploadURL(c.Request.Context(), req.FileType, req.ContentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "INTERNAL_ERROR",
			"message": "Failed to generate presigned URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Presigned URL generated successfully",
		"data":    response,
	})
}
