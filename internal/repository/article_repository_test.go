package repository

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

// TestClassifyCreateError memastikan pelanggaran constraint INSERT artikel
// dipetakan ke sentinel yang tepat, dan error lain diteruskan apa adanya.
func TestClassifyCreateError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error // nil = passthrough (instance sama dikembalikan)
	}{
		{
			name: "source url index violation",
			err:  &pgconn.PgError{Code: "23505", ConstraintName: "idx_articles_source_url"},
			want: ErrDuplicateArticle,
		},
		{
			name: "slug constraint violation",
			err:  &pgconn.PgError{Code: "23505", ConstraintName: "articles_slug_key"},
			want: ErrSlugTaken,
		},
		{
			name: "wrapped slug violation",
			err:  fmt.Errorf("insert: %w", &pgconn.PgError{Code: "23505", ConstraintName: "articles_slug_key"}),
			want: ErrSlugTaken,
		},
		{
			name: "other unique constraint passthrough",
			err:  &pgconn.PgError{Code: "23505", ConstraintName: "article_tags_pkey"},
		},
		{
			name: "non-unique pg error passthrough",
			err:  &pgconn.PgError{Code: "23503", ConstraintName: "idx_articles_source_url"},
		},
		{
			name: "non pg error passthrough",
			err:  errors.New("connection refused"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyCreateError(tt.err)
			if tt.want != nil {
				if !errors.Is(got, tt.want) {
					t.Fatalf("classifyCreateError(%v) = %v, want %v", tt.err, got, tt.want)
				}
				return
			}
			if got != tt.err {
				t.Fatalf("classifyCreateError(%v) harus passthrough, got %v", tt.err, got)
			}
		})
	}
}

// TestIsConstraintViolation memastikan pencocokan nama constraint presisi.
func TestIsConstraintViolation(t *testing.T) {
	slugErr := &pgconn.PgError{Code: "23505", ConstraintName: "articles_slug_key"}
	if !isConstraintViolation(slugErr, "articles_slug_key") {
		t.Fatal("harus mendeteksi articles_slug_key")
	}
	if isConstraintViolation(slugErr, "idx_articles_source_url") {
		t.Fatal("tidak boleh mendeteksi nama constraint yang berbeda")
	}
	if isConstraintViolation(errors.New("bukan pg error"), "articles_slug_key") {
		t.Fatal("error non-pg tidak boleh terdeteksi")
	}
}
