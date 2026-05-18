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

func TestDST_MultiMemberIsolationAndReplay(t *testing.T) {
	rng := rand.New(rand.NewSource(42)) // Fixed seed
	
	numMembers := 10
	numOrgs := 3
	orgs := make([]string, numOrgs)
	for i := 0; i < numOrgs; i++ {
		orgs[i] = fmt.Sprintf("org-%d", i)
	}

	type storedEvent struct {
		id        string
		eventType string
		payload   []byte
		ts        time.Time
	}
	var allEvents []storedEvent
	groundTruth := make(map[string]*domain.Member)
	currentTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	
	for i := 0; i < 500; i++ {
		currentTime = currentTime.Add(time.Duration(rng.Intn(10)) * time.Minute)
		memberIdx := rng.Intn(numMembers)
		memberID := fmt.Sprintf("member-%d", memberIdx)
		
		var ev domain.Event
		var eventType string
		m, exists := groundTruth[memberID]
		
		if !exists {
			reg := domain.MemberRegistered{
				ID:        memberID,
				Name:      fmt.Sprintf("Member %d", memberIdx),
				Email:     fmt.Sprintf("m%d@example.com", memberIdx),
				OrgID:     orgs[rng.Intn(numOrgs)],
				Timestamp: currentTime,
			}
			ev = reg
			eventType = domain.EventTypeMemberRegistered
			groundTruth[memberID] = &domain.Member{}
		} else if m.Status == "SHREDDED" {
			continue
		} else if rng.Intn(10) == 0 {
			ev = domain.MemberShredded{ID: memberID, Timestamp: currentTime}
			eventType = domain.EventTypeMemberShredded
		} else {
			fields := make(map[string]any)
			fields["name"] = fmt.Sprintf("Updated Name %d", i)
			ev = domain.MemberUpdated{ID: memberID, UpdatedFields: fields, Timestamp: currentTime}
			eventType = domain.EventTypeMemberUpdated
		}
		
		domain.ApplyEvent(groundTruth[memberID], ev)
		payload, _ := json.Marshal(ev)
		allEvents = append(allEvents, storedEvent{id: memberID, eventType: eventType, payload: payload, ts: currentTime})
	}

	view1 := make(InMemoryView)
	for _, e := range allEvents {
		view1.Apply(e.id, e.eventType, e.payload, e.ts)
	}

	view2 := make(InMemoryView)
	for _, e := range allEvents {
		view2.Apply(e.id, e.eventType, e.payload, e.ts)
	}

	for id, expected := range groundTruth {
		actual1 := view1[id]
		actual2 := view2[id]
		
		expected.CreatedAt = time.Time{}
		expected.UpdatedAt = time.Time{}
		actual1.CreatedAt = time.Time{}
		actual1.UpdatedAt = time.Time{}
		actual2.CreatedAt = time.Time{}
		actual2.UpdatedAt = time.Time{}

		if !reflect.DeepEqual(actual1, expected) {
			t.Errorf("Isolation error for member %s", id)
		}
		if !reflect.DeepEqual(actual1, actual2) {
			t.Errorf("Determinism error for member %s", id)
		}
	}
}
