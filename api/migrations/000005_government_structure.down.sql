-- Reverse Government Structure Migration

-- Drop triggers
DROP TRIGGER IF EXISTS update_political_parties_updated_at ON political_parties;
DROP TRIGGER IF EXISTS update_government_positions_updated_at ON government_positions;

-- Drop indexes on politicians
DROP INDEX IF EXISTS idx_politicians_level;
DROP INDEX IF EXISTS idx_politicians_branch;
DROP INDEX IF EXISTS idx_politicians_position_id;
DROP INDEX IF EXISTS idx_politicians_party_id;
DROP INDEX IF EXISTS idx_politicians_district_id;
DROP INDEX IF EXISTS idx_politicians_level_branch;

-- Remove columns from politicians
ALTER TABLE politicians
DROP COLUMN IF EXISTS level,
DROP COLUMN IF EXISTS branch,
DROP COLUMN IF EXISTS position_id,
DROP COLUMN IF EXISTS party_id,
DROP COLUMN IF EXISTS district_id;

-- Drop politician_jurisdictions table
DROP TABLE IF EXISTS politician_jurisdictions;

-- Drop government_positions table
DROP TABLE IF EXISTS government_positions;

-- Drop political_parties table
DROP TABLE IF EXISTS political_parties;

-- Drop enum types
DROP TYPE IF EXISTS government_branch;
DROP TYPE IF EXISTS government_level;
