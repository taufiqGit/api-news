-- 000006_add_updated_at_trigger.down.sql
DROP TRIGGER IF EXISTS trg_articles_updated_at ON articles;
DROP TRIGGER IF EXISTS trg_categories_updated_at ON categories;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;

DROP FUNCTION IF EXISTS set_updated_at();
