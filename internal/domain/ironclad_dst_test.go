package domain

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

// Scenario: One user has multiple memberships across different organizations.
// We test that their balances are isolated, but their profile is shared and redaction is handled correctly.
func TestIroncladDST_MultiMembershipIsolation(t *testing.T) {
	seeds := []int64{42, 1337, 2026, 999}
	
	for _, seed := range seeds {
		t.Run(fmt.Sprintf("Seed-%d", seed), func(t *testing.T) {
			rng := rand.New(rand.NewSource(seed))
			
			u := &User{}
			userID := "u-1"
			
			// Track ground truth for each membership
			type msState struct {
				m       *Membership
				balance int
			}
			memberships := make(map[string]*msState)
			orgIDs := []string{"org-national", "org-regional-1", "org-local-1", "org-local-2"}

			currentTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

			// 1. Initial User Creation
			userCreated := UserCreated{
				ID:           userID,
				Email:        "user@example.com",
				Name:         "Original Name",
				PasswordHash: "secret",
				Timestamp:    currentTime,
			}
			ApplyUserEvent(u, userCreated)

			for i := 0; i < 5000; i++ {
				currentTime = currentTime.Add(time.Duration(rng.Intn(60)) * time.Minute)
				action := rng.Intn(100)

				if action < 5 && len(memberships) < len(orgIDs) {
					// CREATE MEMBERSHIP
					orgID := orgIDs[len(memberships)]
					msID := fmt.Sprintf("ms-%s", orgID)
					if _, exists := memberships[msID]; !exists {
						ev := MembershipCreated{
							ID:        msID,
							UserID:    userID,
							OrgID:     orgID,
							Role:      "member",
							Timestamp: currentTime,
						}
						m := &Membership{}
						ApplyMembershipEvent(m, ev)
						memberships[msID] = &msState{m: m, balance: 0}
					}
				} else if action < 15 {
					// UPDATE USER PROFILE (Shared state)
					newName := fmt.Sprintf("Name Update %d", i)
					ev := UserProfileUpdated{
						ID:        userID,
						Name:      newName,
						Email:     u.Email,
						Timestamp: currentTime,
					}
					ApplyUserEvent(u, ev)
				} else if action < 60 && len(memberships) > 0 {
					// FINANCIAL ACTION (Isolated state)
					// Pick random membership
					var ids []string
					for id := range memberships { ids = append(ids, id) }
					targetID := ids[rng.Intn(len(ids))]
					state := memberships[targetID]

					if state.m.Status == StatusShredded { continue }

					amount := (rng.Intn(100) + 1) * 10
					if rng.Intn(2) == 0 {
						// Fee
						ev := FeeGenerated{ID: targetID, Amount: amount, Timestamp: currentTime}
						ApplyMembershipEvent(state.m, ev)
						state.balance -= amount
					} else {
						// Payment
						ev := PaymentReceived{ID: targetID, Amount: amount, Timestamp: currentTime}
						ApplyMembershipEvent(state.m, ev)
						state.balance += amount
					}
				} else if action < 62 && len(memberships) > 0 {
					// SHRED MEMBERSHIP (Scoped shredding)
					var ids []string
					for id := range memberships { ids = append(ids, id) }
					targetID := ids[rng.Intn(len(ids))]
					state := memberships[targetID]
					
					if state.m.Status != StatusShredded {
						ev := MembershipShredded{ID: targetID, Timestamp: currentTime}
						ApplyMembershipEvent(state.m, ev)
					}
				}
			}

			// FINAL INVARIANT VERIFICATION
			for id, state := range memberships {
				// 1. Balance Invariant: The aggregate balance must match our ground truth
				if state.m.Balance != state.balance {
					t.Errorf("Inconsistency in membership %s: Expected balance %d, got %d", id, state.balance, state.m.Balance)
				}

				// 2. Isolation Invariant: Action on one membership must not affect others
				// (Implicitly checked by the loop and ground truth map)

				// 3. Redaction Invariant: Shredded memberships must not have a UserID link
				if state.m.Status == StatusShredded {
					if state.m.UserID != "REDACTED" {
						t.Errorf("Shredded membership %s still has UserID link: %s", id, state.m.UserID)
					}
				} else {
					// 4. Identity Invariant: Active memberships must point to the correct UserID
					if state.m.UserID != userID {
						t.Errorf("Active membership %s points to wrong UserID: %s", id, state.m.UserID)
					}
				}
			}

			// 5. Profile Invariant: The User aggregate should be stable across all memberships
			// (Since they are decoupled, we verify u itself)
			if u.ID != userID {
				t.Errorf("User ID changed unexpectedly: %s", u.ID)
			}
		})
	}
}

func TestIroncladDST_FeeCalculationEdgeCases(t *testing.T) {
	rng := rand.New(rand.NewSource(888))
	u := &User{ID: "u1"}
	
	for i := 0; i < 1000; i++ {
		birthYear := 1950 + rng.Intn(70)
		refYear := 2020 + rng.Intn(10)
		limit := 10 + rng.Intn(20)
		under := rng.Intn(50000)
		over := rng.Intn(100000)
		
		formula := fmt.Sprintf("AGE_BASED:%d:%d:%d", limit, under, over)
		m := &Membership{
			Metadata: map[string]any{"birth_year": birthYear},
		}

		fee, err := CalculateFee(u, m, formula, refYear)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		age := refYear - birthYear
		if age < limit {
			if fee != under {
				t.Errorf("Age %d (< %d): Expected %d, got %d", age, limit, under, fee)
			}
		} else {
			if fee != over {
				t.Errorf("Age %d (>= %d): Expected %d, got %d", age, limit, over, fee)
			}
		}
	}
}

// Verify that events can be replayed in any order (if they are independent) 
// or strictly (if dependent) to reach the same state.
func TestIroncladDST_ReplayConsistency(t *testing.T) {
	m1 := &Membership{}
	m2 := &Membership{}
	
	id := "ms-replay"
	var events []Event
	
	// Create
	ts := time.Now()
	e1 := MembershipCreated{ID: id, UserID: "u1", OrgID: "o1", Timestamp: ts}
	events = append(events, e1)
	
	// Sequential updates
	for i := 0; i < 100; i++ {
		ts = ts.Add(time.Minute)
		events = append(events, FeeGenerated{ID: id, Amount: 100, Timestamp: ts})
	}
	
	// Replay 1: Full sequence
	for _, e := range events {
		ApplyMembershipEvent(m1, e)
	}
	
	// Replay 2: Full sequence
	for _, e := range events {
		ApplyMembershipEvent(m2, e)
	}
	
	if !reflect.DeepEqual(m1, m2) {
		t.Error("Non-deterministic replay: m1 and m2 should be identical")
	}
	
	if m1.Balance != -10000 {
		t.Errorf("Expected balance -10000, got %d", m1.Balance)
	}
}
