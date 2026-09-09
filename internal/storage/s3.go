package storage

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
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
	MaxImageSize    = 5 * 1024 * 1024 // 5MB
	MaxCoverSize    = 10 * 1024 * 1024 // 10MB untuk cover artikel
	MaxUploadSize   = 20 * 1024 * 1024 // 20MB global
)

var allowedImageTypes = map[string]bool{
	"image/jpeg":            true,
	"image/png":             true,
	"image/webp":            true,
	"image/gif":             true,
	"image/avif":            true,
}

// Storage adalah interface untuk upload file ke object storage
type Storage interface {
	Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (*FileInfo, error)
	Delete(ctx context.Context, key string) error
	URL(key string) string
}

// FileInfo berisi metadata file yang berhasil diupload
type FileInfo struct {
	Key       string `json:"key"`
	URL       string `json:"url"`
	Size      int64  `json:"size"`
	ContentType string `json:"content_type"`
}

type s3Storage struct {
	client    *minio.Client
	cfg       config.S3Config
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
	// Validasi ukuran
	if header.Size > MaxUploadSize {
		return nil, ErrFileTooLarge
	}

	// Validasi tipe dari header Content-Type (fallback ke ekstensi)
	contentType := header.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		// Fallback: cek ekstensi file
		ext := strings.ToLower(filepath.Ext(header.Filename))
		extTypes := map[string]string{
			".jpg": "image/jpeg", ".jpeg": "image/jpeg",
			".png": "image/png", ".webp": "image/webp",
			".gif": "image/gif", ".avif": "image/avif",
		}
		detected, ok := extTypes[ext]
		if !ok {
			return nil, ErrUnsupportedType
		}
		contentType = detected
	}

	// Buat key unik: uploads/2026/08/<uuid>.<ext>
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		case "image/gif":
			ext = ".gif"
		case "image/avif":
			ext = ".avif"
		}
	}
	now := time.Now()
	key := fmt.Sprintf("%s/%04d/%02d/%s%s", folder, now.Year(), now.Month(), uuid.New().String(), ext)

	// Upload ke bucket
	_, err := s.client.PutObject(ctx, s.cfg.Bucket, key, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	return &FileInfo{
		Key:         key,
		URL:         s.URL(key),
		Size:        header.Size,
		ContentType: contentType,
	}, nil
}

func (s *s3Storage) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.cfg.Bucket, key, minio.RemoveObjectOptions{})
}

// URL mengembalikan URL publik file. Jika S3_PUBLIC_URL di-set, pakai itu.
func (s *s3Storage) URL(key string) string {
	if s.cfg.PublicURL != "" {
		return strings.TrimSuffix(s.cfg.PublicURL, "/") + "/" + key
	}
	// Default: URL dari endpoint S3
	scheme := "http"
	if s.cfg.UseSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, s.cfg.Endpoint, s.cfg.Bucket, key)
}

var _ Storage = (*s3Storage)(nil)
