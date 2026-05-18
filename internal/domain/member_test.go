package domain

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestDeterministicSimulation_MemberEvents(t *testing.T) {
	rng := rand.New(rand.NewSource(42)) // Deterministic random source

	m := &Member{}
	memberID := "uuid-1234"
	orgID := "org-1"
	
	var events []Event
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	
	// Initial event
	events = append(events, MemberRegistered{
		ID:        memberID,
		Name:      "Original Name",
		Email:     "original@test.com",
		OrgID:     orgID,
		Metadata:  map[string]any{"role": "user"},
		Timestamp: baseTime,
	})
	
	expectedName := "Original Name"
	expectedEmail := "original@test.com"
	expectedStatus := "ACTIVE"
	
	// Generate 999 random update events
	for i := 0; i < 999; i++ {
		baseTime = baseTime.Add(time.Duration(rng.Intn(3600)) * time.Second)
		
		updatedFields := make(map[string]any)
		action := rng.Intn(3)
		
		if action == 0 {
			newName := fmt.Sprintf("Updated Name %d", i)
			updatedFields["name"] = newName
			expectedName = newName
		} else if action == 1 {
			newEmail := fmt.Sprintf("updated%d@test.com", i)
			updatedFields["email"] = newEmail
			expectedEmail = newEmail
		} else {
			newStatus := "INACTIVE"
			updatedFields["status"] = newStatus
			expectedStatus = newStatus
		}

		events = append(events, MemberUpdated{
			ID:            memberID,
			UpdatedFields: updatedFields,
			Timestamp:     baseTime,
		})
	}
	
	// Apply all events sequentially
	for _, e := range events {
		err := ApplyEvent(m, e)
		if err != nil {
			t.Fatalf("Failed to apply event: %v", err)
		}
	}
	
	// Verify final state
	if m.ID != memberID {
		t.Errorf("Expected ID %s, got %s", memberID, m.ID)
	}
	if m.OrgID != orgID {
		t.Errorf("Expected OrgID %s, got %s", orgID, m.OrgID)
	}
	if m.Name != expectedName {
		t.Errorf("Expected Name %s, got %s", expectedName, m.Name)
	}
	if m.Email != expectedEmail {
		t.Errorf("Expected Email %s, got %s", expectedEmail, m.Email)
	}
	if m.Status != expectedStatus {
		t.Errorf("Expected Status %s, got %s", expectedStatus, m.Status)
	}
	if m.UpdatedAt.Unix() != baseTime.Unix() {
		t.Errorf("Expected UpdatedAt %v, got %v", baseTime, m.UpdatedAt)
	}
}
