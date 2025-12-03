-- Rollback: 000001_initial_schema

-- Drop messaging system first
DROP TRIGGER IF EXISTS update_conversations_updated_at ON conversations;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversations;

-- Drop comment system triggers and functions first
DROP TRIGGER IF EXISTS enforce_single_level_threading ON comments;
DROP FUNCTION IF EXISTS check_single_level_threading();
DROP TRIGGER IF EXISTS update_comments_updated_at ON comments;

-- Drop comment tables (must be before articles/authors due to FKs)
DROP TABLE IF EXISTS comment_mentions;
DROP TABLE IF EXISTS comment_reactions;
DROP TABLE IF EXISTS comments;

-- Drop comment status enum type
DROP TYPE IF EXISTS comment_status;

-- Drop other triggers
DROP TRIGGER IF EXISTS update_articles_updated_at ON articles;
DROP TRIGGER IF EXISTS update_tags_updated_at ON tags;
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;
DROP TRIGGER IF EXISTS update_authors_updated_at ON authors;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS article_tags;
DROP TABLE IF EXISTS articles;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS authors;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
