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

	// 1. Create Hierarchy with Static UUIDs for Idempotency
	nationalID := "00000000-0000-0000-0000-000000000001"
	regionID   := "00000000-0000-0000-0000-000000000002"
	localID    := "00000000-0000-0000-0000-000000000003"

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
	adminEmail := "admin@chronos.no"
	var adminID string
	err = db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", adminEmail).Scan(&adminID)
	if err == sql.ErrNoRows {
		adminID = uuid.New().String()
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		
		adminEvent := domain.UserCreated{
			ID:           adminID,
			Email:        adminEmail,
			Name:         "System Admin",
			PasswordHash: string(hashedPassword),
			Timestamp:    time.Now(),
		}
		_ = es.Append(ctx, adminID, 1, adminEvent)
		
		// Update role and org for admin
		_, _ = db.ExecContext(ctx, "UPDATE users SET role = 'admin', org_id = $1 WHERE id = $2", nationalID, adminID)
		log.Println("Created admin user (admin@chronos.no / admin123)")
	} else {
		log.Println("Admin user already exists")
	}

	// 3. Seed Users and Memberships
	rng := rand.New(rand.NewSource(42))
	names := []string{"Audun", "Bente", "Christian", "Dorthe", "Erik", "Frida", "Gunnar", "Hanne", "Ivar", "Janne"}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	
	for i := 0; i < 50; i++ {
		email := fmt.Sprintf("user%d@eksempel.no", i)
		var userID string
		err = db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&userID)
		if err == sql.ErrNoRows {
			userID = uuid.New().String()
			name := fmt.Sprintf("%s %s-sen", names[rng.Intn(len(names))], names[rng.Intn(len(names))])
			
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
				Metadata:  map[string]any{"seeded": true},
				Timestamp: time.Now().AddDate(0, 0, -rng.Intn(30)),
			}
			_ = es.Append(ctx, membershipID, 1, membershipEvent)

			// Random financial history
			for j := 0; j < rng.Intn(5); j++ {
				feeEvent := domain.FeeGenerated{
					ID:        membershipID,
					Amount:    rng.Intn(5) * 100,
					Period:    "2026",
					Timestamp: time.Now().AddDate(0, 0, -rng.Intn(100)),
				}
				_ = es.Append(ctx, membershipID, j+2, feeEvent)
			}
		}
	}
	log.Println("Seeded users with membership and financial history")

	// 4. Email Templates
	_, _ = db.ExecContext(ctx, `
		INSERT INTO email_templates (org_id, name, subject, body_html) 
		VALUES ($1, 'Velkomst', 'Velkommen til {{.OrgName}}', '<h1>Hei {{.UserName}}!</h1><p>Velkommen som medlem.</p>')
		ON CONFLICT DO NOTHING`,
		nationalID)
	log.Println("Created default email template")

	log.Println("Bootstrapping complete!")
}
