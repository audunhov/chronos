-- Legg til trigger_aggregate_id for mer spesifikke pipelines
ALTER TABLE event_reactions ADD COLUMN IF NOT EXISTS trigger_aggregate_id UUID;
CREATE INDEX IF NOT EXISTS idx_reaction_aggregate ON event_reactions(trigger_aggregate_id);
