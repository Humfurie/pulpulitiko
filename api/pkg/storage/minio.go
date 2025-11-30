package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	client         *minio.Client
	bucketName     string
	endpoint       string
	publicEndpoint string
	useSSL         bool
}

type UploadResult struct {
	Key      string `json:"key"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

func NewMinioStorage(endpoint, publicEndpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinioStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create bucket if it doesn't exist
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}

		// Set bucket policy to allow public read access
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, bucket)

		err = client.SetBucketPolicy(ctx, bucket, policy)
		if err != nil {
			return nil, fmt.Errorf("failed to set bucket policy: %w", err)
		}
	}

	return &MinioStorage{
		client:         client,
		bucketName:     bucket,
		endpoint:       endpoint,
		publicEndpoint: publicEndpoint,
		useSSL:         useSSL,
	}, nil
}

func (s *MinioStorage) Upload(ctx context.Context, reader io.Reader, fileName string, contentType string, size int64) (*UploadResult, error) {
	// Generate unique key
	ext := filepath.Ext(fileName)
	key := fmt.Sprintf("%s/%s%s", time.Now().Format("2006/01"), uuid.New().String(), ext)

	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	info, err := s.client.PutObject(ctx, s.bucketName, key, reader, size, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	protocol := "http"
	if s.useSSL {
		protocol = "https"
	}

	return &UploadResult{
		Key:      key,
		URL:      fmt.Sprintf("%s://%s/%s/%s", protocol, s.publicEndpoint, s.bucketName, key),
		Size:     info.Size,
		MimeType: contentType,
	}, nil
}

func (s *MinioStorage) Delete(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *MinioStorage) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucketName, key, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}

func (s *MinioStorage) GetURL(key string) string {
	protocol := "http"
	if s.useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, s.publicEndpoint, s.bucketName, key)
}

func (s *MinioStorage) KeyFromURL(fileURL string) string {
	prefix := fmt.Sprintf("/%s/", s.bucketName)
	idx := strings.Index(fileURL, prefix)
	if idx == -1 {
		return ""
	}
	return fileURL[idx+len(prefix):]
}

func IsAllowedMimeType(mimeType string) bool {
	allowed := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/gif":       true,
		"image/webp":      true,
		"application/pdf": true,
	}
	return allowed[mimeType]
}

func GetMaxFileSize() int64 {
	return 10 * 1024 * 1024 // 10MB
}
