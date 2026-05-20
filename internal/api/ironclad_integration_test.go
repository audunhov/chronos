package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"register/internal/storage"
	"testing"

	"github.com/google/uuid"
)

func TestIroncladIntegration_HierarchicalGatekeeper(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, err := storage.NewDB()
	if err != nil {
		t.Skipf("Skipping: DB not available: %v", err)
		return
	}
	defer db.Close()
	es := storage.NewEventStore(db)
	server := NewServer(db, es)

	ctx := context.Background()
	
	// Setup test orgs
	exclusiveParentID := uuid.New().String()
	childA_ID := uuid.New().String()
	childB_ID := uuid.New().String()

	// Parent org (Exclusive)
	_, _ = db.ExecContext(ctx, "INSERT INTO organization_hierarchy (id, name, path, policy) VALUES ($1, $2, $3, $4)",
		exclusiveParentID, "Exclusive Federation", "EXCL", `{"allow_multiple": false}`)
	
	// Children (Default to Additive, but parent is Exclusive)
	_, _ = db.ExecContext(ctx, "INSERT INTO organization_hierarchy (id, name, parent_id, path, policy) VALUES ($1, $2, $3, $4, $5)",
		childA_ID, "Local A", exclusiveParentID, "EXCL.A", `{"allow_multiple": true}`)
	_, _ = db.ExecContext(ctx, "INSERT INTO organization_hierarchy (id, name, parent_id, path, policy) VALUES ($1, $2, $3, $4, $5)",
		childB_ID, "Local B", exclusiveParentID, "EXCL.B", `{"allow_multiple": true}`)

	email := fmt.Sprintf("test-%s@gatekeeper.no", uuid.New().String())

	// 1. First registration (Success)
	req1 := RegisterMemberRequest{
		Name:  "Test User",
		Email: email,
		OrgID: childA_ID,
	}
	body1, _ := json.Marshal(req1)
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest("POST", "/api/commands/register-member", bytes.NewBuffer(body1))
	server.RegisterMemberHandler(w1, r1)

	if w1.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d. Body: %s", w1.Code, w1.Body.String())
	}

	// 2. Second registration in same EXCLUSIVE branch (Conflict)
	req2 := RegisterMemberRequest{
		Name:  "Test User",
		Email: email,
		OrgID: childB_ID,
	}
	body2, _ := json.Marshal(req2)
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("POST", "/api/commands/register-member", bytes.NewBuffer(body2))
	server.RegisterMemberHandler(w2, r2)

	if w2.Code != http.StatusConflict {
		t.Errorf("Expected 409 Conflict for exclusive branch, got %d. Body: %s", w2.Code, w2.Body.String())
	}

	if !bytes.Contains(w2.Body.Bytes(), []byte("controlled by Exclusive Federation")) {
		t.Errorf("Error message did not mention the controlling org: %s", w2.Body.String())
	}
}

func TestIroncladIntegration_MagicLinkFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, err := storage.NewDB()
	if err != nil {
		t.Skipf("Skipping integration test: DB not available: %v", err)
		return
	}
	defer db.Close()
	es := storage.NewEventStore(db)
	server := NewServer(db, es)
	
	email := fmt.Sprintf("magic-%s@test.no", uuid.New().String())
	
	// 1. Request Magic Link (User doesn't exist yet)
	req1 := map[string]string{"email": email}
	body1, _ := json.Marshal(req1)
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest("POST", "/api/auth/request-magic-link", bytes.NewBuffer(body1))
	server.RequestMagicLinkHandler(w1, r1)
	
	if w1.Code != http.StatusOK {
		t.Errorf("Expected 200 even for non-existent user, got %d", w1.Code)
	}

	// 2. Create the user manually first
	userID := uuid.New().String()
	_, _ = db.ExecContext(context.Background(), "INSERT INTO users (id, email, name, password_hash) VALUES ($1, $2, $3, $4)",
		userID, email, "Magic User", "hash")
	
	// 3. Request again (Success)
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("POST", "/api/auth/request-magic-link", bytes.NewBuffer(body1))
	server.RequestMagicLinkHandler(w2, r2)
	
	if w2.Code != http.StatusOK {
		t.Fatalf("Failed to request magic link: %s", w2.Body.String())
	}
	
	// 4. Find token in DB
	var token string
	err = db.QueryRowContext(context.Background(), "SELECT token FROM magic_links WHERE user_id = $1", userID).Scan(&token)
	if err != nil {
		t.Fatalf("Magic link token not found in DB: %v", err)
	}
	
	// 5. Login with token
	w3 := httptest.NewRecorder()
	r3 := httptest.NewRequest("GET", "/api/auth/magic-login?token="+token, nil)
	server.LoginWithMagicLinkHandler(w3, r3)
	
	if w3.Code != http.StatusOK {
		t.Errorf("Login failed: %s", w3.Body.String())
	}
	
	var authResp struct {
		AccessToken string `json:"access_token"`
	}
	json.Unmarshal(w3.Body.Bytes(), &authResp)
	if authResp.AccessToken == "" {
		t.Error("Missing access token in magic login response")
	}
}
