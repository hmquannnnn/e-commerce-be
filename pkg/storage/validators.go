package storage

import (
	"errors"
	"mime/multipart"
	"path/filepath"
	"slices"
)

type FileValidator struct {
	MaxSize           int64
	AllowedTypes      []string
	AllowedExtensions []string
}

func NewFileValidator(maxSize int64, allowedTypes []string, allowedExtensions []string) *FileValidator {
	return &FileValidator{
		MaxSize:           maxSize,
		AllowedTypes:      allowedTypes,
		AllowedExtensions: allowedExtensions,
	}
}

func (v *FileValidator) Validate(file *multipart.File, header *multipart.FileHeader) error {
	if header.Size > v.MaxSize {
		return errors.New("file size exceeds the maximum allowed size")
	}

	if !slices.Contains(v.AllowedTypes, header.Header.Get("Content-Type")) {
		return errors.New("file type not allowed")
	}

	if !slices.Contains(v.AllowedExtensions, filepath.Ext(header.Filename)) {
		return errors.New("file extension not allowed")
	}

	return nil
}

func ImageValidator(maxSize int64) *FileValidator {
	return NewFileValidator(maxSize, []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/svg+xml",
	}, []string{
		".jpg",
		".jpeg",
		".png",
		".gif",
		".webp",
		".svg",
	})
}

func VideoValidator(maxSize int64) *FileValidator {
	return NewFileValidator(maxSize, []string{
		"video/mp4",
		"video/mpeg",
		"video/quicktime",
		"video/x-msvideo",
		"video/x-matroska",
		"video/x-flv",
		"video/x-ms-wmv",
	}, []string{
		".mp4",
		".mpeg",
		".quicktime",
		".avi",
		".mkv",
		".flv",
		".wmv",
	})
}

func DocumentValidator(maxSize int64) *FileValidator {
	return NewFileValidator(maxSize, []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}, []string{
		".pdf",
		".doc",
		".docx",
	})
}

func AudioValidator(maxSize int64) *FileValidator {
	return NewFileValidator(maxSize, []string{
		"audio/mpeg",
		"audio/mp3",
		"audio/mp4",
		"audio/ogg",
		"audio/wav",
		"audio/webm",
	}, []string{
		".mp3",
		".mp4",
		".ogg",
		".wav",
		".webm",
	})
}
