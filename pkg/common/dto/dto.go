package dto

import (
	"time"

	"github.com/hmquannnnn/e-commerce/pkg/storage/minio"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	TotalItems int64       `json:"total_items"`
	TotalPages int         `json:"total_pages"`
}

type FileUploadResponse struct {
	FileName   string `json:"file_name"` // Original filename
	URL        string `json:"url"`       // Full URL for client access (MINIO_URL + relative path)
	Size       int64  `json:"size"`
	UploadedAt string `json:"uploaded_at"`
}

// ToFileUploadResponse converts UploadResult to FileUploadResponse
// urlGenerator should be a function that takes filePath (relative path) and returns full URL
func ToFileUploadResponse(result minio.UploadResult, urlGenerator func(string) string) FileUploadResponse {
	return FileUploadResponse{
		FileName:   result.FileName,
		URL:        urlGenerator(result.FilePath),
		Size:       result.Size,
		UploadedAt: result.UploadedAt.Format(time.RFC3339),
	}
}
