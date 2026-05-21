-- Add last_run_at to event_reactions for timed schedules
ALTER TABLE event_reactions ADD COLUMN last_run_at TIMESTAMP WITH TIME ZONE;
