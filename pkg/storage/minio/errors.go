package minio

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidBucketName = errors.New("invalid bucket name")
	ErrNoFile            = errors.New("no file provided")
	ErrInvalidFileSize   = errors.New("file size exceeds the maximum allowed")
)

type ErrFileTooLarge struct {
	Size    int64
	MaxSize int64
}

func (e *ErrFileTooLarge) Error() string {
	return fmt.Sprintf("file too large: %d bytes (max: %d bytes)", e.Size, e.MaxSize)
}

type ErrInvalidContentType struct {
	Type string
}

func (e *ErrInvalidContentType) Error() string {
	return fmt.Sprintf("invalid content type: %s", e.Type)
}
