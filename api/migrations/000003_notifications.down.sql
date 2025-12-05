-- Rollback: 000003_notifications

-- Drop notifications
DROP TABLE IF EXISTS notifications;
DROP TYPE IF EXISTS notification_type;

-- Drop politician comment mentions
DROP TABLE IF EXISTS politician_comment_mentions;

-- Drop article comment user mentions
DROP TABLE IF EXISTS comment_user_mentions;
