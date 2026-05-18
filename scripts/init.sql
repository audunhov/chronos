-- Opprett schemaer hvis de ikke finnes
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS extensions;

-- Opprett roller hvis de ikke finnes
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'anon') THEN
        CREATE ROLE anon NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'authenticated') THEN
        CREATE ROLE authenticated NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'service_role') THEN
        CREATE ROLE service_role NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'supabase_admin') THEN
        CREATE ROLE supabase_admin LOGIN SUPERUSER;
    END IF;
END
$$;

-- Gi tilgang til schemaer
GRANT USAGE ON SCHEMA auth TO anon, authenticated, service_role;
GRANT USAGE ON SCHEMA public TO anon, authenticated, service_role;
GRANT ALL ON ALL TABLES IN SCHEMA public TO postgres, supabase_admin;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO postgres, supabase_admin;

-- En enkel auth.jwt() funksjon for å unngå feil i policies før GoTrue har kjørt migrasjoner
CREATE OR REPLACE FUNCTION auth.jwt() RETURNS jsonb AS $$
  -- Returner en tom JSON hvis den ikke er satt av GoTrue
  SELECT coalesce(current_setting('request.jwt.claims', true), '{}')::jsonb;
$$ LANGUAGE sql STABLE;

CREATE TABLE IF NOT EXISTS event_store (
    id BIGSERIAL PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    version INT NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_event_store_aggregate_id ON event_store(aggregate_id);

-- Aktiver ltree utvidelsen for hierarkiske stier
-- Bruker DO-blokk for å håndtere potensielle rettighetsproblemer i supabase/postgres image
DO $$
BEGIN
    CREATE EXTENSION IF NOT EXISTS ltree SCHEMA extensions;
EXCEPTION
    WHEN OTHERS THEN
        -- Hvis den allerede finnes eller vi ikke har tilgang, ignorer feilen
        -- I supabase/postgres er den ofte allerede tilgjengelig
        NULL;
END
$$;

-- Sørg for at extensions er i search_path
ALTER DATABASE postgres SET search_path TO "$user", public, extensions, auth;
SET search_path TO "$user", public, extensions, auth;

-- Tabell for organisasjonsstruktur
CREATE TABLE IF NOT EXISTS organization_hierarchy (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    parent_id UUID REFERENCES organization_hierarchy(id),
    path ltree NOT NULL, -- Lagrer stien, f.eks. "Hovedkontor.FylkeA.Lokallag1"
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_org_path ON organization_hierarchy USING GIST (path);

-- Oppdater member_view til å bruke org_id i stedet for team_id (mer generisk)
CREATE TABLE IF NOT EXISTS member_view (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES organization_hierarchy(id),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    metadata JSONB DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Aktiver RLS
ALTER TABLE member_view ENABLE ROW LEVEL SECURITY;

-- Policy: Hierarkisk tilgang
-- En bruker kan se medlemmer hvis de er i samme org, 
-- ELLER hvis brukeren er leder i en overliggende org (ancestor i path).
CREATE POLICY hierarchical_access_policy ON member_view
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM organization_hierarchy user_org
            JOIN organization_hierarchy member_org ON member_org.id = member_view.org_id
            WHERE user_org.id = ((auth.jwt() ->> 'org_id')::uuid)
            AND (
                -- Brukeren er i samme organisasjon
                user_org.id = member_org.id 
                OR 
                -- Brukeren er i en overliggende organisasjon (f.eks. fylke over lokallag)
                -- og har rollen 'admin' eller 'leader'
                (
                    user_org.path @> member_org.path 
                    AND 
                    (auth.jwt() ->> 'role') IN ('admin', 'leader')
                )
            )
        )
    );

-- Tabell for Crypto-shredding nøkler
CREATE TABLE IF NOT EXISTS encryption_keys (
    aggregate_id UUID PRIMARY KEY,
    key_value BYTEA NOT NULL, -- Kryptert eller beskyttet nøkkel
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
