-- Migration: 000010_polls_location
-- Add location/region filtering to polls

-- Add location fields to polls table
ALTER TABLE polls
    ADD COLUMN region_id UUID REFERENCES regions(id) ON DELETE SET NULL,
    ADD COLUMN province_id UUID REFERENCES provinces(id) ON DELETE SET NULL,
    ADD COLUMN city_municipality_id UUID REFERENCES cities_municipalities(id) ON DELETE SET NULL,
    ADD COLUMN barangay_id UUID REFERENCES barangays(id) ON DELETE SET NULL;

-- Add indexes for location filtering
CREATE INDEX idx_polls_region ON polls(region_id) WHERE region_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_polls_province ON polls(province_id) WHERE province_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_polls_city ON polls(city_municipality_id) WHERE city_municipality_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_polls_barangay ON polls(barangay_id) WHERE barangay_id IS NOT NULL AND deleted_at IS NULL;

-- Add comment explaining the location hierarchy
COMMENT ON COLUMN polls.region_id IS 'Optional: Scope poll to a specific region';
COMMENT ON COLUMN polls.province_id IS 'Optional: Scope poll to a specific province';
COMMENT ON COLUMN polls.city_municipality_id IS 'Optional: Scope poll to a specific city/municipality';
COMMENT ON COLUMN polls.barangay_id IS 'Optional: Scope poll to a specific barangay';
