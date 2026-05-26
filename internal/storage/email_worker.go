package storage

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"time"
)

type EmailWorker struct {
	db *sql.DB
}

func NewEmailWorker(db *sql.DB) *EmailWorker {
	return &EmailWorker{db: db}
}

func (w *EmailWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.processOutbox(ctx); err != nil {
				log.Printf("EmailWorker error: %v", err)
			}
		}
	}
}

func (w *EmailWorker) processOutbox(ctx context.Context) error {
	rows, err := w.db.QueryContext(ctx, `
		SELECT 
			o.id, 
			o.recipient_email, 
			o.context, 
			COALESCE(t.subject, o.subject) as subject, 
			COALESCE(t.body_html, o.body_html) as body_tpl
		FROM email_outbox o
		LEFT JOIN email_templates t ON o.template_id = t.id
		WHERE o.status = 'PENDING'
		FOR UPDATE OF o SKIP LOCKED
		LIMIT 10`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, recipient, contextJSON string
		var subject, bodyTpl sql.NullString
		if err := rows.Scan(&id, &recipient, &contextJSON, &subject, &bodyTpl); err != nil {
			continue
		}

		if !subject.Valid || !bodyTpl.Valid {
			w.markFailed(ctx, id, "Missing subject or body")
			continue
		}

		// 1. Parse Context
		var templateData map[string]any
		_ = json.Unmarshal([]byte(contextJSON), &templateData)

		// 2. Render Template (even ad-hoc emails can have placeholders if context is provided)
		body, err := renderTemplate(bodyTpl.String, templateData)
		if err != nil {
			w.markFailed(ctx, id, "Template error: "+err.Error())
			continue
		}

		// 3. Send via SMTP
		if err := sendEmail(recipient, subject.String, body); err != nil {
			w.markFailed(ctx, id, "SMTP error: "+err.Error())
			continue
		}

		// 4. Mark Success
		w.markSent(ctx, id)
	}

	return nil
}

func (w *EmailWorker) markSent(ctx context.Context, id string) {
	_, _ = w.db.ExecContext(ctx, "UPDATE email_outbox SET status = 'SENT', processed_at = NOW() WHERE id = $1", id)
}

func (w *EmailWorker) markFailed(ctx context.Context, id string, err string) {
	_, _ = w.db.ExecContext(ctx, "UPDATE email_outbox SET status = 'FAILED', error_message = $1, processed_at = NOW() WHERE id = $2", err, id)
}

func renderTemplate(tplStr string, data any) (string, error) {
	tmpl, err := template.New("email").Parse(tplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func sendEmail(to string, subject string, body string) error {
	host := os.Getenv("SMTP_HOST")
	if host == "" { host = "mailpit" }
	port := os.Getenv("SMTP_PORT")
	if port == "" { port = "1025" }

	msg := "From: Chronos <noreply@chronos.no>\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n" +
		"MIME-version: 1.0;\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\n\n" +
		body

	return smtp.SendMail(host+":"+port, nil, "noreply@chronos.no", []string{to}, []byte(msg))
}
