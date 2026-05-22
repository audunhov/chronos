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

	log.Println("Wiping existing data for full historical reset...")
	tables := []string{
		"form_responses", "forms", "email_outbox", "email_templates",
		"role_assignments", "organs", "membership_view", "event_store",
		"users", "organization_hierarchy", "audit_logs", "magic_links",
		"payment_pipelines", "invoice_view", "event_reactions",
	}
	for _, table := range tables {
		_, err := db.ExecContext(ctx, "TRUNCATE TABLE "+table+" CASCADE")
		if err != nil {
			log.Printf("Warning: failed to truncate %s: %v", table, err)
		}
	}

	// 1. Create Hierarchy
	nationalID := "00000000-0000-0000-0000-000000000001"
	regionID := "00000000-0000-0000-0000-000000000002"
	localID := "00000000-0000-0000-0000-000000000003"

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
			"INSERT INTO organization_hierarchy (id, name, parent_id, path, policy) VALUES ($1, $2, $3, $4, $5)",
			o.id, o.name, o.parentID, o.path, o.policy)
		if err != nil {
			log.Fatalf("Failed to create org %s: %v", o.name, err)
		}
	}
	log.Println("Created organization hierarchy")

	// 2. Create Admin User
	adminID := "00000000-0000-0000-0000-ffffffffffff"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	
	adminEvent := domain.UserCreated{
		ID:           adminID,
		Email:        "admin@chronos.no",
		Name:         "System Admin",
		PasswordHash: string(hashedPassword),
		Timestamp:    time.Now().AddDate(-3, 0, 0), // Born 3 years ago
	}
	_ = es.Append(ctx, adminID, 1, adminEvent)
	
	_, _ = db.ExecContext(ctx, "INSERT INTO role_assignments (user_id, org_id, role_type) VALUES ($1, $2, 'admin')", adminID, nationalID)
	log.Println("Created admin user (admin@chronos.no / admin123)")

	// 3. Seed Users and Memberships (Varied over 3 years)
	rng := rand.New(rand.NewSource(1337))
	names := []string{"Audun", "Bente", "Christian", "Dorthe", "Erik", "Frida", "Gunnar", "Hanne", "Ivar", "Janne", "Knut", "Lise", "Morten", "Nina", "Ole", "Pia"}
	
	startDate := time.Now().AddDate(-3, 0, 0)
	
	for i := 0; i < 150; i++ {
		// Random creation date between 3 years ago and now
		daysOffset := rng.Intn(3 * 365)
		userBorn := startDate.AddDate(0, 0, daysOffset)
		
		userID := uuid.New().String()
		name := fmt.Sprintf("%s %s-sen", names[rng.Intn(len(names))], names[rng.Intn(len(names))])
		email := fmt.Sprintf("user%d@eksempel.no", i)
		
		userEvent := domain.UserCreated{
			ID:           userID,
			Email:        email,
			Name:         name,
			PasswordHash: string(hashedPassword),
			Timestamp:    userBorn,
		}
		_ = es.Append(ctx, userID, 1, userEvent)

		// Join an org shortly after
		targetOrg := localID
		if rng.Intn(10) > 8 { targetOrg = regionID }
		
		membershipID := uuid.New().String()
		version := 1
		msEvent := domain.MembershipCreated{
			ID:        membershipID,
			UserID:    userID,
			OrgID:     targetOrg,
			Role:      "member",
			Metadata:  map[string]any{"birth_year": 1970 + rng.Intn(40)},
			Timestamp: userBorn.Add(time.Hour * 2),
		}
		_ = es.Append(ctx, membershipID, version, msEvent)
		version++

		// Generate yearly fees and random payments
		current := userBorn.AddDate(0, 1, 0) // First fee 1 month after join
		for current.Before(time.Now()) {
			feeAmount := 50000 // 500 NOK
			_ = es.Append(ctx, membershipID, version, domain.FeeGenerated{
				ID:        membershipID,
				Amount:    feeAmount,
				Period:    fmt.Sprintf("%d", current.Year()),
				Timestamp: current,
			})
			version++

			// 80% chance of paying
			if rng.Intn(10) < 8 {
				payDate := current.AddDate(0, 0, rng.Intn(30))
				if payDate.Before(time.Now()) {
					_ = es.Append(ctx, membershipID, version, domain.PaymentReceived{
						ID:        membershipID,
						Amount:    feeAmount,
						Timestamp: payDate,
					})
					version++
				}
			}

			current = current.AddDate(1, 0, 0) // Yearly fee
		}

		// Churn: 15% chance of leaving (Shredding)
		if rng.Intn(100) < 15 {
			leaveDate := userBorn.AddDate(0, rng.Intn(24), rng.Intn(30))
			if leaveDate.Before(time.Now()) {
				_ = es.Append(ctx, membershipID, version, domain.MembershipShredded{
					ID:        membershipID,
					Timestamp: leaveDate,
				})
			}
		}
	}
	log.Println("Seeded 150 users with 3 years of history and simulated churn")

	// 4. Default Form
	formID := uuid.New().String()
	formEvent := domain.FormCreated{
		ID:    formID,
		OrgID: nationalID,
		Title: "Medlemsundersøkelse 2026",
		Schema: map[string]any{
			"fields": []map[string]any{
				{"name": "trivsel", "label": "Hvor bra trives du?", "type": "select", "options": []string{"Veldig bra", "Bra", "Ok", "Dårlig"}},
				{"name": "kommentar", "label": "Noe mer på hjertet?", "type": "textarea"},
			},
		},
		Timestamp: time.Now().AddDate(-1, 0, 0),
	}
	_ = es.Append(ctx, formID, 1, formEvent)
	log.Println("Created default survey")

	// 5. Default Pipeline
	reactionID := uuid.New().String()
	defaultConfig := `{"nodes": [{"id": "trigger", "type": "trigger", "position": {"x": 50, "y": 50}, "data": {"event": "MembershipCreated", "outputs": ["user_id", "org_id"]}}], "edges": []}`
	_, _ = db.ExecContext(ctx, `
		INSERT INTO event_reactions (id, org_id, trigger_event, action_type, config)
		VALUES ($1, $2, 'MembershipCreated', 'PIPELINE_DAG', $3)`,
		reactionID, nationalID, defaultConfig)
	log.Println("Created default pipeline")

	log.Println("Bootstrapping complete!")
}
