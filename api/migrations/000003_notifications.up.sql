-- Migration: 000003_notifications
-- Adds notification system and comment mentions

-- =====================================================
-- COMMENT MENTIONS (Article Comments)
-- =====================================================

-- Article comment user mentions (track @mentions to users)
CREATE TABLE comment_user_mentions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comment_id UUID NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    mentioned_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (comment_id, mentioned_user_id)
);

CREATE INDEX idx_comment_user_mentions_comment ON comment_user_mentions(comment_id);
CREATE INDEX idx_comment_user_mentions_user ON comment_user_mentions(mentioned_user_id);

-- =====================================================
-- POLITICIAN COMMENT MENTIONS
-- =====================================================

-- Politician comment mentions (track @mentions)
CREATE TABLE politician_comment_mentions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comment_id UUID NOT NULL REFERENCES politician_comments(id) ON DELETE CASCADE,
    mentioned_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (comment_id, mentioned_user_id)
);

CREATE INDEX idx_politician_comment_mentions_comment ON politician_comment_mentions(comment_id);
CREATE INDEX idx_politician_comment_mentions_user ON politician_comment_mentions(mentioned_user_id);

-- =====================================================
-- NOTIFICATIONS SYSTEM
-- =====================================================

-- Notification types enum
CREATE TYPE notification_type AS ENUM (
    'mention_article_comment',      -- Someone mentioned you in an article comment
    'mention_politician_comment',   -- Someone mentioned you in a politician comment
    'reply_article_comment',        -- Someone replied to your article comment
    'reply_politician_comment',     -- Someone replied to your politician comment
    'comment_reaction'              -- Someone reacted to your comment
);

-- Notifications table
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,  -- Recipient
    type notification_type NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT,

    -- Reference to the source (polymorphic)
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,  -- Who triggered the notification

    -- Optional references based on type
    article_id UUID REFERENCES articles(id) ON DELETE CASCADE,
    politician_id UUID REFERENCES politicians(id) ON DELETE CASCADE,
    comment_id UUID,  -- Could be either article or politician comment

    -- State
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP,

    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for notifications
CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_user_unread ON notifications(user_id) WHERE is_read = FALSE;
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_actor ON notifications(actor_id);
