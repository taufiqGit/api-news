package source

// DefaultRSSFeeds mengembalikan daftar sumber berita default:
// Google News + beberapa feed RSS internasional berbahasa Inggris.
func DefaultRSSFeeds() []NewsSource {
	return []NewsSource{
		NewGoogleNews(),
		NewRSS("BBC News World", "https://feeds.bbci.co.uk/news/world/rss.xml"),
		NewRSS("The Guardian World", "https://www.theguardian.com/world/rss"),
		NewRSS("CNN Top Stories", "http://rss.cnn.com/rss/edition.rss"),
		NewRSS("NPR News", "https://feeds.npr.org/1001/rss.xml"),
		NewRSS("Al Jazeera", "https://www.aljazeera.com/xml/rss/all.xml"),
	}
}

// NewDefaultRegistry membangun registry berisi semua sumber default.
func NewDefaultRegistry() *Registry {
	r := NewRegistry()
	for _, s := range DefaultRSSFeeds() {
		r.Add(s)
	}
	return r
}
