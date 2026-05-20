package domain

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestDST_MembershipInvariants(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))
	
	for i := 0; i < 100; i++ {
		membershipID := fmt.Sprintf("ms-%d", i)
		userID := fmt.Sprintf("u-%d", i)
		m := &Membership{}
		
		var totalCharges int
		var totalPayments int
		
		isShredded := false
		
		for j := 0; j < 200; j++ {
			ts := time.Now().Add(time.Duration(j) * time.Hour)
			var ev Event
			
			action := rng.Intn(100)
			
			if j == 0 {
				ev = MembershipCreated{ID: membershipID, UserID: userID, OrgID: "org-1", Timestamp: ts}
			} else if isShredded {
				ev = MembershipUpdated{ID: membershipID, UpdatedFields: map[string]any{"role": "admin"}, Timestamp: ts}
			} else if action < 60 {
				ev = MembershipUpdated{ID: membershipID, UpdatedFields: map[string]any{"role": fmt.Sprintf("Role %d", j)}, Timestamp: ts}
			} else if action < 80 {
				amount := (rng.Intn(100) + 1) * 100
				if rng.Intn(2) == 0 {
					ev = FeeGenerated{ID: membershipID, Amount: amount, Timestamp: ts}
					totalCharges += amount
				} else {
					ev = PaymentReceived{ID: membershipID, Amount: amount, Timestamp: ts}
					totalPayments += amount
				}
			} else if action < 90 {
				ev = MembershipOrgMoved{ID: membershipID, ToOrgID: fmt.Sprintf("org-%d", rng.Intn(5)), Timestamp: ts}
			} else {
				ev = MembershipShredded{ID: membershipID, Timestamp: ts}
				isShredded = true
			}
			
			err := ApplyMembershipEvent(m, ev)
			
			if isShredded && j > 0 && ev.EventType() != EventTypeMembershipShredded {
				if err == nil {
					t.Errorf("Membership %d, step %d: Expected error when updating shredded membership, but got nil", i, j)
				}
				continue
			}
			
			if err != nil {
				t.Fatalf("Unexpected error at step %d: %v", j, err)
			}
			
			expectedBalance := totalPayments - totalCharges
			if m.Balance != expectedBalance {
				t.Errorf("Membership %d, step %d: Ledger mismatch! Expected %d, got %d", i, j, expectedBalance, m.Balance)
			}
		}

		if isShredded {
			if m.Status != StatusShredded || m.UserID != "REDACTED" {
				t.Errorf("Membership %d: Shredding failed to redact! State: %+v", i, m)
			}
		}
	}
}

func TestDeterministicSimulation_MembershipEvents(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	m := &Membership{}
	membershipID := "ms-1234"
	userID := "u-5678"
	orgID := "org-1"
	
	var events []Event
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	
	events = append(events, MembershipCreated{
		ID:        membershipID,
		UserID:    userID,
		OrgID:     orgID,
		Role:      "member",
		Timestamp: baseTime,
	})
	
	expectedRole := "member"
	expectedStatus := "ACTIVE"
	
	for i := 0; i < 999; i++ {
		baseTime = baseTime.Add(time.Duration(rng.Intn(3600)) * time.Second)
		
		updatedFields := make(map[string]any)
		action := rng.Intn(2)
		
		if action == 0 {
			newRole := fmt.Sprintf("Role %d", i)
			updatedFields["role"] = newRole
			expectedRole = newRole
		} else {
			newStatus := "INACTIVE"
			updatedFields["status"] = newStatus
			expectedStatus = newStatus
		}

		events = append(events, MembershipUpdated{
			ID:            membershipID,
			UpdatedFields: updatedFields,
			Timestamp:     baseTime,
		})
	}
	
	for _, e := range events {
		err := ApplyMembershipEvent(m, e)
		if err != nil {
			t.Fatalf("Failed to apply event: %v", err)
		}
	}
	
	if m.ID != membershipID {
		t.Errorf("Expected ID %s, got %s", membershipID, m.ID)
	}
	if m.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, m.UserID)
	}
	if m.Role != expectedRole {
		t.Errorf("Expected Role %s, got %s", expectedRole, m.Role)
	}
	if m.Status != expectedStatus {
		t.Errorf("Expected Status %s, got %s", expectedStatus, m.Status)
	}
}

func TestDST_MembershipAgingAndFees(t *testing.T) {
	u := &User{ID: "u1"}
	birthYear := 2015
	m := &Membership{
		ID: "ms1",
		Metadata: map[string]any{"birth_year": birthYear},
	}

	policies := []struct {
		year    int
		formula string
	}{
		{2020, "AGE_BASED:18:20000:50000"},
		{2030, "FIXED:75000"},
	}

	for _, p := range policies {
		fee, _ := CalculateFee(u, m, p.formula, p.year)
		if strings.HasPrefix(p.formula, "FIXED") && fee != 75000 {
			t.Errorf("Year %d: Expected 75000, got %d", p.year, fee)
		}
	}
}
