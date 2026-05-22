CREATE TABLE IF NOT EXISTS pipeline_executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pipeline_id UUID NOT NULL REFERENCES event_reactions(id) ON DELETE CASCADE,
    trigger_event VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL, -- SUCCESS, FAILED
    logs JSONB NOT NULL DEFAULT '[]',
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pipeline_executions_pipeline ON pipeline_executions(pipeline_id);
