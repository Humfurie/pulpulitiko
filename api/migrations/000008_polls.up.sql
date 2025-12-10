-- Polls & Surveys System Migration
-- Community engagement feature for user-created and admin polls

-- Poll Status enum
CREATE TYPE poll_status AS ENUM (
    'draft',
    'pending_approval',
    'active',
    'closed',
    'rejected'
);

-- Poll Category enum
CREATE TYPE poll_category AS ENUM (
    'general',
    'election',
    'legislation',
    'politician',
    'policy',
    'local_issue',
    'national_issue'
);

-- Polls table
CREATE TABLE polls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    title VARCHAR(300) NOT NULL,
    slug VARCHAR(300) UNIQUE NOT NULL,
    description TEXT,
    category poll_category NOT NULL DEFAULT 'general',
    status poll_status NOT NULL DEFAULT 'draft',
    -- Optional associations
    politician_id UUID REFERENCES politicians(id) ON DELETE SET NULL,
    election_id UUID REFERENCES elections(id) ON DELETE SET NULL,
    bill_id UUID REFERENCES bills(id) ON DELETE SET NULL,
    -- Settings
    is_anonymous BOOLEAN DEFAULT TRUE,
    allow_multiple_votes BOOLEAN DEFAULT FALSE,
    show_results_before_vote BOOLEAN DEFAULT FALSE,
    is_featured BOOLEAN DEFAULT FALSE,
    -- Scheduling
    starts_at TIMESTAMP,
    ends_at TIMESTAMP,
    -- Moderation
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP,
    rejection_reason TEXT,
    -- Stats
    total_votes INTEGER DEFAULT 0,
    view_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_polls_user ON polls(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_polls_status ON polls(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_polls_category ON polls(category) WHERE deleted_at IS NULL;
CREATE INDEX idx_polls_slug ON polls(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_polls_featured ON polls(is_featured) WHERE is_featured = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_polls_active ON polls(status, starts_at, ends_at) WHERE status = 'active' AND deleted_at IS NULL;
CREATE INDEX idx_polls_politician ON polls(politician_id) WHERE politician_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_polls_election ON polls(election_id) WHERE election_id IS NOT NULL AND deleted_at IS NULL;

-- Poll Options table
CREATE TABLE poll_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    poll_id UUID NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    text VARCHAR(500) NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 0,
    vote_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_poll_options_poll ON poll_options(poll_id);

-- Poll Votes table
CREATE TABLE poll_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    poll_id UUID NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    option_id UUID NOT NULL REFERENCES poll_options(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL, -- NULL if anonymous
    ip_hash VARCHAR(64), -- Hashed IP for anonymous vote tracking
    created_at TIMESTAMP DEFAULT NOW(),
    -- Prevent duplicate votes (by user or IP hash)
    UNIQUE(poll_id, user_id),
    UNIQUE(poll_id, ip_hash)
);

CREATE INDEX idx_poll_votes_poll ON poll_votes(poll_id);
CREATE INDEX idx_poll_votes_option ON poll_votes(option_id);
CREATE INDEX idx_poll_votes_user ON poll_votes(user_id) WHERE user_id IS NOT NULL;

-- Poll Comments table
CREATE TABLE poll_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    poll_id UUID NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    parent_id UUID REFERENCES poll_comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    -- Moderation
    moderated_by UUID REFERENCES users(id),
    moderated_at TIMESTAMP,
    moderation_reason TEXT,
    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_poll_comments_poll ON poll_comments(poll_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_poll_comments_user ON poll_comments(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_poll_comments_parent ON poll_comments(parent_id) WHERE parent_id IS NOT NULL AND deleted_at IS NULL;

-- Poll Comment Reactions table
CREATE TABLE poll_comment_reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comment_id UUID NOT NULL REFERENCES poll_comments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reaction VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(comment_id, user_id, reaction)
);

CREATE INDEX idx_poll_comment_reactions_comment ON poll_comment_reactions(comment_id);

-- Triggers
CREATE TRIGGER update_polls_updated_at
    BEFORE UPDATE ON polls
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_poll_comments_updated_at
    BEFORE UPDATE ON poll_comments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to update poll vote counts
CREATE OR REPLACE FUNCTION update_poll_vote_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- Update option vote count
        UPDATE poll_options SET vote_count = vote_count + 1 WHERE id = NEW.option_id;
        -- Update poll total votes
        UPDATE polls SET total_votes = total_votes + 1 WHERE id = NEW.poll_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        -- Update option vote count
        UPDATE poll_options SET vote_count = vote_count - 1 WHERE id = OLD.option_id;
        -- Update poll total votes
        UPDATE polls SET total_votes = total_votes - 1 WHERE id = OLD.poll_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER poll_vote_count_trigger
    AFTER INSERT OR DELETE ON poll_votes
    FOR EACH ROW
    EXECUTE FUNCTION update_poll_vote_counts();

-- Function to update poll comment counts
CREATE OR REPLACE FUNCTION update_poll_comment_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE polls SET comment_count = comment_count + 1 WHERE id = NEW.poll_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE polls SET comment_count = comment_count - 1 WHERE id = OLD.poll_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER poll_comment_count_trigger
    AFTER INSERT OR DELETE ON poll_comments
    FOR EACH ROW
    EXECUTE FUNCTION update_poll_comment_counts();

-- Seed some sample polls
INSERT INTO polls (user_id, title, slug, description, category, status, is_featured, starts_at)
SELECT
    u.id,
    'Should the Philippines shift to a federal form of government?',
    'federal-form-government',
    'A poll to gauge public opinion on the proposed shift to federalism in the Philippines.',
    'policy',
    'active',
    TRUE,
    NOW()
FROM users u JOIN roles r ON u.role_id = r.id WHERE r.name = 'admin' LIMIT 1;

INSERT INTO poll_options (poll_id, text, display_order)
SELECT p.id, option_text, option_order
FROM polls p, (
    VALUES
        ('Yes, it would benefit the country', 1),
        ('No, the current system is fine', 2),
        ('Need more information to decide', 3),
        ('Indifferent / No opinion', 4)
) AS options(option_text, option_order)
WHERE p.slug = 'federal-form-government';

INSERT INTO polls (user_id, title, slug, description, category, status, is_featured, starts_at)
SELECT
    u.id,
    'What is the most important issue for the 2025 elections?',
    'important-issue-2025-elections',
    'Help identify the key issues that should be prioritized in the upcoming elections.',
    'election',
    'active',
    TRUE,
    NOW()
FROM users u JOIN roles r ON u.role_id = r.id WHERE r.name = 'admin' LIMIT 1;

INSERT INTO poll_options (poll_id, text, display_order)
SELECT p.id, option_text, option_order
FROM polls p, (
    VALUES
        ('Economy and Jobs', 1),
        ('Education', 2),
        ('Healthcare', 3),
        ('Infrastructure', 4),
        ('Peace and Security', 5),
        ('Environment and Climate Change', 6),
        ('Anti-Corruption', 7),
        ('Other', 8)
) AS options(option_text, option_order)
WHERE p.slug = 'important-issue-2025-elections';
