package storage

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/hmquannnnn/e-commerce/pkg/storage/minio"
)

type Manager struct {
	minioClient *minio.Client
	validators  map[string]*FileValidator
}

// GetPublicURL generates full URL from filePath
func (m *Manager) GetPublicURL(filePath string) string {
	return m.minioClient.GetPublicURL(filePath)
}

func (m *Manager) GetPresignedPutURL(
	ctx context.Context,
	bucketName, fileName string,
	expiredTime time.Duration,
	contentType string,
) (string, error) {
	return m.minioClient.GetPresignedPutURL(ctx, bucketName, fileName, expiredTime, contentType)
}

func NewManager(minioClient *minio.Client) *Manager {
	return &Manager{
		minioClient: minioClient,
		validators: map[string]*FileValidator{
			"image":    ImageValidator(10 * 1024 * 1024),
			"video":    VideoValidator(10 * 1024 * 1024),
			"document": DocumentValidator(10 * 1024 * 1024),
			"audio":    AudioValidator(10 * 1024 * 1024),
		},
	}
}

func (m *Manager) UploadImage(
	ctx context.Context,
	bucketName string,
	file *multipart.File,
	header *multipart.FileHeader,
	prefix string,
) (minio.UploadResult, error) {
	if err := m.validators["image"].Validate(file, header); err != nil {
		return minio.UploadResult{}, err
	}

	return m.minioClient.Upload(ctx, minio.UploadOptions{
		BucketName:  bucketName,
		File:        *file,
		FileName:    header.Filename,
		FileSize:    header.Size,
		ContentType: header.Header.Get("Content-Type"),
		Prefix:      prefix,
	})
}
