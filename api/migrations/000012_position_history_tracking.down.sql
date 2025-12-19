-- Rollback Migration: Position History Tracking
-- Description: Drops all tables and triggers created in 000012_position_history_tracking.up.sql

-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_politician_position_history_updated_at ON politician_position_history;
DROP TRIGGER IF EXISTS trigger_election_events_updated_at ON election_events;

-- Drop trigger functions
DROP FUNCTION IF EXISTS update_politician_position_history_updated_at();
DROP FUNCTION IF EXISTS update_election_events_updated_at();

-- Drop tables in reverse order of creation (respecting foreign key dependencies)
DROP TABLE IF EXISTS politician_import_logs CASCADE;
DROP TABLE IF EXISTS politician_position_history CASCADE;
DROP TABLE IF EXISTS election_events CASCADE;

-- Note: We don't restore the original politician.position_id and politician.party_id data
-- because this is a destructive operation. If you need to rollback, ensure you have backups.
