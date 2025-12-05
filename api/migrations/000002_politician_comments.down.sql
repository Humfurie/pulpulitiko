-- Rollback: 000002_politician_comments

-- Drop triggers first
DROP TRIGGER IF EXISTS enforce_politician_comment_single_level_threading ON politician_comments;
DROP FUNCTION IF EXISTS check_politician_comment_single_level_threading();
DROP TRIGGER IF EXISTS update_politician_comments_updated_at ON politician_comments;

-- Drop tables
DROP TABLE IF EXISTS politician_comment_reactions;
DROP TABLE IF EXISTS politician_comments;
