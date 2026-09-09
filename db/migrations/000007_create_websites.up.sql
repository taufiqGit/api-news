-- 000007_create_websites.up.sql
-- Tabel website (multi-tenant: satu API, banyak website)
CREATE TABLE IF NOT EXISTS websites (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(120) NOT NULL UNIQUE,
    domain      VARCHAR(255),
    description TEXT,
    logo_url    TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_websites_slug ON websites(slug);
CREATE INDEX idx_websites_domain ON websites(domain);

-- Tambahkan website_id ke articles (nullable dulu, backfill, lalu NOT NULL)
ALTER TABLE articles ADD COLUMN IF NOT EXISTS website_id UUID REFERENCES websites(id) ON DELETE CASCADE;

CREATE INDEX idx_articles_website_id ON articles(website_id);
