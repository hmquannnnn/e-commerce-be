package minio

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client *minio.Client
	config Config
}

func NewClient(config Config) (*Client, error) {
	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			config.AccessKey,
			config.SecretAccessKey,
			"",
		),
		Secure: config.UseSSL,
		Region: config.Region,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create Minio client: %w", err)
	}

	return &Client{
		client: minioClient,
		config: config,
	}, nil
}

func (c *Client) EnsureBucket(ctx context.Context, bucketName string) error {
	exists, err := c.client.BucketExists(ctx, bucketName)

	if err != nil {
		return fmt.Errorf("failed to check bucket exists: %w", err)
	}

	if !exists {
		return c.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
			Region: c.config.Region,
		})
	}

	return nil
}

func (c *Client) Upload(ctx context.Context, options UploadOptions) (UploadResult, error) {
	if err := options.Validate(); err != nil {
		return UploadResult{}, err
	}

	info, err := c.client.PutObject(
		ctx,
		options.BucketName,
		options.FileName,
		options.File,
		options.FileSize,
		minio.PutObjectOptions{
			ContentType: options.ContentType,
		},
	)

	if err != nil {
		return UploadResult{}, fmt.Errorf("failed to upload file: %w", err)
	}

	// FilePath format: "bucket-name/prefix/filename.ext" (relative path for DB storage)
	filePath := fmt.Sprintf("%s/%s", options.BucketName, info.Key)

	return UploadResult{
		FilePath:   filePath, // Store relative path, not full URL
		FileName:   info.Key,
		Size:       info.Size,
		UploadedAt: time.Now().UTC(),
	}, nil
}

// GetPublicURL generates full URL from filePath (format: "bucket-name/prefix/filename.ext")
// Uses MINIO_ENDPOINT and MINIO_USE_SSL to build the base URL, then appends filePath (relative path)
func (c *Client) GetPublicURL(filePath string) string {
	// Build base URL from endpoint and SSL setting
	scheme := "http"
	if c.config.UseSSL {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, c.config.Endpoint)

	// Ensure baseURL doesn't end with /
	baseURL = strings.TrimSuffix(baseURL, "/")
	// Ensure filePath starts with /
	if !strings.HasPrefix(filePath, "/") {
		filePath = "/" + filePath
	}
	return baseURL + filePath
}

// GetPublicURLFromParts generates full URL from bucket and filename
// Uses MINIO_ENDPOINT and MINIO_USE_SSL to build the base URL
func (c *Client) GetPublicURLFromParts(bucketName, fileName string) string {
	// Build filePath from parts
	filePath := fmt.Sprintf("%s/%s", bucketName, fileName)
	return c.GetPublicURL(filePath)
}

func (c *Client) GetPresignedURL(
	ctx context.Context,
	bucketName, fileName string,
	expiredTime time.Duration,
) (string, error) {
	presignedURL, err := c.client.PresignedGetObject(
		ctx,
		bucketName,
		fileName,
		expiredTime,
		nil,
	)

	if err != nil {
		return "", fmt.Errorf("failed to get presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

func (c *Client) GetPresignedPutURL(
	ctx context.Context,
	bucketName, fileName string,
	expiredTime time.Duration,
	contentType string,
) (string, error) {
	presignedURL, err := c.client.PresignedPutObject(
		ctx,
		bucketName,
		fileName,
		expiredTime,
	)

	if err != nil {
		return "", fmt.Errorf("failed to get presigned PUT URL: %w", err)
	}

	if contentType != "" {
		presignedURL.RawQuery += fmt.Sprintf("&Content-Type=%s", contentType)
	}

	return presignedURL.String(), nil
}
