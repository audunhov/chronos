-- Add subject and body_html to email_outbox for ad-hoc emails
ALTER TABLE email_outbox ADD COLUMN subject VARCHAR(255);
ALTER TABLE email_outbox ADD COLUMN body_html TEXT;
