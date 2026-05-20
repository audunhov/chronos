package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"register/internal/domain"
	"register/internal/storage"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	db, err := storage.NewDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	es := storage.NewEventStore(db)
	ctx := context.Background()

	log.Println("Bootstrapping system...")

	// 1. Create Hierarchy
	nationalID := uuid.New().String()
	regionID := uuid.New().String()
	localID := uuid.New().String()

	orgs := []struct {
		id       string
		name     string
		parentID sql.NullString
		path     string
		policy   string
	}{
		{nationalID, "Norges Spillforbund", sql.NullString{}, "NSF", `{"allow_multiple": true}`},
		{regionID, "Viken Spillregion", sql.NullString{String: nationalID, Valid: true}, "NSF.Viken", `{"allow_multiple": true}`},
		{localID, "Oslo Brettspillklubb", sql.NullString{String: regionID, Valid: true}, "NSF.Viken.Oslo", `{"allow_multiple": false}`},
	}

	for _, o := range orgs {
		_, err := db.ExecContext(ctx, 
			"INSERT INTO organization_hierarchy (id, name, parent_id, path, policy) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (id) DO NOTHING",
			o.id, o.name, o.parentID, o.path, o.policy)
		if err != nil {
			log.Fatalf("Failed to create org %s: %v", o.name, err)
		}
	}
	log.Println("Created organization hierarchy")

	// 2. Create Admin User
	adminID := uuid.New().String()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	
	adminEvent := domain.UserCreated{
		ID:           adminID,
		Email:        "admin@chronos.no",
		Name:         "System Admin",
		PasswordHash: string(hashedPassword),
		Timestamp:    time.Now(),
	}
	_ = es.Append(ctx, adminID, 1, adminEvent)
	
	// Assign Admin Role to National Org
	_, _ = db.ExecContext(ctx, "INSERT INTO role_assignments (user_id, org_id, role_type) VALUES ($1, $2, $3)", adminID, nationalID, "admin")
	log.Println("Created admin user (admin@chronos.no / admin123)")

	// 3. Seed Users and Memberships
	rng := rand.New(rand.NewSource(42))
	names := []string{"Audun", "Bente", "Christian", "Dorthe", "Erik", "Frida", "Gunnar", "Hanne", "Ivar", "Janne"}
	
	for i := 0; i < 50; i++ {
		userID := uuid.New().String()
		name := fmt.Sprintf("%s %s-sen", names[rng.Intn(len(names))], names[rng.Intn(len(names))])
		email := fmt.Sprintf("user%d@eksempel.no", i)
		
		userEvent := domain.UserCreated{
			ID:           userID,
			Email:        email,
			Name:         name,
			PasswordHash: string(hashedPassword),
			Timestamp:    time.Now().AddDate(0, 0, -rng.Intn(365)),
		}
		_ = es.Append(ctx, userID, 1, userEvent)

		// Join either local, region or national
		targetOrg := localID
		if rng.Intn(10) > 8 { targetOrg = regionID }
		if rng.Intn(10) > 9 { targetOrg = nationalID }

		membershipID := uuid.New().String()
		membershipEvent := domain.MembershipCreated{
			ID:        membershipID,
			UserID:    userID,
			OrgID:     targetOrg,
			Role:      "member",
			Metadata:  map[string]any{"birth_year": 1980 + rng.Intn(30)},
			Timestamp: userEvent.Timestamp.Add(time.Hour),
		}
		_ = es.Append(ctx, membershipID, 1, membershipEvent)

		// Generate some financial history
		for month := 1; month <= 6; month++ {
			feeDate := time.Date(2026, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
			if feeDate.After(time.Now()) { break }

			feeEvent := domain.FeeGenerated{
				ID:        membershipID,
				Amount:    10000, // 100 kr
				Period:    fmt.Sprintf("2026-%02d", month),
				Timestamp: feeDate,
			}
			_ = es.Append(ctx, membershipID, month+1, feeEvent)

			// Randomly pay
			if rng.Intn(10) > 2 {
				payEvent := domain.PaymentReceived{
					ID:        membershipID,
					Amount:    10000,
					Timestamp: feeDate.AddDate(0, 0, rng.Intn(15)),
				}
				_ = es.Append(ctx, membershipID, month+2, payEvent)
			}
		}
	}
	log.Println("Seeded 50 users with membership and financial history")

	// 4. Create an Email Template
	templateID := uuid.New().String()
	_, _ = db.ExecContext(ctx, `
		INSERT INTO email_templates (id, name, subject, body_html) 
		VALUES ($1, 'Velkomst', 'Velkommen til {{.OrgName}}', '<h1>Hei {{.UserName}}!</h1><p>Velkommen som medlem.</p>')`,
		templateID)
	log.Println("Created default email template")

	log.Println("Bootstrapping complete!")
}
