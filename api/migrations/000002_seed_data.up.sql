-- Migration: 002_seed_data
-- Description: Seed initial data for Philippine Politics Blog

-- =====================================================
-- PERMISSIONS
-- =====================================================
INSERT INTO permissions (name, slug, description, category) VALUES
    -- Articles permissions
    ('View Articles', 'view_articles', 'Can view articles in admin panel', 'articles'),
    ('Create Articles', 'create_articles', 'Can create new articles', 'articles'),
    ('Edit Articles', 'edit_articles', 'Can edit existing articles', 'articles'),
    ('Delete Articles', 'delete_articles', 'Can delete articles', 'articles'),
    ('Publish Articles', 'publish_articles', 'Can publish/unpublish articles', 'articles'),

    -- Categories permissions
    ('View Categories', 'view_categories', 'Can view categories in admin panel', 'categories'),
    ('Create Categories', 'create_categories', 'Can create new categories', 'categories'),
    ('Edit Categories', 'edit_categories', 'Can edit existing categories', 'categories'),
    ('Delete Categories', 'delete_categories', 'Can delete categories', 'categories'),

    -- Tags permissions
    ('View Tags', 'view_tags', 'Can view tags in admin panel', 'tags'),
    ('Create Tags', 'create_tags', 'Can create new tags', 'tags'),
    ('Edit Tags', 'edit_tags', 'Can edit existing tags', 'tags'),
    ('Delete Tags', 'delete_tags', 'Can delete tags', 'tags'),

    -- Users permissions
    ('View Users', 'view_users', 'Can view users in admin panel', 'users'),
    ('Create Users', 'create_users', 'Can create new users', 'users'),
    ('Edit Users', 'edit_users', 'Can edit existing users', 'users'),
    ('Delete Users', 'delete_users', 'Can delete users', 'users'),

    -- Roles permissions
    ('View Roles', 'view_roles', 'Can view roles in admin panel', 'roles'),
    ('Create Roles', 'create_roles', 'Can create new roles', 'roles'),
    ('Edit Roles', 'edit_roles', 'Can edit existing roles', 'roles'),
    ('Delete Roles', 'delete_roles', 'Can delete roles', 'roles'),

    -- Metrics permissions
    ('View Metrics', 'view_metrics', 'Can view dashboard metrics', 'metrics'),

    -- Upload permissions
    ('Upload Files', 'upload_files', 'Can upload files', 'uploads')
ON CONFLICT (slug) DO NOTHING;

-- =====================================================
-- ROLES
-- =====================================================
INSERT INTO roles (name, slug, description, is_system) VALUES
    ('Administrator', 'admin', 'Full access to all features and settings', TRUE),
    ('Author', 'author', 'Can manage articles, categories, and tags', TRUE),
    ('User', 'user', 'Basic user with limited access', TRUE)
ON CONFLICT (slug) DO NOTHING;

-- =====================================================
-- ROLE PERMISSIONS
-- =====================================================

-- Admin gets ALL permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.slug = 'admin'
ON CONFLICT DO NOTHING;

-- Author permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.slug = 'author' AND p.slug IN (
    'view_articles', 'create_articles', 'edit_articles', 'delete_articles', 'publish_articles',
    'view_categories', 'create_categories', 'edit_categories',
    'view_tags', 'create_tags', 'edit_tags',
    'view_metrics',
    'upload_files'
)
ON CONFLICT DO NOTHING;

-- User permissions (very limited)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.slug = 'user' AND p.slug IN (
    'view_articles',
    'view_categories',
    'view_tags'
)
ON CONFLICT DO NOTHING;

-- =====================================================
-- NOTE: Categories, Tags, Authors, and Articles are now
-- seeded via the seed command (cmd/seed/main.go).
-- This keeps the migration focused on system-level data
-- (permissions, roles) and allows flexible content seeding.
-- =====================================================
