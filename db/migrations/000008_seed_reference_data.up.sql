-- 000008_seed_reference_data.up.sql
-- Seed data awal untuk fitur scheduler berita:
--   1) system user (pemilik artikel hasil AI, bukan untuk login)
--   2) kategori berita (English)
--   3) tag umum (English)
--   4) website contoh (target publikasi / multi-tenant)
--
-- Semua UUID deterministik supaya mudah dirujuk dari kode.
-- INSERT idempoten (ON CONFLICT DO NOTHING) agar aman dijalankan ulang.

-- 1) System user -------------------------------------------------------------
INSERT INTO users (id, name, email, password, role, avatar_url, is_active)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'System',
    'system@news-api.internal',
    '$2a$10$zmGQLAcMUQ/ZRZJy1RdwUeLfMzHYJOD3EreIH/0.vHFXrg0C0XMo6',
    'system',
    NULL,
    FALSE
)
ON CONFLICT DO NOTHING;

-- 2) Kategori berita (English, flat / tanpa parent) -------------------------
INSERT INTO categories (id, name, slug, description, parent_id, is_active) VALUES
    ('10000000-0000-0000-0000-000000000001', 'World',         'world',         NULL, NULL, TRUE),
    ('10000000-0000-0000-0000-000000000002', 'Business',      'business',      NULL, NULL, TRUE),
    ('10000000-0000-0000-0000-000000000003', 'Technology',    'technology',    NULL, NULL, TRUE),
    ('10000000-0000-0000-0000-000000000004', 'Science',       'science',       NULL, NULL, TRUE),
    ('10000000-0000-0000-0000-000000000005', 'Health',        'health',        NULL, NULL, TRUE),
    ('10000000-0000-0000-0000-000000000006', 'Sports',        'sports',        NULL, NULL, TRUE),
    ('10000000-0000-0000-0000-000000000007', 'Entertainment', 'entertainment', NULL, NULL, TRUE),
    ('10000000-0000-0000-0000-000000000008', 'Politics',      'politics',      NULL, NULL, TRUE),
    ('10000000-0000-0000-0000-000000000009', 'Economy',       'economy',       NULL, NULL, TRUE),
    ('10000000-0000-0000-0000-00000000000a', 'Lifestyle',     'lifestyle',     NULL, NULL, TRUE)
ON CONFLICT DO NOTHING;

-- 3) Tag umum (English) ------------------------------------------------------
INSERT INTO tags (id, name, slug) VALUES
    ('20000000-0000-0000-0000-000000000001', 'Breaking',    'breaking'),
    ('20000000-0000-0000-0000-000000000002', 'Viral',       'viral'),
    ('20000000-0000-0000-0000-000000000003', 'Trending',    'trending'),
    ('20000000-0000-0000-0000-000000000004', 'World News',  'world-news'),
    ('20000000-0000-0000-0000-000000000005', 'Technology',  'technology'),
    ('20000000-0000-0000-0000-000000000006', 'AI',          'ai'),
    ('20000000-0000-0000-0000-000000000007', 'Economy',     'economy'),
    ('20000000-0000-0000-0000-000000000008', 'Climate',     'climate')
ON CONFLICT DO NOTHING;

-- 4) Website contoh (target publikasi) ---------------------------------------
-- Ganti domain/deskripsi dengan data produksi kamu saat sudah siap.
INSERT INTO websites (id, name, slug, domain, description, logo_url, is_active) VALUES
    (
        '30000000-0000-0000-0000-000000000001',
        'Global News Hub',
        'global-news-hub',
        'https://globalnewshub.example.com',
        'Portal berita internasional (contoh).',
        NULL,
        TRUE
    ),
    (
        '30000000-0000-0000-0000-000000000002',
        'Tech Daily',
        'tech-daily',
        'https://techdaily.example.com',
        'Portal berita teknologi (contoh).',
        NULL,
        TRUE
    ),
    (
        '30000000-0000-0000-0000-000000000003',
        'World Report',
        'world-report',
        'https://worldreport.example.com',
        'Portal laporan dunia (contoh).',
        NULL,
        TRUE
    )
ON CONFLICT DO NOTHING;