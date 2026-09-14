package storage

import (
	"path/filepath"
	"strings"
)

var extToContentType = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
	".gif":  "image/gif",
	".avif": "image/avif",
}

// resolveImageType menentukan contentType & ekstensi file yang valid.
// contentType boleh kosong — akan dideteksi dari ekstensi filename.
func resolveImageType(contentType, filename string) (string, string, error) {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	ext := strings.ToLower(filepath.Ext(filename))

	if !allowedImageTypes[contentType] {
		detected, ok := extToContentType[ext]
		if !ok {
			return "", "", ErrUnsupportedType
		}
		contentType = detected
	}

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

	return contentType, ext, nil
}
