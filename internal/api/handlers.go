package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"register/internal/domain"
	"register/internal/reports"
	"register/internal/storage"
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

func (s *Server) GetMembersHandler(w http.ResponseWriter, r *http.Request) {
	orgID, _ := r.Context().Value(OrgIDKey).(string)

	query := "SELECT id, org_id, name, email, status, metadata FROM member_view"
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

	var members = []domain.Member{}
	for rows.Next() {
		var m domain.Member
		var metadata []byte
		if err := rows.Scan(&m.ID, &m.OrgID, &m.Name, &m.Email, &m.Status, &metadata); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.Unmarshal(metadata, &m.Metadata)
		members = append(members, m)
	}

	// DEBUG LOG
	fmt.Printf("GetMembersHandler: Found %d members for OrgID: '%s'\n", len(members), orgID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}


func (s *Server) GetOrganizationsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), "SELECT DISTINCT org_id::text FROM member_view UNION SELECT id::text FROM organization_hierarchy")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orgs []string
	for rows.Next() {
		var org string
		if err := rows.Scan(&org); err != nil {
			continue
		}
		orgs = append(orgs, org)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orgs)
}

type RegisterMemberRequest struct {
	Name     string         `json:"name"`
	Email    string         `json:"email"`
	OrgID    string         `json:"org_id"`
	Metadata map[string]any `json:"metadata"`
}

func (s *Server) RegisterMemberHandler(w http.ResponseWriter, r *http.Request) {
	userOrgID, _ := r.Context().Value(OrgIDKey).(string)
	// Fjernet streng sjekk på userOrgID for å tillate globale admins å registrere medlemmer

	var req RegisterMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	targetOrgID := req.OrgID
	if targetOrgID == "" {
		targetOrgID = userOrgID
	}

	memberID := uuid.New().String()
	event := domain.MemberRegistered{
		ID:        memberID,
		Name:      req.Name,
		Email:     req.Email,
		OrgID:     targetOrgID,
		Metadata:  req.Metadata,
		Timestamp: time.Now(),
	}

	if err := s.eventStore.Append(r.Context(), memberID, 1, event); err != nil {
		http.Error(w, "Failed to save event: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": memberID})
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

	// Get current version for optimistic concurrency
	var version int
	err := s.db.QueryRowContext(r.Context(), "SELECT COALESCE(MAX(version), 0) FROM event_store WHERE aggregate_id = $1", req.ID).Scan(&version)
	if err != nil {
		http.Error(w, "Member not found", http.StatusNotFound)
		return
	}

	event := domain.MemberUpdated{
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
		http.Error(w, "Member not found", http.StatusNotFound)
		return
	}

	event := domain.MemberShredded{
		ID:        req.ID,
		Timestamp: time.Now(),
	}

	if err := s.eventStore.Append(r.Context(), req.ID, version+1, event); err != nil {
		http.Error(w, "Failed to save event: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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

	var members = []domain.Member{}
	for _, m := range membersMap {
		// Hvis userOrgID er tom, er det en global admin som kan se alt.
		// Ellers må org_id matche.
		if userOrgID == "" || m.OrgID == userOrgID {
			members = append(members, *m)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}
