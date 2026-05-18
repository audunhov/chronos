-- Aktiver nødvendige utvidelser
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "ltree";

-- Tabell for brukere (Autentisering)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    org_id VARCHAR(255), -- Kan være NULL for globale admins
    role VARCHAR(50) DEFAULT 'user',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tabell for hendelseslager (Event Store)
CREATE TABLE IF NOT EXISTS event_store (
    id BIGSERIAL PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    version INT NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_event_store_aggregate_id ON event_store(aggregate_id);

-- Tabell for organisasjonsstruktur
CREATE TABLE IF NOT EXISTS organization_hierarchy (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    parent_id UUID REFERENCES organization_hierarchy(id),
    path ltree NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_org_path ON organization_hierarchy USING GIST (path);

-- Medlemsvisning (Read Model)
CREATE TABLE IF NOT EXISTS member_view (
    id UUID PRIMARY KEY,
    org_id VARCHAR(255) NOT NULL, -- Endret fra UUID til VARCHAR for enkelhets skyld
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    metadata JSONB DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tabell for Crypto-shredding nøkler
CREATE TABLE IF NOT EXISTS encryption_keys (
    aggregate_id UUID PRIMARY KEY,
    key_value BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
