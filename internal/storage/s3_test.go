package storage

import (
	"testing"

	"github.com/taufiqgit/news-api/internal/config"
)

func TestS3StorageURL(t *testing.T) {
	const key = "articles/2026/09/image.jpg"

	tests := []struct {
		name string
		cfg  config.S3Config
		want string
	}{
		{
			name: "public NOS base URL uses path-style bucket",
			cfg: config.S3Config{
				PublicURL: "https://nos.jkt-1.neo.id",
				Bucket:    "news",
				UseSSL:    true,
			},
			want: "https://nos.jkt-1.neo.id/news/articles/2026/09/image.jpg",
		},
		{
			name: "public URL trailing slash is normalized",
			cfg: config.S3Config{
				PublicURL: "https://nos.jkt-1.neo.id/",
				Bucket:    "news",
			},
			want: "https://nos.jkt-1.neo.id/news/articles/2026/09/image.jpg",
		},
		{
			name: "fallback endpoint uses HTTPS path-style bucket",
			cfg: config.S3Config{
				Endpoint: "nos.jkt-1.neo.id",
				Bucket:   "news",
				UseSSL:   true,
			},
			want: "https://nos.jkt-1.neo.id/news/articles/2026/09/image.jpg",
		},
		{
			name: "leading slash in key is normalized",
			cfg: config.S3Config{
				PublicURL: "https://nos.jkt-1.neo.id",
				Bucket:    "news",
			},
			want: "https://nos.jkt-1.neo.id/news/articles/2026/09/image.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := (&s3Storage{cfg: tt.cfg}).URL(key)
			if tt.name == "leading slash in key is normalized" {
				got = (&s3Storage{cfg: tt.cfg}).URL("/" + key)
			}
			if got != tt.want {
				t.Fatalf("URL() = %q, want %q", got, tt.want)
			}
		})
	}
}
