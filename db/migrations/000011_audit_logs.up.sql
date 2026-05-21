-- Legg til correlation_id i event_store
ALTER TABLE event_store ADD COLUMN IF NOT EXISTS correlation_id UUID;
CREATE INDEX IF NOT EXISTS idx_event_store_correlation ON event_store(correlation_id);

-- Tabell for Audit Logs
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    correlation_id UUID NOT NULL,
    actor_id UUID REFERENCES users(id), -- Kan være NULL for anonyme/signup handlinger
    org_id UUID REFERENCES organization_hierarchy(id),
    action VARCHAR(255) NOT NULL, -- f.eks 'MEMBERSHIP_REGISTERED', 'REPORT_VIEWED'
    target_id UUID, -- Valgfri ID til objektet som ble påvirket
    detail JSONB NOT NULL DEFAULT '{}', -- Kryptert hvis det inneholder PII
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_correlation ON audit_logs(correlation_id);
CREATE INDEX IF NOT EXISTS idx_audit_actor ON audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_org ON audit_logs(org_id);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action);
