package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/humfurie/pulpulitiko/api/pkg/storage"
)

type UploadService struct {
	storage *storage.MinioStorage
}

func NewUploadService(storage *storage.MinioStorage) *UploadService {
	return &UploadService{storage: storage}
}

func (s *UploadService) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*storage.UploadResult, error) {
	if header.Size > storage.GetMaxFileSize() {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of 10MB")
	}

	contentType := header.Header.Get("Content-Type")
	if !storage.IsAllowedMimeType(contentType) {
		return nil, fmt.Errorf("file type not allowed. Allowed types: JPEG, PNG, GIF, WebP, PDF")
	}

	result, err := s.storage.Upload(ctx, file, header.Filename, contentType, header.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return result, nil
}

func (s *UploadService) UploadReader(ctx context.Context, reader io.Reader, filename, contentType string, size int64) (*storage.UploadResult, error) {
	if size > storage.GetMaxFileSize() {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of 10MB")
	}

	if !storage.IsAllowedMimeType(contentType) {
		return nil, fmt.Errorf("file type not allowed. Allowed types: JPEG, PNG, GIF, WebP, PDF")
	}

	result, err := s.storage.Upload(ctx, reader, filename, contentType, size)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return result, nil
}

func (s *UploadService) DeleteFile(ctx context.Context, fileURL string) error {
	key := s.storage.KeyFromURL(fileURL)
	if key == "" {
		return fmt.Errorf("invalid file URL")
	}
	return s.storage.Delete(ctx, key)
}
