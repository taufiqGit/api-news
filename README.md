# News API

Starter project REST API website berita dengan **layered architecture** yang siap dikembangkan menjadi skala kompleks.

## Tech Stack

| Layer | Teknologi |
|---|---|
| Language | Go 1.26 |
| HTTP Framework | Gin |
| Database | PostgreSQL |
| SQL Generator | sqlc (type-safe queries) |
| DB Driver | pgx/v5 |
| Auth | JWT (HS256) |

## Arsitektur (Layered)

```
cmd/server/               → entry point (main.go)
internal/
├── config/               → konfigurasi dari .env
├── database/             → koneksi pool PostgreSQL + generated sqlc code
│   └── db/               → (auto-generated oleh sqlc — jangan edit)
├── domain/
│   ├── entity/           → model & request/response DTO
│   └── repository/       → interface repository (kontrak data layer)
├── repository/           → implementasi repository (sqlc + pgx)
├── usecase/              → business logic (validasi, transformasi)
├── handler/              → HTTP handlers (bind request, parse response)
│   └── response/         → format response API standar
├── middleware/           → JWT auth, CORS, role-based access
└── router/               → registrasi route + dependency injection
db/
├── migrations/           → SQL schema (urut: 000001, 000002, ...)
├── queries/              → SQL queries untuk sqlc
└── sqlc.yaml             → konfigurasi sqlc
```

**Alur request:** `HTTP → Router → Middleware → Handler → Usecase → Repository → PostgreSQL`

## Fitur

- ✅ Auth: register, login, JWT (role: admin, editor, author)
- ✅ Role-based access control (RBAC)
- ✅ CRUD Artikel (draft/published/archived, slug, view count auto-increment)
- ✅ CRUD Kategori (hierarki via parent_id)
- ✅ CRUD Tag + relasi many-to-many dengan artikel
- ✅ Komentar (reply via parent_id, approval)
- ✅ Pagination + filter (status, kategori)
- ✅ Swagger UI docs (interaktif)
- ✅ Upload gambar via S3 storage (AWS S3, MinIO, R2, DO Spaces)
- ✅ Graceful shutdown
- ✅ CORS

## Swagger Docs

Dokumentasi API interaktif tersedia di:

```
http://localhost:8080/swagger/index.html
```

Setelah mengubah handler, regenerate docs:

```bash
make docs
```

Tombol **Authorize** di Swagger UI dipakai untuk memasukkan token JWT
(format: `Bearer <token>`) sehingga endpoint protected bisa dicoba langsung.

## Quick Start

### 1. Prasyarat

- Go 1.21+
- PostgreSQL 14+

### 2. Setup

```bash
# Clone & masuk folder
cd news-api

# Copy env & sesuaikan kredensial database
cp .env.example .env

# Install dependency
go mod tidy
```

### 3. Database

```bash
# Buat database
createdb news_db

# Jalankan migrations
make migrate-up
# atau manual:
psql -d news_db -f db/migrations/000001_create_users.sql
psql -d news_db -f db/migrations/000002_create_categories.sql
# ... dst
```

### 4. Regenerate sqlc (setelah edit db/queries/)

```bash
make sqlc
```

### 5. Jalankan

```bash
make run        # go run ./cmd/server
# atau
make build && ./bin/news-api
```

## API Endpoints

### Public
| Method | Path | Deskripsi |
|---|---|---|
| GET | `/healthz` | Health check |
| POST | `/api/v1/auth/register` | Register user |
| POST | `/api/v1/auth/login` | Login → dapat JWT |
| GET | `/api/v1/articles/published` | Artikel published (publik) |
| GET | `/api/v1/articles/slug/:slug` | Detail artikel by slug (publik) |
| GET | `/api/v1/articles` | List artikel (filter status/kategori) |
| GET | `/api/v1/articles/:id` | Detail artikel by id |
| GET | `/api/v1/articles/:id/comments` | Komentar artikel |
| GET | `/api/v1/categories` | List kategori |
| GET | `/api/v1/categories/:id` | Detail kategori |
| GET | `/api/v1/categories/slug/:slug` | Kategori by slug |
| GET | `/api/v1/tags` | List tag |

### Protected (butuh `Authorization: Bearer <token>`)
| Method | Path | Role |
|---|---|---|
| GET | `/api/v1/users/me` | all |
| GET/PUT/DELETE | `/api/v1/users` | admin |
| POST/PUT/DELETE | `/api/v1/categories` | editor+ |
| POST/PUT/DELETE | `/api/v1/tags` | editor+ |
| POST/PUT/PATCH/DELETE | `/api/v1/articles` | author+ |
| POST/DELETE | `/api/v1/comments` | all |

### Contoh Request

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@mail.com","password":"secret123","role":"author"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@mail.com","password":"secret123"}'
# → {"data":{"token":"eyJ..."}}

# Buat artikel (pakai token)
curl -X POST http://localhost:8080/api/v1/articles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"title":"Judul Berita","slug":"judul-berita","content":"Isi berita...","status":"published"}'
```

## Konfigurasi (.env)

| Variabel | Default | Deskripsi |
|---|---|---|
| `APP_PORT` | 8080 | Port server |
| `APP_ENV` | development | development/production |
| `DB_HOST` | localhost | Host PostgreSQL |
| `DB_PORT` | 5432 | Port PostgreSQL |
| `DB_USER` | postgres | User DB |
| `DB_PASSWORD` | postgres | Password DB |
| `DB_NAME` | news_db | Nama database |
| `DB_SSLMODE` | disable | SSL mode |
| `JWT_SECRET` | change-me | Secret JWT (ganti di production!) |
| `JWT_EXPIRES_IN_HOURS` | 24 | Umur token |
| `S3_ENDPOINT` | s3.amazonaws.com | Endpoint S3 (MinIO: localhost:9000) |
| `S3_REGION` | ap-southeast-1 | Region bucket |
| `S3_BUCKET` | news-api | Nama bucket |
| `S3_ACCESS_KEY` | (kosong) | Access key — kosong = local storage |
| `S3_SECRET_KEY` | (kosong) | Secret key — kosong = local storage |
| `S3_USE_SSL` | true | Pakai HTTPS untuk S3 |
| `S3_PUBLIC_URL` | (kosong) | URL publik CDN (opsional, contoh: https://cdn.deployaja.web.id) |

## Upload Gambar

Endpoint: `POST /api/v1/uploads/images` (auth: author+)

```bash
curl -X POST https://api-news.deployaja.web.id/api/v1/uploads/images \
  -H "Authorization: Bearer <TOKEN>" \
  -F "file=@cover.jpg" \
  -F "folder=covers"
```

Response:

```json
{
  "success": true,
  "data": {
    "key": "covers/2026/08/uuid.jpg",
    "url": "https://cdn.deployaja.web.id/covers/2026/08/uuid.jpg",
    "size": 123456,
    "content_type": "image/jpeg"
  }
}
```

- Format yang diizinkan: `jpeg/png/webp/gif/avif` (max 20MB)
- URL hasil upload bisa langsung dipakai di field `cover_image` saat buat/update artikel
- Tanpa kredensial S3, file tersimpan di folder `uploads/` lokal (dev mode)

## Perintah Makefile

```bash
make run        # jalankan server
make build      # build binary ke bin/
make sqlc       # regenerate sqlc
make docs       # regenerate swagger docs
make migrate-up # jalankan migrations (butuh golang-migrate)
make test       # jalankan test
make tidy       # go mod tidy
```

## Struktur Database

```
users (id, name, email, password, role, avatar_url, is_active)
categories (id, name, slug, description, parent_id, is_active)  ← parent_id untuk hierarki
tags (id, name, slug)
articles (id, title, slug, excerpt, content, cover_image, status, view_count, published_at, created_by, category_id)
article_tags (article_id, tag_id)  ← many-to-many
comments (id, article_id, user_id, parent_id, content, is_approved)  ← parent_id untuk reply
```

## Roadmap Pengembangan

- [ ] Refresh token / logout
- [ ] Upload gambar (cover, avatar) — S3/local storage
- [ ] Search artikel (full-text search / trigram)
- [ ] Slug otomatis dari title
- [ ] Rate limiting
- [ ] Swagger/OpenAPI docs
- [ ] Caching (Redis) untuk artikel populer
- [ ] Webhook/subscription newsletter
- [ ] Audit log
