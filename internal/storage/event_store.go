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

	// Check for version conflict (Optimistic Concurrency)
	var lastVersion int
	err = tx.QueryRow("SELECT COALESCE(MAX(version), 0) FROM event_store WHERE aggregate_id = $1", aggregateID).Scan(&lastVersion)
	if err != nil {
		return err
	}

	if lastVersion != version-1 {
		return fmt.Errorf("concurrency conflict: expected version %d, got %d", version-1, lastVersion)
	}

	_, err = tx.Exec(`
		INSERT INTO event_store (aggregate_id, version, event_type, payload, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		aggregateID, version, event.EventType(), payload, time.Now())
	if err != nil {
		return err
	}

	return tx.Commit()
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
