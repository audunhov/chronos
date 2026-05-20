-- 1. Legg til policy på organisasjoner
ALTER TABLE organization_hierarchy ADD COLUMN IF NOT EXISTS policy JSONB DEFAULT '{"allow_multiple": false}'::jsonb;

-- 2. Gi nytt navn til member_view -> membership_view
-- Vi sletter og gjenoppretter den for å sikre riktig struktur for det nye designet
DROP TABLE IF EXISTS member_view;

CREATE TABLE IF NOT EXISTS membership_view (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    org_id UUID NOT NULL,
    user_name VARCHAR(255) NOT NULL,
    user_email VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    role VARCHAR(50) DEFAULT 'member',
    balance INTEGER DEFAULT 0,
    fee_formula VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_membership_user ON membership_view(user_id);
CREATE INDEX IF NOT EXISTS idx_membership_org ON membership_view(org_id);

-- 3. Oppdater users-tabellen
-- Vi fjerner org_id og role herfra etter hvert, men lar dem stå inntil videre 
-- for å ikke brekke eksisterende kode før Phase 2.
-- Men vi legger til et navn-felt hvis det manglet (sjekker 000001_init.up.sql)
ALTER TABLE users ADD COLUMN IF NOT EXISTS name VARCHAR(255);
