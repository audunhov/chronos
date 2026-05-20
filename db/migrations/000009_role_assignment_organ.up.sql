-- Add organ_id to role_assignments to allow linking roles to specific organs (boards, etc.)
ALTER TABLE role_assignments ADD COLUMN organ_id UUID REFERENCES organs(id) ON DELETE CASCADE;

CREATE INDEX idx_role_assignment_organ ON role_assignments(organ_id);
