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
-- CATEGORIES
-- =====================================================
INSERT INTO categories (name, slug, description) VALUES
    ('National Politics', 'national-politics', 'News and analysis about national government and politics'),
    ('Local Government', 'local-government', 'Coverage of local government units and regional political developments'),
    ('Elections', 'elections', 'Election news, candidates, and voting information'),
    ('Policy', 'policy', 'Analysis of government policies and their impact'),
    ('Opinion', 'opinion', 'Editorial opinions and commentary on political matters'),
    ('Investigations', 'investigations', 'In-depth investigative reports on political matters')
ON CONFLICT (slug) DO NOTHING;

-- =====================================================
-- TAGS
-- =====================================================
INSERT INTO tags (name, slug) VALUES
    ('Breaking News', 'breaking-news'),
    ('Senate', 'senate'),
    ('Congress', 'congress'),
    ('Budget', 'budget'),
    ('2025 Elections', '2025-elections'),
    ('Executive', 'executive'),
    ('Policy', 'policy'),
    ('Corruption', 'corruption'),
    ('Supreme Court', 'supreme-court'),
    ('Human Rights', 'human-rights')
ON CONFLICT (slug) DO NOTHING;

-- =====================================================
-- AUTHORS
-- =====================================================
INSERT INTO authors (name, slug, bio, email, role_id, social_links)
SELECT
    'Juan dela Cruz',
    'juan-dela-cruz',
    'Senior political correspondent covering national politics and governance.',
    'juan@pulpulitiko.com',
    (SELECT id FROM roles WHERE slug = 'author'),
    '{"twitter": "https://twitter.com/juandelacruz", "facebook": "https://facebook.com/juandelacruz"}'
ON CONFLICT (slug) DO NOTHING;

INSERT INTO authors (name, slug, bio, email, role_id, social_links)
SELECT
    'Maria Santos',
    'maria-santos',
    'Investigative journalist specializing in electoral processes and local government.',
    'maria@pulpulitiko.com',
    (SELECT id FROM roles WHERE slug = 'author'),
    '{"twitter": "https://twitter.com/mariasantos"}'
ON CONFLICT (slug) DO NOTHING;

INSERT INTO authors (name, slug, bio, email, role_id, social_links)
SELECT
    'Pedro Reyes',
    'pedro-reyes',
    'Opinion columnist and political analyst with 15 years of experience.',
    'pedro@pulpulitiko.com',
    (SELECT id FROM roles WHERE slug = 'author'),
    '{"linkedin": "https://linkedin.com/in/pedroreyes"}'
ON CONFLICT (slug) DO NOTHING;

-- =====================================================
-- ARTICLES
-- =====================================================
INSERT INTO articles (slug, title, summary, content, status, view_count, published_at, category_id, author_id)
SELECT
    'senate-passes-2025-national-budget',
    'Senate Passes 2025 National Budget After Marathon Session',
    'The Philippine Senate approved the P5.768 trillion national budget for 2025 following weeks of deliberation and a marathon plenary session.',
    '<p>The Philippine Senate approved the P5.768 trillion national budget for 2025 following weeks of deliberation and a marathon plenary session that lasted until the early hours of Wednesday morning.</p><p>Senate President Juan Miguel Zubiri expressed satisfaction with the outcome, noting that the budget prioritizes education, infrastructure, and social services.</p><p>"This budget reflects our commitment to the Filipino people," Zubiri said in a press conference. "We have ensured that funds are allocated where they are needed most."</p><h2>Key Allocations</h2><p>The approved budget includes:</p><ul><li>P924.3 billion for the Department of Education</li><li>P789.8 billion for the Department of Public Works and Highways</li><li>P293.4 billion for the Department of Health</li><li>P249.1 billion for the Department of Social Welfare and Development</li></ul><p>The budget will now go through a bicameral conference to reconcile differences with the House version before being sent to the President for signing.</p>',
    'published',
    1250,
    NOW() - INTERVAL '2 days',
    (SELECT id FROM categories WHERE slug = 'national-politics'),
    (SELECT id FROM authors WHERE slug = 'juan-dela-cruz')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO articles (slug, title, summary, content, status, view_count, published_at, category_id, author_id)
SELECT
    'comelec-sets-filing-period-2025-elections',
    'COMELEC Sets October Filing Period for 2025 Midterm Elections',
    'The Commission on Elections announces the certificate of candidacy filing period for the May 2025 midterm elections.',
    '<p>The Commission on Elections (COMELEC) has officially set the certificate of candidacy (COC) filing period for the May 2025 midterm elections from October 1 to October 8, 2024.</p><p>COMELEC Chairman George Garcia announced the schedule during a press briefing, emphasizing the importance of an orderly filing process.</p><h2>Positions Up for Grabs</h2><p>The 2025 midterm elections will see Filipinos voting for:</p><ul><li>12 Senate seats</li><li>All 316 congressional district representatives</li><li>All local government positions from governor down to councilors</li></ul><p>Garcia reminded prospective candidates to ensure their documents are complete before filing to avoid delays.</p>',
    'published',
    980,
    NOW() - INTERVAL '3 days',
    (SELECT id FROM categories WHERE slug = 'elections'),
    (SELECT id FROM authors WHERE slug = 'maria-santos')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO articles (slug, title, summary, content, status, view_count, published_at, category_id, author_id)
SELECT
    'infrastructure-push-build-better-more',
    'Infrastructure Push Continues Under Build Better More Program',
    'The government reports significant progress in flagship infrastructure projects under the Build Better More initiative.',
    '<p>The Department of Public Works and Highways (DPWH) reported significant progress in flagship infrastructure projects under the Build Better More program, the current administration''s continuation of infrastructure development.</p><p>Secretary Manuel Bonoan highlighted that over 12,000 kilometers of roads have been constructed or rehabilitated since the program''s inception.</p><h2>Major Projects</h2><p>Key infrastructure projects nearing completion include:</p><ul><li>Metro Manila Subway - 60% complete</li><li>North-South Commuter Railway - Phase 1 operational by Q2 2025</li><li>Mindanao Railway - Initial segment under construction</li></ul><p>The infrastructure push aims to improve connectivity and boost economic growth across the archipelago.</p>',
    'published',
    756,
    NOW() - INTERVAL '5 days',
    (SELECT id FROM categories WHERE slug = 'national-politics'),
    (SELECT id FROM authors WHERE slug = 'juan-dela-cruz')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO articles (slug, title, summary, content, status, view_count, published_at, category_id, author_id)
SELECT
    'political-dynasties-reform-proposals',
    'Political Dynasties Reform Proposals Gain Traction in Congress',
    'Multiple bills seeking to regulate political dynasties are now pending in both chambers of Congress.',
    '<p>The long-standing debate on political dynasties has gained renewed momentum as multiple bills seeking to regulate the practice are now pending in both chambers of Congress.</p><p>Senator Risa Hontiveros, author of one such bill, argues that political dynasties perpetuate inequality and limit democratic participation.</p><h2>Proposed Regulations</h2><p>The various bills propose different approaches:</p><ul><li>Prohibiting relatives within the second degree of consanguinity from running simultaneously</li><li>Limiting family members holding elective positions within the same province</li><li>Requiring a "cooling off" period between terms for family members</li></ul><p>Critics argue that such regulations may infringe on the right to run for public office, while supporters maintain that they are necessary for genuine democratic representation.</p>',
    'published',
    542,
    NOW() - INTERVAL '7 days',
    (SELECT id FROM categories WHERE slug = 'opinion'),
    (SELECT id FROM authors WHERE slug = 'pedro-reyes')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO articles (slug, title, summary, content, status, view_count, published_at, category_id, author_id)
SELECT
    'lgu-performance-challenge-fund',
    'High-Performing LGUs to Receive Performance Challenge Fund Bonus',
    'DILG announces the list of local government units qualified for the Performance Challenge Fund.',
    '<p>The Department of the Interior and Local Government (DILG) has announced the list of local government units (LGUs) that qualified for the Performance Challenge Fund (PCF) bonus for fiscal year 2024.</p><p>Interior Secretary Benhur Abalos commended the qualifying LGUs for their excellent performance in governance and service delivery.</p><h2>Qualifying Criteria</h2><p>LGUs were evaluated based on:</p><ul><li>Financial administration and sustainability</li><li>Disaster preparedness</li><li>Social protection programs implementation</li><li>Business-friendliness and competitiveness</li><li>Peace and order maintenance</li></ul><p>The PCF provides additional funding for development projects in high-performing LGUs.</p>',
    'published',
    423,
    NOW() - INTERVAL '10 days',
    (SELECT id FROM categories WHERE slug = 'local-government'),
    (SELECT id FROM authors WHERE slug = 'maria-santos')
ON CONFLICT (slug) DO NOTHING;

-- =====================================================
-- ARTICLE TAGS
-- =====================================================
INSERT INTO article_tags (article_id, tag_id)
SELECT a.id, t.id FROM articles a, tags t
WHERE a.slug = 'senate-passes-2025-national-budget' AND t.slug IN ('senate', 'budget')
ON CONFLICT DO NOTHING;

INSERT INTO article_tags (article_id, tag_id)
SELECT a.id, t.id FROM articles a, tags t
WHERE a.slug = 'comelec-sets-filing-period-2025-elections' AND t.slug IN ('2025-elections')
ON CONFLICT DO NOTHING;

INSERT INTO article_tags (article_id, tag_id)
SELECT a.id, t.id FROM articles a, tags t
WHERE a.slug = 'infrastructure-push-build-better-more' AND t.slug IN ('executive', 'policy')
ON CONFLICT DO NOTHING;

INSERT INTO article_tags (article_id, tag_id)
SELECT a.id, t.id FROM articles a, tags t
WHERE a.slug = 'political-dynasties-reform-proposals' AND t.slug IN ('congress', 'policy')
ON CONFLICT DO NOTHING;

INSERT INTO article_tags (article_id, tag_id)
SELECT a.id, t.id FROM articles a, tags t
WHERE a.slug = 'lgu-performance-challenge-fund' AND t.slug IN ('policy')
ON CONFLICT DO NOTHING;
