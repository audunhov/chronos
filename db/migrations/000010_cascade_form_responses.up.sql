-- Add ON DELETE CASCADE to form_responses and any other missing references

ALTER TABLE form_responses DROP CONSTRAINT form_responses_form_id_fkey;
ALTER TABLE form_responses ADD CONSTRAINT form_responses_form_id_fkey FOREIGN KEY (form_id) REFERENCES forms(id) ON DELETE CASCADE;

ALTER TABLE form_responses DROP CONSTRAINT form_responses_user_id_fkey;
ALTER TABLE form_responses ADD CONSTRAINT form_responses_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
