package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/taufiqgit/news-api/internal/domain/entity"
)

const (
	defaultHTTPTimeout = 20 * time.Second
	defaultLimit       = 5
	userAgent          = "news-api-scheduler/1.0 (+https://github.com/taufiqgit/news-api)"
)

// rssSource adalah sumber berita berbasis RSS/Atom feed apa pun.
type rssSource struct {
	name   string
	url    string
	client *http.Client
}

// NewRSS membuat sumber RSS dengan nama tampilan dan URL feed.
func NewRSS(name, url string) NewsSource {
	return &rssSource{
		name:   name,
		url:    url,
		client: &http.Client{Timeout: defaultHTTPTimeout},
	}
}

func (s *rssSource) Name() string { return s.name }
func (s *rssSource) Type() string { return TypeRSS }

func (s *rssSource) Fetch(ctx context.Context, opts SourceOptions) ([]entity.SourceItem, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	return fetchRSS(ctx, s.client, s.url, s.name, TypeRSS, limit)
}

// fetchRSS mengambil dan memetakan item dari satu feed URL.
func fetchRSS(ctx context.Context, client *http.Client, feedURL, sourceName, sourceType string, limit int) ([]entity.SourceItem, error) {
	feed, err := parseFeed(ctx, client, feedURL)
	if err != nil {
		return nil, err
	}

	items := make([]entity.SourceItem, 0, limit)
	for _, it := range feed.Items {
		if len(items) >= limit {
			break
		}
		items = append(items, toSourceItem(sourceName, sourceType, it))
	}
	return items, nil
}

// parseFeed mengunduh dan mem-parse feed RSS/Atom.
func parseFeed(ctx context.Context, client *http.Client, feedURL string) (*gofeed.Feed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("source: build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("source: fetch %s: %w", feedURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Baca sebagian body untuk diagnostik, lalu buang.
		_, _ = io.CopyN(io.Discard, resp.Body, 512)
		return nil, fmt.Errorf("source: %s returned status %d", feedURL, resp.StatusCode)
	}

	feed, err := gofeed.NewParser().Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("source: parse %s: %w", feedURL, err)
	}
	return feed, nil
}

// toSourceItem memetakan gofeed.Item menjadi entity.SourceItem.
func toSourceItem(sourceName, sourceType string, it *gofeed.Item) entity.SourceItem {
	item := entity.SourceItem{
		Title:      it.Title,
		URL:        it.Link,
		Excerpt:    it.Description,
		Content:    it.Content,
		SourceName: sourceName,
		SourceType: sourceType,
	}

	// Waktu terbit: prioritaskan PublishedParsed, fallback ke UpdatedParsed.
	if it.PublishedParsed != nil {
		item.PublishedAt = *it.PublishedParsed
	} else if it.UpdatedParsed != nil {
		item.PublishedAt = *it.UpdatedParsed
	}

	// Gambar: cek field Image lalu enclosure bertipe image.
	if it.Image != nil && it.Image.URL != "" {
		item.ImageURL = it.Image.URL
	} else {
		for _, enc := range it.Enclosures {
			if enc != nil && isImageType(enc.Type) && enc.URL != "" {
				item.ImageURL = enc.URL
				break
			}
		}
	}

	// Kategori pertama.
	if len(it.Categories) > 0 {
		item.Category = it.Categories[0]
	}

	return item
}

// isImageType mengecek apakah tipe MIME merupakan gambar.
func isImageType(mime string) bool {
	return len(mime) >= 6 && mime[:6] == "image/"
}
