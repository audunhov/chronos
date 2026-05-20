ALTER TABLE users DROP COLUMN IF EXISTS name;

DROP TABLE IF EXISTS membership_view;

-- Gjenopprett gammel member_view hvis nødvendig
CREATE TABLE IF NOT EXISTS member_view (
    id UUID PRIMARY KEY,
    org_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    metadata JSONB DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE organization_hierarchy DROP COLUMN IF EXISTS policy;
