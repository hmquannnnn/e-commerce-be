package minio

import (
	"io"
	"time"
)

type UploadOptions struct {
	BucketName  string
	File        io.Reader
	FileName    string
	FileSize    int64
	ContentType string
	Prefix      string            // Folder prefix: "avatars/", "products/"
	Metadata    map[string]string // Custom metadata for the file
}

func (o *UploadOptions) Validate() error {
	if o.BucketName == "" {
		return ErrInvalidBucketName
	}
	if o.File == nil {
		return ErrNoFile
	}
	if o.FileSize <= 0 {
		return ErrInvalidFileSize
	}
	return nil
}

type UploadResult struct {
	FilePath   string // Relative path: "bucket-name/prefix/filename.ext" - for storage in DB
	FileName   string // Original filename
	Size       int64
	UploadedAt time.Time
}

type FileInfo struct {
	Name         string
	Size         int64
	ContentType  string
	LastModified time.Time
	ETag         string
}
