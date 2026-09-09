-- 000007_create_websites.down.sql
ALTER TABLE articles DROP COLUMN IF EXISTS website_id;
DROP TABLE IF EXISTS websites;
