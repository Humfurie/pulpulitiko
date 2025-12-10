-- Reverse Polls & Surveys System Migration

-- Drop triggers
DROP TRIGGER IF EXISTS update_polls_updated_at ON polls;
DROP TRIGGER IF EXISTS update_poll_comments_updated_at ON poll_comments;
DROP TRIGGER IF EXISTS poll_vote_count_trigger ON poll_votes;
DROP TRIGGER IF EXISTS poll_comment_count_trigger ON poll_comments;

-- Drop functions
DROP FUNCTION IF EXISTS update_poll_vote_counts();
DROP FUNCTION IF EXISTS update_poll_comment_counts();

-- Drop tables in reverse order
DROP TABLE IF EXISTS poll_comment_reactions;
DROP TABLE IF EXISTS poll_comments;
DROP TABLE IF EXISTS poll_votes;
DROP TABLE IF EXISTS poll_options;
DROP TABLE IF EXISTS polls;

-- Drop enum types
DROP TYPE IF EXISTS poll_category;
DROP TYPE IF EXISTS poll_status;
