package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/taufiqgit/news-api/internal/repository"
)

var slugCleaner = regexp.MustCompile(`[^a-z0-9]+`)

// cleanSlug menormalkan slug: lowercase, non-alnum jadi "-", trim "-".
func cleanSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugCleaner.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "article"
	}
	return s
}

// uniqueSlug menghasilkan slug yang unik di DB, dengan suffix numerik bila tabrakan.
func (p *Pipeline) uniqueSlug(ctx context.Context, preferred, title string) (string, error) {
	base := preferred
	if strings.TrimSpace(base) == "" {
		base = title
	}
	base = cleanSlug(base)

	slug := base
	for i := 2; ; i++ {
		_, err := p.articleRepo.GetBySlug(ctx, slug)
		if errors.Is(err, repository.ErrNotFound) {
			return slug, nil
		}
		if err != nil {
			return "", err
		}
		slug = base + "-" + itoa(i)
	}
}

// itoa mengubah int positif menjadi string desimal.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
