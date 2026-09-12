-- 000009_add_article_source_and_scheduler_runs.up.sql
-- Fitur scheduler berita: metadata sumber & penanda AI pada articles,
-- index dedup, dan tabel audit scheduler_runs.

ALTER TABLE articles
    ADD COLUMN IF NOT EXISTS source_url       TEXT,
    ADD COLUMN IF NOT EXISTS source_type      VARCHAR(50),   -- google_news / rss / manual
    ADD COLUMN IF NOT EXISTS source_name      VARCHAR(255),  -- mis. "BBC News"
    ADD COLUMN IF NOT EXISTS source_hash      VARCHAR(64),   -- dedup fallback tanpa URL
    ADD COLUMN IF NOT EXISTS is_ai_generated  BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS image_credit     TEXT,
    ADD COLUMN IF NOT EXISTS image_source_url TEXT,
    ADD COLUMN IF NOT EXISTS image_license    VARCHAR(255);

-- Dedup: URL sumber unik secara global (lintas website)
CREATE UNIQUE INDEX IF NOT EXISTS idx_articles_source_url
    ON articles (source_url)
    WHERE source_url IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_articles_source_hash
    ON articles (source_hash)
    WHERE source_hash IS NOT NULL;

-- Index komposit: daftar artikel terbaru per website
CREATE INDEX IF NOT EXISTS idx_articles_website_created
    ON articles (website_id, created_at DESC);

-- Audit run scheduler
CREATE TABLE IF NOT EXISTS scheduler_runs (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    website_id       UUID REFERENCES websites(id) ON DELETE SET NULL,
    status           VARCHAR(20) NOT NULL DEFAULT 'running', -- running, success, partial, failed
    started_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at      TIMESTAMPTZ,
    articles_created INTEGER NOT NULL DEFAULT 0,
    articles_skipped INTEGER NOT NULL DEFAULT 0,
    articles_failed  INTEGER NOT NULL DEFAULT 0,
    error            TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scheduler_runs_website ON scheduler_runs(website_id);
CREATE INDEX IF NOT EXISTS idx_scheduler_runs_started_at ON scheduler_runs(started_at DESC);