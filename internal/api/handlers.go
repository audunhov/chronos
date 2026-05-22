package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"register/internal/domain"
	"register/internal/reports"
	"register/internal/storage"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Server struct {
	db          *sql.DB
	eventStore  *storage.EventStore
	auditLogger *storage.AuditLogger
}

func NewServer(db *sql.DB, es *storage.EventStore) *Server {
	return &Server{
		db:          db,
		eventStore:  es,
		auditLogger: storage.NewAuditLogger(db),
	}
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

func (s *Server) GetUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Only admins can see other users
	role, _ := r.Context().Value(RoleKey).(string)
	if role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	if _, err := uuid.Parse(userID); err != nil {
		http.Error(w, "Invalid UUID format", http.StatusBadRequest)
		return
	}

	var u domain.User
	err := s.db.QueryRowContext(r.Context(), "SELECT id, email, name, created_at FROM users WHERE id = $1", userID).
		Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt)
	
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func (s *Server) GetUserMembershipsHandler(w http.ResponseWriter, r *http.Request) {
	// Only admins can see other users
	role, _ := r.Context().Value(RoleKey).(string)
	if role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	if _, err := uuid.Parse(userID); err != nil {
		http.Error(w, "Invalid UUID format", http.StatusBadRequest)
		return
	}

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

	type UserMembershipResponse struct {
		ID        string    `json:"id"`
		OrgID     string    `json:"org_id"`
		OrgName   string    `json:"org_name"`
		Status    string    `json:"status"`
		Role      string    `json:"role"`
		Balance   int       `json:"balance"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	var memberships []UserMembershipResponse
	for rows.Next() {
		var m UserMembershipResponse
		if err := rows.Scan(&m.ID, &m.OrgID, &m.OrgName, &m.Status, &m.Role, &m.Balance, &m.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		memberships = append(memberships, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(memberships)
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

	s.auditLogger.Log(r.Context(), "MEMBER_LIST_VIEWED", "", map[string]any{
		"org_id": orgID,
	})

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
		SELECT er.id, er.org_id, er.trigger_event, er.trigger_aggregate_id, er.action_type, er.config,
		       (er.org_id != $1) as is_inherited
		FROM event_reactions er
		JOIN organization_hierarchy o ON er.org_id = o.id
		WHERE o.path @> $2::ltree`
	
	args := []any{orgID, path}
	
	triggerAggregateID := r.URL.Query().Get("trigger_aggregate_id")
	if triggerAggregateID != "" {
		query += " AND (er.trigger_aggregate_id = $3 OR er.trigger_aggregate_id IS NULL)"
		args = append(args, triggerAggregateID)
	}
	
	query += " ORDER BY o.path ASC"

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ReactionResponse struct {
		ID                 string         `json:"id"`
		OrgID              string         `json:"org_id"`
		TriggerEvent       string         `json:"trigger_event"`
		TriggerAggregateID *string        `json:"trigger_aggregate_id"`
		ActionType         string         `json:"action_type"`
		Config             map[string]any `json:"config"`
		IsInherited        bool           `json:"is_inherited"`
	}

	var reactions []ReactionResponse
	for rows.Next() {
		var r ReactionResponse
		var configJSON []byte
		var aggID sql.NullString
		if err := rows.Scan(&r.ID, &r.OrgID, &r.TriggerEvent, &aggID, &r.ActionType, &configJSON, &r.IsInherited); err != nil {
			continue
		}
		if aggID.Valid { r.TriggerAggregateID = &aggID.String }
		json.Unmarshal(configJSON, &r.Config)
		reactions = append(reactions, r)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reactions)
}

func (s *Server) CreateReactionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrgID              string         `json:"org_id"`
		TriggerEvent       string         `json:"trigger_event"`
		TriggerAggregateID *string        `json:"trigger_aggregate_id"`
		ActionType         string         `json:"action_type"`
		Config             map[string]any `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	event := domain.ReactionCreated{
		ID:                 id,
		OrgID:              req.OrgID,
		TriggerEvent:       req.TriggerEvent,
		TriggerAggregateID: req.TriggerAggregateID,
		ActionType:         req.ActionType,
		Config:             req.Config,
		Timestamp:          time.Now(),
	}

	if err := s.eventStore.Append(r.Context(), id, 1, event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}
func (s *Server) GlobalSearchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if len(q) < 2 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	searchQuery := "%" + strings.ToLower(q) + "%"

	query := `
		SELECT id, 'user' as type, name as title, email as subtitle
		FROM users
		WHERE LOWER(name) LIKE $1 OR LOWER(email) LIKE $1
		UNION ALL
		SELECT id, 'org' as type, name as title, path::text as subtitle
		FROM organization_hierarchy
		WHERE LOWER(name) LIKE $1
		UNION ALL
		SELECT id, 'pipeline' as type, trigger_event as title, id::text as subtitle
		FROM event_reactions
		WHERE LOWER(trigger_event) LIKE $1 OR id::text LIKE $1
		LIMIT 20
	`

	rows, err := s.db.QueryContext(r.Context(), query, searchQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type SearchResult struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Title    string `json:"title"`
		Subtitle string `json:"subtitle"`
	}

	var results []SearchResult
	for rows.Next() {
		var res SearchResult
		if err := rows.Scan(&res.ID, &res.Type, &res.Title, &res.Subtitle); err == nil {
			results = append(results, res)
		}
	}

	if results == nil {
		results = []SearchResult{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *Server) GetSecretsHandler(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		http.Error(w, "Missing org_id", http.StatusBadRequest)
		return
	}

	rows, err := s.db.QueryContext(r.Context(), "SELECT id, name FROM secrets WHERE org_id = $1 ORDER BY name", orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Secret struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	var results []Secret
	for rows.Next() {
		var sec Secret
		if err := rows.Scan(&sec.ID, &sec.Name); err == nil {
			results = append(results, sec)
		}
	}

	if results == nil {
		results = []Secret{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *Server) CreateSecretHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrgID string `json:"org_id"`
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	encrypted, err := storage.Encrypt([]byte(req.Value))
	if err != nil {
		http.Error(w, "Encryption failed", http.StatusInternalServerError)
		return
	}

	var id string
	err = s.db.QueryRowContext(r.Context(), `
		INSERT INTO secrets (org_id, name, encrypted_value)
		VALUES ($1, $2, $3)
		RETURNING id`,
		req.OrgID, req.Name, encrypted).Scan(&id)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (s *Server) GetOrgansHandler(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		http.Error(w, "Missing org_id", http.StatusBadRequest)
		return
	}

	if _, err := uuid.Parse(orgID); err != nil {
		http.Error(w, "Invalid UUID format", http.StatusBadRequest)
		return
	}

	query := `
		SELECT o.id, o.org_id, o.name, o.parent_organ_id,
		       (SELECT COUNT(*) FROM role_assignments WHERE organ_id = o.id) as member_count
		FROM organs o 
		WHERE o.org_id = $1`

	rows, err := s.db.QueryContext(r.Context(), query, orgID)
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
		MemberCount   int     `json:"member_count"`
	}

	var organs []OrganResponse
	for rows.Next() {
		var o OrganResponse
		var parentID sql.NullString
		if err := rows.Scan(&o.ID, &o.OrgID, &o.Name, &parentID, &o.MemberCount); err != nil {
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

func (s *Server) GetOrganMembersHandler(w http.ResponseWriter, r *http.Request) {
	organID := r.URL.Query().Get("organ_id")
	if organID == "" {
		http.Error(w, "Missing organ_id", http.StatusBadRequest)
		return
	}

	query := `
		SELECT ra.id, ra.user_id, ra.role_type, u.name, u.email, ra.created_at
		FROM role_assignments ra
		JOIN users u ON ra.user_id = u.id
		WHERE ra.organ_id = $1`
	
	rows, err := s.db.QueryContext(r.Context(), query, organID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type MemberRoleResponse struct {
		ID        string    `json:"id"`
		UserID    string    `json:"user_id"`
		RoleType  string    `json:"role_type"`
		Name      string    `json:"name"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
	}

	var members []MemberRoleResponse
	for rows.Next() {
		var m MemberRoleResponse
		if err := rows.Scan(&m.ID, &m.UserID, &m.RoleType, &m.Name, &m.Email, &m.CreatedAt); err != nil {
			continue
		}
		members = append(members, m)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func (s *Server) AssignOrganMemberHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrganID  string `json:"organ_id"`
		UserID   string `json:"user_id"`
		RoleType string `json:"role_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	// Finn org_id for organet
	var orgID string
	err := s.db.QueryRowContext(r.Context(), "SELECT org_id FROM organs WHERE id = $1", req.OrganID).Scan(&orgID)
	if err != nil {
		http.Error(w, "Organ not found", http.StatusNotFound)
		return
	}

	id := uuid.New().String()
	event := domain.RoleAssigned{
		ID:        id,
		UserID:    req.UserID,
		OrgID:     orgID,
		OrganID:   &req.OrganID,
		RoleType:  req.RoleType,
		Timestamp: time.Now(),
	}

	if err := s.eventStore.Append(r.Context(), id, 1, event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) RevokeOrganMemberHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing assignment ID", http.StatusBadRequest)
		return
	}

	event := domain.RoleRevoked{
		ID:        id,
		Timestamp: time.Now(),
	}

	// Vi append-er til den samme aggregate_id (assignment id)
	// Trenger å finne gjeldende versjon
	var version int
	s.db.QueryRowContext(r.Context(), "SELECT COALESCE(MAX(version), 0) FROM event_store WHERE aggregate_id = $1", id).Scan(&version)

	if err := s.eventStore.Append(r.Context(), id, version+1, event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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

	id := uuid.New().String()
	event := domain.OrganCreated{
		ID:            id,
		OrgID:         req.OrgID,
		Name:          req.Name,
		ParentOrganID: req.ParentOrganID,
		Timestamp:     time.Now(),
	}

	if err := s.eventStore.Append(r.Context(), id, 1, event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) DeleteOrganHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	event := domain.OrganDeleted{
		ID:        id,
		Timestamp: time.Now(),
	}

	var version int
	s.db.QueryRowContext(r.Context(), "SELECT COALESCE(MAX(version), 0) FROM event_store WHERE aggregate_id = $1", id).Scan(&version)

	if err := s.eventStore.Append(r.Context(), id, version+1, event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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
			// Siden s.eventStore.Append bruker sin egen tx, og vi senere må bruke
			// en tx for hierarkisjekk, burde vi ideelt sett starte tx helt øverst.
			// Men for å gjøre dette enkelt og unngå stor refaktorering akkurat her:
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
		var controllingOrgID string
		var controllingOrgName string
		queryPolicy := `
			SELECT id, name 
			FROM organization_hierarchy 
			WHERE path @> $1 AND (policy->>'allow_multiple')::boolean = false
			ORDER BY nlevel(path) DESC
			LIMIT 1`
		err = tx.QueryRowContext(r.Context(), queryPolicy, orgPath).Scan(&controllingOrgID, &controllingOrgName)
		if err != nil {
			http.Error(w, "Policy lookup failed", http.StatusInternalServerError)
			return
		}

		var conflictID string
		var conflictOrgName string
		queryConflict := `
			SELECT m.id, o.name 
			FROM membership_view m
			JOIN organization_hierarchy o ON m.org_id = o.id
			WHERE m.user_id = $1 AND (o.path <@ (
				SELECT path FROM organization_hierarchy WHERE id = $2
			) OR o.path @> (
				SELECT path FROM organization_hierarchy WHERE id = $2
			))`
		
		err = tx.QueryRowContext(r.Context(), queryConflict, userID, controllingOrgID).Scan(&conflictID, &conflictOrgName)
		if err == nil {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"error": fmt.Sprintf("User is already a member in the exclusive branch controlled by %s (Member in: %s)", controllingOrgName, conflictOrgName),
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

	if err := s.eventStore.AppendWithTx(r.Context(), tx, membershipID, 1, membershipEvent); err != nil {
		http.Error(w, "Failed to create membership: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit registration", http.StatusInternalServerError)
		return
	}

	s.auditLogger.Log(r.Context(), "MEMBERSHIP_REGISTERED", membershipID, map[string]any{
		"org_id": req.OrgID,
		"email":  req.Email,
	})

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

	s.auditLogger.Log(r.Context(), "MEMBERSHIP_SHREDDED", req.ID, nil)

	w.WriteHeader(http.StatusOK)
	}

	func (s *Server) DeleteFormHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	event := domain.FormDeleted{
		ID:        id,
		Timestamp: time.Now(),
	}

	var version int
	s.db.QueryRowContext(r.Context(), "SELECT COALESCE(MAX(version), 0) FROM event_store WHERE aggregate_id = $1", id).Scan(&version)

	if err := s.eventStore.Append(r.Context(), id, version+1, event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) GetFormResponsesHandler(w http.ResponseWriter, r *http.Request) {
	// Only admins can see form responses
	role, _ := r.Context().Value(RoleKey).(string)
	if role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	formID := r.URL.Query().Get("form_id")
	if formID == "" {
		http.Error(w, "Missing form_id", http.StatusBadRequest)
		return
	}

	if _, err := uuid.Parse(formID); err != nil {
		http.Error(w, "Invalid UUID format", http.StatusBadRequest)
		return
	}

	query := `
		SELECT fr.id, fr.form_id, fr.user_id, u.name as user_name, u.email as user_email, fr.answers, fr.created_at
		FROM form_responses fr
		JOIN users u ON fr.user_id = u.id
		WHERE fr.form_id = $1
		ORDER BY fr.created_at DESC`
	
	rows, err := s.db.QueryContext(r.Context(), query, formID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type FormResponse struct {
		ID        string         `json:"id"`
		FormID    string         `json:"form_id"`
		UserID    string         `json:"user_id"`
		UserName  string         `json:"user_name"`
		UserEmail string         `json:"user_email"`
		Answers   map[string]any `json:"answers"`
		CreatedAt time.Time      `json:"created_at"`
	}

	var responses []FormResponse
	for rows.Next() {
		var resp FormResponse
		var answersJSON []byte
		if err := rows.Scan(&resp.ID, &resp.FormID, &resp.UserID, &resp.UserName, &resp.UserEmail, &answersJSON, &resp.CreatedAt); err != nil {
			continue
		}
		json.Unmarshal(answersJSON, &resp.Answers)
		responses = append(responses, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
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

	var parentID *string
	if req.ParentID != "" {
		parentID = &req.ParentID
	}

	event := domain.OrganizationCreated{
		ID:        id,
		Name:      req.Name,
		ParentID:  parentID,
		Path:      path,
		Policy:    req.Policy,
		Timestamp: time.Now(),
	}

	if err := s.eventStore.Append(r.Context(), id, 1, event); err != nil {
		http.Error(w, "Failed to save event: "+err.Error(), http.StatusInternalServerError)
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

	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "Invalid UUID format", http.StatusBadRequest)
		return
	}

	// Find current version
	var version int
	err := s.db.QueryRowContext(r.Context(), "SELECT COALESCE(MAX(version), 0) FROM event_store WHERE aggregate_id = $1", id).Scan(&version)
	if err != nil {
		version = 0 // Might not be event-sourced yet if legacy
	}

	event := domain.OrganizationDeleted{
		ID:        id,
		Timestamp: time.Now(),
	}

	if err := s.eventStore.Append(r.Context(), id, version+1, event); err != nil {
		http.Error(w, "Failed to save event: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully deleted organization %s (Cascaded dependencies via event)", id)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) GetAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	actorID := r.URL.Query().Get("actor_id")
	targetID := r.URL.Query().Get("target_id")
	
	query := `
		SELECT 
			a.id, a.correlation_id, a.action, a.target_id, a.detail, 
			a.ip_address, a.user_agent, a.created_at,
			u.email as actor_email, o.name as org_name
		FROM audit_logs a
		LEFT JOIN users u ON a.actor_id = u.id
		LEFT JOIN organization_hierarchy o ON a.org_id = o.id`
	
	var args []any
	var conditions []string

	// Hjelpefunksjon for å validere UUID før den sendes til databasen
	isUUID := func(s string) bool {
		_, err := uuid.Parse(s)
		return err == nil
	}

	if orgID != "" && isUUID(orgID) {
		conditions = append(conditions, fmt.Sprintf("a.org_id = $%d", len(args)+1))
		args = append(args, orgID)
	}
	if actorID != "" && isUUID(actorID) {
		conditions = append(conditions, fmt.Sprintf("a.actor_id = $%d", len(args)+1))
		args = append(args, actorID)
	}
	if targetID != "" && isUUID(targetID) {
		conditions = append(conditions, fmt.Sprintf("a.target_id = $%d", len(args)+1))
		args = append(args, targetID)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	
	query += " ORDER BY a.created_at DESC LIMIT 100"

	rows, err := s.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type AuditEntry struct {
		ID            string         `json:"id"`
		CorrelationID string         `json:"correlation_id"`
		Action        string         `json:"action"`
		TargetID      *string        `json:"target_id"`
		Detail        map[string]any `json:"detail"`
		IP            string         `json:"ip_address"`
		UserAgent     string         `json:"user_agent"`
		CreatedAt     time.Time      `json:"created_at"`
		ActorEmail    *string        `json:"actor_email"`
		OrgName       *string        `json:"org_name"`
	}

	var logs []AuditEntry
	for rows.Next() {
		var l AuditEntry
		var targetID, actorEmail, orgName sql.NullString
		var detailJSON []byte
		if err := rows.Scan(&l.ID, &l.CorrelationID, &l.Action, &targetID, &detailJSON, &l.IP, &l.UserAgent, &l.CreatedAt, &actorEmail, &orgName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.Unmarshal(detailJSON, &l.Detail)
		
		if targetID.Valid { l.TargetID = &targetID.String }
		if actorEmail.Valid { l.ActorEmail = &actorEmail.String }
		if orgName.Valid { l.OrgName = &orgName.String }

		logs = append(logs, l)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
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

func (s *Server) GetPipelineExecutionsHandler(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	if orgID == "" {
		http.Error(w, "Missing org_id", http.StatusBadRequest)
		return
	}

	query := `
		SELECT e.id, e.pipeline_id, e.trigger_event, e.status, e.logs, e.executed_at
		FROM pipeline_executions e
		JOIN event_reactions r ON e.pipeline_id = r.id
		WHERE r.org_id = $1
		ORDER BY e.executed_at DESC
		LIMIT 50`

	rows, err := s.db.QueryContext(r.Context(), query, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ExecutionLog struct {
		ID           string           `json:"id"`
		PipelineID   string           `json:"pipeline_id"`
		TriggerEvent string           `json:"trigger_event"`
		Status       string           `json:"status"`
		Logs         []map[string]any `json:"logs"`
		ExecutedAt   time.Time        `json:"executed_at"`
	}

	var executions []ExecutionLog
	for rows.Next() {
		var exec ExecutionLog
		var logsJSON []byte
		if err := rows.Scan(&exec.ID, &exec.PipelineID, &exec.TriggerEvent, &exec.Status, &logsJSON, &exec.ExecutedAt); err != nil {
			continue
		}
		json.Unmarshal(logsJSON, &exec.Logs)
		executions = append(executions, exec)
	}

	if executions == nil {
		executions = []ExecutionLog{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(executions)
}

func (s *Server) TestPipelineHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PipelineID  string                `json:"pipeline_id"`
		Config      *domain.PipelineConfig `json:"config,omitempty"`
		TriggerData map[string]any        `json:"trigger_data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	var dag domain.PipelineConfig
	if req.Config != nil {
		dag = *req.Config
	} else {
		// 1. Fetch the pipeline config from DB
		var configJSON []byte
		err := s.db.QueryRowContext(r.Context(), "SELECT config FROM event_reactions WHERE id = $1", req.PipelineID).Scan(&configJSON)
		if err != nil {
			http.Error(w, "Pipeline not found", http.StatusNotFound)
			return
		}

		if err := json.Unmarshal(configJSON, &dag); err != nil {
			http.Error(w, "Invalid config", http.StatusInternalServerError)
			return
		}
	}

	// 2. Start Transaction
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // Always rollback test runs

	// 3. Setup Test Executor
	worker := storage.NewReactionWorker(s.db) // Just for the operations
	executor := &domain.PipelineExecutor{
		Operations: worker.GetOperations(tx),
	}

	// 4. Run
	trace, err := executor.Execute(dag, req.TriggerData)
	
	res := struct {
		Success bool     `json:"success"`
		Error   string   `json:"error,omitempty"`
		Logs    []string `json:"logs"`
	}{
		Success: err == nil,
		Logs:    []string{"Test run started", "Rolling back changes..."},
	}
	
	for _, t := range trace {
		tBytes, _ := json.Marshal(t)
		res.Logs = append(res.Logs, string(tBytes))
	}

	if err != nil {
		res.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *Server) UpdateReactionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string         `json:"id"`
		Config map[string]any `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	event := domain.ReactionUpdated{
		ID:        req.ID,
		Config:    req.Config,
		Timestamp: time.Now(),
	}

	var version int
	s.db.QueryRowContext(r.Context(), "SELECT COALESCE(MAX(version), 0) FROM event_store WHERE aggregate_id = $1", req.ID).Scan(&version)

	if err := s.eventStore.Append(r.Context(), req.ID, version+1, event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) DeleteReactionHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	event := domain.ReactionDeleted{
		ID:        id,
		Timestamp: time.Now(),
	}

	var version int
	s.db.QueryRowContext(r.Context(), "SELECT COALESCE(MAX(version), 0) FROM event_store WHERE aggregate_id = $1", id).Scan(&version)

	if err := s.eventStore.Append(r.Context(), id, version+1, event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) ReconcileTreasuryHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Parse CSV and look for UUIDs
	body, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	content := string(body)
	
	// Find all PENDING invoices
	rows, err := s.db.QueryContext(r.Context(), "SELECT id, membership_id, amount FROM invoice_view WHERE status = 'PENDING'")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Invoice struct {
		ID           string
		MembershipID string
		Amount       int
	}
	var invoices []Invoice
	for rows.Next() {
		var inv Invoice
		if err := rows.Scan(&inv.ID, &inv.MembershipID, &inv.Amount); err == nil {
			invoices = append(invoices, inv)
		}
	}

	matchedCount := 0
	totalAmount := 0

	for _, inv := range invoices {
		// Very simple matching for prototype: check if the UUID exists in the CSV text
		if strings.Contains(content, inv.ID) {
			// Trigger PaymentReceived event
			event := domain.PaymentReceived{
				ID:        inv.ID,
				Amount:    inv.Amount,
				Timestamp: time.Now(),
			}
			err := s.eventStore.Append(r.Context(), inv.MembershipID, 0, event) // Use version 0 or ignore versioning for this simple implementation
			if err == nil {
				matchedCount++
				totalAmount += inv.Amount
				
				// Update invoice_view status (in a real system the projection worker would do this, but we force it here for immediate feedback)
				s.db.ExecContext(r.Context(), "UPDATE invoice_view SET status = 'PAID' WHERE id = $1", inv.ID)
				s.db.ExecContext(r.Context(), "UPDATE membership_view SET balance = balance + $1 WHERE id = $2", inv.Amount, inv.MembershipID)
			}
		}
	}

	res := map[string]any{
		"matched": matchedCount,
		"amount":  totalAmount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *Server) GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	query := `
		WITH months AS (
			SELECT TO_CHAR(date_trunc('month', d), 'YYYY-MM') as month
			FROM generate_series(NOW() - INTERVAL '35 months', NOW(), '1 month') d
		),
		stats AS (
			SELECT 
				TO_CHAR(created_at, 'YYYY-MM') as month,
				COUNT(*) FILTER (WHERE event_type = 'MembershipCreated') as new_count,
				COUNT(*) FILTER (WHERE event_type = 'MembershipShredded') as churn_count,
				COUNT(*) FILTER (WHERE event_type = 'PaymentReceived') as pay_count
			FROM event_store
			WHERE created_at > NOW() - INTERVAL '36 months'
			GROUP BY month
		)
		SELECT 
			m.month,
			COALESCE(s.new_count, 0),
			COALESCE(s.churn_count, 0),
			SUM(COALESCE(s.new_count, 0) - COALESCE(s.churn_count, 0)) OVER (ORDER BY m.month) as total_active,
			COALESCE(s.pay_count, 0)
		FROM months m
		LEFT JOIN stats s ON m.month = s.month
		ORDER BY m.month`
	
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

	type MultiStats struct {
		NewMembers  []StatRow `json:"new_members"`
		Churn       []StatRow `json:"churn"`
		TotalActive []StatRow `json:"total_active"`
		Payments    []StatRow `json:"payments"`
	}

	var res MultiStats
	for rows.Next() {
		var month string
		var n, c, t, p int
		if err := rows.Scan(&month, &n, &c, &t, &p); err != nil {
			continue
		}
		res.NewMembers = append(res.NewMembers, StatRow{month, n})
		res.Churn = append(res.Churn, StatRow{month, c})
		res.TotalActive = append(res.TotalActive, StatRow{month, t})
		res.Payments = append(res.Payments, StatRow{month, p})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *Server) GetFormsHandler(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("org_id")
	
	if orgID != "" {
		if _, err := uuid.Parse(orgID); err != nil {
			http.Error(w, "Invalid UUID format", http.StatusBadRequest)
			return
		}
	}

	query := "SELECT id, org_id, title, schema FROM forms"
	var args []any
	var conditions []string

	if orgID != "" {
		conditions = append(conditions, "org_id = $1")
		args = append(args, orgID)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
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

	responseID := uuid.New().String()
	event := domain.FormResponseSubmitted{
		FormID:    req.FormID,
		UserID:    userID,
		Answers:   req.Answers,
		Timestamp: time.Now(),
	}

	// We use a unique ResponseID as the aggregate ID to avoid version conflicts
	if err := s.eventStore.Append(r.Context(), responseID, 1, event); err != nil {
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

	id := uuid.New().String()
	event := domain.FormCreated{
		ID:        id,
		OrgID:     req.OrgID,
		Title:     req.Title,
		Schema:    req.Schema,
		Timestamp: time.Now(),
	}

	if err := s.eventStore.Append(r.Context(), id, 1, event); err != nil {
		log.Printf("CreateForm error: %v (OrgID: %s)\n", err, req.OrgID)
		http.Error(w, "Failed to create form: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

type UpdateFormRequest struct {
	ID     string         `json:"id"`
	Title  string         `json:"title"`
	Schema map[string]any `json:"schema"`
}

func (s *Server) UpdateFormHandler(w http.ResponseWriter, r *http.Request) {
	var req UpdateFormRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	// Finn gjeldende versjon
	var version int
	err := s.db.QueryRowContext(r.Context(), "SELECT COALESCE(MAX(version), 0) FROM event_store WHERE aggregate_id = $1", req.ID).Scan(&version)
	if err != nil {
		http.Error(w, "Form not found", http.StatusNotFound)
		return
	}

	event := domain.FormUpdated{
		ID:        req.ID,
		Title:     req.Title,
		Schema:    req.Schema,
		Timestamp: time.Now(),
	}

	if err := s.eventStore.Append(r.Context(), req.ID, version+1, event); err != nil {
		http.Error(w, "Failed to save event: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) SendEmailHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserIDs []string `json:"user_ids"`
		Subject string   `json:"subject"`
		Body    string   `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if len(req.UserIDs) == 0 {
		http.Error(w, "No recipients specified", http.StatusBadRequest)
		return
	}

	// 1. Resolve Emails
	query := `SELECT email FROM users WHERE id = ANY($1)`
	rows, err := s.db.QueryContext(r.Context(), query, pq.Array(req.UserIDs))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err == nil {
			emails = append(emails, email)
		}
	}

	if len(emails) == 0 {
		http.Error(w, "No valid recipients found", http.StatusNotFound)
		return
	}

	// 2. Queue in Outbox
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for _, email := range emails {
		_, err = tx.ExecContext(r.Context(), `
			INSERT INTO email_outbox (recipient_email, subject, body_html, status)
			VALUES ($1, $2, $3, 'PENDING')`,
			email, req.Subject, req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
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
