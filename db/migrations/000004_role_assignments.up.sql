-- Tabell for Rolletildelinger (Staff, Styre, etc.)
CREATE TABLE IF NOT EXISTS role_assignments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    org_id UUID NOT NULL REFERENCES organization_hierarchy(id),
    role_type VARCHAR(50) NOT NULL, -- f.eks. 'leader', 'secretary', 'board_member'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_role_assignment_user ON role_assignments(user_id);
CREATE INDEX IF NOT EXISTS idx_role_assignment_org ON role_assignments(org_id);
