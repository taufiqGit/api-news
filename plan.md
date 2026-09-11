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