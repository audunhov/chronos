package reports

import (
	"context"
	"database/sql"
	"encoding/json"
	"register/internal/domain"
	"time"
)

func GetMembersAsOf(ctx context.Context, db *sql.DB, targetDate time.Time) (map[string]*domain.Member, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT aggregate_id, event_type, payload, created_at 
		FROM event_store 
		WHERE created_at <= $1 
		ORDER BY id ASC`, targetDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make(map[string]*domain.Member)

	for rows.Next() {
		var aggregateID string
		var eventType string
		var payload []byte
		var createdAt time.Time
		if err := rows.Scan(&aggregateID, &eventType, &payload, &createdAt); err != nil {
			return nil, err
		}

		m, ok := members[aggregateID]
		if !ok {
			m = &domain.Member{}
			members[aggregateID] = m
		}

		var event domain.Event
		switch eventType {
		case domain.EventTypeMemberRegistered:
			var e domain.MemberRegistered
			json.Unmarshal(payload, &e)
			event = e
		case domain.EventTypeMemberUpdated:
			var e domain.MemberUpdated
			json.Unmarshal(payload, &e)
			event = e
		case domain.EventTypeMembershipFeeFormulaDefined:
			var e domain.MembershipFeeFormulaDefined
			json.Unmarshal(payload, &e)
			event = e
		case domain.EventTypeFeeGenerated:
			var e domain.FeeGenerated
			json.Unmarshal(payload, &e)
			event = e
		case domain.EventTypePaymentReceived:
			var e domain.PaymentReceived
			json.Unmarshal(payload, &e)
			event = e
		case domain.EventTypeMemberShredded:
			var e domain.MemberShredded
			json.Unmarshal(payload, &e)
			event = e
		}

		if event != nil {
			if err := domain.ApplyEvent(m, event); err != nil {
				return nil, err
			}
		}
	}

	return members, nil
}
