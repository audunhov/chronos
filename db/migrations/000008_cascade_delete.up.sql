-- Add ON DELETE CASCADE to organization references

-- Forms
ALTER TABLE forms DROP CONSTRAINT forms_org_id_fkey;
ALTER TABLE forms ADD CONSTRAINT forms_org_id_fkey FOREIGN KEY (org_id) REFERENCES organization_hierarchy(id) ON DELETE CASCADE;

-- Role Assignments
ALTER TABLE role_assignments DROP CONSTRAINT role_assignments_org_id_fkey;
ALTER TABLE role_assignments ADD CONSTRAINT role_assignments_org_id_fkey FOREIGN KEY (org_id) REFERENCES organization_hierarchy(id) ON DELETE CASCADE;

-- Role Types
ALTER TABLE role_types DROP CONSTRAINT role_types_org_id_fkey;
ALTER TABLE role_types ADD CONSTRAINT role_types_org_id_fkey FOREIGN KEY (org_id) REFERENCES organization_hierarchy(id) ON DELETE CASCADE;

-- Organs
ALTER TABLE organs DROP CONSTRAINT organs_org_id_fkey;
ALTER TABLE organs ADD CONSTRAINT organs_org_id_fkey FOREIGN KEY (org_id) REFERENCES organization_hierarchy(id) ON DELETE CASCADE;

-- Email Templates
ALTER TABLE email_templates DROP CONSTRAINT email_templates_org_id_fkey;
ALTER TABLE email_templates ADD CONSTRAINT email_templates_org_id_fkey FOREIGN KEY (org_id) REFERENCES organization_hierarchy(id) ON DELETE CASCADE;

-- Payment Pipelines
ALTER TABLE payment_pipelines DROP CONSTRAINT payment_pipelines_org_id_fkey;
ALTER TABLE payment_pipelines ADD CONSTRAINT payment_pipelines_org_id_fkey FOREIGN KEY (org_id) REFERENCES organization_hierarchy(id) ON DELETE CASCADE;

-- Event Reactions
ALTER TABLE event_reactions DROP CONSTRAINT event_reactions_org_id_fkey;
ALTER TABLE event_reactions ADD CONSTRAINT event_reactions_org_id_fkey FOREIGN KEY (org_id) REFERENCES organization_hierarchy(id) ON DELETE CASCADE;

-- Self-reference for Hierarchy (Child orgs)
ALTER TABLE organization_hierarchy DROP CONSTRAINT organization_hierarchy_parent_id_fkey;
ALTER TABLE organization_hierarchy ADD CONSTRAINT organization_hierarchy_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES organization_hierarchy(id) ON DELETE CASCADE;
