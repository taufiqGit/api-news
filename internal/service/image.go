package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/taufiqgit/news-api/internal/storage"
)

const maxImageDownload = 10 * 1024 * 1024 // 10MB (cover)

// downloadAndUpload mengunduh gambar dari URL, menambahkan watermark kredit,
// lalu meng-uploadnya ke storage. Mengembalikan FileInfo + credit.
// Error dikembalikan agar pemanggil bisa graceful-degrade (artikel tetap disimpan).
func (p *Pipeline) downloadAndUpload(ctx context.Context, imageURL, credit, siteName string) (*storage.FileInfo, *string, error) {
	data, contentType, err := p.download(ctx, imageURL)
	if err != nil {
		return nil, nil, err
	}

	creditText := credit
	if creditText == "" {
		creditText = siteName
	}

	// Watermark credit pada gambar (best-effort; gambar tetap dipakai bila gagal).
	watermarked, wmErr := applyWatermark(data, contentType, creditText)
	if wmErr != nil {
		p.logger.Warn("watermark failed, using original", "url", imageURL, "error", wmErr)
		watermarked = data
	}

	filename := imageFilename(imageURL, contentType)
	info, err := p.storage.UploadReader(ctx, bytes.NewReader(watermarked), int64(len(watermarked)), filename, contentType, p.imageFolder)
	if err != nil {
		return nil, nil, fmt.Errorf("service: upload image: %w", err)
	}

	c := creditText
	return info, &c, nil
}

// download mengunduh gambar dan mengembalikan bytes + content type.
func (p *Pipeline) download(ctx context.Context, imageURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("service: build image request: %w", err)
	}
	req.Header.Set("User-Agent", "news-api-scheduler/1.0")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("service: download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("service: image %s returned status %d", imageURL, resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxImageDownload+1))
	if err != nil {
		return nil, "", fmt.Errorf("service: read image: %w", err)
	}
	if len(data) > maxImageDownload {
		return nil, "", fmt.Errorf("service: image too large (%d bytes)", len(data))
	}

	contentType := resp.Header.Get("Content-Type")
	// Normalisasi; deteksi dari header. mime.ParseMediaType membersihkan charset.
	if mt, _, err := mime.ParseMediaType(contentType); err == nil {
		contentType = mt
	}
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	return data, contentType, nil
}

// imageFilename menurunkan nama file dari URL + content type.
func imageFilename(imageURL, contentType string) string {
	u, err := url.Parse(imageURL)
	base := "image"
	if err == nil && u.Path != "" {
		base = path.Base(u.Path)
	}
	base = strings.TrimSpace(base)
	if base == "" || base == "." || base == "/" {
		base = "image"
	}
	// Pastikan ada ekstensi sesuai content type.
	if !strings.Contains(base, ".") {
		if ext, err := extForContentType(contentType); err == nil {
			base += ext
		}
	}
	return base
}

// extForContentType mengembalikan ekstensi untuk content type gambar.
func extForContentType(contentType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg":
		return ".jpg", nil
	case "image/png":
		return ".png", nil
	case "image/webp":
		return ".webp", nil
	case "image/gif":
		return ".gif", nil
	case "image/avif":
		return ".avif", nil
	default:
		return "", fmt.Errorf("unsupported image content type %q", contentType)
	}
}

// watermarkText mengembalikan teks kredit untuk watermark.
func watermarkText(credit string) string {
	if credit == "" {
		return ""
	}
	return "Credit: " + credit + " · " + time.Now().Format("2006-01-02")
}
