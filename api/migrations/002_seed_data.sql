-- Migration: 002_seed_data
-- Description: Seed initial data for Philippine Politics Blog

-- Insert default categories
INSERT INTO categories (name, slug, description) VALUES
    ('National Politics', 'national-politics', 'News and analysis about national government and politics'),
    ('Local Politics', 'local-politics', 'Coverage of local government units and regional political developments'),
    ('Elections', 'elections', 'Election news, candidates, and voting information'),
    ('Policy', 'policy', 'Analysis of government policies and their impact'),
    ('Opinion', 'opinion', 'Editorial opinions and commentary on political matters'),
    ('Fact Check', 'fact-check', 'Verification of political claims and statements')
ON CONFLICT (slug) DO NOTHING;

-- Insert default tags
INSERT INTO tags (name, slug) VALUES
    ('Breaking News', 'breaking-news'),
    ('Analysis', 'analysis'),
    ('Investigation', 'investigation'),
    ('Economy', 'economy'),
    ('Health', 'health'),
    ('Education', 'education'),
    ('Environment', 'environment'),
    ('Foreign Policy', 'foreign-policy'),
    ('Human Rights', 'human-rights'),
    ('Corruption', 'corruption')
ON CONFLICT (slug) DO NOTHING;
