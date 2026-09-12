package entity

// RewriteRequest adalah input untuk AI rewriter.
type RewriteRequest struct {
	Title       string // judul asli
	Content     string // isi/ringkasan asli
	SourceName  string // nama publikasi sumber (mis. "BBC News")
	SourceURL   string // URL artikel asli
	Language    string // bahasa output, selalu "en"
	WebsiteName string // nama website target publikasi (untuk konteks)
}

// RewriteResult adalah hasil rewrite AI dalam format terstruktur.
type RewriteResult struct {
	Title    string
	Slug     string
	Excerpt  string
	Content  string
	Category *string  // nama kategori (opsional)
	Tags     []string // tag yang disarankan AI (opsional)
}
