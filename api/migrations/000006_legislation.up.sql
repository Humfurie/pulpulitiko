-- Legislation/Bills Tracker Migration
-- Comprehensive tracking of Philippine legislative bills and voting records

-- Bill Status enum type
CREATE TYPE bill_status AS ENUM (
    'filed',
    'pending_committee',
    'in_committee',
    'reported_out',
    'pending_second_reading',
    'approved_second_reading',
    'pending_third_reading',
    'approved_third_reading',
    'transmitted',
    'consolidated',
    'ratified',
    'signed_into_law',
    'vetoed',
    'lapsed',
    'withdrawn',
    'archived'
);

-- Legislative Chamber enum type
CREATE TYPE legislative_chamber AS ENUM (
    'senate',
    'house'
);

-- Vote Type enum type
CREATE TYPE vote_type AS ENUM (
    'yea',
    'nay',
    'abstain',
    'absent'
);

-- Legislative Sessions table (e.g., 19th Congress, 1st Regular Session)
CREATE TABLE legislative_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    congress_number INTEGER NOT NULL,
    session_number INTEGER NOT NULL,
    session_type VARCHAR(50) NOT NULL DEFAULT 'regular', -- regular, special
    start_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(congress_number, session_number, session_type)
);

CREATE INDEX idx_legislative_sessions_current ON legislative_sessions(is_current) WHERE is_current = TRUE;
CREATE INDEX idx_legislative_sessions_congress ON legislative_sessions(congress_number);

-- Committees table
CREATE TABLE committees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chamber legislative_chamber NOT NULL,
    name VARCHAR(300) NOT NULL,
    slug VARCHAR(300) UNIQUE NOT NULL,
    description TEXT,
    chairperson_id UUID REFERENCES politicians(id),
    vice_chairperson_id UUID REFERENCES politicians(id),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_committees_chamber ON committees(chamber) WHERE deleted_at IS NULL;
CREATE INDEX idx_committees_slug ON committees(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_committees_is_active ON committees(is_active) WHERE deleted_at IS NULL;

-- Bills table
CREATE TABLE bills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES legislative_sessions(id),
    chamber legislative_chamber NOT NULL,
    bill_number VARCHAR(50) NOT NULL, -- e.g., "SB 1234" or "HB 5678"
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) UNIQUE NOT NULL,
    short_title VARCHAR(200),
    summary TEXT,
    full_text TEXT,
    significance VARCHAR(100), -- local, national, urgent
    status bill_status NOT NULL DEFAULT 'filed',
    filed_date DATE NOT NULL,
    last_action_date DATE,
    date_signed DATE,
    republic_act_number VARCHAR(50), -- If signed into law
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP DEFAULT NULL,
    UNIQUE(session_id, chamber, bill_number)
);

CREATE INDEX idx_bills_session_id ON bills(session_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_bills_chamber ON bills(chamber) WHERE deleted_at IS NULL;
CREATE INDEX idx_bills_status ON bills(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_bills_filed_date ON bills(filed_date DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_bills_slug ON bills(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_bills_last_action_date ON bills(last_action_date DESC) WHERE deleted_at IS NULL;

-- Bill Authors table (principal authors and co-authors)
CREATE TABLE bill_authors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bill_id UUID NOT NULL REFERENCES bills(id) ON DELETE CASCADE,
    politician_id UUID NOT NULL REFERENCES politicians(id),
    is_principal_author BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(bill_id, politician_id)
);

CREATE INDEX idx_bill_authors_bill_id ON bill_authors(bill_id);
CREATE INDEX idx_bill_authors_politician_id ON bill_authors(politician_id);
CREATE INDEX idx_bill_authors_principal ON bill_authors(is_principal_author) WHERE is_principal_author = TRUE;

-- Bill Status History table (timeline of status changes)
CREATE TABLE bill_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bill_id UUID NOT NULL REFERENCES bills(id) ON DELETE CASCADE,
    status bill_status NOT NULL,
    action_description TEXT,
    action_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_bill_status_history_bill_id ON bill_status_history(bill_id);
CREATE INDEX idx_bill_status_history_date ON bill_status_history(action_date DESC);

-- Bill Committees table (committee referrals)
CREATE TABLE bill_committees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bill_id UUID NOT NULL REFERENCES bills(id) ON DELETE CASCADE,
    committee_id UUID NOT NULL REFERENCES committees(id),
    referred_date DATE NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE,
    status VARCHAR(50) DEFAULT 'pending', -- pending, approved, disapproved
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(bill_id, committee_id)
);

CREATE INDEX idx_bill_committees_bill_id ON bill_committees(bill_id);
CREATE INDEX idx_bill_committees_committee_id ON bill_committees(committee_id);

-- Bill Votes table (voting sessions per bill)
CREATE TABLE bill_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bill_id UUID NOT NULL REFERENCES bills(id) ON DELETE CASCADE,
    chamber legislative_chamber NOT NULL,
    reading VARCHAR(20) NOT NULL, -- second, third
    vote_date DATE NOT NULL,
    yeas INTEGER NOT NULL DEFAULT 0,
    nays INTEGER NOT NULL DEFAULT 0,
    abstentions INTEGER NOT NULL DEFAULT 0,
    absent INTEGER NOT NULL DEFAULT 0,
    is_passed BOOLEAN NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_bill_votes_bill_id ON bill_votes(bill_id);
CREATE INDEX idx_bill_votes_vote_date ON bill_votes(vote_date DESC);

-- Politician Votes table (individual voting records)
CREATE TABLE politician_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bill_vote_id UUID NOT NULL REFERENCES bill_votes(id) ON DELETE CASCADE,
    politician_id UUID NOT NULL REFERENCES politicians(id),
    vote vote_type NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(bill_vote_id, politician_id)
);

CREATE INDEX idx_politician_votes_bill_vote_id ON politician_votes(bill_vote_id);
CREATE INDEX idx_politician_votes_politician_id ON politician_votes(politician_id);
CREATE INDEX idx_politician_votes_vote ON politician_votes(vote);

-- Bill Topics/Tags table (for categorization)
CREATE TABLE bill_topics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_bill_topics_slug ON bill_topics(slug);

-- Bill-Topic junction table
CREATE TABLE bill_topic_assignments (
    bill_id UUID NOT NULL REFERENCES bills(id) ON DELETE CASCADE,
    topic_id UUID NOT NULL REFERENCES bill_topics(id) ON DELETE CASCADE,
    PRIMARY KEY (bill_id, topic_id)
);

CREATE INDEX idx_bill_topic_assignments_topic_id ON bill_topic_assignments(topic_id);

-- Triggers for updated_at
CREATE TRIGGER update_legislative_sessions_updated_at
    BEFORE UPDATE ON legislative_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_committees_updated_at
    BEFORE UPDATE ON committees
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_bills_updated_at
    BEFORE UPDATE ON bills
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Seed current legislative session
INSERT INTO legislative_sessions (congress_number, session_number, session_type, start_date, is_current) VALUES
(19, 1, 'regular', '2022-07-25', FALSE),
(19, 2, 'regular', '2023-07-24', FALSE),
(19, 3, 'regular', '2024-07-22', TRUE);

-- Seed common bill topics
INSERT INTO bill_topics (name, slug, description) VALUES
('Agriculture', 'agriculture', 'Bills related to farming, fisheries, and food production'),
('Education', 'education', 'Bills related to schools, universities, and educational policies'),
('Health', 'health', 'Bills related to healthcare, hospitals, and public health'),
('Economy', 'economy', 'Bills related to economic policies, trade, and commerce'),
('Environment', 'environment', 'Bills related to environmental protection and natural resources'),
('Infrastructure', 'infrastructure', 'Bills related to roads, bridges, and public works'),
('Labor', 'labor', 'Bills related to workers rights and employment'),
('Justice', 'justice', 'Bills related to the judiciary and legal reforms'),
('Defense', 'defense', 'Bills related to national defense and security'),
('Social Welfare', 'social-welfare', 'Bills related to social services and welfare programs'),
('Governance', 'governance', 'Bills related to government structure and administration'),
('Finance', 'finance', 'Bills related to taxation, budget, and financial policies'),
('Technology', 'technology', 'Bills related to IT, telecommunications, and digital policies'),
('Transportation', 'transportation', 'Bills related to public transport and mobility'),
('Housing', 'housing', 'Bills related to housing and urban development'),
('Human Rights', 'human-rights', 'Bills related to civil liberties and human rights'),
('Local Government', 'local-government', 'Bills related to LGU powers and local governance'),
('Indigenous Peoples', 'indigenous-peoples', 'Bills related to IP rights and ancestral domains'),
('Women and Children', 'women-and-children', 'Bills related to gender equality and child protection'),
('Senior Citizens', 'senior-citizens', 'Bills related to elderly welfare and benefits');

-- Seed major Senate committees
INSERT INTO committees (chamber, name, slug, description) VALUES
('senate', 'Committee on Finance', 'senate-finance', 'Oversees matters relating to appropriations, taxation, and financial policies'),
('senate', 'Committee on Justice and Human Rights', 'senate-justice-human-rights', 'Handles justice system reforms and human rights issues'),
('senate', 'Committee on Education, Arts and Culture', 'senate-education', 'Oversees educational policies and cultural affairs'),
('senate', 'Committee on Health and Demography', 'senate-health', 'Handles healthcare policies and population concerns'),
('senate', 'Committee on Public Order and Dangerous Drugs', 'senate-public-order', 'Oversees law enforcement and anti-drug policies'),
('senate', 'Committee on National Defense and Security', 'senate-defense', 'Handles matters of national defense'),
('senate', 'Committee on Ways and Means', 'senate-ways-means', 'Oversees revenue measures and tax policies'),
('senate', 'Committee on Agriculture, Food and Agrarian Reform', 'senate-agriculture', 'Handles agricultural policies and agrarian reform'),
('senate', 'Committee on Environment, Natural Resources and Climate Change', 'senate-environment', 'Oversees environmental protection policies'),
('senate', 'Committee on Labor, Employment and Human Resources Development', 'senate-labor', 'Handles labor policies and employment issues');

-- Seed major House committees
INSERT INTO committees (chamber, name, slug, description) VALUES
('house', 'Committee on Appropriations', 'house-appropriations', 'Handles the national budget and appropriations'),
('house', 'Committee on Ways and Means', 'house-ways-means', 'Oversees revenue and taxation measures'),
('house', 'Committee on Justice', 'house-justice', 'Handles judicial reforms and legal matters'),
('house', 'Committee on Education', 'house-education', 'Oversees educational policies'),
('house', 'Committee on Health', 'house-health', 'Handles healthcare policies'),
('house', 'Committee on Public Order and Safety', 'house-public-order', 'Oversees law enforcement matters'),
('house', 'Committee on National Defense and Security', 'house-defense', 'Handles national defense matters'),
('house', 'Committee on Agriculture and Food', 'house-agriculture', 'Oversees agricultural policies'),
('house', 'Committee on Natural Resources', 'house-natural-resources', 'Handles environmental and natural resources'),
('house', 'Committee on Labor and Employment', 'house-labor', 'Oversees labor and employment policies');
