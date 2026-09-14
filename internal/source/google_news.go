package source

import (
	"context"
	"net/http"
	"net/url"

	"github.com/taufiqgit/news-api/internal/domain/entity"
)

const googleNewsBase = "https://news.google.com/rss?hl=en-US&gl=US&ceid=US:en"

// googleNewsSource mengambil berita teratas dari Google News RSS.
type googleNewsSource struct {
	client *http.Client
}

// NewGoogleNews membuat sumber Google News.
func NewGoogleNews() NewsSource {
	return &googleNewsSource{
		client: &http.Client{Timeout: defaultHTTPTimeout},
	}
}

func (s *googleNewsSource) Name() string { return "Google News" }
func (s *googleNewsSource) Type() string { return TypeGoogleNews }

func (s *googleNewsSource) Fetch(ctx context.Context, opts SourceOptions) ([]entity.SourceItem, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	return fetchRSS(ctx, s.client, s.feedURL(opts.Topic), s.Name(), TypeGoogleNews, limit)
}

// feedURL membangun URL feed Google News, dengan topik bila disediakan.
func (s *googleNewsSource) feedURL(topic string) string {
	if topic == "" {
		return googleNewsBase
	}
	q := url.QueryEscape(topic)
	return "https://news.google.com/rss/search?q=" + q + "&hl=en-US&gl=US&ceid=US:en"
}
