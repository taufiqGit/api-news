# News API Development Plan

> **Related PRD:** [`prd.md`](./prd.md)
>
> **Implementation mode:** kerjakan berurutan; setiap task menghasilkan perubahan kecil yang dapat diverifikasi dan di-commit.

## Goal

Membawa News API dari MVP yang sudah berjalan menjadi backend berita multi-website yang production-ready: aman, teruji, terdokumentasi, mudah dideploy, dan siap dipakai frontend berita.

## Current Context

- Backend menggunakan Go, Gin, PostgreSQL, pgx/v5, sqlc, JWT, Docker, dan Swagger.
- Arsitektur saat ini: `cmd/server`, `internal/{config,database,domain,repository,usecase,handler,middleware,router}`, dan `db/{migrations,queries}`.
- Fitur yang sudah tersedia mencakup auth/RBAC, artikel, kategori, tag, komentar, website, upload, pagination, CORS, Swagger, dan graceful shutdown.
- Repository: `https://github.com/taufiqGit/api-news.git`.
- File upload runtime harus tetap di luar Git; secret hanya melalui environment variable.

## Prinsip Implementasi

1. Pertahankan layered architecture dan interface repository.
2. Jangan mengubah kontrak API tanpa versioning atau backward compatibility.
3. Gunakan TDD untuk business rule dan handler penting.
4. Semua perubahan database dimulai dari migration, lalu query sqlc, repository, usecase, handler, route, dan docs.
5. Setiap task harus diverifikasi sebelum lanjut.
6. Hindari memasukkan `.env`, JWT, credential, dan `uploads/` ke repository.

---

## Phase 0 — Baseline dan Keamanan Repository

### Task 0.1: Validasi baseline lokal

**Files:** tidak ada perubahan.

**Actions:**

```bash
go test ./...
go vet ./...
go build -o /tmp/news-api ./cmd/server
```

**Acceptance:** seluruh command selesai tanpa error, atau error dicatat sebagai blocker sebelum fitur baru.

### Task 0.2: Audit secret dan file runtime

**Files:** `.gitignore`, `.env.example`, dokumentasi bila diperlukan.

**Actions:**

- Pastikan `.env`, credential, token, `uploads/`, binary build, dan log tidak dilacak Git.
- Cari hardcoded secret di source dan sample config.
- Pastikan `.env.example` hanya berisi placeholder aman.

**Verification:**

```bash
git status --short
git grep -nEi 'jwt.*secret|password|access.?key|token' -- ':!*.md' ':!docs/*'
```

### Task 0.3: Tambahkan CI dasar

**Create:** `.github/workflows/ci.yml`

**Actions:**

- Setup Go versi yang digunakan project.
- Jalankan `go test ./...`, `go vet ./...`, dan build.
- Tambahkan pemeriksaan format `gofmt -l`.

**Acceptance:** workflow berhasil pada push dan pull request.

---

## Phase 1 — Test Foundation

### Task 1.1: Inventarisasi endpoint dan response contract

**Files:** `README.md`, `docs/` atau `docs/api-contract.md`.

**Actions:**

- Catat status code sukses/error tiap endpoint.
- Catat bentuk pagination, validation error, auth error, dan not-found error.
- Cocokkan dokumentasi dengan route aktual.

### Task 1.2: Test response helper

**Create:** `internal/handler/response/response_test.go`

**Cases:** success response, error response, validation error, empty list, pagination metadata.

**Verification:**

```bash
go test ./internal/handler/response -v
```

### Task 1.3: Test JWT middleware dan role guard

**Create:** `internal/middleware/auth_test.go`

**Cases:** token valid, expired, malformed, missing, role cukup, role kurang, user inactive.

**Verification:**

```bash
go test ./internal/middleware -v
```

### Task 1.4: Tambahkan mock repository/usecase untuk unit test

**Files:** interface di `internal/domain/repository/`, mock di package test atau `internal/mocks/`.

**Actions:**

- Mock hanya method yang diperlukan test.
- Jangan mengubah generated sqlc files secara manual.
- Gunakan dependency injection yang sudah dipakai router.

---

## Phase 2 — Data Integrity dan Multi-Website

### Task 2.1: Audit schema website dan article

**Files:** `db/migrations/`, `db/queries/`, `internal/domain/entity/article.go`, `internal/domain/entity/website.go`.

**Actions:**

- Pastikan `website_id` wajib untuk artikel.
- Tambahkan foreign key/index bila belum ada.
- Pastikan slug website unik.
- Pastikan query artikel selalu dapat difilter website.

**Verification:** jalankan migration dari database kosong dan `make sqlc`.

### Task 2.2: Validasi website pada create/update artikel

**Files:** `internal/usecase/article_usecase.go`, repository terkait, test usecase.

**Rules:**

- Website harus ada dan aktif.
- Author/editor tidak boleh membuat artikel untuk website invalid.
- Update tidak boleh menghilangkan website yang wajib.

**Tests:** valid website, nonexistent website, inactive website, missing website ID.

### Task 2.3: Tambahkan filter website pada list dan published query

**Files:** `db/queries/articles.sql`, generated sqlc melalui `make sqlc`, `internal/domain/entity/article.go`, handler/usecase.

**Acceptance:**

```text
GET /api/v1/articles?website_id=<uuid>
GET /api/v1/articles/published?website_id=<uuid>
```

mengembalikan artikel dari website yang diminta saja.

### Task 2.4: Test isolasi data antar-website

**Create:** integration/usecase test yang membuat dua website dan dua artikel.

**Acceptance:** tidak ada artikel website A muncul pada query website B.

---

## Phase 3 — Article Workflow

### Task 3.1: Tegakkan state transition artikel

**Files:** `internal/usecase/article_usecase.go`, test usecase.

**Rules:**

- Draft dapat diedit author pemiliknya.
- Editor/admin dapat publish/archive.
- Artikel archived tidak muncul di endpoint publik.
- `published_at` diisi saat publish bila kosong.
- Status invalid ditolak.

### Task 3.2: Validasi slug dan content

**Files:** entity/request validation, article usecase, tests.

**Rules:**

- title, slug, content, website wajib sesuai kontrak.
- slug hanya karakter URL-safe.
- slug unik per kontrak yang dipilih; default aman: unik global.
- content Markdown tidak boleh kosong.

### Task 3.3: Pagination, sorting, dan filter artikel

**Files:** `db/queries/articles.sql`, handler, usecase, tests.

**Parameters:** `page`, `limit`, `status`, `category_id`, `website_id`, `author_id`, `search` bila search sudah disetujui.

**Rules:** limit default dan maksimum harus dibatasi; sorting default harus deterministik.

### Task 3.4: View count yang aman

**Files:** query increment, repository, handler/usecase, test.

**Actions:**

- Increment hanya pada pembacaan detail publik yang sesuai aturan.
- Hindari race condition dengan atomic SQL update.
- Pastikan response menampilkan count terbaru.

---

## Phase 4 — Categories, Tags, Comments, dan Moderation

### Task 4.1: Validasi hierarki kategori

**Files:** `internal/usecase/category_usecase.go`, tests.

**Rules:** parent harus ada, kategori tidak boleh menjadi parent dirinya sendiri, dan cycle harus ditolak.

### Task 4.2: Pastikan relasi tag idempotent

**Files:** `db/queries/article_tags.sql`, repository/usecase, tests.

**Rules:** tag duplikat tidak membuat relasi duplikat; update artikel harus menyinkronkan relasi secara atomik.

### Task 4.3: Workflow komentar dan approval

**Files:** comment entity, migration/query bila perlu, usecase, handler, tests.

**Rules:** komentar baru pending/non-public sesuai desain; admin/editor dapat approve/delete; reply memvalidasi parent artikel yang sama.

### Task 4.4: Tambahkan pagination komentar

**Files:** query dan handler komentar.

**Acceptance:** endpoint komentar memiliki limit maksimum dan tidak memuat seluruh thread tanpa batas.

---

## Phase 5 — Media Storage dan Validation

### Task 5.1: Validasi upload berdasarkan content type dan size

**Files:** `internal/handler/upload_handler.go`, storage package, tests.

**Rules:** allowlist jpeg/png/webp/gif/avif sesuai kebutuhan; maksimum file configurable; nama file tidak berasal langsung dari user.

### Task 5.2: Uji local storage dan S3-compatible storage

**Files:** `internal/storage/local.go`, `internal/storage/s3.go`, tests.

**Cases:** upload sukses, credential kosong fallback local, object storage error, duplicate filename, delete/cleanup bila tersedia.

### Task 5.3: Dokumentasikan operasi media

**Files:** `README.md`, Swagger annotations.

**Acceptance:** contoh curl tidak mengandung token nyata dan menjelaskan konfigurasi S3.

---

## Phase 6 — API Quality dan Observability

### Task 6.1: Standardisasi error response

**Files:** `internal/handler/response/response.go`, handlers, middleware.

**Actions:** gunakan error code stabil, message aman, detail validation terstruktur, dan jangan expose SQL/internal error ke client.

### Task 6.2: Request ID dan structured logging

**Create/Modify:** `internal/middleware/request_id.go`, logger/config files.

**Acceptance:** setiap request memiliki request ID; error log menyertakan method, path, status, duration, request ID.

### Task 6.3: Rate limiting dan body limit

**Files:** middleware/router/config.

**Prioritas endpoint:** login, register, comments, upload.

**Tests:** request di bawah limit diterima; request di atas limit mendapat status yang sesuai.

### Task 6.4: Health dan readiness endpoint

**Files:** `internal/handler/health_handler.go`, router, tests.

**Actions:** pisahkan liveness (`/healthz`) dan readiness database (`/readyz`) bila deployment memerlukannya.

---

## Phase 7 — Swagger, Documentation, dan Deployment

### Task 7.1: Regenerate Swagger

**Files:** handler annotations dan `docs/` generated output.

```bash
make docs
git diff -- docs/
```

**Acceptance:** route, auth Bearer, request body, response, query parameter, dan status code terdokumentasi.

### Task 7.2: Perbarui README operasional

**File:** `README.md`

**Isi:** quick start Docker, migration, sqlc, test, config production, S3, Swagger, troubleshooting, backup database.

### Task 7.3: Hardening Docker

**Files:** `Dockerfile`, `docker-compose.yml`, `.dockerignore`.

**Actions:** multi-stage build, non-root runtime user, healthcheck, no secret baked into image, persistent database volume.

### Task 7.4: CI/CD staging

**Files:** `.github/workflows/ci.yml`, workflow deployment terpisah bila diperlukan.

**Gates:** test, vet, build, migration validation, image build, deploy hanya dari branch/tag yang disetujui.

---

## Phase 8 — Performance dan Fitur Pasca-MVP

### Task 8.1: Index dan query profiling

**Actions:** ukur query list/detail/published; gunakan `EXPLAIN ANALYZE`; tambahkan index berdasarkan bukti, bukan perkiraan.

### Task 8.2: Full-text search

**Files:** migration/query/entity/handler/usecase/tests.

**Scope awal:** search judul dan excerpt/content PostgreSQL; pagination dan sanitasi input.

### Task 8.3: Scheduled publishing

**Files:** migration/entity/query/usecase/background worker/config/tests.

**Rules:** artikel dengan waktu publish masa depan tidak tampil sebelum waktunya; worker idempotent dan aman saat restart.

### Task 8.4: Analytics dan cache

**Actions:** definisikan event view, agregasi harian, endpoint artikel populer, kemudian evaluasi Redis setelah baseline performa tersedia.

---

## Verification Matrix

Jalankan pada setiap pull request:

```bash
gofmt -w $(find . -name '*.go' -not -path './internal/database/db/*')
go test ./...
go vet ./...
go build -o /tmp/news-api ./cmd/server
```

Jalankan saat ada perubahan query/migration:

```bash
make sqlc
make migrate-up
make docs
```

Manual smoke test:

1. Start PostgreSQL dan API.
2. `GET /healthz` mendapat 200.
3. Register/login mendapat JWT.
4. Buat artikel dengan `website_id` valid.
5. Publish artikel sesuai role.
6. Ambil artikel publik melalui website filter dan slug.
7. Pastikan artikel website lain tidak ikut.
8. Upload gambar dengan file valid dan invalid.
9. Buat komentar dan reply.
10. Buka Swagger dan jalankan endpoint protected memakai Bearer token.

## Release Checklist

- [ ] Migration tested from empty database.
- [ ] Backward-compatible API contract reviewed.
- [ ] No secret/token/upload tracked by Git.
- [ ] Unit and integration tests pass.
- [ ] Swagger regenerated.
- [ ] Docker image builds and healthcheck passes.
- [ ] Environment variables documented.
- [ ] Database backup/restore procedure tested.
- [ ] Rollback plan prepared.
- [ ] Production logs and error monitoring verified.

## Open Decisions

1. Apakah slug harus unik global atau unik per website?
2. Apakah komentar baru langsung tampil atau selalu memerlukan approval?
3. Apakah public article detail wajib menaikkan view count untuk bot/crawler?
4. Apakah search cukup menggunakan PostgreSQL atau perlu OpenSearch?
5. Apakah workflow editorial memerlukan scheduled publishing pada rilis terdekat?
6. Apakah API publik memerlukan API key selain endpoint yang benar-benar public?

**Target urutan rilis:** Phase 0–3 untuk baseline production, Phase 4–7 untuk hardening, Phase 8 untuk scale dan fitur tambahan.

**Dokumen:** News API Development Plan  
**Versi:** 1.0  
**Tanggal:** 2026-09-11
