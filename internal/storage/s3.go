package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/taufiqgit/news-api/internal/config"
)

var (
	ErrUnsupportedType = errors.New("unsupported file type")
	ErrFileTooLarge    = errors.New("file too large")
)

// Ukuran & tipe yang diizinkan
const (
	MaxImageSize  = 5 * 1024 * 1024  // 5MB
	MaxCoverSize  = 10 * 1024 * 1024 // 10MB untuk cover artikel
	MaxUploadSize = 20 * 1024 * 1024 // 20MB global
)

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
	"image/avif": true,
}

// Storage adalah interface untuk upload file ke object storage
type Storage interface {
	Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (*FileInfo, error)
	// UploadReader meng-upload data dari io.Reader (mis. gambar hasil download).
	UploadReader(ctx context.Context, r io.Reader, size int64, filename, contentType, folder string) (*FileInfo, error)
	Delete(ctx context.Context, key string) error
	URL(key string) string
}

// FileInfo berisi metadata file yang berhasil diupload
type FileInfo struct {
	Key         string `json:"key"`
	URL         string `json:"url"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

type s3Storage struct {
	client *minio.Client
	cfg    config.S3Config
}

// New membuat instance storage.
// Jika kredensial S3 diisi → pakai S3; jika kosong → fallback ke local disk (dev).
func New(cfg config.S3Config) (Storage, error) {
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return NewLocal("uploads")
	}
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init s3 client: %w", err)
	}
	return &s3Storage{client: client, cfg: cfg}, nil
}

func (s *s3Storage) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (*FileInfo, error) {
	if header.Size > MaxUploadSize {
		return nil, ErrFileTooLarge
	}
	return s.putObject(ctx, file, header.Size, header.Filename, header.Header.Get("Content-Type"), folder)
}

// UploadReader meng-upload data dari io.Reader ke S3.
func (s *s3Storage) UploadReader(ctx context.Context, r io.Reader, size int64, filename, contentType, folder string) (*FileInfo, error) {
	if size > MaxUploadSize {
		return nil, ErrFileTooLarge
	}
	return s.putObject(ctx, r, size, filename, contentType, folder)
}

func (s *s3Storage) putObject(ctx context.Context, r io.Reader, size int64, filename, contentType, folder string) (*FileInfo, error) {
	contentType, ext, err := resolveImageType(contentType, filename)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	key := fmt.Sprintf("%s/%04d/%02d/%s%s", folder, now.Year(), now.Month(), uuid.New().String(), ext)

	if _, err := s.client.PutObject(ctx, s.cfg.Bucket, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	}); err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	return &FileInfo{
		Key:         key,
		URL:         s.URL(key),
		Size:        size,
		ContentType: contentType,
	}, nil
}

func (s *s3Storage) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.cfg.Bucket, key, minio.RemoveObjectOptions{})
}

// URL mengembalikan URL publik path-style: {base}/{bucket}/{key}.
// S3_PUBLIC_URL, jika diisi, harus berupa base endpoint/CDN tanpa nama bucket,
// misalnya https://nos.jkt-1.neo.id (bukan https://news.nos.jkt-1.neo.id).
func (s *s3Storage) URL(key string) string {
	baseURL := strings.TrimSuffix(s.cfg.PublicURL, "/")
	if baseURL == "" {
		scheme := "http"
		if s.cfg.UseSSL {
			scheme = "https"
		}
		baseURL = fmt.Sprintf("%s://%s", scheme, s.cfg.Endpoint)
	}
	return fmt.Sprintf("%s/%s/%s", baseURL, s.cfg.Bucket, strings.TrimPrefix(key, "/"))
}

var _ Storage = (*s3Storage)(nil)
