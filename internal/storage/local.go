package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
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
	if header.Size > MaxUploadSize {
		return nil, ErrFileTooLarge
	}
	return s.save(ctx, file, header.Size, header.Filename, header.Header.Get("Content-Type"), folder)
}

// UploadReader meng-upload data dari io.Reader ke disk.
func (s *localStorage) UploadReader(ctx context.Context, r io.Reader, size int64, filename, contentType, folder string) (*FileInfo, error) {
	if size > MaxUploadSize {
		return nil, ErrFileTooLarge
	}
	return s.save(ctx, r, size, filename, contentType, folder)
}

func (s *localStorage) save(ctx context.Context, r io.Reader, size int64, filename, contentType, folder string) (*FileInfo, error) {
	contentType, ext, err := resolveImageType(contentType, filename)
	if err != nil {
		return nil, err
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

	if _, err := io.Copy(dst, r); err != nil {
		return nil, err
	}

	return &FileInfo{
		Key:         relPath,
		URL:         s.URL(relPath),
		Size:        size,
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
