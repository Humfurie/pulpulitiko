-- Migration: 000014_politician_import_logs
-- Creates table for tracking Excel import operations

CREATE TABLE IF NOT EXISTS politician_import_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filename VARCHAR(500) NOT NULL,
    uploaded_by UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Statistics
    total_rows INTEGER DEFAULT 0,
    successful_imports INTEGER DEFAULT 0,
    failed_imports INTEGER DEFAULT 0,
    politicians_created INTEGER DEFAULT 0,
    politicians_updated INTEGER DEFAULT 0,
    positions_archived INTEGER DEFAULT 0,

    -- Status: 'pending', 'processing', 'completed', 'failed'
    status VARCHAR(20) DEFAULT 'pending' NOT NULL,
    error_log TEXT,
    validation_errors JSONB DEFAULT '[]',

    -- Optional election linkage
    election_id UUID REFERENCES election_events(id) ON DELETE SET NULL,

    -- Timestamps
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for performance (idempotent)
CREATE INDEX IF NOT EXISTS idx_politician_import_logs_uploaded_by ON politician_import_logs(uploaded_by);
CREATE INDEX IF NOT EXISTS idx_politician_import_logs_status ON politician_import_logs(status);
CREATE INDEX IF NOT EXISTS idx_politician_import_logs_election_id ON politician_import_logs(election_id) WHERE election_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_politician_import_logs_created_at ON politician_import_logs(created_at DESC);
