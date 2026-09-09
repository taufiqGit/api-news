-- name: CreateCategory :one
INSERT INTO categories (name, slug, description, parent_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCategoryByID :one
SELECT * FROM categories WHERE id = $1;

-- name: GetCategoryBySlug :one
SELECT * FROM categories WHERE slug = $1;

-- name: ListCategories :many
SELECT * FROM categories
ORDER BY name ASC;

-- name: ListCategoriesPaginated :many
SELECT * FROM categories
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateCategory :one
UPDATE categories
SET name = COALESCE(sqlc.narg('name'), name),
    slug = COALESCE(sqlc.narg('slug'), slug),
    description = COALESCE(sqlc.narg('description'), description),
    parent_id = COALESCE(sqlc.narg('parent_id'), parent_id),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = $1;

-- name: CountCategories :one
SELECT COUNT(*) FROM categories;
