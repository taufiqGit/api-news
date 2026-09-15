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

**Status: SELESAI** — migration `000009` di-apply, sqlc di-regenerate, build hijau.

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
- [x] `make sqlc` sukses, tidak ada kolom mapper yang hilang.
- [x] `make migrate-up` berhasil & system user ada.
- [x] Test build hijau.

**Catatan implementasi:**
- Migration skema memakai nomor `000009` (`000008` sudah dipakai seed).
- Entity baru: `SchedulerRun` (+ konstanta status), `SourceItem`, `RewriteRequest`/`RewriteResult`.
- Interface repository baru: `SchedulerRunRepository`; `ArticleRepository` + `GetBySourceURL`/`GetBySourceHash`/`GetExistingSourceURLs`; `WebsiteRepository` + `ListActive`.
- Perbaikan kecil: `toArticleEntity` kini juga memetakan `WebsiteID` (sebelumnya terlewat).

---

## Phase 2 — Storage: Upload dari io.Reader

**Status: SELESAI** — `UploadReader` ditambahkan & logic divalidasi.

**Tujuan:** mendukung upload gambar yang di-*download* (bukan di-upload lewat HTTP multipart).

**Deliverable:**
- [x] Tambah method ke interface `storage.Storage`: `UploadReader(ctx, r io.Reader, size int64, filename, contentType, folder) (*FileInfo, error)`.
- [x] Implementasi di `s3Storage` (minio `PutObject` dengan streaming) & `localStorage` (tulis ke disk).
- [x] Reuse validasi tipe/ukuran existing.

**Files:** `internal/storage/s3.go`, `internal/storage/local.go`, `internal/storage/helpers.go` (baru).

**Acceptance:**
- [x] Upload from reader menghasilkan `FileInfo` & file benar di S3/local.
- [x] Ukuran & tipe tervalidasi, error konsisten.

**Catatan implementasi:**
- Logic validasi (size, deteksi tipe dari header/ekstensi, penentuan ekstensi) diekstrak ke helper `resolveImageType` di `helpers.go`, dipakai bersama oleh `s3Storage` & `localStorage` (hilangkan duplikasi).

---

## Phase 3 — News Source Adapter

**Status: SELESAI** — interface, Google News RSS, RSS generik, dan registry selesai; build & vet green.

**Tujuan:** fetch kandidat berita dari sumber eksternal via interface pluggable.

**Deliverable:**
- [x] Define interface `NewsSource` + `SourceItem` (lihat `prd.md` §8.1).
- [x] Implementasi **Google News RSS** — parse feed agregasi internasional (region US, bahasa Inggris, max 5 item).
- [x] Implementasi **RSS generik** (parse `rss_url`, ambil 5 item terbaru) — gunakan `github.com/mmcdole/gofeed`.
- [x] Daftar feed **situs langsung** (BBC, The Guardian, Reuters, dst.) sebagai config.
- [x] Registry sumber (`source.Registry`) untuk memilih sumber berdasarkan config/website.
- [x] Timeout & error handling per source.

**Files:** `internal/source/source.go`, `internal/source/google_news.go`, `internal/source/rss.go`, `internal/source/defaults.go` (baru).

**Acceptance:**
- [x] Bisa fetch ≤5 item dari Google News RSS & RSS feed langsung (tanpa API key wajib).
- [x] Setiap item punya `URL` unik untuk dedup.
- [x] Error source tidak panic; di-return.

**Catatan implementasi:**
- Dependency `github.com/mmcdole/gofeed@v1.4.2` ditambahkan (mem-pull upgrade minor `golang.org/x/*`).
- `gofeed.Item` tidak punya field `Source`, jadi atribusi gambar default memakai nama sumber (`sourceName`).
- Waktu terbit memakai `PublishedParsed` dengan fallback `UpdatedParsed`; gambar memakai `Image.URL` dengan fallback `Enclosures` bertipe `image/*`.
- Timeout HTTP default 20s; User-Agent custom diset agar tidak di-block sebagian feed.

---

## Phase 4 — AI Rewriter

**Status: SELESAI** — interface `Rewriter`, provider [OI]-compatible, prompt anti-slop, parsing JSON + retry selesai; build & vet green.

**Tujuan:** integrasi AI untuk rewrite artikel (anti slop).

**Deliverable:**
- [x] Define interface `Rewriter` + `RewriteRequest`/`RewriteResult` (lihat `prd.md` §8.2).
- [x] Implementasi provider **[OI]-compatible** (client HTTP sendiri, tanpa dependency eksternal). Dukung `AI_BASE_URL`, `AI_MODEL`.
- [ ] (Opsional) provider Anthropic / Gemini.
- [x] **Prompt editorial** untuk anti-slop + output **JSON** terstruktur (title, slug, excerpt, content, category, tags).
- [x] Parse & validasi output JSON; ada fallback error jika format tidak valid (retry 1x).
- [x] Batas `AI_MAX_TOKENS`, `AI_TEMPERATURE` (≤0.7).
- [x] Paksa bahasa Inggris (`Language: en`).

**Files:** `internal/ai/rewriter.go`, `internal/ai/openai.go`, `internal/ai/prompt.go`.

**Acceptance:**
- [x] Input artikel → output JSON terstruktur & bahasa Inggris.
- [x] Prompt mencantumkan aturan anti-slop, panjang, struktur.
- [x] API key dari env, tidak di-log/hardcode.

**Catatan implementasi:**
- Provider [OI]-compatible diimplementasi dengan client HTTP stdlib (`net/http`) — tidak menambah dependency. Constructor: `ai.New[OI]Compatible(cfg ai.Config)`.
- Endpoint: `{AI_BASE_URL}/chat/completions`, format chat completions ([OI]-compatible), `Authorization: Bearer` hanya jika key non-kosong.
- Prompt sistem berisi daftar kata/frasa terlarang (slop) + aturan 5W+1H, 300–600 kata, tanpa markdown di body.
- Output diparse dari JSON; toleran terhadap pembungkus markdown fence (` ```json `). Retry 1x bila parse gagal.
- `AI_TEMPERATURE` hanya dikirim bila != 0; `AI_MAX_TOKENS` hanya bila > 0. Timeout default 60s.

---

## Phase 5 — Pipeline Orkestrasi

**Status: SELESAI** — `Pipeline.ProcessWebsite` (fetch→dedup→rewrite→image→save), download/watermark gambar, slug dedup, system user; build/vet/test green.

**Tujuan:** gabungkan semua langkah jadi satu alur per artikel.

**Deliverable:**
- [x] Buat `internal/service` (atau `pipeline`) dengan fungsi utama:
  ```
  ProcessWebsite(ctx, website) → (created, skipped int, err error)
  ```
  Langkah:
  1. fetch kandidat (source)
  2. dedup (cek source_url/hash)
  3. rewrite (AI)
  4. (jika ada gambar) download → credit/watermark → upload S3
  5. simpan artikel (existing `ArticleRepository`)
- [x] Helper **download & kredit gambar** (`internal/service/image.go`): HTTP GET, validasi tipe/ukuran, ambil attribution, panggil `UploadReader`.
- [x] Graceful degradation: gambar gagal → artikel tetap disimpan tanpa gambar (warning).
- [x] Slug generation dengan suffix dedup (reuse/ekstend logika existing).
- [x] `created_by` = system user ID.

**Files:** `internal/service/pipeline.go`, `internal/service/image.go`, `internal/service/slug.go`, `internal/service/watermark.go` (baru).

**Acceptance:**
- [x] Artikel tersimpan lengkap dengan metadata sumber & credit gambar.
- [x] Dedup berfungsi (artikel sama tidak dobel).
- [x] Error gambar tidak menggagalkan artikel.

**Catatan implementasi:**
- `entity.SourceItem` ditambah field `SourceName`/`SourceType` (sebelumnya `ImageCredit` menampung nama sumber, kini dipisah). `toSourceItem` + `fetchRSS` diperbarui.
- `Pipeline` memakai `source.Registry`, `ai.Rewriter`, `repository.ArticleRepository/TagRepository/CategoryRepository/SchedulerRunRepository`, `storage.Storage`.
- `ProcessWebsite` selalu mencatat `scheduler_runs` (status success/partial/failed via defer).
- Kategori di-resolve dari hasil AI (cocok by slug/name) — bila tidak cocok, artikel tetap disimpan dengan kategori `nil` (kategori bebas).
- Watermark: `golang.org/x/image/font/basicfont` (tanpa file font eksternal); teks kredit + latar semi-transparan di kiri-bawah; JPEG/PNG/GIF; gagal watermark → pakai gambar asli.
- Dependency baru: `golang.org/x/image` (menaikkan minor `golang.org/x/*`).

---

## Phase 6 — Scheduler & Wiring

**Status: SELESAI** — scheduler in-process goroutine + config env + wiring main.go; build/vet/test green.

**Tujuan:** jalankan pipeline secara periodik & aman.

**Deliverable:**
- [x] `internal/scheduler/scheduler.go`:
  - ticker dengan `SCHEDULER_INTERVAL_MINUTES`.
  - guard anti-overlap (mutex/atomic).
  - worker pool concurrency (batas `SCHEDULER_MAX_CONCURRENCY`).
  - grace shutdown via context.
  - catat `scheduler_runs` (dilakukan oleh `Pipeline.ProcessWebsite`).
- [x] Config baru (`internal/config`): `SCHEDULER_*`, `AI_*`, `NEWS_*`.
- [x] Wiring di `cmd/server/main.go`: jalankan scheduler hanya jika `SCHEDULER_ENABLED=true` (**in-process goroutine** — sesuai keputusan Phase 0).

**Files:** `internal/scheduler/scheduler.go` (baru), `internal/config/config.go`, `cmd/server/main.go`.

**Acceptance:**
- [x] Job jalan tiap 10 menit (default), tidak overlap.
- [x] `SIGTERM/SIGINT` graceful.
- [x] Banyak website diproses paralel dengan concurrency terbatas.

**Catatan implementasi:**
- Config baru di `config.go`: `SchedulerConfig{Enabled,IntervalMinutes,MaxConcurrency,DefaultStatus}`, `AIConfig{Provider,APIKey,BaseURL,Model,MaxTokens,Temperature}`, `NewsConfig{ItemsPerWebsite}`; helper `getEnvFloat` ditambah.
- Env vars: `SCHEDULER_ENABLED` (default false), `SCHEDULER_INTERVAL_MINUTES` (10), `SCHEDULER_MAX_CONCURRENCY` (3), `SCHEDULER_DEFAULT_STATUS` (published), `NEWS_ITEMS_PER_WEBSITE` (5), `AI_PROVIDER/AI_API_KEY/AI_BASE_URL/AI_MODEL/AI_MAX_TOKENS/AI_TEMPERATURE`.
- `Scheduler.Start(ctx)` idempoten (atomic guard); `Stop()` graceful; `RunNow(ctx)` untuk trigger manual (dipakai Phase 7 admin endpoint).
- `Pipeline` kini punya `SetItemsPerWebsite(n)` & `SetDefaultStatus(s)` (default 5 / published).
- Scheduler dijalankan di `main.go` setelah DB connect, hanya bila `SCHEDULER_ENABLED=true`; dihentikan sebelum server shutdown.
- `NEWS_*` saat ini hanya `NEWS_ITEMS_PER_WEBSITE` (feed sumber default dari `source.DefaultRSSFeeds()`).

---

## Phase 7 — Observability, Retry & Admin (opsional)

**Status: SELESAI** — structured log (slog), retry download gambar 1x, admin endpoint scheduler; build/vet/test green.

**Tujuan:** operasional & monitoring.

**Deliverable:**
- [x] Structured log per run & per artikel (level info/warn/error).
- [x] Retry terbatas (1x) untuk panggilan AI / download gagal sementara.
- [x] Endpoint admin (protected `admin`):
  - `POST /api/v1/admin/scheduler/run` — trigger manual (async).
  - `GET /api/v1/admin/scheduler/runs` — riwayat run (dengan pagination).
- [x] (wajib) reset kredensial: pastikan secret tidak ter-commit (sudah ada `.gitignore` untuk `.env`).

**Files:** `internal/handler/scheduler_handler.go` (baru), `internal/router/router.go`, `internal/service/image.go`, `cmd/server/main.go`.

**Acceptance:**
- [x] Riwayat run terbaca; trigger manual berfungsi.
- [x] Log jelas untuk debug.

**Catatan implementasi:**
- `SchedulerHandler` (`NewSchedulerHandler(sched, runRepo)`) — `Run` memicu scheduler async (goroutine + `context.WithTimeout` 30m) agar request tidak terblokir; balas `202 Accepted`. `ListRuns` pakai pagination `?page=&limit=`.
- `router.Setup` sekarang menerima parameter `sched *scheduler.Scheduler` (boleh nil): bila nil, endpoint trigger manual balas `503 scheduler disabled`; `GET /runs` tetap jalan.
- `main.go` membangun scheduler **sebelum** `router.Setup`, lalu diserahkan ke router.
- Retry download gambar (1x) untuk error transien (network / status 5xx) — 4xx dianggap permanen. Implementasi via sentinel `transientError` + `errors.As`.
- Logging memakai `log/slog` (structured, level info/warn/error) di pipeline/scheduler/main.

---

## Phase 8 — Test, Docs & Packaging

**Status: SELESAI** — unit test, README, `.env.example`, docker-compose; build/sqlc/test hijau.

**Tujuan:** kualitas & kemudahan deploy.

**Deliverable:**
- [x] Unit test: slug dedup, dedup source_url, parse output AI, image validation.
- [ ] (Opsional) integration test pakai mock source & mock rewriter.
- [x] Update `README.md`: cara setup API key, env baru, cara run scheduler.
- [x] Update `.env.example` (var scheduler/AI/news source).
- [x] Update `Makefile` (target dev tetap Air; tidak perlu worker terpisah — scheduler in-process).
- [x] Update `docker-compose.yml` (tambah `SCHEDULER_*`, `AI_*`, `NEWS_*`).

**Files:** `internal/service/slug_test.go`, `internal/service/pipeline_test.go`, `internal/ai/openai_test.go`, `internal/storage/helpers_test.go`, `README.md`, `.env.example`, `docker-compose.yml` (semua baru/ubah).

**Acceptance:**
- [x] `go test ./...` hijau (termasuk test existing).
- [x] `make build` & `make sqlc` sukses.
- [x] Dokumentasi env baru lengkap.

**Catatan implementasi:**
- 4 file test baru: `slug_test.go` (cleanSlug+itoa), `pipeline_test.go` (sourceHash+firstNonEmpty+strPtr), `openai_test.go` (stripFences+parseResult+anti-slop prompt), `helpers_test.go` (resolveImageType).
- Fix bug minor: `uniqueSlug` kini benar-benar fallback ke title bila slug kosong (sebelumnya `cleanSlug("")` langsung balas "article" sehingga fallback title tidak pernah jalan).
- `Makefile` tidak diubah (target `dev` Air & `build`/`sqlc` sudah ada; scheduler in-process tidak butuh target worker terpisah).
- `docker-compose.yml` menambah env `SCHEDULER_*`/`AI_*`/`NEWS_*` dengan placeholder `AI_API_KEY` kosong (tidak hardcode secret); nilai di-inject dari env host via `${VAR:-default}`.
- `.env.example` ditambah blok konfigurasi scheduler & AI (AI_API_KEY dikosongkan).

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
