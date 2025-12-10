-- Migration: 000009_location_seed_data (DOWN)
-- Remove seeded location data

-- Remove congressional districts
DELETE FROM congressional_districts;

-- Remove cities/municipalities
DELETE FROM cities_municipalities;

-- Remove provinces (except keeping the regions)
DELETE FROM provinces;
