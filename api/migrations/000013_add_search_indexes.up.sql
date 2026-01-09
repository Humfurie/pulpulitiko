-- Add search indexes for tags, categories, and users for better search performance

-- Tags: index for searching by name
CREATE INDEX IF NOT EXISTS idx_tags_search
ON tags USING gin(to_tsvector('english', name));

-- Categories: index for searching by name and description
CREATE INDEX IF NOT EXISTS idx_categories_search
ON categories USING gin(to_tsvector('english', name || ' ' || COALESCE(description, '')));

-- Users: composite index for searching by name and email
CREATE INDEX IF NOT EXISTS idx_users_name_email
ON users(name, email) WHERE deleted_at IS NULL;

-- Users: text search index for name
CREATE INDEX IF NOT EXISTS idx_users_name_search
ON users USING gin(to_tsvector('english', name)) WHERE deleted_at IS NULL;
