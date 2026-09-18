package service

import (
	"strings"
	"testing"
)

func TestSourceHash(t *testing.T) {
	// Hash deterministik & berbeda untuk input berbeda.
	a := sourceHash("title-a", "content-a")
	b := sourceHash("title-a", "content-a")
	c := sourceHash("title-b", "content-b")

	if a != b {
		t.Fatalf("sourceHash harus deterministik: %q != %q", a, b)
	}
	if a == "" {
		t.Fatal("sourceHash tidak boleh kosong")
	}
	if a == c {
		t.Fatalf("sourceHash input berbeda harus berbeda: %q == %q", a, c)
	}
	// Panjang SHA-256 hex = 64 karakter.
	if len(a) != 64 {
		t.Fatalf("sourceHash length = %d, want 64", len(a))
	}
}

func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		name string
		vals []string
		want string
	}{
		{"first non-empty", []string{"", "hello", "world"}, "hello"},
		{"all empty", []string{"", "  ", "\t"}, ""},
		{"empty slice", nil, ""},
		{"first wins", []string{"a", "b"}, "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := firstNonEmpty(tt.vals...)
			if got != tt.want {
				t.Fatalf("firstNonEmpty(%v) = %q, want %q", tt.vals, got, tt.want)
			}
		})
	}
}

func TestStrPtr(t *testing.T) {
	if got := strPtr(""); got != nil {
		t.Fatalf("strPtr(\"\") harus nil, got %q", *got)
	}
	if got := strPtr("hi"); got == nil || *got != "hi" {
		t.Fatalf("strPtr(\"hi\") salah: %v", got)
	}
}

func TestSlugWithRaceSuffix(t *testing.T) {
	base := "houthi-attacks-threaten-saudi-oil"
	a := slugWithRaceSuffix(base)
	b := slugWithRaceSuffix(base)

	// Prefix asli tetap utuh + suffix pendek (4 hex).
	if !strings.HasPrefix(a, base+"-") {
		t.Fatalf("suffix hilangkan prefix: %q", a)
	}
	if len(a) != len(base)+5 {
		t.Fatalf("panjang slug = %d, want %d", len(a), len(base)+5)
	}
	// Dua panggilan menghasilkan suffix berbeda (tahan retry paralel).
	if a == b {
		t.Fatalf("suffix harus acak: %q == %q", a, b)
	}
	// Slug panjang dipotong agar total <= 550 (kolom VARCHAR(550)).
	long := strings.Repeat("a", 600)
	got := slugWithRaceSuffix(long)
	if len(got) != 550 {
		t.Fatalf("panjang slug terpotong = %d, want 550", len(got))
	}
	if len(slugWithRaceSuffix("")) != 5 {
		t.Fatalf("slug kosong harus jadi suffix saja: %q", slugWithRaceSuffix(""))
	}
}
