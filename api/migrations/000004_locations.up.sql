-- Migration: 000004_locations
-- Philippine Geographic Locations (PSGC-based hierarchy)

-- =====================================================
-- LOCATION TABLES (Philippine Standard Geographic Code)
-- =====================================================

-- Regions (17 regions including NCR, CAR, BARMM)
CREATE TABLE regions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) UNIQUE NOT NULL,  -- PSGC code (e.g., "130000000" for NCR)
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);

-- Provinces (82 provinces)
CREATE TABLE provinces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region_id UUID NOT NULL REFERENCES regions(id) ON DELETE CASCADE,
    code VARCHAR(20) UNIQUE NOT NULL,  -- PSGC code
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);

-- Cities and Municipalities
CREATE TABLE cities_municipalities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    province_id UUID NOT NULL REFERENCES provinces(id) ON DELETE CASCADE,
    code VARCHAR(20) UNIQUE NOT NULL,  -- PSGC code
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) UNIQUE NOT NULL,
    is_city BOOLEAN DEFAULT FALSE,
    is_capital BOOLEAN DEFAULT FALSE,
    is_huc BOOLEAN DEFAULT FALSE,  -- Highly Urbanized City (independent from province)
    is_icc BOOLEAN DEFAULT FALSE,  -- Independent Component City
    population INTEGER,  -- Latest census population
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);

-- Barangays (42,000+ barangays)
CREATE TABLE barangays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    city_municipality_id UUID NOT NULL REFERENCES cities_municipalities(id) ON DELETE CASCADE,
    code VARCHAR(20) UNIQUE NOT NULL,  -- PSGC code
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) UNIQUE NOT NULL,
    population INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);

-- Congressional Districts (for House of Representatives)
CREATE TABLE congressional_districts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    province_id UUID REFERENCES provinces(id) ON DELETE CASCADE,
    city_municipality_id UUID REFERENCES cities_municipalities(id) ON DELETE CASCADE,  -- For lone/HUC districts
    district_number INTEGER NOT NULL,
    name VARCHAR(200) NOT NULL,  -- e.g., "1st District of Cebu"
    slug VARCHAR(200) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL,
    -- Either province_id or city_municipality_id must be set
    CHECK (province_id IS NOT NULL OR city_municipality_id IS NOT NULL)
);

-- Junction table: Districts can cover multiple cities/municipalities
CREATE TABLE district_coverage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    district_id UUID NOT NULL REFERENCES congressional_districts(id) ON DELETE CASCADE,
    city_municipality_id UUID NOT NULL REFERENCES cities_municipalities(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (district_id, city_municipality_id)
);

-- =====================================================
-- INDEXES
-- =====================================================

-- Regions
CREATE INDEX idx_regions_code ON regions(code);
CREATE INDEX idx_regions_slug ON regions(slug);
CREATE INDEX idx_regions_deleted_at ON regions(deleted_at);

-- Provinces
CREATE INDEX idx_provinces_region ON provinces(region_id);
CREATE INDEX idx_provinces_code ON provinces(code);
CREATE INDEX idx_provinces_slug ON provinces(slug);
CREATE INDEX idx_provinces_deleted_at ON provinces(deleted_at);

-- Cities/Municipalities
CREATE INDEX idx_cities_province ON cities_municipalities(province_id);
CREATE INDEX idx_cities_code ON cities_municipalities(code);
CREATE INDEX idx_cities_slug ON cities_municipalities(slug);
CREATE INDEX idx_cities_is_city ON cities_municipalities(is_city);
CREATE INDEX idx_cities_is_huc ON cities_municipalities(is_huc);
CREATE INDEX idx_cities_deleted_at ON cities_municipalities(deleted_at);

-- Barangays
CREATE INDEX idx_barangays_city ON barangays(city_municipality_id);
CREATE INDEX idx_barangays_code ON barangays(code);
CREATE INDEX idx_barangays_slug ON barangays(slug);
CREATE INDEX idx_barangays_deleted_at ON barangays(deleted_at);

-- Congressional Districts
CREATE INDEX idx_districts_province ON congressional_districts(province_id);
CREATE INDEX idx_districts_city ON congressional_districts(city_municipality_id);
CREATE INDEX idx_districts_slug ON congressional_districts(slug);
CREATE INDEX idx_districts_deleted_at ON congressional_districts(deleted_at);

-- District Coverage
CREATE INDEX idx_district_coverage_district ON district_coverage(district_id);
CREATE INDEX idx_district_coverage_city ON district_coverage(city_municipality_id);

-- =====================================================
-- TRIGGERS
-- =====================================================

-- Updated_at triggers
CREATE TRIGGER update_regions_updated_at BEFORE UPDATE ON regions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_provinces_updated_at BEFORE UPDATE ON provinces
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cities_updated_at BEFORE UPDATE ON cities_municipalities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_barangays_updated_at BEFORE UPDATE ON barangays
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_districts_updated_at BEFORE UPDATE ON congressional_districts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- SEED DATA: Philippine Regions
-- =====================================================

INSERT INTO regions (code, name, slug) VALUES
('010000000', 'Ilocos Region', 'region-i'),
('020000000', 'Cagayan Valley', 'region-ii'),
('030000000', 'Central Luzon', 'region-iii'),
('040000000', 'CALABARZON', 'region-iv-a'),
('170000000', 'MIMAROPA', 'region-iv-b'),
('050000000', 'Bicol Region', 'region-v'),
('060000000', 'Western Visayas', 'region-vi'),
('070000000', 'Central Visayas', 'region-vii'),
('080000000', 'Eastern Visayas', 'region-viii'),
('090000000', 'Zamboanga Peninsula', 'region-ix'),
('100000000', 'Northern Mindanao', 'region-x'),
('110000000', 'Davao Region', 'region-xi'),
('120000000', 'SOCCSKSARGEN', 'region-xii'),
('130000000', 'National Capital Region', 'ncr'),
('140000000', 'Cordillera Administrative Region', 'car'),
('160000000', 'Caraga', 'region-xiii'),
('190000000', 'Bangsamoro Autonomous Region in Muslim Mindanao', 'barmm');
