package projection

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"register/internal/domain"
	"testing"
	"time"
)

// In-memory projection state for DST
type InMemoryView map[string]*domain.Member

func (v InMemoryView) Apply(id string, eventType string, payload []byte, ts time.Time) error {
	m, ok := v[id]
	if !ok && eventType != domain.EventTypeMemberRegistered {
		return fmt.Errorf("member not found: %s", id)
	}

	if eventType == domain.EventTypeMemberRegistered {
		var e domain.MemberRegistered
		if err := json.Unmarshal(payload, &e); err != nil {
			return err
		}
		m = &domain.Member{}
		if err := domain.ApplyEvent(m, e); err != nil {
			return err
		}
		m.UpdatedAt = ts // Force consistent timestamp
		v[id] = m
		return nil
	}

	// For other events, we need the current state
	var event domain.Event
	switch eventType {
	case domain.EventTypeMemberUpdated:
		var e domain.MemberUpdated
		if err := json.Unmarshal(payload, &e); err != nil {
			return err
		}
		event = e
	case domain.EventTypeMemberShredded:
		var e domain.MemberShredded
		if err := json.Unmarshal(payload, &e); err != nil {
			return err
		}
		event = e
	default:
		return nil // Ignore unknown events for this test
	}

	if err := domain.ApplyEvent(m, event); err != nil {
		return err
	}
	m.UpdatedAt = ts
	return nil
}

func TestDST_ProjectionConsistency(t *testing.T) {
	rng := rand.New(rand.NewSource(1337)) // Fixed seed
	view := make(InMemoryView)
	
	memberID := "member-dst-1"
	orgID := "org-dst-1"
	
	var events []domain.Event

	// 1. Register
	reg := domain.MemberRegistered{
		ID:        memberID,
		Name:      "DST Member",
		Email:     "dst@example.com",
		OrgID:     orgID,
		Timestamp: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
	}
	events = append(events, reg)
	payload, _ := json.Marshal(reg)
	view.Apply(memberID, domain.EventTypeMemberRegistered, payload, reg.Timestamp)

	// 2. Random Updates
	currentTime := reg.Timestamp
	for i := 0; i < 100; i++ {
		currentTime = currentTime.Add(time.Duration(rng.Intn(60)) * time.Minute)
		
		updatedFields := make(map[string]any)
		if rng.Intn(2) == 0 {
			updatedFields["name"] = fmt.Sprintf("Name %d", i)
		} else {
			updatedFields["status"] = []string{"ACTIVE", "INACTIVE", "PENDING"}[rng.Intn(3)]
		}

		upd := domain.MemberUpdated{
			ID:            memberID,
			UpdatedFields: updatedFields,
			Timestamp:     currentTime,
		}
		events = append(events, upd)
		payload, _ := json.Marshal(upd)
		if err := view.Apply(memberID, domain.EventTypeMemberUpdated, payload, currentTime); err != nil {
			t.Fatalf("Failed at step %d: %v", i, err)
		}
	}

	// 3. Shred
	currentTime = currentTime.Add(1 * time.Hour)
	shred := domain.MemberShredded{
		ID:        memberID,
		Timestamp: currentTime,
	}
	events = append(events, shred)
	payload, _ = json.Marshal(shred)
	view.Apply(memberID, domain.EventTypeMemberShredded, payload, currentTime)

	// Final verification
	m := view[memberID]
	
	// Recreate state from scratch to compare
	replayed := &domain.Member{}
	for _, e := range events {
		domain.ApplyEvent(replayed, e)
	}

	// Compare using DeepEqual
	// Note: We need to normalize time for comparison
	m_copy := *m
	replayed_copy := *replayed
	
	m_copy.UpdatedAt = time.Time{}
	replayed_copy.UpdatedAt = time.Time{}
	m_copy.CreatedAt = time.Time{}
	replayed_copy.CreatedAt = time.Time{}

	if !reflect.DeepEqual(&m_copy, &replayed_copy) {
		t.Errorf("DeepEqual consistency failed!\nProjected: %+v\nReplayed:  %+v", &m_copy, &replayed_copy)
	}

	if m.Status != "SHREDDED" {
		t.Errorf("Expected status SHREDDED, got %s", m.Status)
	}
	if m.Name != "REDACTED" {
		t.Errorf("Expected name REDACTED, got %s", m.Name)
	}
}
