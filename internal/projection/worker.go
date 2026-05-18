package projection

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"register/internal/domain"
	"time"
)

type Worker struct {
	db *sql.DB
}

func NewWorker(db *sql.DB) *Worker {
	return &Worker{db: db}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastProcessedID int64
	// Try to get last processed ID from somewhere or start from 0
	// For MVP we just start from 0 or use a dedicated table for pointers.
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.processNewEvents(ctx, &lastProcessedID); err != nil {
				log.Printf("Projection worker error: %v", err)
			}
		}
	}
}

func (w *Worker) processNewEvents(ctx context.Context, lastID *int64) error {
	rows, err := w.db.QueryContext(ctx, "SELECT id, aggregate_id, event_type, payload FROM event_store WHERE id > $1 ORDER BY id ASC", *lastID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var aggregateID string
		var eventType string
		var payload []byte
		if err := rows.Scan(&id, &aggregateID, &eventType, &payload); err != nil {
			return err
		}

		if err := w.applyToView(ctx, aggregateID, eventType, payload); err != nil {
			return err
		}
		*lastID = id
	}
	return nil
}

func (w *Worker) applyToView(ctx context.Context, aggregateID string, eventType string, payload []byte) error {
	switch eventType {
	case domain.EventTypeMemberRegistered:
		var e domain.MemberRegistered
		if err := json.Unmarshal(payload, &e); err != nil {
			return err
		}
		
		_, err := w.db.ExecContext(ctx, `
			INSERT INTO member_view (id, org_id, name, email, status, metadata, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				email = EXCLUDED.email,
				status = EXCLUDED.status,
				metadata = EXCLUDED.metadata,
				updated_at = EXCLUDED.updated_at`,
			e.ID, e.OrgID, e.Name, e.Email, "ACTIVE", payload, time.Now())
		return err

	case domain.EventTypeMemberUpdated:
		var e domain.MemberUpdated
		if err := json.Unmarshal(payload, &e); err != nil {
			return err
		}

		// First, get the current view state
		var currentMetadata []byte
		var currentName, currentEmail, currentStatus string
		err := w.db.QueryRowContext(ctx, "SELECT name, email, status, metadata FROM member_view WHERE id = $1", aggregateID).
			Scan(&currentName, &currentEmail, &currentStatus, &currentMetadata)
		if err != nil {
			return err
		}

		// Reuse domain logic to apply updates
		m := &domain.Member{
			ID:     aggregateID,
			Name:   currentName,
			Email:  currentEmail,
			Status: currentStatus,
		}
		json.Unmarshal(currentMetadata, &m.Metadata)

		if err := domain.ApplyEvent(m, e); err != nil {
			return err
		}

		metadataJSON, _ := json.Marshal(m.Metadata)
		_, err = w.db.ExecContext(ctx, `
			UPDATE member_view 
			SET name = $1, email = $2, status = $3, metadata = $4, updated_at = $5
			WHERE id = $6`,
			m.Name, m.Email, m.Status, metadataJSON, time.Now(), aggregateID)
		return err
	}

	return nil
}
