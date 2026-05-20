# System Expansion Plan

This document outlines the architectural and implementation strategy for the massive feature expansion requested.

## 1. Configurable Organs and Role Types
*   **Database:** Add `role_types` (id, org_id, name, permissions JSONB) and `organs` (id, org_id, name, parent_organ_id) tables.
*   **Backend:** CRUD API endpoints for managing these structures per organization.
*   **Frontend:** An "Organs & Roles" management tab in the Admin view.

## 2. Email Sending Flow with Templating & Outbox Pattern
*   **Architecture:** To maintain determinism and decouple I/O, we will use the **Transactional Outbox Pattern**.
*   **Database:** Add `email_templates` (id, name, subject, body_html, org_id) and `email_outbox` (id, status, to, template_id, context JSONB) tables.
*   **Backend:** A Go background worker that polls the `email_outbox` and sends emails via SMTP (to the existing Mailpit container).
*   **Templating:** Use Go's `html/template`.

## 3. Pipeline-Inspired Payment Methods
*   **Domain:** Introduce a new `Invoice` aggregate (event-sourced).
*   **Pipeline Logic:** A state machine driven by a background cron/worker.
    *   State 1: `Attempt_Card` (Stripe/Nets). If failed -> State 2.
    *   State 2: `Attempt_Vipps`. If failed -> State 3.
    *   State 3: `Send_Email_Invoice`.
*   **Database:** `payment_pipelines` table defining the fallback sequence for an organization.

## 4. Customizable Forms (Member Surveys)
*   **Database:** 
    *   `forms` (id, org_id, title, schema JSONB)
    *   `form_responses` (id, form_id, user_id, answers JSONB)
*   **Frontend:** A dynamic form builder/renderer that reads the JSON schema and outputs native HTML inputs, saving responses to the backend.

## 5. Configurable Event Handling Pipelines
*   **Architecture:** A lightweight Rule/Reaction Engine hooked into the Event Store.
*   **Database:** `event_reactions` (id, trigger_event, action_type, config JSONB).
    *   *Example:* `trigger_event: MembershipCreated`, `action_type: send_email`, `config: { role: 'leader', template: 'new_member_alert' }`.
*   **Backend:** The projection worker (or a new Reaction worker) reads new events, matches them against `event_reactions`, and queues the corresponding action (e.g., writing to `email_outbox`).

## 6. Statistics Pages with Graphs
*   **Backend:** New reporting endpoints (e.g., `/api/reports/stats/growth`, `/api/reports/stats/revenue`) that aggregate historical data using SQL over the `event_store` or `membership_view`.
*   **Frontend:** Integrate a lightweight charting library (e.g., Chart.js) to display visual growth curves and fee collection rates in the Admin Dashboard.

## 7. UI/UX Redesign: Tiny Brutalism CSS
*   **Frontend:** Strip out Tailwind CSS.
*   **Integration:** Import `tiny-brutalism-css` (via npm or CDN).
*   **Refactoring:** Rewrite all Vue components (`App.vue`, `LoginForm.vue`, `RegisterForm.vue`) to use semantic HTML and the brutalist aesthetic (bold borders, monospace fonts, `.btn`, `.container` classes).

## 8. Complete Bootstrapping
*   **Implementation:** Create a `cmd/seeder/main.go` script.
*   **Data:** Automatically generate a realistic hierarchy (National -> Regional -> Local), seed 50+ users with diverse statuses, generate fee events spanning several years (so graphs work immediately), and set up the default Admin account.
*   **Docker:** Configure `docker-compose.yml` to run the seeder automatically on fresh starts.
