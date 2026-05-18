package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"register/internal/domain"
	"time"
)

type EventStore struct {
	db *sql.DB
}

func NewEventStore(db *sql.DB) *EventStore {
	return &EventStore{db: db}
}

type StoredEvent struct {
	ID          int64
	AggregateID string
	Version     int
	EventType   string
	Payload     json.RawMessage
	CreatedAt   time.Time
}

func (s *EventStore) Append(ctx context.Context, aggregateID string, version int, event domain.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Sjekk for versjonskonflikt (Optimistic Concurrency)
	var lastVersion int
	err = tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM event_store WHERE aggregate_id = $1", aggregateID).Scan(&lastVersion)
	if err != nil {
		return err
	}

	if lastVersion != version-1 {
		return fmt.Errorf("concurrency conflict: expected version %d, got %d", version-1, lastVersion)
	}

	// 2. Lagre eventen i event_store
	now := time.Now()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO event_store (aggregate_id, version, event_type, payload, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		aggregateID, version, event.EventType(), payload, now)
	if err != nil {
		return err
	}

	// 3. Synkron projeksjon: Oppdater member_view i samme transaksjon
	if err := s.projectSynchronously(ctx, tx, aggregateID, event, now); err != nil {
		return fmt.Errorf("failed to project event synchronously: %w", err)
	}

	return tx.Commit()
}

func (s *EventStore) projectSynchronously(ctx context.Context, tx *sql.Tx, aggregateID string, event domain.Event, now time.Time) error {
	switch e := event.(type) {
	case domain.MemberRegistered:
		_, err := tx.ExecContext(ctx, `
			INSERT INTO member_view (id, org_id, name, email, status, metadata, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO UPDATE SET
				name = EXCLUDED.name,
				email = EXCLUDED.email,
				status = EXCLUDED.status,
				metadata = EXCLUDED.metadata,
				updated_at = EXCLUDED.updated_at`,
			e.ID, e.OrgID, e.Name, e.Email, "ACTIVE", "{}", now)
		return err

	case domain.MemberUpdated:
		// Hent nåværende tilstand for å bruke domain logic (ApplyEvent)
		var m domain.Member
		var metadataJSON []byte
		err := tx.QueryRowContext(ctx, "SELECT id, org_id, name, email, status, metadata FROM member_view WHERE id = $1", aggregateID).
			Scan(&m.ID, &m.OrgID, &m.Name, &m.Email, &m.Status, &metadataJSON)
		
		if err == sql.ErrNoRows {
			return nil // Kan skje hvis eventer er ut av rekkefølge eller vi replays
		}
		if err != nil {
			return err
		}
		json.Unmarshal(metadataJSON, &m.Metadata)

		if err := domain.ApplyEvent(&m, e); err != nil {
			return err
		}

		newMetadata, _ := json.Marshal(m.Metadata)
		_, err = tx.ExecContext(ctx, `
			UPDATE member_view 
			SET name = $1, email = $2, status = $3, metadata = $4, updated_at = $5
			WHERE id = $6`,
			m.Name, m.Email, m.Status, newMetadata, now, aggregateID)
		return err

	case domain.MemberShredded:
		_, err := tx.ExecContext(ctx, `
			UPDATE member_view 
			SET name = 'REDACTED', email = 'redacted@example.com', status = 'SHREDDED', metadata = '{}', updated_at = $1
			WHERE id = $2`,
			now, aggregateID)
		return err
	}

	return nil
}

func (s *EventStore) GetEvents(ctx context.Context, aggregateID string) ([]StoredEvent, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, aggregate_id, version, event_type, payload, created_at FROM event_store WHERE aggregate_id = $1 ORDER BY version ASC", aggregateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []StoredEvent
	for rows.Next() {
		var e StoredEvent
		if err := rows.Scan(&e.ID, &e.AggregateID, &e.Version, &e.EventType, &e.Payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}
