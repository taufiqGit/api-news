package entity

import "time"

// SourceItem adalah kandidat berita mentah dari sumber eksternal
// (Google News RSS / RSS feed situs langsung) sebelum di-rewrite AI.
type SourceItem struct {
	Title       string    // judul asli
	URL         string    // URL kanonik — kunci dedup
	Excerpt     string    // ringkasan asli
	Content     string    // isi asli (jika tersedia dari sumber)
	ImageURL    string    // URL gambar (jika ada)
	ImageCredit string    // atribusi gambar dari sumber (jika ada)
	PublishedAt time.Time // waktu terbit di sumber (zero time jika tidak diketahui)
	Category    string    // kategori dari sumber (opsional)
}
