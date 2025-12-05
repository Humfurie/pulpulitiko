-- Migration: 000002_politician_comments
-- Adds discussion/comment system for politician pages

-- Politician comments table (similar to article comments but for politician pages)
CREATE TABLE politician_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    politician_id UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES politician_comments(id) ON DELETE CASCADE, -- NULL for root comments, set for replies
    content TEXT NOT NULL, -- Supports markdown: **bold**, *italic*, ~~strikethrough~~, > quotes, @mentions
    status comment_status DEFAULT 'active' NOT NULL, -- Reuses existing comment_status enum
    moderated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    moderated_at TIMESTAMP,
    moderation_reason TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);

-- Politician comment reactions
CREATE TABLE politician_comment_reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comment_id UUID NOT NULL REFERENCES politician_comments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reaction VARCHAR(50) NOT NULL, -- 'heart', 'thumbsup', 'thumbsdown', 'laugh', 'fire', 'eyes'
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (comment_id, user_id)
);

-- Indexes for politician comments
CREATE INDEX idx_politician_comments_politician ON politician_comments(politician_id);
CREATE INDEX idx_politician_comments_user ON politician_comments(user_id);
CREATE INDEX idx_politician_comments_parent ON politician_comments(parent_id);
CREATE INDEX idx_politician_comments_created_at ON politician_comments(created_at DESC);
CREATE INDEX idx_politician_comments_deleted_at ON politician_comments(deleted_at);
CREATE INDEX idx_politician_comments_status ON politician_comments(status);
CREATE INDEX idx_politician_comment_reactions_comment ON politician_comment_reactions(comment_id);
CREATE INDEX idx_politician_comment_reactions_user ON politician_comment_reactions(user_id);

-- Trigger for updated_at
CREATE TRIGGER update_politician_comments_updated_at BEFORE UPDATE ON politician_comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to enforce single-level threading for politician comments
CREATE OR REPLACE FUNCTION check_politician_comment_single_level_threading()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.parent_id IS NOT NULL THEN
        IF EXISTS (SELECT 1 FROM politician_comments WHERE id = NEW.parent_id AND parent_id IS NOT NULL) THEN
            RAISE EXCEPTION 'Replies cannot have replies. Only single-level threading is allowed.';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to enforce single-level threading
CREATE TRIGGER enforce_politician_comment_single_level_threading
    BEFORE INSERT OR UPDATE ON politician_comments
    FOR EACH ROW
    EXECUTE FUNCTION check_politician_comment_single_level_threading();
