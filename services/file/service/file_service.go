package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/pkg/storage"
)

type FileService interface {
	GetPresignedUploadURL(ctx context.Context, fileType, contentType string) (PresignedUploadResponse, error)
}

type fileService struct {
	storageManager *storage.Manager
	bucketName     string
}

type PresignedUploadResponse struct {
	PresignedURL string `json:"presigned_url"`
	FilePath     string `json:"file_path"`
	ExpiresIn    int    `json:"expires_in"`
}

func NewFileService(
	storageManager *storage.Manager,
	bucketName string,
) FileService {
	return &fileService{
		storageManager: storageManager,
		bucketName:     bucketName,
	}
}

func (s *fileService) GetPresignedUploadURL(
	ctx context.Context,
	fileType, contentType string,
) (PresignedUploadResponse, error) {
	fileID := uuid.New().String()

	var prefix string
	switch fileType {
	case "avatar", "image":
		prefix = "avatars"
	case "product":
		prefix = "products"
	case "document":
		prefix = "documents"
	default:
		prefix = "uploads"
	}

	// Generate file key: prefix/uuid.ext
	// Extract extension from content type if possible
	ext := getExtensionFromContentType(contentType)
	fileName := fmt.Sprintf("%s/%s%s", prefix, fileID, ext)

	// Generate presigned URL (valid for 15 minutes)
	expiredTime := 15 * time.Minute
	presignedURL, err := s.storageManager.GetPresignedPutURL(
		ctx,
		s.bucketName,
		fileName,
		expiredTime,
		contentType,
	)
	if err != nil {
		return PresignedUploadResponse{}, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// FilePath format: "bucket-name/prefix/filename.ext" (for DB storage)
	filePath := fmt.Sprintf("%s/%s", s.bucketName, fileName)

	return PresignedUploadResponse{
		PresignedURL: presignedURL,
		FilePath:     filePath,
		ExpiresIn:    int(expiredTime.Seconds()),
	}, nil
}

// getExtensionFromContentType extracts file extension from content type
func getExtensionFromContentType(contentType string) string {
	extMap := map[string]string{
		"image/jpeg":      ".jpg",
		"image/png":       ".png",
		"image/gif":       ".gif",
		"image/webp":      ".webp",
		"image/svg+xml":   ".svg",
		"application/pdf": ".pdf",
		"video/mp4":       ".mp4",
		"video/quicktime": ".mov",
	}

	if ext, ok := extMap[contentType]; ok {
		return ext
	}
	return ".bin" // Default extension
}
