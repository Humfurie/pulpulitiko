-- Migration: 000010_polls_location (DOWN)
-- Remove location fields from polls

-- Drop indexes
DROP INDEX IF EXISTS idx_polls_region;
DROP INDEX IF EXISTS idx_polls_province;
DROP INDEX IF EXISTS idx_polls_city;
DROP INDEX IF EXISTS idx_polls_barangay;

-- Remove columns
ALTER TABLE polls
    DROP COLUMN IF EXISTS region_id,
    DROP COLUMN IF EXISTS province_id,
    DROP COLUMN IF EXISTS city_municipality_id,
    DROP COLUMN IF EXISTS barangay_id;
