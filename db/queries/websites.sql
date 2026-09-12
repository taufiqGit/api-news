-- name: CreateWebsite :one
INSERT INTO websites (name, slug, domain, description, logo_url, is_active)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetWebsiteByID :one
SELECT * FROM websites WHERE id = $1;

-- name: GetWebsiteBySlug :one
SELECT * FROM websites WHERE slug = $1;

-- name: GetWebsiteByDomain :one
SELECT * FROM websites WHERE domain = $1;

-- name: ListWebsites :many
SELECT * FROM websites
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountWebsites :one
SELECT COUNT(*) FROM websites;

-- name: UpdateWebsite :one
UPDATE websites
SET name = COALESCE(sqlc.narg('name'), name),
    slug = COALESCE(sqlc.narg('slug'), slug),
    domain = COALESCE(sqlc.narg('domain'), domain),
    description = COALESCE(sqlc.narg('description'), description),
    logo_url = COALESCE(sqlc.narg('logo_url'), logo_url),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteWebsite :exec
DELETE FROM websites WHERE id = $1;

-- name: ListActiveWebsites :many
SELECT * FROM websites
WHERE is_active = TRUE
ORDER BY created_at ASC;
