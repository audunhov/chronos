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
	return s.AppendWithTx(ctx, nil, aggregateID, version, event)
}

func (s *EventStore) AppendWithTx(ctx context.Context, tx *sql.Tx, aggregateID string, version int, event domain.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	var ownsTx bool
	if tx == nil {
		tx, err = s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		ownsTx = true
	}

	// 1. Sjekk for versjonskonflikt
	var lastVersion int
	err = tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM event_store WHERE aggregate_id = $1", aggregateID).Scan(&lastVersion)
	if err != nil {
		return err
	}

	if lastVersion != version-1 {
		return fmt.Errorf("concurrency conflict: expected version %d, got %d", version-1, lastVersion)
	}

	// 2. Lagre eventen
	now := time.Now()
	correlationID, _ := ctx.Value(domain.CorrelationIDKey).(string)
	
	_, err = tx.ExecContext(ctx, `
		INSERT INTO event_store (aggregate_id, version, event_type, payload, created_at, correlation_id)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		aggregateID, version, event.EventType(), payload, now, sql.NullString{String: correlationID, Valid: correlationID != ""})
	if err != nil {
		return err
	}

	// 3. Synkron projeksjon
	if err := s.projectSynchronously(ctx, tx, aggregateID, event, now); err != nil {
		return fmt.Errorf("failed to project event synchronously: %w", err)
	}

	if ownsTx {
		return tx.Commit()
	}
	return nil
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
		// Sjekk om det er en Organ-rolle eller Organisasjons-rolle
		var query string
		if e.OrganID != nil && *e.OrganID != "" {
			query = `
				INSERT INTO role_assignments (id, user_id, org_id, organ_id, role_type, created_at)
				VALUES ($1, $2, $3, $4, $5, $6)`
			_, err := tx.ExecContext(ctx, query, e.ID, e.UserID, e.OrgID, *e.OrganID, e.RoleType, now)
			return err
		} else {
			query = `
				INSERT INTO role_assignments (id, user_id, org_id, role_type, created_at)
				VALUES ($1, $2, $3, $4, $5)`
			_, err := tx.ExecContext(ctx, query, e.ID, e.UserID, e.OrgID, e.RoleType, now)
			return err
		}

	case domain.RoleRevoked:
		_, err := tx.ExecContext(ctx, "DELETE FROM role_assignments WHERE id = $1", aggregateID)
		return err

	case domain.OrganizationCreated:
		policyJSON, _ := json.Marshal(e.Policy)
		var parentID sql.NullString
		if e.ParentID != nil && *e.ParentID != "" {
			parentID.String = *e.ParentID
			parentID.Valid = true
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO organization_hierarchy (id, name, parent_id, path, policy, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			e.ID, e.Name, parentID, e.Path, policyJSON, now)
		return err

	case domain.OrganizationDeleted:
		_, err := tx.ExecContext(ctx, "DELETE FROM organization_hierarchy WHERE id = $1", aggregateID)
		return err

	case domain.OrganCreated:
		var parentID sql.NullString
		if e.ParentOrganID != nil && *e.ParentOrganID != "" {
			parentID.String = *e.ParentOrganID
			parentID.Valid = true
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO organs (id, org_id, name, parent_organ_id, created_at)
			VALUES ($1, $2, $3, $4, $5)`,
			e.ID, e.OrgID, e.Name, parentID, now)
		return err

	case domain.FormCreated:
		schemaJSON, _ := json.Marshal(e.Schema)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO forms (id, org_id, title, schema, created_at)
			VALUES ($1, $2, $3, $4, $5)`,
			e.ID, e.OrgID, e.Title, schemaJSON, now)
		return err

	case domain.FormResponseSubmitted:
		answersJSON, _ := json.Marshal(e.Answers)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO form_responses (form_id, user_id, answers, created_at)
			VALUES ($1, $2, $3, $4)`,
			e.FormID, e.UserID, answersJSON, now)
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
