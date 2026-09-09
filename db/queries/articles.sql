-- name: CreateArticle :one
INSERT INTO articles (title, slug, excerpt, content, cover_image, status, published_at, created_by, category_id, website_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetArticleByID :one
SELECT a.*, u.name AS author_name, u.avatar_url AS author_avatar,
       c.name AS category_name, c.slug AS category_slug,
       w.name AS website_name, w.slug AS website_slug
FROM articles a
JOIN users u ON u.id = a.created_by
LEFT JOIN categories c ON c.id = a.category_id
LEFT JOIN websites w ON w.id = a.website_id
WHERE a.id = $1;

-- name: GetArticleBySlug :one
SELECT a.*, u.name AS author_name, u.avatar_url AS author_avatar,
       c.name AS category_name, c.slug AS category_slug,
       w.name AS website_name, w.slug AS website_slug
FROM articles a
JOIN users u ON u.id = a.created_by
LEFT JOIN categories c ON c.id = a.category_id
LEFT JOIN websites w ON w.id = a.website_id
WHERE a.slug = $1;

-- name: ListArticles :many
SELECT a.*, u.name AS author_name, u.avatar_url AS author_avatar,
       c.name AS category_name, c.slug AS category_slug,
       w.name AS website_name, w.slug AS website_slug
FROM articles a
JOIN users u ON u.id = a.created_by
LEFT JOIN categories c ON c.id = a.category_id
LEFT JOIN websites w ON w.id = a.website_id
WHERE ($1::text = '' OR a.status = $1)
  AND (CAST(sqlc.narg('category_id') AS uuid) IS NULL OR a.category_id = CAST(sqlc.narg('category_id') AS uuid))
  AND (CAST(sqlc.narg('website_id') AS uuid) IS NULL OR a.website_id = CAST(sqlc.narg('website_id') AS uuid))
ORDER BY a.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListPublishedArticles :many
SELECT a.*, u.name AS author_name, u.avatar_url AS author_avatar,
       c.name AS category_name, c.slug AS category_slug,
       w.name AS website_name, w.slug AS website_slug
FROM articles a
JOIN users u ON u.id = a.created_by
LEFT JOIN categories c ON c.id = a.category_id
LEFT JOIN websites w ON w.id = a.website_id
WHERE a.status = 'published'
  AND a.published_at <= NOW()
  AND (CAST(sqlc.narg('website_id') AS uuid) IS NULL OR a.website_id = CAST(sqlc.narg('website_id') AS uuid))
ORDER BY a.published_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateArticle :one
UPDATE articles
SET title = COALESCE(sqlc.narg('title'), title),
    slug = COALESCE(sqlc.narg('slug'), slug),
    excerpt = COALESCE(sqlc.narg('excerpt'), excerpt),
    content = COALESCE(sqlc.narg('content'), content),
    cover_image = COALESCE(sqlc.narg('cover_image'), cover_image),
    status = COALESCE(sqlc.narg('status'), status),
    published_at = CASE WHEN sqlc.narg('status') = 'published' AND published_at IS NULL THEN NOW() ELSE published_at END,
    category_id = COALESCE(sqlc.narg('category_id'), category_id),
    website_id = COALESCE(sqlc.narg('website_id'), website_id),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: UpdateArticleStatus :one
UPDATE articles
SET status = $2,
    published_at = CASE WHEN $2 = 'published' THEN COALESCE(published_at, NOW()) ELSE published_at END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: IncrementViewCount :one
UPDATE articles
SET view_count = view_count + 1
WHERE id = $1
RETURNING view_count;

-- name: DeleteArticle :exec
DELETE FROM articles WHERE id = $1;

-- name: CountArticles :one
SELECT COUNT(*) FROM articles
WHERE ($1::text = '' OR status = $1)
  AND (CAST(sqlc.narg('website_id') AS uuid) IS NULL OR website_id = CAST(sqlc.narg('website_id') AS uuid));
