-- Migration: 000004_locations (DOWN)
-- Rollback Philippine Geographic Locations

-- Drop triggers
DROP TRIGGER IF EXISTS update_districts_updated_at ON congressional_districts;
DROP TRIGGER IF EXISTS update_barangays_updated_at ON barangays;
DROP TRIGGER IF EXISTS update_cities_updated_at ON cities_municipalities;
DROP TRIGGER IF EXISTS update_provinces_updated_at ON provinces;
DROP TRIGGER IF EXISTS update_regions_updated_at ON regions;

-- Drop tables in reverse order (respecting foreign key constraints)
DROP TABLE IF EXISTS district_coverage;
DROP TABLE IF EXISTS congressional_districts;
DROP TABLE IF EXISTS barangays;
DROP TABLE IF EXISTS cities_municipalities;
DROP TABLE IF EXISTS provinces;
DROP TABLE IF EXISTS regions;
