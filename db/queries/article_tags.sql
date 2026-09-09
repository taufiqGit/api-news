-- name: CreateArticleTag :exec
INSERT INTO article_tags (article_id, tag_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListArticleTags :many
SELECT t.* FROM tags t
JOIN article_tags at ON at.tag_id = t.id
WHERE at.article_id = $1
ORDER BY t.name ASC;

-- name: DeleteArticleTags :exec
DELETE FROM article_tags WHERE article_id = $1;

-- name: ListArticlesByTag :many
SELECT a.*, u.name AS author_name, u.avatar_url AS author_avatar,
       c.name AS category_name, c.slug AS category_slug,
       w.name AS website_name, w.slug AS website_slug
FROM articles a
JOIN article_tags at ON at.article_id = a.id
JOIN users u ON u.id = a.created_by
LEFT JOIN categories c ON c.id = a.category_id
LEFT JOIN websites w ON w.id = a.website_id
WHERE at.tag_id = $1 AND a.status = 'published'
ORDER BY a.published_at DESC
LIMIT $2 OFFSET $3;
