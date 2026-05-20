package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"register/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	OrgID    string `json:"org_id"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	User        User   `json:"user"`
}

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	OrgID string `json:"org_id"`
	Role  string `json:"role"`
}

func (s *Server) SignupHandler(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	var userID = uuid.New().String()
	event := domain.UserCreated{
		ID:           userID,
		Email:        req.Email,
		Name:         req.Email, // Bruk email som navn inntil videre
		PasswordHash: string(hashedPassword),
		Timestamp:    time.Now(),
	}

	if err := s.eventStore.Append(r.Context(), userID, 1, event); err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Hvis de sendte med en OrgID, opprett et medlemskap også
	if req.OrgID != "" {
		membershipID := uuid.New().String()
		membershipEvent := domain.MembershipCreated{
			ID:        membershipID,
			UserID:    userID,
			OrgID:     req.OrgID,
			Role:      "member",
			Metadata:  nil,
			Timestamp: time.Now(),
		}
		if err := s.eventStore.Append(r.Context(), membershipID, 1, membershipEvent); err != nil {
			// Vi logger feilen, men brukeren er allerede opprettet
			fmt.Printf("Failed to create initial membership: %v\n", err)
		}
	}

	// For simplicity, we just return success. In a real app, we might issue a token here.
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created. Please log in."})
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var user User
	var passwordHash string

	err := s.db.QueryRowContext(r.Context(),
		"SELECT id, email, password_hash FROM users WHERE email = $1",
		req.Email).Scan(&user.ID, &user.Email, &passwordHash)

	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// For nå setter vi ingen org_id i tokenet ved login, 
	// brukeren må sende X-Org-ID for å aksessere spesifikke orger.
	// En fremtidig forbedring er å sette en "default" org her.
	user.Role = "user" 

	token, err := s.generateJWT(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		AccessToken: token,
		User:        user,
	})
}

func (s *Server) RequestMagicLinkHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var userID string
	err := s.db.QueryRowContext(r.Context(), "SELECT id FROM users WHERE email = $1", req.Email).Scan(&userID)
	if err == sql.ErrNoRows {
		// Vi returnerer 200 OK uansett for å ikke lekke e-poster
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "If you have an account, a link has been sent."})
		return
	}

	token := uuid.New().String()
	expiresAt := time.Now().Add(15 * time.Minute)

	_, err = s.db.ExecContext(r.Context(), "INSERT INTO magic_links (token, user_id, expires_at) VALUES ($1, $2, $3)", token, userID, expiresAt)
	if err != nil {
		http.Error(w, "Failed to create magic link", http.StatusInternalServerError)
		return
	}

	// I en ekte app ville vi sendt en e-post her. 
	// For nå logger vi det til konsollen så brukeren kan simulere det.
	fmt.Printf("MAGIC LINK for %s: http://localhost:5173/magic-login?token=%s\n", req.Email, token)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Magic link generated (Check server logs)."})
}

func (s *Server) LoginWithMagicLinkHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	var userID string
	err := s.db.QueryRowContext(r.Context(), `
		DELETE FROM magic_links 
		WHERE token = $1 AND expires_at > NOW() 
		RETURNING user_id`, token).Scan(&userID)
	
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	var user User
	err = s.db.QueryRowContext(r.Context(), "SELECT id, email FROM users WHERE id = $1", userID).Scan(&user.ID, &user.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	user.Role = "user"
	jwtToken, err := s.generateJWT(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		AccessToken: jwtToken,
		User:        user,
	})
}

func (s *Server) generateJWT(user User) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET not set")
	}

	claims := jwt.MapClaims{
		"sub":     user.ID,
		"email":   user.Email,
		"org_id":  user.OrgID,
		"role":    user.Role,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
