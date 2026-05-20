-- Create forms table
CREATE TABLE IF NOT EXISTS forms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organization_hierarchy(id),
    title VARCHAR(255) NOT NULL,
    schema JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create form_responses table
CREATE TABLE IF NOT EXISTS form_responses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    form_id UUID NOT NULL REFERENCES forms(id),
    user_id UUID NOT NULL REFERENCES users(id),
    answers JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_form_org ON forms(org_id);
CREATE INDEX IF NOT EXISTS idx_response_form ON form_responses(form_id);
CREATE INDEX IF NOT EXISTS idx_response_user ON form_responses(user_id);
