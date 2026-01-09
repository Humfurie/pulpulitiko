-- Remove search indexes for tags, categories, and users

DROP INDEX IF EXISTS idx_users_name_search;
DROP INDEX IF EXISTS idx_users_name_email;
DROP INDEX IF EXISTS idx_categories_search;
DROP INDEX IF EXISTS idx_tags_search;
