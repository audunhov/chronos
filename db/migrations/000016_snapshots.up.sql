CREATE TABLE IF NOT EXISTS event_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_id UUID NOT NULL,
    version INTEGER NOT NULL,
    state JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_snapshots_aggregate ON event_snapshots(aggregate_id);
CREATE UNIQUE INDEX idx_snapshots_agg_version ON event_snapshots(aggregate_id, version);
