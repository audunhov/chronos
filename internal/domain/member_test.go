package domain

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func TestDST_AdvancedInvariants(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))
	
	for i := 0; i < 100; i++ {
		memberID := fmt.Sprintf("m-%d", i)
		m := &Member{}
		
		var history []Event
		var totalCharges int
		var totalPayments int
		
		isShredded := false
		
		// Simuler opptil 200 hendelser per medlem
		for j := 0; j < 200; j++ {
			ts := time.Now().Add(time.Duration(j) * time.Hour)
			var ev Event
			
			action := rng.Intn(100)
			
			if j == 0 {
				ev = MemberRegistered{ID: memberID, Name: "Initial", Timestamp: ts}
			} else if isShredded {
				// Hvis medlemmet er shreddet, prøv å sende tilfeldige hendelser
				// for å se om vi bryter invarianten
				ev = MemberUpdated{ID: memberID, UpdatedFields: map[string]any{"name": "Should Fail"}, Timestamp: ts}
			} else if action < 60 {
				// Vanlig oppdatering
				ev = MemberUpdated{ID: memberID, UpdatedFields: map[string]any{"name": fmt.Sprintf("Name %d", j)}, Timestamp: ts}
			} else if action < 80 {
				// Finansiell hendelse
				amount := (rng.Intn(100) + 1) * 100
				if rng.Intn(2) == 0 {
					ev = FeeGenerated{ID: memberID, Amount: amount, Timestamp: ts}
					totalCharges += amount
				} else {
					ev = PaymentReceived{ID: memberID, Amount: amount, Timestamp: ts}
					totalPayments += amount
				}
			} else if action < 90 {
				// Flytt organisasjon
				ev = MemberOrgMoved{ID: memberID, ToOrgID: fmt.Sprintf("org-%d", rng.Intn(5)), Timestamp: ts}
			} else {
				// SHRED!
				ev = MemberShredded{ID: memberID, Timestamp: ts}
				isShredded = true
			}
			
			err := ApplyEvent(m, ev)
			
			// Verifiser Invariant 1: Shredded medlemmer kan ikke endres
			if isShredded && j > 0 && ev.EventType() != EventTypeMemberShredded {
				if err == nil {
					t.Errorf("Member %d, step %d: Expected error when updating shredded member, but got nil", i, j)
				}
				continue // Forventet feil, hopp over resten av sjekkene for denne iterasjonen
			}
			
			if err != nil {
				t.Fatalf("Unexpected error at step %d: %v", j, err)
			}
			
			history = append(history, ev)

			// Verifiser Invariant 2: Finansiell integritet (Ledger check)
			// Balanse = Betalinger - Krav (siden krav lagres som negative i vår nåværende impl)
			expectedBalance := totalPayments - totalCharges
			if m.Balance != expectedBalance {
				t.Errorf("Member %d, step %d: Ledger mismatch! Expected %d, got %d", i, j, expectedBalance, m.Balance)
			}
		}

		// Verifiser Invariant 3: GDPR integritet
		if isShredded {
			if m.Status != StatusShredded || m.Name != "REDACTED" || m.Email != "redacted@example.com" {
				t.Errorf("Member %d: Shredding failed to redact data! State: %+v", i, m)
			}
		}
	}
}

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

func TestDST_MemberAgingAndFees(t *testing.T) {
	rng := rand.New(rand.NewSource(99)) // Fast seed
	
	// Initial state: Child member born in 2015
	birthYear := 2015
	m := &Member{
		ID: "m1",
		Metadata: map[string]any{"birth_year": birthYear},
	}

	// Vi simulerer 30 år med kontingent-endringer og aldring
	currentYear := 2020
	
	type policy struct {
		year    int
		formula string
	}

	// Organisasjonens pris-historikk
	policies := []policy{
		{2020, "AGE_BASED:18:20000:50000"}, // Barn 200, Voksen 500
		{2025, "AGE_BASED:18:25000:60000"}, // Prisstigning i 2025
		{2030, "FIXED:75000"},              // Flat avgift fra 2030
		{2040, "AGE_BASED:25:30000:90000"}, // Ny aldersgrense i 2040
	}

	policyIdx := 0
	
	for i := 0; i < 30; i++ {
		year := currentYear + i
		
		// Oppdater gjeldende pris-policy hvis vi har nådd et nytt år i historikken
		if policyIdx+1 < len(policies) && year >= policies[policyIdx+1].year {
			policyIdx++
		}
		
		currentFormula := policies[policyIdx].formula
		age := year - birthYear
		
		fee, err := CalculateFee(m, currentFormula, year)
		if err != nil {
			t.Fatalf("Calculation failed in year %d: %v", year, err)
		}

		// Deterministisk sjekk av logikken
		if strings.HasPrefix(currentFormula, "FIXED") {
			if fee != 75000 {
				t.Errorf("Year %d: Expected 75000, got %d", year, fee)
			}
		} else if strings.HasPrefix(currentFormula, "AGE_BASED") {
			if policyIdx == 0 { // 2020 policy
				if age < 18 && fee != 20000 {
					t.Errorf("Year %d (age %d): Expected 20000, got %d", year, age, fee)
				}
				if age >= 18 && fee != 50000 {
					t.Errorf("Year %d (age %d): Expected 50000, got %d", year, age, fee)
				}
			} else if policyIdx == 1 { // 2025 policy
				if age < 18 && fee != 25000 {
					t.Errorf("Year %d (age %d): Expected 25000, got %d", year, age, fee)
				}
				if age >= 18 && fee != 60000 {
					t.Errorf("Year %d (age %d): Expected 60000, got %d", year, age, fee)
				}
			}
		}

		// Simuler tilfeldige (men deterministiske) endringer i medlemsdata
		if rng.Intn(5) == 0 {
			m.Name = "Random Name Update"
		}
	}
}
