package source

import (
	"context"
	"errors"

	"github.com/taufiqgit/news-api/internal/domain/entity"
)

// Tipe sumber berita yang didukung.
const (
	TypeGoogleNews = "google_news"
	TypeRSS        = "rss"
)

// ErrSourceNotFound dikembalikan ketika sumber dengan nama tertentu tidak ada.
var ErrSourceNotFound = errors.New("source not found")

// SourceOptions mengontrol perilaku Fetch.
type SourceOptions struct {
	Limit int    // jumlah item maksimum (0 = default)
	Topic string // topik pencarian (hanya Google News)
}

// NewsSource adalah kontrak untuk satu sumber berita eksternal.
type NewsSource interface {
	Name() string
	Type() string
	// Fetch mengambil item berita terbaru dari sumber.
	Fetch(ctx context.Context, opts SourceOptions) ([]entity.SourceItem, error)
}

// Registry menyimpan sumber berita berdasarkan nama.
type Registry struct {
	sources map[string]NewsSource
}

// NewRegistry membuat registry kosong.
func NewRegistry() *Registry {
	return &Registry{sources: make(map[string]NewsSource)}
}

// Add mendaftarkan satu sumber. Menimpa sumber dengan nama yang sama.
func (r *Registry) Add(s NewsSource) {
	r.sources[s.Name()] = s
}

// Get mengembalikan sumber berdasarkan nama.
func (r *Registry) Get(name string) (NewsSource, error) {
	s, ok := r.sources[name]
	if !ok {
		return nil, ErrSourceNotFound
	}
	return s, nil
}

// List mengembalikan daftar nama semua sumber terdaftar.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.sources))
	for name := range r.sources {
		names = append(names, name)
	}
	return names
}
