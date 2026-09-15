package storage

import "testing"

func TestResolveImageType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		filename    string
		wantCT      string
		wantExt     string
		wantErr     bool
	}{
		{
			name:        "jpeg from content type",
			contentType: "image/jpeg",
			filename:    "photo.jpg",
			wantCT:      "image/jpeg",
			wantExt:     ".jpg",
		},
		{
			name:        "png from content type",
			contentType: "image/png",
			filename:    "img.png",
			wantCT:      "image/png",
			wantExt:     ".png",
		},
		{
			name:        "webp empty content type detects from ext",
			contentType: "",
			filename:    "img.webp",
			wantCT:      "image/webp",
			wantExt:     ".webp",
		},
		{
			name:        "missing ext detected from content type",
			contentType: "image/gif",
			filename:    "anim",
			wantCT:      "image/gif",
			wantExt:     ".gif",
		},
		{
			name:        "avif from content type",
			contentType: "image/avif",
			filename:    "pic.avif",
			wantCT:      "image/avif",
			wantExt:     ".avif",
		},
		{
			name:        "yolo unsupported type and no ext",
			contentType: "video/mp4",
			filename:    "movie",
			wantErr:     true,
		},
		{
			name:        "content type with charset",
			contentType: "image/png; charset=utf-8",
			filename:    "x.png",
			wantCT:      "image/png",
			wantExt:     ".png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct, ext, err := resolveImageType(tt.contentType, tt.filename)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("resolveImageType(%q, %q) diharapkan error", tt.contentType, tt.filename)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveImageType(%q, %q) error: %v", tt.contentType, tt.filename, err)
			}
			if ct != tt.wantCT {
				t.Errorf("contentType = %q, want %q", ct, tt.wantCT)
			}
			if ext != tt.wantExt {
				t.Errorf("ext = %q, want %q", ext, tt.wantExt)
			}
		})
	}
}
