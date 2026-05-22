package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"register/internal/domain"
	"time"
)

type SnapshotWorker struct {
	db         *sql.DB
	eventStore *EventStore
}

func NewSnapshotWorker(db *sql.DB, eventStore *EventStore) *SnapshotWorker {
	return &SnapshotWorker{
		db:         db,
		eventStore: eventStore,
	}
}

func (w *SnapshotWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.processSnapshots(ctx); err != nil {
				log.Printf("SnapshotWorker error: %v", err)
			}
		}
	}
}

func (w *SnapshotWorker) processSnapshots(ctx context.Context) error {
	// Find aggregates that have more than 100 events since their last snapshot
	// (or total events if no snapshot exists)
	query := `
		SELECT e.aggregate_id, MAX(e.version) as current_version, COALESCE(s.version, 0) as last_snapshot_version
		FROM event_store e
		LEFT JOIN (
			SELECT aggregate_id, MAX(version) as version
			FROM event_snapshots
			GROUP BY aggregate_id
		) s ON e.aggregate_id = s.aggregate_id
		GROUP BY e.aggregate_id, s.version
		HAVING MAX(e.version) - COALESCE(s.version, 0) >= 100
		LIMIT 50
	`

	rows, err := w.db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var aggregateID string
		var currentVersion, lastSnapshotVersion int
		if err := rows.Scan(&aggregateID, &currentVersion, &lastSnapshotVersion); err != nil {
			continue
		}

		w.createSnapshot(ctx, aggregateID, currentVersion)
	}

	return nil
}

func (w *SnapshotWorker) createSnapshot(ctx context.Context, aggregateID string, targetVersion int) {
	// 1. Fetch events up to targetVersion
	events, err := w.eventStore.GetEvents(ctx, aggregateID)
	if err != nil {
		log.Printf("Failed to get events for snapshot %s: %v", aggregateID, err)
		return
	}

	if len(events) == 0 {
		return
	}

	// 2. We need to know what type of aggregate this is.
	// Let's determine it from the first event
	firstEvent := events[0]

	var state any

	switch firstEvent.EventType {
	case domain.EventTypeMembershipCreated:
		m := &domain.Membership{}
		for _, e := range events {
			if e.Version > targetVersion {
				break
			}
			switch e.EventType {
			case domain.EventTypeMembershipCreated:
				var ev domain.MembershipCreated
				json.Unmarshal(e.Payload, &ev)
				domain.ApplyMembershipEvent(m, ev)
			case domain.EventTypeMembershipUpdated:
				var ev domain.MembershipUpdated
				json.Unmarshal(e.Payload, &ev)
				domain.ApplyMembershipEvent(m, ev)
			case domain.EventTypeFeeGenerated:
				var ev domain.FeeGenerated
				json.Unmarshal(e.Payload, &ev)
				domain.ApplyMembershipEvent(m, ev)
			case domain.EventTypePaymentReceived:
				var ev domain.PaymentReceived
				json.Unmarshal(e.Payload, &ev)
				domain.ApplyMembershipEvent(m, ev)
			case domain.EventTypeMembershipShredded:
				var ev domain.MembershipShredded
				json.Unmarshal(e.Payload, &ev)
				domain.ApplyMembershipEvent(m, ev)
			}
		}
		state = m
	case domain.EventTypeUserCreated:
		u := &domain.User{}
		for _, e := range events {
			if e.Version > targetVersion {
				break
			}
			switch e.EventType {
			case domain.EventTypeUserCreated:
				var ev domain.UserCreated
				json.Unmarshal(e.Payload, &ev)
				domain.ApplyUserEvent(u, ev)
			case domain.EventTypeUserProfileUpdated:
				var ev domain.UserProfileUpdated
				json.Unmarshal(e.Payload, &ev)
				domain.ApplyUserEvent(u, ev)
			}
		}
		state = u
	default:
		// Not an aggregate we snapshot yet
		return
	}

	stateJSON, err := json.Marshal(state)
	if err != nil {
		log.Printf("Failed to marshal snapshot state for %s: %v", aggregateID, err)
		return
	}

	_, err = w.db.ExecContext(ctx, `
		INSERT INTO event_snapshots (aggregate_id, version, state)
		VALUES ($1, $2, $3)
		ON CONFLICT (aggregate_id, version) DO NOTHING`,
		aggregateID, targetVersion, stateJSON)
	
	if err != nil {
		log.Printf("Failed to save snapshot for %s: %v", aggregateID, err)
	} else {
		log.Printf("Created snapshot for %s at version %d", aggregateID, targetVersion)
	}
}
