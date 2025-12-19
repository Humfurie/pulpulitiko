-- Migration: Position History Tracking with Election Events and Import Logs
-- Description: Implements historical tracking of politician positions, election events, and Excel import logging

-- ============================================================================
-- Table: politician_position_history
-- Description: Tracks all position assignments over time with jurisdiction
-- ============================================================================
CREATE TABLE IF NOT EXISTS politician_position_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    politician_id UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    position_id UUID NOT NULL REFERENCES government_positions(id) ON DELETE RESTRICT,
    party_id UUID REFERENCES political_parties(id) ON DELETE SET NULL,

    -- Jurisdiction fields (polymorphic relationship)
    region_id UUID REFERENCES regions(id) ON DELETE SET NULL,
    province_id UUID REFERENCES provinces(id) ON DELETE SET NULL,
    city_id UUID REFERENCES cities_municipalities(id) ON DELETE SET NULL,
    barangay_id UUID REFERENCES barangays(id) ON DELETE SET NULL,
    district_id UUID REFERENCES congressional_districts(id) ON DELETE SET NULL,
    is_national BOOLEAN DEFAULT FALSE NOT NULL,

    -- Term dates
    term_start DATE NOT NULL,
    term_end DATE,

    -- Status tracking
    is_current BOOLEAN DEFAULT TRUE NOT NULL,
    ended_reason VARCHAR(100), -- 'term_expired', 'resigned', 'replaced', 'election', 'deceased', etc.

    -- Metadata
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW() NOT NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Election linkage (optional)
    election_id UUID, -- FK added after election_events table is created

    -- Ensure valid jurisdiction type: only one location type OR is_national
    CONSTRAINT check_history_jurisdiction_type CHECK (
        (is_national = TRUE AND region_id IS NULL AND province_id IS NULL AND city_id IS NULL AND barangay_id IS NULL AND district_id IS NULL) OR
        (is_national = FALSE AND (
            (region_id IS NOT NULL AND province_id IS NULL AND city_id IS NULL AND barangay_id IS NULL AND district_id IS NULL) OR
            (province_id IS NOT NULL AND region_id IS NULL AND city_id IS NULL AND barangay_id IS NULL AND district_id IS NULL) OR
            (city_id IS NOT NULL AND region_id IS NULL AND province_id IS NULL AND barangay_id IS NULL AND district_id IS NULL) OR
            (barangay_id IS NOT NULL AND region_id IS NULL AND province_id IS NULL AND city_id IS NULL AND district_id IS NULL) OR
            (district_id IS NOT NULL AND region_id IS NULL AND province_id IS NULL AND city_id IS NULL AND barangay_id IS NULL)
        ))
    ),

    -- Ensure term_end is after term_start if specified
    CONSTRAINT check_term_dates CHECK (term_end IS NULL OR term_end >= term_start)
);

-- Indexes for efficient queries
CREATE INDEX idx_politician_position_history_politician ON politician_position_history(politician_id);
CREATE INDEX idx_politician_position_history_position ON politician_position_history(position_id);
CREATE INDEX idx_politician_position_history_party ON politician_position_history(party_id) WHERE party_id IS NOT NULL;
CREATE INDEX idx_politician_position_history_current ON politician_position_history(is_current) WHERE is_current = TRUE;
CREATE INDEX idx_politician_position_history_dates ON politician_position_history(term_start, term_end);
CREATE INDEX idx_politician_position_history_region ON politician_position_history(region_id) WHERE region_id IS NOT NULL;
CREATE INDEX idx_politician_position_history_province ON politician_position_history(province_id) WHERE province_id IS NOT NULL;
CREATE INDEX idx_politician_position_history_city ON politician_position_history(city_id) WHERE city_id IS NOT NULL;
CREATE INDEX idx_politician_position_history_barangay ON politician_position_history(barangay_id) WHERE barangay_id IS NOT NULL;
CREATE INDEX idx_politician_position_history_district ON politician_position_history(district_id) WHERE district_id IS NOT NULL;

-- Unique partial indexes to ensure only one current holder per position-jurisdiction combination
-- National positions (e.g., President, Senators)
CREATE UNIQUE INDEX idx_unique_current_position_national
ON politician_position_history(position_id)
WHERE is_current = TRUE AND is_national = TRUE;

-- Regional positions
CREATE UNIQUE INDEX idx_unique_current_position_region
ON politician_position_history(position_id, region_id)
WHERE is_current = TRUE AND is_national = FALSE AND region_id IS NOT NULL;

-- Provincial positions
CREATE UNIQUE INDEX idx_unique_current_position_province
ON politician_position_history(position_id, province_id)
WHERE is_current = TRUE AND is_national = FALSE AND province_id IS NOT NULL;

-- City/Municipal positions
CREATE UNIQUE INDEX idx_unique_current_position_city
ON politician_position_history(position_id, city_id)
WHERE is_current = TRUE AND is_national = FALSE AND city_id IS NOT NULL;

-- Barangay positions
CREATE UNIQUE INDEX idx_unique_current_position_barangay
ON politician_position_history(position_id, barangay_id)
WHERE is_current = TRUE AND is_national = FALSE AND barangay_id IS NOT NULL;

-- Congressional district positions
CREATE UNIQUE INDEX idx_unique_current_position_district
ON politician_position_history(position_id, district_id)
WHERE is_current = TRUE AND is_national = FALSE AND district_id IS NOT NULL;

-- ============================================================================
-- Table: election_events
-- Description: Tracks elections that trigger bulk position changes
-- ============================================================================
CREATE TABLE IF NOT EXISTS election_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    election_date DATE NOT NULL,

    -- Election level
    level VARCHAR(50) NOT NULL, -- 'national', 'local', 'barangay', 'regional', etc.

    -- Status
    status VARCHAR(50) DEFAULT 'scheduled' NOT NULL, -- 'scheduled', 'in_progress', 'completed', 'cancelled'

    -- Metadata
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW() NOT NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Ensure valid status
    CONSTRAINT check_election_status CHECK (status IN ('scheduled', 'in_progress', 'completed', 'cancelled'))
);

-- Indexes
CREATE INDEX idx_election_events_date ON election_events(election_date DESC);
CREATE INDEX idx_election_events_status ON election_events(status);
CREATE INDEX idx_election_events_level ON election_events(level);
CREATE INDEX idx_election_events_created_by ON election_events(created_by) WHERE created_by IS NOT NULL;

-- ============================================================================
-- Table: politician_import_logs
-- Description: Tracks Excel import operations with detailed error logging
-- ============================================================================
CREATE TABLE IF NOT EXISTS politician_import_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filename VARCHAR(500) NOT NULL,
    uploaded_by UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Statistics
    total_rows INTEGER NOT NULL DEFAULT 0,
    successful_imports INTEGER DEFAULT 0 NOT NULL,
    failed_imports INTEGER DEFAULT 0 NOT NULL,
    politicians_created INTEGER DEFAULT 0 NOT NULL,
    politicians_updated INTEGER DEFAULT 0 NOT NULL,
    positions_archived INTEGER DEFAULT 0 NOT NULL,

    -- Status
    status VARCHAR(50) DEFAULT 'pending' NOT NULL, -- 'pending', 'processing', 'completed', 'failed'
    error_log TEXT,
    validation_errors JSONB, -- Array of validation errors with row numbers

    -- Election linkage (optional - if import is for an election)
    election_id UUID REFERENCES election_events(id) ON DELETE SET NULL,

    -- Timestamps
    started_at TIMESTAMP DEFAULT NOW() NOT NULL,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,

    -- Ensure valid status
    CONSTRAINT check_import_status CHECK (status IN ('pending', 'processing', 'completed', 'failed'))
);

-- Indexes
CREATE INDEX idx_politician_import_logs_status ON politician_import_logs(status);
CREATE INDEX idx_politician_import_logs_uploaded_by ON politician_import_logs(uploaded_by) WHERE uploaded_by IS NOT NULL;
CREATE INDEX idx_politician_import_logs_election ON politician_import_logs(election_id) WHERE election_id IS NOT NULL;
CREATE INDEX idx_politician_import_logs_created_at ON politician_import_logs(created_at DESC);
CREATE INDEX idx_politician_import_logs_started_at ON politician_import_logs(started_at DESC);

-- ============================================================================
-- Add foreign key constraint for election_id in politician_position_history
-- ============================================================================
ALTER TABLE politician_position_history
ADD CONSTRAINT fk_position_history_election
FOREIGN KEY (election_id) REFERENCES election_events(id) ON DELETE SET NULL;

CREATE INDEX idx_politician_position_history_election ON politician_position_history(election_id) WHERE election_id IS NOT NULL;

-- ============================================================================
-- Trigger: Auto-update updated_at timestamp
-- ============================================================================
CREATE OR REPLACE FUNCTION update_politician_position_history_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_politician_position_history_updated_at
BEFORE UPDATE ON politician_position_history
FOR EACH ROW
EXECUTE FUNCTION update_politician_position_history_updated_at();

CREATE OR REPLACE FUNCTION update_election_events_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_election_events_updated_at
BEFORE UPDATE ON election_events
FOR EACH ROW
EXECUTE FUNCTION update_election_events_updated_at();

-- ============================================================================
-- Data Migration: Migrate existing politician positions to history table
-- ============================================================================
-- Migrate current positions from politicians table to politician_position_history
-- This ensures existing data is preserved in the new historical tracking system

INSERT INTO politician_position_history (
    politician_id,
    position_id,
    party_id,
    region_id,
    province_id,
    city_id,
    barangay_id,
    district_id,
    is_national,
    term_start,
    term_end,
    is_current,
    created_at,
    updated_at
)
SELECT
    p.id as politician_id,
    p.position_id,
    p.party_id,
    -- Get jurisdiction from politician_jurisdictions if exists, otherwise NULL
    (SELECT pj.region_id FROM politician_jurisdictions pj WHERE pj.politician_id = p.id LIMIT 1) as region_id,
    (SELECT pj.province_id FROM politician_jurisdictions pj WHERE pj.politician_id = p.id LIMIT 1) as province_id,
    (SELECT pj.city_id FROM politician_jurisdictions pj WHERE pj.politician_id = p.id LIMIT 1) as city_id,
    (SELECT pj.barangay_id FROM politician_jurisdictions pj WHERE pj.politician_id = p.id LIMIT 1) as barangay_id,
    p.district_id,
    COALESCE((SELECT pj.is_national FROM politician_jurisdictions pj WHERE pj.politician_id = p.id LIMIT 1), p.level = 'national') as is_national,
    COALESCE(p.term_start, p.created_at::date) as term_start,
    p.term_end,
    TRUE as is_current, -- All existing positions are marked as current
    p.created_at,
    p.updated_at
FROM politicians p
WHERE p.position_id IS NOT NULL -- Only migrate politicians with positions
  AND p.deleted_at IS NULL; -- Don't migrate soft-deleted politicians

-- ============================================================================
-- Comments for documentation
-- ============================================================================
COMMENT ON TABLE politician_position_history IS 'Historical tracking of politician position assignments with jurisdiction details';
COMMENT ON TABLE election_events IS 'Election events that trigger bulk position changes';
COMMENT ON TABLE politician_import_logs IS 'Logs of Excel import operations with statistics and error details';

COMMENT ON COLUMN politician_position_history.is_current IS 'Only one position per position-jurisdiction can have is_current=TRUE (enforced by unique indexes)';
COMMENT ON COLUMN politician_position_history.ended_reason IS 'Reason for term ending: term_expired, resigned, replaced, election, deceased, etc.';
COMMENT ON COLUMN politician_position_history.is_national IS 'True for national positions (no specific jurisdiction), false for regional/local positions';
COMMENT ON COLUMN politician_import_logs.validation_errors IS 'JSONB array of validation errors: [{"row": 5, "field": "position", "error": "Position not found"}]';
