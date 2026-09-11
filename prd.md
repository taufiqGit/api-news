# PRD — Automated News Scheduler (AI-Generated International News)

> Status: Draft v1 — untuk direview sebelum dibangun dengan vibecode.
> Scope: fitur baru pada service `news-api` (Go + Gin + PostgreSQL + sqlc + S3).

---

## 1. Ringkasan

Membangun **backend scheduler** yang secara otomatis:

1. Menjalankan job setiap **10 menit** (interval bisa dikonfigurasi).
2. Mengambil **berita terbaru / viral internasional** dari sumber eksternal (Google Trends dan/atau news API/RSS).
3. Untuk **setiap `website` aktif** di database, membuat **5 berita** dengan kategori bebas.
4. **Menulis ulang (rewrite)** setiap artikel dengan AI agar unik & natural — **bukan AI slop**.
5. Jika ada **gambar sumber**, mengunduhnya, **memberi credit/atribusi**, lalu **menyimpan ke S3**.
6. Menyimpan artikel hasil ke database, **semua artikel berbahasa Inggris**.

Hasil akhir: artikel otomatis terpublikasi (atau masuk draft/review) di bawah tiap website, siap dipakai frontend/publik.

---

## 2. Latar Belakang & Masalah

- Saat ini `articles` dibuat manual lewat API (`POST /articles`), tidak ada ingesti otomatis.
- Tidak ada integrasi sumber berita eksternal, tidak ada AI, tidak ada job scheduler.
- Ingin mengisi website dengan konten segar secara konsisten tanpa kerja manual.
- Konten hasil AI harus tetap **berkualitas**, tidak terdeteksi "AI slop" (generik, berbunga-bunga, tidak informatif).

---

## 3. Goals & Non-Goals

### Goals
- [ ] Scheduler berjalan otomatis dengan interval terkonfigurasi (default 10 menit).
- [ ] Pipeline lengkap: **fetch → dedup → AI rewrite → unduh gambar → upload S3 → simpan DB**.
- [ ] Mendukung banyak sumber berita via interface pluggable (mulai dari 1–2 sumber).
- [ ] Integrasi AI dengan provider yang bisa dikonfigurasi (OpenAI-compatible/Anthropic/Gemini).
- [ ] Atribusi gambar (credit) tersimpan & tampil di data artikel.
- [ ] Dedup: artikel yang sama tidak dibuat dua kali.
- [ ] Semua output artikel berbahasa Inggris.
- [ ] Dapat dimonitor (log run, jumlah artikel dibuat/dilewati, error).

### Non-Goals (di luar scope v1)
- Frontend / CMS untuk mengelola konten otomatis.
- Scraping HTML penuh secara langsung (raw crawl) — gunakan API/RSS yang resmi/legal dulu.
- Sistem penjadwalan terdistribusi (multi-instance lock) — v1 cukup single instance.
- Moderasi/editorial workflow yang kompleks (cukup draft vs published + flag).

---

## 4. Persona & Alur

**Aktor sistem:**
- **Scheduler worker** — menjadwalkan & menjalankan pipeline.
- **News Source** — penyedia berita eksternal (GNews, RSS, Google Trends).
- **AI Rewriter** — menulis ulang konten.
- **Storage (S3)** — penyimpanan gambar.

**Alur utama (happy path):**

```
Ticker 10m
  → ambil semua website is_active = true
  → untuk tiap website (berjalan paralel/berurutan):
      → fetch kandidat berita viral/terbaru (5 item)
      → dedup: skip jika source_url sudah ada di DB
      → AI rewrite → hasil (title, excerpt, content, slug, kategori, tags)
      → jika ada gambar: download → beri credit → upload S3 → dapat URL
      → simpan artikel (created_by = system user, website_id = website)
  → catat hasil ke scheduler_runs
```

---

## 5. Functional Requirements (FR)

### FR-1: Scheduler & Interval
- FR-1.1 Job terjadwal berjalan setiap interval tertentu, default **10 menit**, bisa diubah via config (`SCHEDULER_INTERVAL_MINUTES`).
- FR-1.2 Ada flag enable/disable (`SCHEDULER_ENABLED`).
- FR-1.3 Job tidak saling tumpang tindih (prevent overlapping run) — pakai lock sederhana (mis. `sync.Mutex` / atomic guard).
- FR-1.4 Ada mekanisme graceful shutdown (tidak memotong artikel di tengah proses).

### FR-2: Pemilihan Website & Kandidat Berita
- FR-2.1 Iterasi **hanya** website `is_active = true`.
- FR-2.2 Per website, ambil maksimal **5 kandidat berita** (bisa dikonfigurasi: `NEWS_ITEMS_PER_WEBSITE`).
- FR-2.3 Kategori bebas — tidak wajib menentukan kategori; bisa `category_id = NULL` atau auto-deteksi (opsional).
- FR-2.4 Sumber berita bersifat pluggable (interface). Sumber v1:
  - **Google News RSS** (agregasi headline internasional, region US, bahasa Inggris).
  - **RSS feed situs langsung** (BBC, The Guardian, Reuters, dst.).
  - (Opsional, fase lanjut) **Google Trends** — hanya discovery topik, bukan isi artikel.
- FR-2.5 Kandidat harus berupa berita **internasional / global** yang sedang ramai (viral) atau berita terbaru.

### FR-3: Dedup
- FR-3.1 Sebelum memproses, cek apakah `source_url` (kanonik) sudah ada di `articles`.
- FR-3.2 Jika sudah ada → **skip** (catat `skipped`).
- FR-3.3 Dedup bersifat **global** per URL asli (bukan hanya per website), untuk mencegah duplikat lintas website.
- FR-3.4 Untuk sumber tanpa URL (mis. topik Google Trends), gunakan hash judul+topik sebagai kunci dedup.

### FR-4: AI Rewrite (anti "AI slop")
- FR-4.1 Artikel asli di-rewrite menjadi konten baru & orisinal.
- FR-4.2 Output wajib **bahasa Inggris** (selalu), terlepas dari bahasa sumber.
- FR-4.3 Output harus terstruktur & terpakai: `title`, `slug`, `excerpt`, `content`, `category` (opsional), `tags` (opsional), dan rekomendasi `image_caption/credit` bila relevan.
- FR-4.4 Ada **pedoman editorial** dalam prompt untuk menghindari AI slop:
  - Hindari kalimat berbunga-bunga, filler, klaim tanpa dasar.
  - Hindari cliché ("in today's fast-paced world...", "delve", "unleash", dll).
  - Fokus pada 5W+1H, fakta, angka konkret.
  - Nada netral, jurnalistik, ringkas.
  - Panjang eksplisit (mis. 300–600 kata) dan struktur jelas.
- FR-4.5 Kontrol **temperature/RNG** agar konsisten (default rendah–sedang, ≤ 0.7).
- FR-4.6 (Opsional, fase lanjut) mekanisme "critique & revise" atau validator yang memblokir output yang terdeteksi slop sebelum disimpan.
- FR-4.7 Output diparse dari format terstruktur (JSON), bukan teks bebas.

### FR-5: Integrasi AI (provider)
- FR-5.1 Abstraksi interface `Rewriter` sehingga provider bisa diganti.
- FR-5.2 Dukungan awal: endpoint **OpenAI-compatible** (bisa dipakai OpenAI, Groq, OpenRouter, Together, Ollama, dsb.) + opsional Anthropic/Gemini.
- FR-5.3 Konfigurasi via env: `AI_PROVIDER`, `AI_API_KEY`, `AI_BASE_URL`, `AI_MODEL`, `AI_MAX_TOKENS`, `AI_TEMPERATURE`.
- FR-5.4 API key **tidak boleh di-hardcode** — selalu dari env/secret.

### FR-6: Gambar & Credit
- FR-6.1 Jika sumber menyediakan URL gambar → **unduh** gambar (HTTP, batasi ukuran/waktu).
- FR-6.2 Validasi tipe & ukuran gambar (sesuai batas existing: jpeg/png/webp/gif/avif, max ukuran).
- FR-6.3 Beri **credit** sebelum disimpan: simpan metadata atribusi (nama sumber/penulis, URL asli). Credit disimpan sebagai field terpisah di DB (bukan dimasukkan ke dalam konten).
- FR-6.4 Upload gambar ke **S3** (pakai key/folder yang jelas, mis. `covers/YYYY/MM/<uuid>.<ext>`).
- FR-6.5 Simpan URL S3 ke `articles.cover_image`, serta `image_credit` & `image_source_url`.
- FR-6.6 Jika download/upload gambar gagal → **tidak menggagalkan** seluruh artikel; artikel tetap disimpan tanpa gambar (graceful degradation), dicatat sebagai warning.

### FR-7: Penyimpanan Artikel
- FR-7.1 Simpan artikel dengan `created_by = system user` (lihat §7).
- FR-7.2 `website_id` diisi sesuai website yang sedang diproses.
- FR-7.3 `status` default bisa diatur (`published` atau `draft`) via config (`SCHEDULER_DEFAULT_STATUS`).
- FR-7.4 `published_at` di-set saat status published.
- FR-7.5 Slug harus unik (menggunakan pola existing; tambahkan suffix jika bentrok).
- FR-7.6 Flag `is_ai_generated = true` untuk menandai konten otomatis.

### FR-8: Observability & Audit
- FR-8.1 Setiap run mencatat ringkasan: waktu, website, jumlah dibuat/dilewati/gagal, error.
- FR-8.2 Log per artikel (info: dibuat; warn: gambar gagal; error: rewrite/API gagal).
- FR-8.3 (Opsional) endpoint admin untuk trigger run manual & lihat riwayat run.

---

## 6. Non-Functional Requirements (NFR)

- **NFR-1 Keandalan (reliability)** — kegagalan 1 website/1 artikel tidak boleh menghentikan seluruh job; harus diisolasi (retry terbatas untuk API AI/download).
- **NFR-2 Timeout** — fetch berita, panggilan AI, dan download gambar punya timeout (mis. 15–30 detik).
- **NFR-3 Rate limit & biaya** — jumlah panggilan AI/download terkendali; interval & jumlah item configurable. Perhitungkan 5 artikel/website/10 menit bisa mahal jika banyak website.
- **NFR-4 Keamanan** — API key (AI, news source) hanya dari secret/env, tidak di-log, tidak di-commit.
- **NFR-5 Observability** — structured logging (log level, context).
- **NFR-6 Performa** — proses website bisa berjalan paralel dengan batas concurrency (worker pool).
- **NFR-7 Konvensi kode** — mengikuti layered architecture existing (entity → repository → usecase → service), tambah layer baru `source`, `ai`, `scheduler`/`worker`.

---

## 7. Perubahan Data Model (Gambaran)

### 7.1 Tabel `articles` — kolom baru
| Kolom | Tipe | Keterangan |
|---|---|---|
| `source_url` | TEXT NULL | URL berita asli (kunci dedup) |
| `source_type` | VARCHAR(50) NULL | `gnews` / `rss` / `google_trends` / `manual` |
| `source_name` | VARCHAR(255) NULL | Nama publikasi asal (mis. "BBC News") |
| `source_hash` | VARCHAR(64) NULL | Hash dedup fallback (tanpa URL) |
| `is_ai_generated` | BOOLEAN NOT NULL DEFAULT FALSE | Penanda konten AI |
| `image_credit` | TEXT NULL | Atribusi gambar (nama penulis/sumber) |
| `image_source_url` | TEXT NULL | URL asli gambar |
| `image_license` | VARCHAR(255) NULL | Info lisensi (opsional) |

- Unique index **partial** pada `source_url` (`WHERE source_url IS NOT NULL`) untuk dedup.
- Tambah index pada `website_id` + `created_at` (kalau belum optimal).

### 7.2 System user (`created_by`)
- `articles.created_by` adalah `NOT NULL REFERENCES users(id)`.
- Tambah **migration seed** user sistem dengan UUID tetap, role khusus `system`, `is_active` berlaku hanya untuk proses internal (bukan login pengguna).
- Role `system` **tidak** diberi akses HTTP (tidak termasuk di middleware existing), dipakai murni sebagai kepemilikan artikel otomatis.

### 7.3 Tabel baru: `scheduler_runs` (audit)
| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | UUID PK | |
| `website_id` | UUID NULL | FK ke websites (NULL = run global) |
| `status` | VARCHAR(20) | `success` / `partial` / `failed` |
| `started_at` | TIMESTAMPTZ | |
| `finished_at` | TIMESTAMPTZ NULL | |
| `articles_created` | INT DEFAULT 0 | |
| `articles_skipped` | INT DEFAULT 0 | |
| `error` | TEXT NULL | Pesan error ringkas |
| `created_at` | TIMESTAMPTZ DEFAULT NOW() | |

### 7.4 Konfigurasi sumber per website (keputusan desain → lihat §10)
- Opsi A: tambah kolom `rss_url` + `source_type` di `websites`.
- Opsi B: tabel `news_sources` terpisah (satu website bisa banyak feed).
- **Rekomendasi awal:** Opsi A (minimal) dulu, karena cukup untuk v1.

---

## 8. Kontrak Integrasi (Interface Baru)

### 8.1 News Source
```go
type SourceItem struct {
    Title       string
    URL         string   // kanonik, untuk dedup
    Except      string   // ringkasan asli
    Content     string   // isi asli (kalau tersedia)
    ImageURL    string
    ImageCredit string
    PublishedAt time.Time
    Category    string   // opsional
}

type NewsSource interface {
    Name() string
    Fetch(ctx context.Context, opts SourceOptions) ([]SourceItem, error)
}
```

### 8.2 AI Rewriter
```go
type RewriteRequest struct {
    Title       string
    Content     string
    SourceName  string
    SourceURL   string
    Language    string // "en"
    WebsiteName string // untuk konteks
}

type RewriteResult struct {
    Title    string
    Slug     string
    Excerpt  string
    Content  string
    Category *string
    Tags     []string
}

type Rewriter interface {
    Rewrite(ctx context.Context, req RewriteRequest) (*RewriteResult, error)
}
```

### 8.3 Storage (extension)
- Tambah method untuk upload dari `io.Reader` + metadata (bukan `multipart.File`), karena gambar hasil download:
```go
UploadReader(ctx context.Context, r io.Reader, size int64, filename, contentType, folder string) (*FileInfo, error)
```
- Implementasi di `s3Storage` dan `localStorage`.

---

## 9. Acceptance Criteria (v1)

1. Menjalankan `make run` (atau command scheduler) → job berjalan tiap 10 menit tanpa tumpang tindih.
2. Website aktif mendapat maksimal 5 artikel per run per website.
3. Artikel yang sama (URL/kanonik sama) tidak dibuat dua kali.
4. Semua artikel tersimpan berbahasa Inggris dan sudah di-rewrite (bukan copy-paste mentah).
5. Artikel AI diberi penanda `is_ai_generated = true` dan `created_by` system user.
6. Gambar sumber diunduh, diberi credit (tersimpan di `image_credit`/`image_source_url`), dan file ada di S3.
7. Kegagalan gambar/1 website tidak menghentikan job.
8. Run tercatat di `scheduler_runs` dengan status & jumlah.
9. Seluruh secret (AI key, news API key) hanya dari env.
10. Tidak ada penurunan pada endpoint existing (build & test tetap hijau).

---

## 10. Decision Log (Final)

| # | Topik | Keputusan |
|---|---|---|
| 1 | Makna `websites` | **Target publikasi** (multi-tenant) |
| 2 | Sumber berita | **Google News RSS** + **RSS feed situs langsung** (BBC/Guardian/Reuters, dst.) |
| 3 | Provider AI | **OpenAI-compatible** (model + base URL configurable via env) |
| 4 | Status artikel default | Langsung **`published`** |
| 5 | Credit gambar | **Watermark di gambar** (selain metadata teks di DB) |
| 6 | Runtime scheduler | **In-process goroutine** (jalan di dalam server, flag `SCHEDULER_ENABLED`) |
| 7 | Bahasa | Sumber **English** → rewrite **English** |
| 8 | Kategori | **Bebas**; seed kategori sudah ada (migration `000008`) |

**Catatan sisa (detail konfigurasi — bukan blocker):**
- Daftar RSS feed konkret (URL BBC/Guardian/Reuters) → diisi saat **Phase 3**.
- `AI_BASE_URL` + `AI_MODEL` yang dipakai → diisi di `.env` saat **Phase 4**.
- Tampilan watermark (posisi/teks/opasitas) → ditentukan saat **Phase 5**.

---

*Dokumen ini akan dipecah roadmap implementasinya di `plan.md`.*