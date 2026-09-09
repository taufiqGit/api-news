package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// localStorage menyimpan file ke disk — fallback saat S3 tidak dikonfigurasi (dev)
type localStorage struct {
	dir string
}

func NewLocal(dir string) (Storage, error) {
	if dir == "" {
		dir = "uploads"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create upload dir: %w", err)
	}
	return &localStorage{dir: dir}, nil
}

func (s *localStorage) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (*FileInfo, error) {
	// Validasi ukuran
	if header.Size > MaxUploadSize {
		return nil, ErrFileTooLarge
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
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
	relPath := fmt.Sprintf("%s/%04d/%02d/%s%s", folder, now.Year(), now.Month(), uuid.New().String(), ext)
	fullPath := filepath.Join(s.dir, relPath)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return nil, err
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, err
	}

	return &FileInfo{
		Key:         relPath,
		URL:         s.URL(relPath),
		Size:        header.Size,
		ContentType: contentType,
	}, nil
}

func (s *localStorage) Delete(ctx context.Context, key string) error {
	return os.Remove(filepath.Join(s.dir, key))
}

func (s *localStorage) URL(key string) string {
	return "/uploads/" + key
}

var _ Storage = (*localStorage)(nil)
