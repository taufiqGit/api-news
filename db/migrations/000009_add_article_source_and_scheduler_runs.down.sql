-- 000009_add_article_source_and_scheduler_runs.down.sql

DROP TABLE IF EXISTS scheduler_runs;

DROP INDEX IF EXISTS idx_articles_website_created;
DROP INDEX IF EXISTS idx_articles_source_hash;
DROP INDEX IF EXISTS idx_articles_source_url;

ALTER TABLE articles
    DROP COLUMN IF EXISTS image_license,
    DROP COLUMN IF EXISTS image_source_url,
    DROP COLUMN IF EXISTS image_credit,
    DROP COLUMN IF EXISTS is_ai_generated,
    DROP COLUMN IF EXISTS source_hash,
    DROP COLUMN IF EXISTS source_name,
    DROP COLUMN IF EXISTS source_type,
    DROP COLUMN IF EXISTS source_url;