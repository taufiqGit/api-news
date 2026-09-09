-- name: CreateComment :one
INSERT INTO comments (article_id, user_id, parent_id, content)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCommentByID :one
SELECT * FROM comments WHERE id = $1;

-- name: ListArticleComments :many
SELECT c.*, u.name AS user_name, u.avatar_url AS user_avatar
FROM comments c
JOIN users u ON u.id = c.user_id
WHERE c.article_id = $1
  AND ($2::boolean OR c.is_approved = TRUE)
ORDER BY c.created_at ASC;

-- name: UpdateComment :one
UPDATE comments
SET content = sqlc.arg('content'),
    is_approved = COALESCE(sqlc.narg('is_approved'), is_approved)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comments WHERE id = $1;

-- name: CountArticleComments :one
SELECT COUNT(*) FROM comments WHERE article_id = $1;
