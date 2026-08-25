package minio

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client       *minio.Client // internal endpoint — dùng cho các thao tác thực (Upload, BucketExists, ...)
	presignClient *minio.Client // public endpoint  — dùng để tạo presigned URL (client-side crypto, không cần kết nối mạng)
	config       Config
}

func NewClient(config Config) (*Client, error) {
	// "iam": chạy trên AWS — lấy credential TẠM từ IAM role của EC2 qua IMDS,
	// không cần access key tĩnh. "static" (mặc định): MinIO/local dùng key như cũ.
	var creds *credentials.Credentials
	if config.AuthType == "iam" {
		creds = credentials.NewIAM("")
	} else {
		creds = credentials.NewStaticV4(config.AccessKey, config.SecretAccessKey, "")
	}

	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  creds,
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Minio client: %w", err)
	}

	// presignClient dùng PublicEndpoint để URL được ký đúng host mà browser sẽ gọi.
	// Presigning là thao tác crypto thuần client-side — không cần kết nối tới MinIO.
	presignEndpoint := config.PublicEndpoint
	if presignEndpoint == "" {
		presignEndpoint = config.Endpoint
	}
	presignClient, err := minio.New(presignEndpoint, &minio.Options{
		Creds:  creds,
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Minio presign client: %w", err)
	}

	log.Printf("[minio] client init — Endpoint=%q  PublicEndpoint=%q  presignEndpoint=%q  UseSSL=%v  Region=%q  AuthType=%q",
		config.Endpoint, config.PublicEndpoint, presignEndpoint, config.UseSSL, config.Region, config.AuthType)

	return &Client{
		client:        minioClient,
		presignClient: presignClient,
		config:        config,
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

// GetPublicURL generates full URL from filePath (format: "bucket-name/prefix/filename.ext").
// Uses PublicEndpoint so the URL is reachable from the browser.
func (c *Client) GetPublicURL(filePath string) string {
	scheme := "http"
	if c.config.UseSSL {
		scheme = "https"
	}
	endpoint := c.config.PublicEndpoint
	if endpoint == "" {
		endpoint = c.config.Endpoint
	}
	baseURL := strings.TrimSuffix(fmt.Sprintf("%s://%s", scheme, endpoint), "/")
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
	presignedURL, err := c.presignClient.PresignedGetObject(
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

// CreateFolder creates a virtual folder in MinIO by uploading a zero-byte object
// with a trailing slash (e.g. "products/uuid/").
func (c *Client) CreateFolder(ctx context.Context, bucketName, folderPath string) error {
	key := strings.TrimSuffix(folderPath, "/") + "/"
	_, err := c.client.PutObject(
		ctx,
		bucketName,
		key,
		strings.NewReader(""),
		0,
		minio.PutObjectOptions{ContentType: "application/x-directory"},
	)
	if err != nil {
		return fmt.Errorf("failed to create folder %q: %w", key, err)
	}
	return nil
}

func (c *Client) GetPresignedPutURL(
	ctx context.Context,
	bucketName, fileName string,
	expiredTime time.Duration,
	contentType string,
) (string, error) {
	log.Printf("[minio] GetPresignedPutURL — presignClient endpoint=%q  bucket=%q  file=%q  contentType=%q",
		c.config.PublicEndpoint, bucketName, fileName, contentType)

	presignedURL, err := c.presignClient.PresignedPutObject(
		ctx,
		bucketName,
		fileName,
		expiredTime,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned PUT URL: %w", err)
	}

	log.Printf("[minio] presigned PUT URL generated — host=%q  url=%q", presignedURL.Host, presignedURL.String())

	return presignedURL.String(), nil
}
