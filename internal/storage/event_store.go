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
	case domain.UserCreated:
		_, err := tx.ExecContext(ctx, `
			INSERT INTO users (id, email, name, password_hash, created_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE SET
				email = EXCLUDED.email,
				name = EXCLUDED.name,
				password_hash = EXCLUDED.password_hash,
				created_at = EXCLUDED.created_at`,
			e.ID, e.Email, e.Name, e.PasswordHash, now)
		return err

	case domain.UserProfileUpdated:
		// Oppdater users-tabellen
		_, err := tx.ExecContext(ctx, `
			UPDATE users SET name = $1, email = $2 WHERE id = $3`,
			e.Name, e.Email, e.ID)
		if err != nil {
			return err
		}
		// Denormalisering: Oppdater alle medlemskap for denne brukeren
		_, err = tx.ExecContext(ctx, `
			UPDATE membership_view SET user_name = $1, user_email = $2 WHERE user_id = $3`,
			e.Name, e.Email, e.ID)
		return err

	case domain.MembershipCreated:
		// Finn bruker-info for denormalisering
		var name, email string
		err := tx.QueryRowContext(ctx, "SELECT name, email FROM users WHERE id = $1", e.UserID).Scan(&name, &email)
		if err != nil {
			return fmt.Errorf("failed to find user for membership: %w", err)
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO membership_view (id, user_id, org_id, user_name, user_email, status, role, metadata, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			e.ID, e.UserID, e.OrgID, name, email, "ACTIVE", e.Role, "{}", now)
		return err

	case domain.MembershipUpdated:
		var m domain.Membership
		var metadataJSON []byte
		err := tx.QueryRowContext(ctx, "SELECT id, user_id, org_id, status, role, metadata FROM membership_view WHERE id = $1", aggregateID).
			Scan(&m.ID, &m.UserID, &m.OrgID, &m.Status, &m.Role, &metadataJSON)
		
		if err == sql.ErrNoRows {
			return nil
		}
		if err != nil {
			return err
		}
		json.Unmarshal(metadataJSON, &m.Metadata)

		if err := domain.ApplyMembershipEvent(&m, e); err != nil {
			return err
		}

		newMetadata, _ := json.Marshal(m.Metadata)
		_, err = tx.ExecContext(ctx, `
			UPDATE membership_view 
			SET status = $1, role = $2, metadata = $3, updated_at = $4
			WHERE id = $5`,
			m.Status, m.Role, newMetadata, now, aggregateID)
		return err

	case domain.MembershipShredded:
		_, err := tx.ExecContext(ctx, `
			UPDATE membership_view 
			SET user_name = 'REDACTED', user_email = 'redacted@example.com', status = 'SHREDDED', metadata = '{}', user_id = '00000000-0000-0000-0000-000000000000', updated_at = $1
			WHERE id = $2`,
			now, aggregateID)
		return err

	case domain.RoleAssigned:
		_, err := tx.ExecContext(ctx, `
			INSERT INTO role_assignments (id, user_id, org_id, role_type, created_at)
			VALUES ($1, $2, $3, $4, $5)`,
			e.ID, e.UserID, e.OrgID, e.RoleType, now)
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
