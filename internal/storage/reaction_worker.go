package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"
)

type ReactionWorker struct {
	db *sql.DB
}

func NewReactionWorker(db *sql.DB) *ReactionWorker {
	return &ReactionWorker{db: db}
}

func (w *ReactionWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lastEventID int64
	_ = w.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(id), 0) FROM event_store").Scan(&lastEventID)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.processReactions(ctx, &lastEventID); err != nil {
				log.Printf("ReactionWorker error: %v", err)
			}
		}
	}
}

func (w *ReactionWorker) processReactions(ctx context.Context, lastID *int64) error {
	rows, err := w.db.QueryContext(ctx, `
		SELECT id, aggregate_id, event_type, payload 
		FROM event_store 
		WHERE id > $1 
		ORDER BY id ASC 
		LIMIT 100`, *lastID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var aggregateID, eventType string
		var payload []byte
		if err := rows.Scan(&id, &aggregateID, &eventType, &payload); err != nil {
			continue
		}

		if err := w.handleEvent(ctx, aggregateID, eventType, payload); err != nil {
			log.Printf("Reaction error for event %d: %v", id, err)
		}
		*lastID = id
	}

	return nil
}

func (w *ReactionWorker) handleEvent(ctx context.Context, aggregateID, eventType string, payload []byte) error {
	// Finn alle reactions for dette eventet
	// Merk: Her forenkler vi og sjekker alle organisasjoner. 
	// I en produksjons-app ville vi kanskje ha filtrert på OrgID i eventet.
	rows, err := w.db.QueryContext(ctx, "SELECT action_type, config FROM event_reactions WHERE trigger_event = $1", eventType)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var actionType string
		var configJSON []byte
		if err := rows.Scan(&actionType, &configJSON); err != nil {
			continue
		}

		var config map[string]any
		json.Unmarshal(configJSON, &config)

		switch actionType {
		case "SEND_EMAIL":
			w.queueEmailReaction(ctx, aggregateID, config)
		}
	}

	return nil
}

func (w *ReactionWorker) queueEmailReaction(ctx context.Context, aggregateID string, config map[string]any) {
	templateID, _ := config["template_id"].(string)
	recipient, _ := config["recipient"].(string)
	
	if templateID == "" || recipient == "" { return }

	_, _ = w.db.ExecContext(ctx, `
		INSERT INTO email_outbox (recipient_email, template_id, context) 
		VALUES ($1, $2, $3)`,
		recipient, templateID, `{"aggregate_id": "`+aggregateID+`"}`)
}
