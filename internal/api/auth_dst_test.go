package api

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

// Simulert organisasjons-struktur (ltree-lignende i minne)
type orgNode struct {
	id   string
	path string // f.eks "Top.Region.Local"
}

func (n orgNode) IsParentOf(other orgNode) bool {
	return strings.HasPrefix(other.path, n.path+".") || other.path == n.path
}

func TestDST_HierarchicalAccess(t *testing.T) {
	rng := rand.New(rand.NewSource(999))
	jwtSecret := "hierarchical-dst-secret"
	os.Setenv("JWT_SECRET", jwtSecret)

	// 1. Bygg et tre
	hierarchy := []orgNode{
		{"root", "Top"},
		{"r1", "Top.Region1"},
		{"r2", "Top.Region2"},
		{"l1", "Top.Region1.Local1"},
		{"l2", "Top.Region1.Local2"},
	}

	// 2. Simuler 200 tilgangs-sjekker
	for i := 0; i < 200; i++ {
		// Tilfeldig bruker-org og rolle
		userOrg := hierarchy[rng.Intn(len(hierarchy))]
		userRole := []string{"admin", "user", "leader"}[rng.Intn(3)]
		
		// Tilfeldig mål-org (dataen brukeren prøver å se)
		targetOrg := hierarchy[rng.Intn(len(hierarchy))]

		// Generer token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":    "u-123",
			"org_id": userOrg.id,
			"role":   userRole,
		})
		tokenString, _ := token.SignedString([]byte(jwtSecret))

		// Simuler request med X-Org-ID (mål-org)
		req := httptest.NewRequest("GET", "/api/members", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		req.Header.Set("X-Org-ID", targetOrg.id)

		// Kjør gjennom AuthMiddleware
		server := &Server{db: nil} // db er nil her, så vi må kanskje mocke den hvis testen krever db tilgang
		handler := server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxOrgID := r.Context().Value(OrgIDKey).(string)
			ctxRole := r.Context().Value(RoleKey).(string)

			// Verifiser Business Logic for Hierarkisk tilgang:
			// I vår nåværende implementasjon i handlers.go er logikken enkel:
			// Hvis userOrgID i token er satt, låses brukeren til den orgen.
			// HVIS de sender en X-Org-ID header, skal den ignoreres hvis de allerede har en i tokenet.
			
			// MEN: En fremtidig implementasjon skal bruke ltree for å sjekke om 
			// userOrg.path @> targetOrg.path.
			
			// For nå verifiserer vi at middlewaren håndterer overstyring korrekt:
			expectedOrg := userOrg.id // Token vinner alltid hvis den finnes
			if ctxOrgID != expectedOrg {
				t.Errorf("AuthMismatch: Expected %s, got %s", expectedOrg, ctxOrgID)
			}
			
			if ctxRole != userRole {
				t.Errorf("RoleMismatch: Expected %s, got %s", userRole, ctxRole)
			}
		}))

		handler.ServeHTTP(httptest.NewRecorder(), req)
	}
}

func TestDST_AccessControl(t *testing.T) {
	rng := rand.New(rand.NewSource(123))
	jwtSecret := "dst-secret-key-12345"
	os.Setenv("JWT_SECRET", jwtSecret)

	// Vi simulerer et sett med organisasjoner og brukere
	type testUser struct {
		id    string
		email string
		orgID string
		role  string
	}

	users := []testUser{
		{"u1", "admin@global.com", "", "admin"},          // Global Admin
		{"u2", "leader@org1.com", "org-1", "leader"},     // Org 1 Leader
		{"u3", "member@org1.com", "org-1", "user"},       // Org 1 User
		{"u4", "leader@org2.com", "org-2", "leader"},     // Org 2 Leader
	}

	// Hjelpefunksjon for å generere token
	generateToken := func(u testUser) string {
		claims := jwt.MapClaims{
			"sub":    u.id,
			"email":  u.email,
			"org_id": u.orgID,
			"role":   u.role,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		ss, _ := token.SignedString([]byte(jwtSecret))
		return ss
	}

	// 1. Simuler en strøm av tilfeldige forespørsler
	for i := 0; i < 100; i++ {
		u := users[rng.Intn(len(users))]
		token := generateToken(u)

		// Simuler en forespørsel mot /api/members
		req := httptest.NewRequest("GET", "/api/members", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		
		// Eventuelt legg til X-Org-ID header (som fra dropdown i UI)
		targetOrg := ""
		if rng.Intn(2) == 0 {
			targetOrg = fmt.Sprintf("org-%d", rng.Intn(3)) // Tilfeldig org
			req.Header.Set("X-Org-ID", targetOrg)
		}

		// Kjør gjennom middleware
		server := &Server{db: nil}
		handler := server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Sjekk context verdier (Dette er det DST verifiserer)
			ctxUserID := r.Context().Value(UserIDKey).(string)
			ctxOrgID := r.Context().Value(OrgIDKey).(string)
			ctxRole := r.Context().Value(RoleKey).(string)

			// Verifiser determinisme
			if ctxUserID != u.id {
				t.Errorf("Iteration %d: Expected UserID %s, got %s", i, u.id, ctxUserID)
			}
			
			// Logikk for OrgID i context:
			// 1. Fra JWT metadata
			// 2. Fra X-Org-ID header hvis JWT er tom
			expectedOrg := u.orgID
			if expectedOrg == "" {
				expectedOrg = targetOrg
			}

			if ctxOrgID != expectedOrg {
				t.Errorf("Iteration %d: Expected OrgID %s, got %s (UserOrg: %s, HeaderOrg: %s)", 
					i, expectedOrg, ctxOrgID, u.orgID, targetOrg)
			}

			if ctxRole != u.role {
				t.Errorf("Iteration %d: Expected Role %s, got %s", i, u.role, ctxRole)
			}
		}))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Iteration %d: Expected status 200, got %d. User: %s", i, w.Code, u.email)
		}
	}
}
