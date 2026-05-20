package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"register/internal/domain"
	"register/internal/reports"
	"register/internal/storage"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Server struct {
	db         *sql.DB
	eventStore *storage.EventStore
}

func NewServer(db *sql.DB, es *storage.EventStore) *Server {
	return &Server{db: db, eventStore: es}
}

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) SwaggerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Chronos API - Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css">
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-standalone-preset.js"></script>
    <script>
    window.onload = function() {
        window.ui = SwaggerUIBundle({
            url: "/openapi.yaml",
            dom_id: '#swagger-ui',
            deepLinking: true,
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIStandalonePreset
            ],
            plugins: [
                SwaggerUIBundle.plugins.DownloadUrl
            ],
            layout: "StandaloneLayout"
        });
    };
    </script>
</body>
</html>
	`))
}

func (s *Server) GetMyProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(UserIDKey).(string)
	
	var u domain.User
	err := s.db.QueryRowContext(r.Context(), "SELECT id, email, name, created_at FROM users WHERE id = $1", userID).
		Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt)
	
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func (s *Server) GetMyMembershipsHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(UserIDKey).(string)

	query := `
		SELECT m.id, m.org_id, o.name as org_name, m.status, m.role, m.balance, m.updated_at 
		FROM membership_view m
		JOIN organization_hierarchy o ON m.org_id = o.id
		WHERE m.user_id = $1`
	
	rows, err := s.db.QueryContext(r.Context(), query, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type MyMembershipResponse struct {
		ID        string    `json:"id"`
		OrgID     string    `json:"org_id"`
		OrgName   string    `json:"org_name"`
		Status    string    `json:"status"`
		Role      string    `json:"role"`
		Balance   int       `json:"balance"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	var memberships []MyMembershipResponse
	for rows.Next() {
		var m MyMembershipResponse
		if err := rows.Scan(&m.ID, &m.OrgID, &m.OrgName, &m.Status, &m.Role, &m.Balance, &m.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		memberships = append(memberships, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(memberships)
}

func (s *Server) GetMembersHandler(w http.ResponseWriter, r *http.Request) {
	orgID, _ := r.Context().Value(OrgIDKey).(string)

	query := "SELECT id, user_id, org_id, user_name, user_email, status, role, balance, fee_formula, metadata, updated_at FROM membership_view"
	var args []any

	if orgID != "" {
		query += " WHERE org_id = $1"
		args = append(args, orgID)
	}

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type MemberResponse struct {
		ID         string         `json:"id"`
		UserID     string         `json:"user_id"`
		OrgID      string         `json:"org_id"`
		Name       string         `json:"name"`
		Email      string         `json:"email"`
		Status     string         `json:"status"`
		Role       string         `json:"role"`
		Balance    int            `json:"balance"`
		FeeFormula string         `json:"fee_formula"`
		Metadata   map[string]any `json:"metadata"`
		UpdatedAt  time.Time      `json:"updated_at"`
	}

	var members = []MemberResponse{}
	for rows.Next() {
		var m MemberResponse
		var metadata []byte
		var feeFormula sql.NullString
		if err := rows.Scan(&m.ID, &m.UserID, &m.OrgID, &m.Name, &m.Email, &m.Status, &m.Role, &m.Balance, &feeFormula, &metadata, &m.UpdatedAt); err != nil {
			log.Printf("Scan error in GetMembers: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		m.FeeFormula = feeFormula.String
		json.Unmarshal(metadata, &m.Metadata)
		members = append(members, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func (s *Server) GetOrganizationsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id, name FROM organization_hierarchy 
		UNION 
		SELECT DISTINCT m.org_id, o.name 
		FROM membership_view m 
		JOIN organization_hierarchy o ON m.org_id = o.id`)

	if err != nil {
		log.Printf("GetOrganizations error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Org struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	var orgs []Org
	for rows.Next() {
		var o Org
		if err := rows.Scan(&o.ID, &o.Name); err != nil {
			continue
		}
		orgs = append(orgs, o)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orgs)
}

func (s *Server) UpdateMyProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if req.Name == "" && req.Email == "" {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	// Vi sender et UserProfileUpdated event til event store for å bevare historikk
	event := domain.UserProfileUpdated{
		ID:        userID,
		Name:      req.Name,
		Email:     req.Email,
		Timestamp: time.Now(),
	}

	// Finn gjeldende versjon
	var version int
	err := s.db.QueryRowContext(r.Context(), "SELECT COALESCE(MAX(version), 0) FROM event_store WHERE aggregate_id = $1", userID).Scan(&version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.eventStore.Append(r.Context(), userID, version+1, event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Oppdater users-tabellen (read model for auth)
	query := "UPDATE users SET "
	var args []any
	if req.Name != "" {
		query += "name = $1"
		args = append(args, req.Name)
	}
	if req.Email != "" {
		if len(args) > 0 { query += ", " }
		query += "email = $" + fmt.Sprint(len(args)+1)
		args = append(args, req.Email)
	}
	query += " WHERE id = $" + fmt.Sprint(len(args)+1)
	args = append(args, userID)

	_, err = s.db.ExecContext(r.Context(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) GetReactionsHandler(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		http.Error(w, "Missing org_id", http.StatusBadRequest)
		return
	}

	// Hent path for org
	var path string
	err := s.db.QueryRowContext(r.Context(), "SELECT path::text FROM organization_hierarchy WHERE id = $1", orgID).Scan(&path)
	if err != nil {
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	// Hent alle reaksjoner som treffer denne path-en (arv oppover i hierarkiet)
	query := `
		SELECT er.id, er.org_id, er.trigger_event, er.action_type, er.config,
		       (er.org_id != $1) as is_inherited
		FROM event_reactions er
		JOIN organization_hierarchy o ON er.org_id = o.id
		WHERE o.path @> $2::ltree
		ORDER BY o.path ASC`

	rows, err := s.db.QueryContext(r.Context(), query, orgID, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ReactionResponse struct {
		ID           string         `json:"id"`
		OrgID        string         `json:"org_id"`
		TriggerEvent string         `json:"trigger_event"`
		ActionType   string         `json:"action_type"`
		Config       map[string]any `json:"config"`
		IsInherited  bool           `json:"is_inherited"`
	}

	var reactions []ReactionResponse
	for rows.Next() {
		var r ReactionResponse
		var configJSON []byte
		if err := rows.Scan(&r.ID, &r.OrgID, &r.TriggerEvent, &r.ActionType, &configJSON, &r.IsInherited); err != nil {
			continue
		}
		json.Unmarshal(configJSON, &r.Config)
		reactions = append(reactions, r)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reactions)
}

func (s *Server) CreateReactionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrgID        string         `json:"org_id"`
		TriggerEvent string         `json:"trigger_event"`
		ActionType   string         `json:"action_type"`
		Config       map[string]any `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	configJSON, _ := json.Marshal(req.Config)
	_, err := s.db.ExecContext(r.Context(), `
		INSERT INTO event_reactions (org_id, trigger_event, action_type, config)
		VALUES ($1, $2, $3, $4)`,
		req.OrgID, req.TriggerEvent, req.ActionType, configJSON)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) GetOrgansHandler(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		http.Error(w, "Missing org_id", http.StatusBadRequest)
		return
	}

	rows, err := s.db.QueryContext(r.Context(), "SELECT id, org_id, name, parent_organ_id FROM organs WHERE org_id = $1", orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type OrganResponse struct {
		ID            string  `json:"id"`
		OrgID         string  `json:"org_id"`
		Name          string  `json:"name"`
		ParentOrganID *string `json:"parent_organ_id"`
	}

	var organs []OrganResponse
	for rows.Next() {
		var o OrganResponse
		var parentID sql.NullString
		if err := rows.Scan(&o.ID, &o.OrgID, &o.Name, &parentID); err != nil {
			continue
		}
		if parentID.Valid {
			o.ParentOrganID = &parentID.String
		}
		organs = append(organs, o)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(organs)
}

func (s *Server) CreateOrganHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrgID         string  `json:"org_id"`
		Name          string  `json:"name"`
		ParentOrganID *string `json:"parent_organ_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	var parentID sql.NullString
	if req.ParentOrganID != nil {
		parentID.String = *req.ParentOrganID
		parentID.Valid = true
	}

	_, err := s.db.ExecContext(r.Context(), `
		INSERT INTO organs (org_id, name, parent_organ_id)
		VALUES ($1, $2, $3)`,
		req.OrgID, req.Name, parentID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type RegisterMemberRequest struct {
	Name     string         `json:"name"`
	Email    string         `json:"email"`
	OrgID    string         `json:"org_id"`
	Metadata map[string]any `json:"metadata"`
}

func (s *Server) RegisterMemberHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 1. Sjekk om bruker finnes fra før
	var userID string
	err := s.db.QueryRowContext(r.Context(), "SELECT id FROM users WHERE email = $1", req.Email).Scan(&userID)
	
	if err == sql.ErrNoRows {
		// Lag ny bruker (vi låser på email-streng for å unngå race condition ved bruker-opprettelse)
		// Bruker pg_advisory_xact_lock med en hash av emailen
		_, _ = s.db.ExecContext(r.Context(), "SELECT pg_advisory_xact_lock(hashtext($1))", req.Email)
		
		// Sjekk igjen etter låsen er tatt
		err = s.db.QueryRowContext(r.Context(), "SELECT id FROM users WHERE email = $1", req.Email).Scan(&userID)
		if err == sql.ErrNoRows {
			userID = uuid.New().String()
			userEvent := domain.UserCreated{
				ID:        userID,
				Email:     req.Email,
				Name:      req.Name,
				Timestamp: time.Now(),
			}
			if err := s.eventStore.Append(r.Context(), userID, 1, userEvent); err != nil {
				http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Ta en lås på UserID for å forhindre race conditions ved medlemskap-sjekk
	// Vi bruker en transaksjonell advisory lock
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "Transaction failed", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Hash UUID til en bigint for pg_advisory_xact_lock
	_, err = tx.ExecContext(r.Context(), "SELECT pg_advisory_xact_lock(hashtext($1))", userID)
	if err != nil {
		http.Error(w, "Locking failed", http.StatusInternalServerError)
		return
	}

	// 3. Finn "Most Restrictive" policy i hierarkiet
	var orgPath string
	err = tx.QueryRowContext(r.Context(), "SELECT path FROM organization_hierarchy WHERE id = $1", req.OrgID).Scan(&orgPath)
	if err != nil {
		orgPath = "" // Fallback
	}

	// Sjekk om noen i grenen (meg selv eller forfedre) har allow_multiple: false
	var isExclusive bool
	if orgPath != "" {
		query := `
			SELECT EXISTS (
				SELECT 1 
				FROM organization_hierarchy 
				WHERE path @> $1 AND (policy->>'allow_multiple')::boolean = false
			)`
		err = tx.QueryRowContext(r.Context(), query, orgPath).Scan(&isExclusive)
	}

	if isExclusive && orgPath != "" {
		// Finn ut hvilken organisasjon som håndhever regelen og blokkerer oss
		var conflictID string
		var conflictOrgName string
		// Vi sjekker om brukeren har et medlemskap i NOEN org som er i konflikt med vår eksklusive gren
		// En konflikt oppstår hvis vi er i samme gren som en eksisterende medlems-org
		query := `
			SELECT m.id, o.name 
			FROM membership_view m
			JOIN organization_hierarchy o ON m.org_id = o.id
			WHERE m.user_id = $1 AND (o.path <@ (
				SELECT path FROM organization_hierarchy WHERE path @> $2 AND (policy->>'allow_multiple')::boolean = false LIMIT 1
			) OR o.path @> (
				SELECT path FROM organization_hierarchy WHERE path @> $2 AND (policy->>'allow_multiple')::boolean = false LIMIT 1
			))`
		
		err = tx.QueryRowContext(r.Context(), query, userID, orgPath).Scan(&conflictID, &conflictOrgName)
		if err == nil {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"error": fmt.Sprintf("User is already a member in the exclusive branch controlled by %s", conflictOrgName),
				"id":    conflictID,
			})
			return
		}
	} else {
		// Standard sjekk: Bare duplikat i nøyaktig samme org
		var existingID string
		err = tx.QueryRowContext(r.Context(), "SELECT id FROM membership_view WHERE user_id = $1 AND org_id = $2", userID, req.OrgID).Scan(&existingID)
		if err == nil {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"error": "User is already a member of this organization", "id": existingID})
			return
		}
	}

	// 4. Opprett medlemskap (innenfor tx for å sikre at låsen holdes til commit)
	membershipID := uuid.New().String()
	membershipEvent := domain.MembershipCreated{
		ID:        membershipID,
		UserID:    userID,
		OrgID:     req.OrgID,
		Role:      "member",
		Metadata:  req.Metadata,
		Timestamp: time.Now(),
	}

	if err := s.eventStore.Append(r.Context(), membershipID, 1, membershipEvent); err != nil {
		http.Error(w, "Failed to create membership: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit registration", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": membershipID, "user_id": userID})
}

type UpdateMemberRequest struct {
	ID            string         `json:"id"`
	UpdatedFields map[string]any `json:"updated_fields"`
}

func (s *Server) UpdateMemberHandler(w http.ResponseWriter, r *http.Request) {
	var req UpdateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Sjekk om det er en medlemskap-oppdatering
	var version int
	err := s.db.QueryRowContext(r.Context(), "SELECT COALESCE(MAX(version), 0) FROM event_store WHERE aggregate_id = $1", req.ID).Scan(&version)
	if err != nil {
		http.Error(w, "Membership not found", http.StatusNotFound)
		return
	}

	event := domain.MembershipUpdated{
		ID:            req.ID,
		UpdatedFields: req.UpdatedFields,
		Timestamp:     time.Now(),
	}

	if err := s.eventStore.Append(r.Context(), req.ID, version+1, event); err != nil {
		http.Error(w, "Failed to save event: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type ShredMemberRequest struct {
	ID string `json:"id"`
}

func (s *Server) ShredMemberHandler(w http.ResponseWriter, r *http.Request) {
	var req ShredMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var version int
	err := s.db.QueryRowContext(r.Context(), "SELECT COALESCE(MAX(version), 0) FROM event_store WHERE aggregate_id = $1", req.ID).Scan(&version)
	if err != nil {
		http.Error(w, "Membership not found", http.StatusNotFound)
		return
	}

	event := domain.MembershipShredded{
		ID:        req.ID,
		Timestamp: time.Now(),
	}

	if err := s.eventStore.Append(r.Context(), req.ID, version+1, event); err != nil {
		http.Error(w, "Failed to save event: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) GetOrganizationHierarchyHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id, name, parent_id, path::text, policy 
		FROM organization_hierarchy 
		ORDER BY path ASC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type OrgNode struct {
		ID       string         `json:"id"`
		Name     string         `json:"name"`
		ParentID sql.NullString `json:"parent_id"`
		Path     string         `json:"path"`
		Policy   map[string]any `json:"policy"`
	}

	var nodes []OrgNode
	for rows.Next() {
		var n OrgNode
		var policyJSON []byte
		if err := rows.Scan(&n.ID, &n.Name, &n.ParentID, &n.Path, &policyJSON); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.Unmarshal(policyJSON, &n.Policy)
		nodes = append(nodes, n)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

type CreateOrgRequest struct {
	Name     string         `json:"name"`
	ParentID string         `json:"parent_id"`
	Policy   map[string]any `json:"policy"`
}

func (s *Server) CreateOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	var path string

	if req.ParentID != "" {
		var parentPath string
		err := s.db.QueryRowContext(r.Context(), "SELECT path FROM organization_hierarchy WHERE id = $1", req.ParentID).Scan(&parentPath)
		if err != nil {
			http.Error(w, "Parent not found", http.StatusBadRequest)
			return
		}
		path = parentPath + "." + strings.ReplaceAll(req.Name, " ", "_")
	} else {
		path = strings.ReplaceAll(req.Name, " ", "_")
	}

	policyJSON, _ := json.Marshal(req.Policy)
	_, err := s.db.ExecContext(r.Context(), `
		INSERT INTO organization_hierarchy (id, name, parent_id, path, policy)
		VALUES ($1, $2, $3, $4, $5)`,
		id, req.Name, sql.NullString{String: req.ParentID, Valid: req.ParentID != ""}, path, policyJSON)
	
	if err != nil {
		http.Error(w, "Failed to create org: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id, "path": path})
}

func (s *Server) DeleteOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	_, err := s.db.ExecContext(r.Context(), "DELETE FROM organization_hierarchy WHERE id = $1", id)
	if err != nil {
		log.Printf("Failed to delete organization %s: %v", id, err)
		http.Error(w, "Failed to delete: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully deleted organization %s (Cascaded dependencies)", id)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) GetTreasuryReportHandler(w http.ResponseWriter, r *http.Request) {
	// Hent alle organisasjoner med deres lokale saldo-sum
	// Vi bruker en ltree-spørring for å aggregere oppover
	query := `
		WITH local_sums AS (
			SELECT org_id, SUM(balance) as local_balance
			FROM membership_view
			GROUP BY org_id
		)
		SELECT 
			o.id, 
			o.name, 
			o.path::text, 
			COALESCE(ls.local_balance, 0) as local_balance,
			COALESCE((
				SELECT SUM(m.balance)
				FROM membership_view m
				JOIN organization_hierarchy child ON m.org_id = child.id
				WHERE child.path <@ o.path
			), 0) as total_branch_balance
		FROM organization_hierarchy o
		LEFT JOIN local_sums ls ON o.id = ls.org_id
		ORDER BY o.path ASC`
	
	rows, err := s.db.QueryContext(r.Context(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type TreasuryResponse struct {
		ID                  string `json:"id"`
		Name                string `json:"name"`
		Path                string `json:"path"`
		LocalBalance        int    `json:"local_balance"`
		TotalBranchBalance int    `json:"total_branch_balance"`
	}

	var report []TreasuryResponse
	for rows.Next() {
		var t TreasuryResponse
		if err := rows.Scan(&t.ID, &t.Name, &t.Path, &t.LocalBalance, &t.TotalBranchBalance); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		report = append(report, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (s *Server) GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Hent medlemsvekst per måned siste år
	query := `
		SELECT 
			TO_CHAR(created_at, 'YYYY-MM') as month,
			COUNT(*) as count
		FROM event_store
		WHERE event_type = 'MembershipCreated'
		AND created_at > NOW() - INTERVAL '12 months'
		GROUP BY month
		ORDER BY month ASC`
	
	rows, err := s.db.QueryContext(r.Context(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type StatRow struct {
		Label string `json:"label"`
		Value int    `json:"value"`
	}

	var stats []StatRow
	for rows.Next() {
		var s StatRow
		if err := rows.Scan(&s.Label, &s.Value); err != nil {
			continue
		}
		stats = append(stats, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) GetFormsHandler(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	query := "SELECT id, org_id, title, schema FROM forms"
	var args []any
	if orgID != "" && orgID != "undefined" {
		query += " WHERE org_id = $1"
		args = append(args, orgID)
	}

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		log.Printf("GetForms error: %v (Query: %s, OrgID: %s)\n", err, query, orgID)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Form struct {
		ID     string         `json:"id"`
		OrgID  string         `json:"org_id"`
		Title  string         `json:"title"`
		Schema map[string]any `json:"schema"`
	}

	var forms []Form
	for rows.Next() {
		var f Form
		var schemaJSON []byte
		if err := rows.Scan(&f.ID, &f.OrgID, &f.Title, &schemaJSON); err != nil {
			continue
		}
		json.Unmarshal(schemaJSON, &f.Schema)
		forms = append(forms, f)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(forms)
}

type FormResponseRequest struct {
	FormID  string         `json:"form_id"`
	Answers map[string]any `json:"answers"`
}

func (s *Server) SubmitFormResponseHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(UserIDKey).(string)
	var req FormResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	answersJSON, _ := json.Marshal(req.Answers)
	_, err := s.db.ExecContext(r.Context(), `
		INSERT INTO form_responses (form_id, user_id, answers)
		VALUES ($1, $2, $3)`,
		req.FormID, userID, answersJSON)
	
	if err != nil {
		http.Error(w, "Failed to submit: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type CreateFormRequest struct {
	OrgID  string         `json:"org_id"`
	Title  string         `json:"title"`
	Schema map[string]any `json:"schema"`
}

func (s *Server) CreateFormHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateFormRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if req.OrgID == "" || req.Title == "" || req.Schema == nil {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	schemaJSON, _ := json.Marshal(req.Schema)
	id := uuid.New().String()

	_, err := s.db.ExecContext(r.Context(), `
		INSERT INTO forms (id, org_id, title, schema)
		VALUES ($1, $2, $3, $4)`,
		id, req.OrgID, req.Title, schemaJSON)

	if err != nil {
		log.Printf("CreateForm error: %v (OrgID: %s)\n", err, req.OrgID)
		http.Error(w, "Failed to create form: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (s *Server) GetMembersAsOfHandler(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	membersMap, err := reports.GetMembersAsOf(r.Context(), s.db, targetDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userOrgID, _ := r.Context().Value(OrgIDKey).(string)

	var members = []reports.HistoricalMember{}
	for _, m := range membersMap {
		if userOrgID == "" || m.OrgID == userOrgID {
			members = append(members, *m)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}
