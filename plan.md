<<<<<<< HEAD
# PLAN — Automated News Scheduler (Roadmap Implementasi)

> Pendamping `prd.md`. Dipecah menjadi **phase** yang bisa dikerjakan bertahap dengan vibecode.
> Setiap phase: tujuan, deliverables, file yang diubah/dibuat, dan acceptance checklist.

---

## Arsitektur Target (ringkas)

```
cmd/server/main.go          → wiring: jalankan scheduler (jika di-enable)
internal/
├── config/                 → tambah konfigurasi scheduler, AI, news source
├── domain/
│   ├── entity/             → + SourceItem, Rewrite*, SchedulerRun, perluasan Article, Website
│   └── repository/         → + interface dedup, scheduler run, article-by-source-url
├── repository/             → implementasi interface di atas
├── usecase/                → (tetap) logika bisnis artikel existing
├── service/ atau pipeline/ → orkestrasi fetch→dedup→rewrite→image→save (BARU)
├── source/                 → adapter sumber berita (interface + Google News RSS + RSS feed) (BARU)
├── ai/                     → adapter AI rewriter (interface + OpenAI-compatible/Anthropic) (BARU)
├── scheduler/              → ticker/worker pool + guard (BARU)
├── storage/                → + UploadReader (io.Reader)
└── router/                 → (opsional) endpoint admin trigger/riwayat
db/
├── migrations/             → migration baru (kolom article, seed system user, scheduler_runs)
├── queries/                → query sqlc baru
└── sqlc.yaml
```

---

## Phase 0 — Keputusan & Prasyarat (BLOCKING)

**Status: SELESAI** — keputusan sudah final (lihat `prd.md` §10).

**Deliverable (terkunci):**
- [x] Makna `websites` → **target publikasi** (multi-tenant).
- [x] Sumber berita → **Google News RSS** + **RSS feed situs langsung** (BBC/Guardian/Reuters, dst.).
- [x] Provider AI → **OpenAI-compatible** (model + base URL via env).
- [x] Status artikel default → **`published`**.
- [x] Credit gambar → **watermark di gambar**.
- [x] Runtime scheduler → **in-process goroutine** (flag `SCHEDULER_ENABLED`).
- [x] Kategori → **bebas** (seed kategori sudah dibuat di migration `000008`).

**Output tersisa (bukan blocker):** daftar RSS feed konkret, `AI_BASE_URL`/`AI_MODEL`, & tampilan watermark — diisi saat phase terkait.

---

## Phase 1 — Schema & Data Layer

**Tujuan:** siapkan database agar bisa menyimpan artikel hasil AI + metadata sumber + dedup + audit.

**Deliverable:**
1. Migration baru:
   - Tambah kolom `articles`: `source_url`, `source_type`, `source_name`, `source_hash`, `is_ai_generated`, `image_credit`, `image_source_url`, `image_license`.
   - Partial unique index `source_url` (`WHERE source_url IS NOT NULL`) + index `source_hash`.
   - ~~Seed system user~~ → **sudah dibuat** di migration `000008_seed_reference_data`.
   - Buat tabel `scheduler_runs`.
2. Query sqlc baru di `db/queries/`:
   - `GetArticleBySourceURL(source_url)`
   - `GetArticleBySourceHash(source_hash)`
   - `CountExistingBySourceURLs(urls[])` / bulk existence check (perf).
   - `CreateSchedulerRun`, `UpdateSchedulerRun`, `ListSchedulerRuns`.
   - (bila perlu) `ListActiveWebsites`.
3. Generate sqlc: `make sqlc`.
4. Perluas entity:
   - `entity.Article` + field source metadata.
   - `entity.SchedulerRun`.
   - `entity.SourceItem`, `entity.RewriteResult` (untuk contract layer baru).
5. Perluas `domain/repository` + implementasi `internal/repository`.

**Files (target):**
- `db/migrations/000009_*.up.sql` / `.down.sql` (skema; `000008` sudah dipakai untuk seed)
- `db/queries/articles.sql`, `db/queries/websites.sql`, `db/queries/scheduler_runs.sql` (baru)
- `internal/domain/entity/article.go`, `scheduler_run.go`, `source_item.go`, `rewrite.go`
- `internal/domain/repository/*.go`
- `internal/repository/*.go`

**Acceptance:**
- [ ] `make sqlc` sukses, tidak ada kolom mapper yang hilang.
- [ ] `make migrate-up` berhasil & system user ada.
- [ ] Test build hijau.

---

## Phase 2 — Storage: Upload dari io.Reader

**Tujuan:** mendukung upload gambar yang di-*download* (bukan di-upload lewat HTTP multipart).

**Deliverable:**
- [ ] Tambah method ke interface `storage.Storage`: `UploadReader(ctx, r io.Reader, size int64, filename, contentType, folder) (*FileInfo, error)`.
- [ ] Implementasi di `s3Storage` (minio `PutObject` dengan streaming) & `localStorage` (tulis ke disk).
- [ ] Reuse validasi tipe/ukuran existing.

**Files:** `internal/storage/s3.go`, `internal/storage/local.go`.

**Acceptance:**
- [ ] Upload from reader menghasilkan `FileInfo` & file benar di S3/local.
- [ ] Ukuran & tipe tervalidasi, error konsisten.

---

## Phase 3 — News Source Adapter

**Tujuan:** fetch kandidat berita dari sumber eksternal via interface pluggable.

**Deliverable:**
- [ ] Define interface `NewsSource` + `SourceItem` (lihat `prd.md` §8.1).
- [ ] Implementasi **Google News RSS** — parse feed agregasi internasional (region US, bahasa Inggris, max 5 item).
- [ ] Implementasi **RSS generik** (parse `rss_url`, ambil 5 item terbaru) — gunakan `github.com/mmcdole/gofeed` atau `encoding/xml`.
- [ ] Daftar feed **situs langsung** (BBC, The Guardian, Reuters, dst.) sebagai config.
- [ ] Registry sumber (`source.Registry`) untuk memilih sumber berdasarkan config/website.
- [ ] Timeout & error handling per source.

**Files:** `internal/source/source.go`, `internal/source/google_news.go`, `internal/source/rss.go`.

**Acceptance:**
- [ ] Bisa fetch ≤5 item dari Google News RSS & RSS feed langsung (tanpa API key wajib).
- [ ] Setiap item punya `URL` unik untuk dedup.
- [ ] Error source tidak panic; di-return.

---

## Phase 4 — AI Rewriter

**Tujuan:** integrasi AI untuk rewrite artikel (anti slop).

**Deliverable:**
- [ ] Define interface `Rewriter` + `RewriteRequest`/`RewriteResult` (lihat `prd.md` §8.2).
- [ ] Implementasi provider **OpenAI-compatible** (pakai HTTP/`github.com/sashabaranov/go-openai` atau client HTTP sendiri). Dukung `AI_BASE_URL`, `AI_MODEL`.
- [ ] (Opsional) provider Anthropic / Gemini.
- [ ] **Prompt editorial** untuk anti-slop + output **JSON** terstruktur (title, slug, excerpt, content, category, tags).
- [ ] Parse & validasi output JSON; ada fallback error jika format tidak valid (retry 1x).
- [ ] Batas `AI_MAX_TOKENS`, `AI_TEMPERATURE` (≤0.7).
- [ ] Paksa bahasa Inggris (`Language: en`).

**Files:** `internal/ai/rewriter.go`, `internal/ai/openai.go`, `internal/ai/prompt.go`, `internal/ai/anthropic.go` (opsional).

**Acceptance:**
- [ ] Input artikel → output JSON terstruktur & bahasa Inggris.
- [ ] Prompt mencantumkan aturan anti-slop, panjang, struktur.
- [ ] API key dari env, tidak di-log/hardcode.

---

## Phase 5 — Pipeline Orkestrasi

**Tujuan:** gabungkan semua langkah jadi satu alur per artikel.

**Deliverable:**
- [ ] Buat `internal/service` (atau `pipeline`) dengan fungsi utama:
  ```
  ProcessWebsite(ctx, website) → (created, skipped int, err error)
  ```
  Langkah:
  1. fetch kandidat (source)
  2. dedup (cek source_url/hash)
  3. rewrite (AI)
  4. (jika ada gambar) download → download gambar → credit → upload S3
  5. simpan artikel (existing `ArticleUsecase`/`ArticleRepository`)
- [ ] Helper **download & kredit gambar** (`internal/service/image.go`): HTTP GET, validasi tipe/ukuran, ambil attribution, panggil `UploadReader`.
- [ ] Graceful degradation: gambar gagal → artikel tetap disimpan tanpa gambar (warning).
- [ ] Slug generation dengan suffix dedup (reuse/ekstend logika existing).
- [ ] `created_by` = system user ID.

**Files:** `internal/service/pipeline.go`, `internal/service/image.go`, kemungkinan helper slug di usecase.

**Acceptance:**
- [ ] Artikel tersimpan lengkap dengan metadata sumber & credit gambar.
- [ ] Dedup berfungsi (artikel sama tidak dobel).
- [ ] Error gambar tidak menggagalkan artikel.

---

## Phase 6 — Scheduler & Wiring

**Tujuan:** jalankan pipeline secara periodik & aman.

**Deliverable:**
- [ ] `internal/scheduler/scheduler.go`:
  - ticker dengan `SCHEDULER_INTERVAL_MINUTES`.
  - guard anti-overlap (mutex/atomic).
  - worker pool concurrency (batas `SCHEDULER_MAX_CONCURRENCY`).
  - grace shutdown via context.
  - catat `scheduler_runs`.
- [ ] Config baru (`internal/config`): `SCHEDULER_*`, `AI_*`, `NEWS_*`.
- [ ] Wiring di `cmd/server/main.go`: jalankan scheduler hanya jika `SCHEDULER_ENABLED=true` (**in-process goroutine** — sesuai keputusan Phase 0).

**Files:** `internal/scheduler/scheduler.go`, `internal/config/config.go`, `cmd/server/main.go`.

**Acceptance:**
- [ ] Job jalan tiap 10 menit (default), tidak overlap.
- [ ] `SIGTERM/SIGINT` graceful.
- [ ] Banyak website diproses paralel dengan concurrency terbatas.

---

## Phase 7 — Observability, Retry & Admin (opsional)

**Tujuan:** operasional & monitoring.

**Deliverable:**
- [ ] Structured log per run & per artikel (level info/warn/error).
- [ ] Retry terbatas (1x) untuk panggilan AI / download gagal sementara.
- [ ] Endpoint admin (protected `admin`):
  - `POST /api/v1/admin/scheduler/run` — trigger manual.
  - `GET /api/v1/admin/scheduler/runs` — riwayat run.
- [ ] (wajib) reset kredensial: pastikan secret tidak ter-commit (sudah ada `.gitignore` untuk `.env`).

**Files:** `internal/handler/scheduler_handler.go`, `internal/router/router.go`, `db/queries/scheduler_runs.sql`.

**Acceptance:**
- [ ] Riwayat run terbaca; trigger manual berfungsi.
- [ ] Log jelas untuk debug.

---

## Phase 8 — Test, Docs & Packaging

**Tujuan:** kualitas & kemudahan deploy.

**Deliverable:**
- [ ] Unit test: slug dedup, dedup source_url, parse output AI, image validation.
- [ ] (Opsional) integration test pakai mock source & mock rewriter.
- [ ] Update `README.md`: cara setup API key, env baru, cara run scheduler.
- [ ] Update `.env.example` (var scheduler/AI/news source).
- [ ] Update `Makefile` (target dev tetap Air; tambah target run-scheduler bila worker terpisah).
- [ ] Update `Dockerfile`/`docker-compose.yml` bila butuh service/worker tambahan.

**Files:** `*_test.go` di tiap package, `README.md`, `.env.example`, `Makefile`, `docker-compose.yml`.

**Acceptance:**
- [ ] `go test ./...` hijau (termasuk test existing).
- [ ] `make build` & `make sqlc` sukses.
- [ ] Dokumentasi env baru lengkap.

---

## Urutan dependency

```
Phase 0 (keputusan)
  → Phase 1 (schema/data)
  → Phase 2 (storage reader)
  → Phase 3 (source) ─┐
  → Phase 4 (ai)      ─┼→ Phase 5 (pipeline)
                      ─┘
  → Phase 6 (scheduler wiring)
  → Phase 7 (observability/admin)
  → Phase 8 (test/docs/packaging)
```

---

## Risiko & Catatan

- **Biaya AI/API**: 5 artikel × N website × tiap 10 menit bisa mahal. Pastikan interval & jumlah configurable (Phase 0 & 6), pertimbangkan mulai dari interval lebih panjang.
- **Akurasi/kualitas AI**: rewrite bisa salah fakta/halusinasi → wajib prompt ketat + opsi status `draft` untuk review.
- **Google Trends ≠ isi berita**: hanya discovery topik; perlu sumber berita aktual untuk konten.
- **Legal/lisensi gambar**: hanya download gambar yang boleh di-redistribute; simpan credit & source URL.
- **Dedup lintas website**: unik pada `source_url` global; pastikan tidak ada bentrok slug antar website (slug unik global existing).

---

*Setelah Phase 0 disepakati, implementasi dimulai dari Phase 1 dan maju bertahap.*
=======
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
>>>>>>> 7dd5edb7343c7f6e7dcea1771f0b04f0c52141b9
