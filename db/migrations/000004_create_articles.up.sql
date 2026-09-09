-- 000004_create_articles.sql
CREATE TABLE IF NOT EXISTS articles (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title        VARCHAR(500) NOT NULL,
    slug         VARCHAR(550) NOT NULL UNIQUE,
    excerpt      TEXT,
    content      TEXT NOT NULL,
    cover_image  TEXT,
    status       VARCHAR(20) NOT NULL DEFAULT 'draft', -- draft, published, archived, deleted
    view_count   INTEGER NOT NULL DEFAULT 0,
    published_at TIMESTAMPTZ,
    created_by   UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    category_id  UUID REFERENCES categories(id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_articles_slug ON articles(slug);
CREATE INDEX idx_articles_status ON articles(status);
CREATE INDEX idx_articles_category ON articles(category_id);
CREATE INDEX idx_articles_created_by ON articles(created_by);
CREATE INDEX idx_articles_published_at ON articles(published_at DESC);
