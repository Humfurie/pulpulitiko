-- Reverse Elections System Migration

-- Drop triggers
DROP TRIGGER IF EXISTS update_elections_updated_at ON elections;
DROP TRIGGER IF EXISTS update_candidates_updated_at ON candidates;
DROP TRIGGER IF EXISTS update_voter_education_updated_at ON voter_education;

-- Drop tables in reverse order
DROP TABLE IF EXISTS voter_education;
DROP TABLE IF EXISTS precinct_results;
DROP TABLE IF EXISTS election_results;
DROP TABLE IF EXISTS candidates;
DROP TABLE IF EXISTS election_positions;
DROP TABLE IF EXISTS elections;

-- Drop enum types
DROP TYPE IF EXISTS candidate_status;
DROP TYPE IF EXISTS election_status;
DROP TYPE IF EXISTS election_type;
