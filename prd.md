# Product Requirements Document (PRD)

## 1. Ringkasan Produk

**Nama produk:** News API

News API adalah backend REST API untuk platform berita multi-website. Sistem menyediakan autentikasi pengguna, manajemen artikel, kategori, tag, komentar, website/brand, upload gambar, serta dokumentasi API untuk digunakan oleh frontend berita seperti Maknews, Bolanet, Gadgetin, dan Kuliner Nusantara.

**Status saat ini:** MVP backend berjalan dan telah digunakan untuk beberapa website berita berbasis tema.

## 2. Tujuan Produk

1. Menyediakan API berita yang stabil, aman, dan mudah digunakan oleh banyak frontend.
2. Mendukung banyak website/brand dengan satu backend dan satu basis data.
3. Memudahkan editor dan author membuat, mengedit, menerbitkan, serta mengarsipkan artikel.
4. Menyediakan konten terstruktur melalui slug, kategori, tag, status publikasi, dan metadata.
5. Menjadi fondasi yang dapat dikembangkan ke pencarian, analitik, moderasi, dan distribusi konten.

## 3. Sasaran Pengguna

| Pengguna | Kebutuhan |
|---|---|
| Admin | Mengelola user, website, kategori, tag, dan seluruh konten |
| Editor | Mengelola dan menerbitkan artikel, kategori, tag, komentar |
| Author | Membuat dan memperbarui artikel serta mengunggah gambar |
| Pembaca | Membaca artikel, kategori, dan memberikan komentar |
| Frontend Developer | Mengambil data berita melalui API publik dan protected |

## 4. Ruang Lingkup MVP

### 4.1 Autentikasi dan Otorisasi

- Register dan login menggunakan email/password.
- JWT dengan masa berlaku yang dapat dikonfigurasi.
- Role: `admin`, `editor`, `author`.
- Middleware autentikasi dan role-based access control.
- Endpoint profil pengguna aktif.
- Admin dapat mengelola pengguna.

### 4.2 Multi-Website

- Setiap artikel terkait dengan satu website melalui `website_id`.
- Website memiliki nama, slug, deskripsi, logo, dan status aktif.
- Frontend dapat memfilter artikel berdasarkan website.
- Satu backend dapat melayani beberapa brand berita dengan tema berbeda.

### 4.3 Manajemen Artikel

- CRUD artikel.
- Field minimum: judul, slug, excerpt, content Markdown, cover image, status, category, website, author, timestamps.
- Status: `draft`, `published`, `archived`.
- Artikel published tersedia melalui endpoint publik.
- Detail artikel berdasarkan ID atau slug.
- View count bertambah saat artikel dibaca.
- Pagination dan filter status, kategori, website, serta author.
- Slug unik dan URL-friendly.

### 4.4 Kategori dan Tag

- CRUD kategori.
- Kategori dapat memiliki `parent_id` untuk hierarki.
- CRUD tag.
- Relasi many-to-many antara artikel dan tag.
- Filter artikel berdasarkan kategori dan tag.

### 4.5 Komentar

- Pembaca terautentikasi dapat membuat komentar.
- Komentar mendukung reply melalui `parent_id`.
- Admin/editor dapat melakukan approval dan penghapusan.
- Komentar dapat diambil berdasarkan artikel.

### 4.6 Upload Media

- Upload cover atau gambar artikel.
- Dukungan object storage S3-compatible.
- Local storage sebagai fallback saat credential S3 kosong.
- Validasi tipe file dan ukuran.
- URL file dikembalikan oleh API.

### 4.7 Dokumentasi dan Operasional

- Swagger UI tersedia untuk endpoint API.
- Health check tersedia.
- CORS dapat dikonfigurasi.
- Graceful shutdown.
- Konfigurasi melalui environment variable.
- Migrasi database menggunakan golang-migrate.
- Docker Compose untuk aplikasi dan PostgreSQL.

## 5. Persyaratan Fungsional

| ID | Persyaratan | Prioritas | Kriteria Penerimaan |
|---|---|---|---|
| FR-001 | User dapat login | P0 | Login valid menghasilkan JWT; kredensial salah ditolak |
| FR-002 | Endpoint protected memvalidasi JWT | P0 | Request tanpa/invalid token mendapat 401 |
| FR-003 | Role membatasi aksi | P0 | User di bawah role minimum mendapat 403 |
| FR-004 | Author dapat membuat draft artikel | P0 | Artikel tersimpan dengan website dan slug valid |
| FR-005 | Editor dapat menerbitkan artikel | P0 | Status berubah menjadi published dan muncul di endpoint publik |
| FR-006 | Pembaca dapat mengambil artikel published | P0 | Endpoint hanya mengembalikan konten published |
| FR-007 | Artikel terisolasi berdasarkan website | P0 | Filter website hanya mengembalikan artikel website tersebut |
| FR-008 | Artikel mendukung kategori dan tag | P1 | Relasi tampil pada response dan dapat difilter |
| FR-009 | Pembaca dapat berkomentar | P1 | Komentar tersimpan dan reply dapat dibuat |
| FR-010 | User dapat mengunggah gambar | P1 | File valid menghasilkan URL; file invalid ditolak |
| FR-011 | API memiliki dokumentasi Swagger | P1 | Endpoint dan schema utama tampil di Swagger UI |
| FR-012 | API menyediakan pagination konsisten | P1 | Response list menyertakan item dan metadata pagination |

## 6. Persyaratan Non-Fungsional

- **Keamanan:** password di-hash; JWT secret tidak boleh di-commit; input tervalidasi; upload dibatasi.
- **Performa:** endpoint list menggunakan pagination; query menggunakan index yang relevan.
- **Ketersediaan:** health check dan graceful shutdown wajib tersedia.
- **Maintainability:** layered architecture dengan kontrak repository dan usecase.
- **Observability:** error response konsisten dan logging server tersedia.
- **Kompatibilitas:** Go 1.26, PostgreSQL, Docker, S3-compatible storage.
- **API consistency:** response JSON menggunakan format sukses/error yang seragam.

## 7. API Utama

### Public

- `GET /healthz`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/articles/published`
- `GET /api/v1/articles/slug/:slug`
- `GET /api/v1/articles`
- `GET /api/v1/articles/:id`
- `GET /api/v1/articles/:id/comments`
- `GET /api/v1/categories`
- `GET /api/v1/categories/:id`
- `GET /api/v1/categories/slug/:slug`
- `GET /api/v1/tags`

### Protected

- User management: `/api/v1/users`
- Article management: `/api/v1/articles`
- Category management: `/api/v1/categories`
- Tag management: `/api/v1/tags`
- Comments: `/api/v1/comments`
- Upload: `/api/v1/uploads/images`
- Website management: `/api/v1/websites`

## 8. Model Data Utama

- `users`
- `websites`
- `articles`
- `categories`
- `tags`
- `article_tags`
- `comments`

Relasi utama: user membuat banyak artikel; website memiliki banyak artikel; kategori memiliki banyak artikel dan dapat bersifat hierarkis; artikel memiliki banyak tag dan komentar; komentar dapat memiliki parent/reply.

## 9. Definition of Done MVP

- Semua endpoint utama tersedia dan terhubung ke PostgreSQL.
- Migration dapat dijalankan dari database kosong.
- Login, RBAC, CRUD artikel, website, kategori, tag, komentar, dan upload teruji.
- Artikel dapat difilter berdasarkan website.
- Swagger dapat dibuka dan mencoba endpoint dengan JWT.
- `go test ./...`, `go vet ./...`, dan build berhasil.
- `.env`, credential, token, dan file upload tidak masuk repository.
- Docker Compose dapat menjalankan API dan PostgreSQL.

## 10. Di Luar Scope MVP

- CMS web admin penuh.
- Aplikasi mobile native.
- Full-text search berbasis Elasticsearch/OpenSearch.
- Notifikasi email/push.
- Sistem rekomendasi berbasis machine learning.
- Monetisasi, iklan, subscription, dan paywall.
- Workflow editorial multi-level yang kompleks.

## 11. Roadmap Berikutnya

1. Automated test suite dan integration test.
2. Search artikel dan sorting lanjutan.
3. Dashboard analytics: views, artikel populer, performa per website.
4. Scheduled publishing.
5. Moderasi komentar yang lebih lengkap.
6. Rate limiting, request ID, audit log, dan monitoring.
7. CI/CD GitHub Actions dan deployment staging/production.
8. Cache Redis untuk endpoint publik.
9. OpenAPI contract testing untuk frontend.

## 12. Risiko dan Mitigasi

| Risiko | Mitigasi |
|---|---|
| Credential production bocor | `.env` di-ignore, secret manager, rotasi credential |
| Spam komentar/upload | Rate limiting, CAPTCHA/moderasi, batas ukuran file |
| Query lambat saat konten bertambah | Index, pagination wajib, profiling query |
| Konten salah website | `website_id` wajib dan validasi ownership/filter |
| Perubahan API mematahkan frontend | Versioning `/api/v1`, contract test, changelog |
| Kegagalan object storage | Local fallback, retry terbatas, monitoring upload |

## 13. Metrik Keberhasilan

- 99.5% uptime endpoint publik.
- Error rate API di bawah 1% pada kondisi normal.
- P95 response endpoint list di bawah 500 ms untuk dataset MVP.
- 100% artikel memiliki website, slug, dan status valid.
- Tidak ada secret production di Git repository.
- Semua endpoint P0 memiliki test otomatis.

---

**Dokumen:** PRD News API  
**Versi:** 1.0  
**Tanggal:** 2026-09-11
