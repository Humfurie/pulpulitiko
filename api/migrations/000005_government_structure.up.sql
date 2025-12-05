-- Government Structure Migration
-- Extends the politician model for comprehensive Philippine government coverage

-- Government Level enum type
CREATE TYPE government_level AS ENUM (
    'national',
    'regional',
    'provincial',
    'city',
    'municipal',
    'barangay'
);

-- Government Branch enum type
CREATE TYPE government_branch AS ENUM (
    'executive',
    'legislative',
    'judicial'
);

-- Political Parties table
CREATE TABLE political_parties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) UNIQUE NOT NULL,
    abbreviation VARCHAR(50),
    logo VARCHAR(500),
    color VARCHAR(20), -- Hex color code
    description TEXT,
    founded_year INTEGER,
    website VARCHAR(500),
    is_major BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_political_parties_slug ON political_parties(slug);
CREATE INDEX idx_political_parties_is_active ON political_parties(is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_political_parties_is_major ON political_parties(is_major) WHERE deleted_at IS NULL AND is_active = TRUE;

-- Government Positions table (normalized position types)
CREATE TABLE government_positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) UNIQUE NOT NULL,
    level government_level NOT NULL,
    branch government_branch NOT NULL,
    display_order INTEGER DEFAULT 0,
    description TEXT,
    max_terms INTEGER, -- NULL means unlimited
    term_years INTEGER DEFAULT 3,
    is_elected BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_government_positions_level ON government_positions(level);
CREATE INDEX idx_government_positions_branch ON government_positions(branch);
CREATE INDEX idx_government_positions_level_branch ON government_positions(level, branch);
CREATE INDEX idx_government_positions_display_order ON government_positions(display_order);

-- Add new columns to politicians table
ALTER TABLE politicians
ADD COLUMN level government_level DEFAULT 'national',
ADD COLUMN branch government_branch DEFAULT 'legislative',
ADD COLUMN position_id UUID REFERENCES government_positions(id),
ADD COLUMN party_id UUID REFERENCES political_parties(id),
ADD COLUMN district_id UUID REFERENCES congressional_districts(id);

CREATE INDEX idx_politicians_level ON politicians(level) WHERE deleted_at IS NULL;
CREATE INDEX idx_politicians_branch ON politicians(branch) WHERE deleted_at IS NULL;
CREATE INDEX idx_politicians_position_id ON politicians(position_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_politicians_party_id ON politicians(party_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_politicians_district_id ON politicians(district_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_politicians_level_branch ON politicians(level, branch) WHERE deleted_at IS NULL;

-- Politician Jurisdictions table (maps politicians to their jurisdiction locations)
-- A politician can have multiple jurisdictions (e.g., a senator represents the whole country)
CREATE TABLE politician_jurisdictions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    politician_id UUID NOT NULL REFERENCES politicians(id) ON DELETE CASCADE,
    -- Only one of these should be set based on jurisdiction type
    region_id UUID REFERENCES regions(id),
    province_id UUID REFERENCES provinces(id),
    city_id UUID REFERENCES cities_municipalities(id),
    barangay_id UUID REFERENCES barangays(id),
    -- National-level politicians have no location reference (represents entire country)
    is_national BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT check_jurisdiction_type CHECK (
        (is_national = TRUE AND region_id IS NULL AND province_id IS NULL AND city_id IS NULL AND barangay_id IS NULL) OR
        (is_national = FALSE AND (
            (region_id IS NOT NULL AND province_id IS NULL AND city_id IS NULL AND barangay_id IS NULL) OR
            (region_id IS NULL AND province_id IS NOT NULL AND city_id IS NULL AND barangay_id IS NULL) OR
            (region_id IS NULL AND province_id IS NULL AND city_id IS NOT NULL AND barangay_id IS NULL) OR
            (region_id IS NULL AND province_id IS NULL AND city_id IS NULL AND barangay_id IS NOT NULL)
        ))
    )
);

CREATE INDEX idx_politician_jurisdictions_politician_id ON politician_jurisdictions(politician_id);
CREATE INDEX idx_politician_jurisdictions_region_id ON politician_jurisdictions(region_id) WHERE region_id IS NOT NULL;
CREATE INDEX idx_politician_jurisdictions_province_id ON politician_jurisdictions(province_id) WHERE province_id IS NOT NULL;
CREATE INDEX idx_politician_jurisdictions_city_id ON politician_jurisdictions(city_id) WHERE city_id IS NOT NULL;
CREATE INDEX idx_politician_jurisdictions_barangay_id ON politician_jurisdictions(barangay_id) WHERE barangay_id IS NOT NULL;
CREATE INDEX idx_politician_jurisdictions_is_national ON politician_jurisdictions(is_national) WHERE is_national = TRUE;

-- Triggers for updated_at
CREATE TRIGGER update_political_parties_updated_at
    BEFORE UPDATE ON political_parties
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_government_positions_updated_at
    BEFORE UPDATE ON government_positions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Seed Government Positions
INSERT INTO government_positions (name, slug, level, branch, display_order, term_years, max_terms, is_elected, description) VALUES
-- National Executive
('President', 'president', 'national', 'executive', 1, 6, 1, TRUE, 'Head of State and Government of the Republic of the Philippines'),
('Vice President', 'vice-president', 'national', 'executive', 2, 6, 2, TRUE, 'Second highest official in the executive branch'),
('Cabinet Secretary', 'cabinet-secretary', 'national', 'executive', 3, NULL, NULL, FALSE, 'Head of a government department'),
('Undersecretary', 'undersecretary', 'national', 'executive', 4, NULL, NULL, FALSE, 'Deputy head of a government department'),

-- National Legislative
('Senator', 'senator', 'national', 'legislative', 10, 6, 2, TRUE, 'Member of the Philippine Senate'),
('House Representative', 'house-representative', 'national', 'legislative', 11, 3, 3, TRUE, 'Member of the House of Representatives'),
('Party-list Representative', 'party-list-representative', 'national', 'legislative', 12, 3, 3, TRUE, 'Party-list member of the House of Representatives'),

-- National Judicial
('Chief Justice', 'chief-justice', 'national', 'judicial', 20, NULL, NULL, FALSE, 'Head of the Supreme Court'),
('Associate Justice', 'associate-justice', 'national', 'judicial', 21, NULL, NULL, FALSE, 'Member of the Supreme Court'),

-- Provincial
('Governor', 'governor', 'provincial', 'executive', 30, 3, 3, TRUE, 'Chief executive of a province'),
('Vice Governor', 'vice-governor', 'provincial', 'executive', 31, 3, 3, TRUE, 'Second highest provincial official'),
('Provincial Board Member', 'provincial-board-member', 'provincial', 'legislative', 32, 3, 3, TRUE, 'Member of the Sangguniang Panlalawigan'),

-- City
('City Mayor', 'city-mayor', 'city', 'executive', 40, 3, 3, TRUE, 'Chief executive of a city'),
('City Vice Mayor', 'city-vice-mayor', 'city', 'executive', 41, 3, 3, TRUE, 'Second highest city official and presiding officer of the city council'),
('City Councilor', 'city-councilor', 'city', 'legislative', 42, 3, 3, TRUE, 'Member of the Sangguniang Panlungsod'),

-- Municipal
('Municipal Mayor', 'municipal-mayor', 'municipal', 'executive', 50, 3, 3, TRUE, 'Chief executive of a municipality'),
('Municipal Vice Mayor', 'municipal-vice-mayor', 'municipal', 'executive', 51, 3, 3, TRUE, 'Second highest municipal official'),
('Municipal Councilor', 'municipal-councilor', 'municipal', 'legislative', 52, 3, 3, TRUE, 'Member of the Sangguniang Bayan'),

-- Barangay
('Barangay Captain', 'barangay-captain', 'barangay', 'executive', 60, 3, NULL, TRUE, 'Chief executive of a barangay (also called Punong Barangay)'),
('Barangay Kagawad', 'barangay-kagawad', 'barangay', 'legislative', 61, 3, NULL, TRUE, 'Member of the Sangguniang Barangay'),
('SK Chairman', 'sk-chairman', 'barangay', 'executive', 62, 3, NULL, TRUE, 'Head of Sangguniang Kabataan (youth council)'),
('SK Kagawad', 'sk-kagawad', 'barangay', 'legislative', 63, 3, NULL, TRUE, 'Member of the Sangguniang Kabataan');

-- Seed Major Political Parties
INSERT INTO political_parties (name, slug, abbreviation, color, is_major, is_active, description, founded_year) VALUES
('Partido Demokratiko Pilipino-Lakas ng Bayan', 'pdp-laban', 'PDP-Laban', '#0066CC', TRUE, TRUE, 'Originally founded as a democratic socialist party, now a major political party in the Philippines', 1982),
('Nacionalista Party', 'nacionalista-party', 'NP', '#FF0000', TRUE, TRUE, 'The oldest political party in the Philippines and one of the oldest in Asia', 1907),
('Liberal Party', 'liberal-party', 'LP', '#FFCC00', TRUE, TRUE, 'One of the major political parties in the Philippines, center to center-left in orientation', 1946),
('Nationalist People''s Coalition', 'nationalist-peoples-coalition', 'NPC', '#006600', TRUE, TRUE, 'A major political party founded by Eduardo Cojuangco Jr.', 1992),
('Lakas-CMD', 'lakas-cmd', 'Lakas', '#0099FF', TRUE, TRUE, 'Lakas-Christian Muslim Democrats, founded by Fidel V. Ramos', 1991),
('National Unity Party', 'national-unity-party', 'NUP', '#663399', TRUE, TRUE, 'A political party in the Philippines', 2011),
('Aksyon Demokratiko', 'aksyon-demokratiko', 'Aksyon', '#FF6600', TRUE, TRUE, 'A reform-oriented political party', 1997),
('Partido Federal ng Pilipinas', 'partido-federal-ng-pilipinas', 'PFP', '#000080', TRUE, TRUE, 'Political party advocating for federalism', 2018),
('Independent', 'independent', 'IND', '#808080', FALSE, TRUE, 'Independent candidates with no party affiliation', NULL);
