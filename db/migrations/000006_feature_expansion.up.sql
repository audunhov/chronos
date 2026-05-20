-- Organs and Role Types
CREATE TABLE IF NOT EXISTS role_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organization_hierarchy(id),
    name VARCHAR(255) NOT NULL,
    permissions JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS organs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organization_hierarchy(id),
    name VARCHAR(255) NOT NULL,
    parent_organ_id UUID REFERENCES organs(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Email Outbox and Templates
CREATE TABLE IF NOT EXISTS email_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID REFERENCES organization_hierarchy(id), -- NULL means system-wide
    name VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    body_html TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS email_outbox (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING', -- PENDING, SENT, FAILED
    recipient_email VARCHAR(255) NOT NULL,
    template_id UUID REFERENCES email_templates(id),
    context JSONB NOT NULL DEFAULT '{}',
    error_message TEXT,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Payment Pipelines and Invoices
CREATE TABLE IF NOT EXISTS payment_pipelines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organization_hierarchy(id),
    sequence JSONB NOT NULL, -- e.g. ["card", "vipps", "email"]
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Invoices are event-sourced, but we need a view
CREATE TABLE IF NOT EXISTS invoice_view (
    id UUID PRIMARY KEY,
    membership_id UUID NOT NULL REFERENCES membership_view(id),
    amount INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL, -- UNPAID, ATTEMPTING_PAYMENT, PAID, CANCELLED
    current_pipeline_step INTEGER DEFAULT 0,
    due_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Customizable Forms
CREATE TABLE IF NOT EXISTS forms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organization_hierarchy(id),
    title VARCHAR(255) NOT NULL,
    schema JSONB NOT NULL, -- JSON Schema for the form
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS form_responses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    form_id UUID NOT NULL REFERENCES forms(id),
    user_id UUID NOT NULL REFERENCES users(id),
    answers JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Event Reactions (Pipelines)
CREATE TABLE IF NOT EXISTS event_reactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organization_hierarchy(id),
    trigger_event VARCHAR(255) NOT NULL,
    action_type VARCHAR(50) NOT NULL, -- e.g. 'SEND_EMAIL', 'WEBHOOK'
    config JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
