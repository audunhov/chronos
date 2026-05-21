DROP TABLE IF EXISTS audit_logs;

ALTER TABLE event_store DROP COLUMN IF EXISTS correlation_id;
