-- Reverse Legislation/Bills Tracker Migration

-- Drop triggers
DROP TRIGGER IF EXISTS update_legislative_sessions_updated_at ON legislative_sessions;
DROP TRIGGER IF EXISTS update_committees_updated_at ON committees;
DROP TRIGGER IF EXISTS update_bills_updated_at ON bills;

-- Drop junction and relationship tables first
DROP TABLE IF EXISTS bill_topic_assignments;
DROP TABLE IF EXISTS politician_votes;
DROP TABLE IF EXISTS bill_votes;
DROP TABLE IF EXISTS bill_committees;
DROP TABLE IF EXISTS bill_status_history;
DROP TABLE IF EXISTS bill_authors;

-- Drop main tables
DROP TABLE IF EXISTS bill_topics;
DROP TABLE IF EXISTS bills;
DROP TABLE IF EXISTS committees;
DROP TABLE IF EXISTS legislative_sessions;

-- Drop enum types
DROP TYPE IF EXISTS vote_type;
DROP TYPE IF EXISTS legislative_chamber;
DROP TYPE IF EXISTS bill_status;
