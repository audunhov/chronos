package reports

import (
	"context"
	"database/sql"
	"encoding/json"
	"register/internal/domain"
	"time"
)

type HistoricalMember struct {
	ID         string         `json:"id"`
	UserID     string         `json:"user_id"`
	OrgID      string         `json:"org_id"`
	Name       string         `json:"name"`
	Email      string         `json:"email"`
	Status     string         `json:"status"`
	Role       string         `json:"role"`
	Balance    int            `json:"balance"`
	FeeFormula string         `json:"fee_formula"`
	Metadata   map[string]any `json:"metadata"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func GetMembersAsOf(ctx context.Context, db *sql.DB, targetDate time.Time) (map[string]*HistoricalMember, error) {
	users := make(map[string]*domain.User)
	memberships := make(map[string]*domain.Membership)

	// 1. Load latest snapshots before targetDate
	snapshotRows, err := db.QueryContext(ctx, `
		SELECT s.aggregate_id, s.version, s.state, e.event_type
		FROM event_snapshots s
		JOIN (
			SELECT aggregate_id, MAX(version) as version
			FROM event_snapshots
			WHERE created_at <= $1
			GROUP BY aggregate_id
		) max_s ON s.aggregate_id = max_s.aggregate_id AND s.version = max_s.version
		JOIN event_store e ON e.aggregate_id = s.aggregate_id AND e.version = 1
	`, targetDate)
	if err != nil {
		return nil, err
	}
	defer snapshotRows.Close()

	aggregateVersions := make(map[string]int)

	for snapshotRows.Next() {
		var aggregateID string
		var version int
		var stateJSON []byte
		var firstEventType string
		if err := snapshotRows.Scan(&aggregateID, &version, &stateJSON, &firstEventType); err != nil {
			continue
		}

		aggregateVersions[aggregateID] = version

		switch firstEventType {
		case domain.EventTypeUserCreated:
			var u domain.User
			json.Unmarshal(stateJSON, &u)
			users[aggregateID] = &u
		case domain.EventTypeMembershipCreated:
			var m domain.Membership
			json.Unmarshal(stateJSON, &m)
			memberships[aggregateID] = &m
		}
	}

	// 2. Load events after the snapshot (or all if no snapshot) up to targetDate
	// We can query all events up to targetDate, but we ignore those <= snapshot version
	rows, err := db.QueryContext(ctx, `
		SELECT aggregate_id, version, event_type, payload, created_at 
		FROM event_store 
		WHERE created_at <= $1 
		ORDER BY id ASC`, targetDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var aggregateID string
		var version int
		var eventType string
		var payload []byte
		var createdAt time.Time
		if err := rows.Scan(&aggregateID, &version, &eventType, &payload, &createdAt); err != nil {
			return nil, err
		}

		// Skip if this event is already covered by a snapshot
		if snapVer, ok := aggregateVersions[aggregateID]; ok && version <= snapVer {
			continue
		}

		switch eventType {
		// USER EVENTS
		case domain.EventTypeUserCreated:
			var e domain.UserCreated
			json.Unmarshal(payload, &e)
			u := users[aggregateID]
			if u == nil {
				u = &domain.User{}
				users[aggregateID] = u
			}
			domain.ApplyUserEvent(u, e)
		case domain.EventTypeUserProfileUpdated:
			var e domain.UserProfileUpdated
			json.Unmarshal(payload, &e)
			if u := users[aggregateID]; u != nil {
				domain.ApplyUserEvent(u, e)
			}

		// MEMBERSHIP EVENTS
		case domain.EventTypeMembershipCreated:
			var e domain.MembershipCreated
			json.Unmarshal(payload, &e)
			m := memberships[aggregateID]
			if m == nil {
				m = &domain.Membership{}
				memberships[aggregateID] = m
			}
			domain.ApplyMembershipEvent(m, e)
		case domain.EventTypeMembershipUpdated:
			var e domain.MembershipUpdated
			json.Unmarshal(payload, &e)
			if m := memberships[aggregateID]; m != nil {
				domain.ApplyMembershipEvent(m, e)
			}
		case domain.EventTypeFeeGenerated:
			var e domain.FeeGenerated
			json.Unmarshal(payload, &e)
			if m := memberships[aggregateID]; m != nil {
				domain.ApplyMembershipEvent(m, e)
			}
		case domain.EventTypePaymentReceived:
			var e domain.PaymentReceived
			json.Unmarshal(payload, &e)
			if m := memberships[aggregateID]; m != nil {
				domain.ApplyMembershipEvent(m, e)
			}
		case domain.EventTypeMembershipShredded:
			var e domain.MembershipShredded
			json.Unmarshal(payload, &e)
			if m := memberships[aggregateID]; m != nil {
				domain.ApplyMembershipEvent(m, e)
			}
		}
	}

	// Join them in memory
	result := make(map[string]*HistoricalMember)
	for id, m := range memberships {
		u := users[m.UserID]
		if u == nil {
			continue // Should not happen if data is consistent
		}

		result[id] = &HistoricalMember{
			ID:         m.ID,
			UserID:     m.UserID,
			OrgID:      m.OrgID,
			Name:       u.Name,
			Email:      u.Email,
			Status:     m.Status,
			Role:       m.Role,
			Balance:    m.Balance,
			FeeFormula: m.FeeFormula,
			Metadata:   m.Metadata,
			UpdatedAt:  m.UpdatedAt,
		}
	}

	return result, nil
}
