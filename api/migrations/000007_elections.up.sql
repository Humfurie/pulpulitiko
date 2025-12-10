-- Elections System Migration
-- Track Philippine elections, candidates, and results

-- Election Type enum
CREATE TYPE election_type AS ENUM (
    'national',        -- Presidential, VP, Senators, Party-list
    'local',           -- Governor, Mayor, Councilors, etc.
    'barangay',        -- Barangay and SK elections
    'special',         -- Special/by-elections
    'plebiscite',      -- Plebiscites and referendums
    'recall'           -- Recall elections
);

-- Election Status enum
CREATE TYPE election_status AS ENUM (
    'upcoming',
    'ongoing',
    'completed',
    'cancelled'
);

-- Candidate Status enum
CREATE TYPE candidate_status AS ENUM (
    'filed',
    'qualified',
    'disqualified',
    'withdrawn',
    'substituted'
);

-- Elections table
CREATE TABLE elections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(300) NOT NULL,
    slug VARCHAR(300) UNIQUE NOT NULL,
    election_type election_type NOT NULL,
    description TEXT,
    election_date DATE NOT NULL,
    registration_start DATE,
    registration_end DATE,
    campaign_start DATE,
    campaign_end DATE,
    status election_status NOT NULL DEFAULT 'upcoming',
    is_featured BOOLEAN DEFAULT FALSE,
    voter_turnout_percentage DECIMAL(5,2),
    total_registered_voters INTEGER,
    total_votes_cast INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_elections_type ON elections(election_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_elections_status ON elections(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_elections_date ON elections(election_date DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_elections_slug ON elections(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_elections_featured ON elections(is_featured) WHERE is_featured = TRUE AND deleted_at IS NULL;

-- Election Positions table (positions being contested in an election)
CREATE TABLE election_positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    election_id UUID NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
    position_id UUID NOT NULL REFERENCES government_positions(id),
    -- Jurisdiction scope (NULL means national)
    region_id UUID REFERENCES regions(id),
    province_id UUID REFERENCES provinces(id),
    city_municipality_id UUID REFERENCES cities_municipalities(id),
    barangay_id UUID REFERENCES barangays(id),
    district_id UUID REFERENCES congressional_districts(id),
    -- Position details
    seats_available INTEGER NOT NULL DEFAULT 1,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(election_id, position_id, region_id, province_id, city_municipality_id, barangay_id, district_id)
);

CREATE INDEX idx_election_positions_election ON election_positions(election_id);
CREATE INDEX idx_election_positions_position ON election_positions(position_id);

-- Candidates table
CREATE TABLE candidates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    election_position_id UUID NOT NULL REFERENCES election_positions(id) ON DELETE CASCADE,
    politician_id UUID NOT NULL REFERENCES politicians(id),
    party_id UUID REFERENCES political_parties(id),
    ballot_number INTEGER,
    ballot_name VARCHAR(200),
    campaign_slogan VARCHAR(500),
    platform TEXT,
    status candidate_status NOT NULL DEFAULT 'filed',
    filing_date DATE,
    is_incumbent BOOLEAN DEFAULT FALSE,
    is_winner BOOLEAN DEFAULT FALSE,
    votes_received INTEGER,
    vote_percentage DECIMAL(5,2),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(election_position_id, politician_id)
);

CREATE INDEX idx_candidates_election_position ON candidates(election_position_id);
CREATE INDEX idx_candidates_politician ON candidates(politician_id);
CREATE INDEX idx_candidates_party ON candidates(party_id);
CREATE INDEX idx_candidates_status ON candidates(status);
CREATE INDEX idx_candidates_winner ON candidates(is_winner) WHERE is_winner = TRUE;

-- Election Results table (aggregate results per position)
CREATE TABLE election_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    election_position_id UUID NOT NULL REFERENCES election_positions(id) ON DELETE CASCADE,
    total_votes INTEGER NOT NULL DEFAULT 0,
    valid_votes INTEGER NOT NULL DEFAULT 0,
    invalid_votes INTEGER NOT NULL DEFAULT 0,
    registered_voters INTEGER,
    turnout_percentage DECIMAL(5,2),
    is_final BOOLEAN DEFAULT FALSE,
    last_updated TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_election_results_position ON election_results(election_position_id);

-- Precinct Results table (optional granular results)
CREATE TABLE precinct_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    candidate_id UUID NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    precinct_id VARCHAR(50) NOT NULL,
    precinct_name VARCHAR(200),
    barangay_id UUID REFERENCES barangays(id),
    votes INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(candidate_id, precinct_id)
);

CREATE INDEX idx_precinct_results_candidate ON precinct_results(candidate_id);
CREATE INDEX idx_precinct_results_precinct ON precinct_results(precinct_id);
CREATE INDEX idx_precinct_results_barangay ON precinct_results(barangay_id);

-- Voter Education content
CREATE TABLE voter_education (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    election_id UUID REFERENCES elections(id) ON DELETE SET NULL,
    title VARCHAR(300) NOT NULL,
    slug VARCHAR(300) UNIQUE NOT NULL,
    content TEXT NOT NULL,
    content_type VARCHAR(50) NOT NULL DEFAULT 'article', -- article, faq, guide, video
    category VARCHAR(100), -- registration, voting_day, requirements, etc.
    is_featured BOOLEAN DEFAULT FALSE,
    is_published BOOLEAN DEFAULT FALSE,
    published_at TIMESTAMP,
    view_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_voter_education_election ON voter_education(election_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_voter_education_slug ON voter_education(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_voter_education_published ON voter_education(is_published) WHERE is_published = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_voter_education_category ON voter_education(category) WHERE deleted_at IS NULL;

-- Triggers
CREATE TRIGGER update_elections_updated_at
    BEFORE UPDATE ON elections
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_candidates_updated_at
    BEFORE UPDATE ON candidates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_voter_education_updated_at
    BEFORE UPDATE ON voter_education
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Seed upcoming/recent elections
INSERT INTO elections (name, slug, election_type, description, election_date, status, is_featured) VALUES
('2025 Midterm Elections', '2025-midterm-elections', 'national', 'Philippine Midterm Elections for Senators, Representatives, and Local Officials', '2025-05-12', 'upcoming', TRUE),
('2025 Barangay and SK Elections', '2025-barangay-sk-elections', 'barangay', 'Synchronized Barangay and Sangguniang Kabataan Elections', '2025-10-27', 'upcoming', TRUE),
('2022 National and Local Elections', '2022-national-local-elections', 'national', 'Philippine National and Local Elections', '2022-05-09', 'completed', FALSE);

-- Seed voter education content
INSERT INTO voter_education (title, slug, content, content_type, category, is_featured, is_published, published_at) VALUES
('How to Register as a Voter', 'how-to-register-voter', 'Complete guide on voter registration in the Philippines...', 'guide', 'registration', TRUE, TRUE, NOW()),
('Election Day Guide', 'election-day-guide', 'What to expect on election day and how to cast your vote...', 'guide', 'voting_day', TRUE, TRUE, NOW()),
('Valid IDs for Voting', 'valid-ids-voting', 'List of valid identification documents accepted at polling precincts...', 'article', 'requirements', TRUE, TRUE, NOW()),
('Understanding the Ballot', 'understanding-ballot', 'How to properly fill out your ballot and avoid invalid votes...', 'guide', 'voting_day', FALSE, TRUE, NOW()),
('Overseas Voting Guide', 'overseas-voting-guide', 'How to vote if you are a Filipino citizen living or working abroad...', 'guide', 'registration', FALSE, TRUE, NOW());
