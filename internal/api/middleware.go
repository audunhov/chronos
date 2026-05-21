package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"register/internal/domain"
)

const (
	UserIDKey        = domain.UserIDKey
	OrgIDKey         = domain.OrgIDKey
	RoleKey          = domain.RoleKey
	CorrelationIDKey = domain.CorrelationIDKey
)

func (s *Server) CorrelationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Correlation-ID")
		if id == "" {
			id = uuid.New().String()
		}
		
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = strings.Split(forwarded, ",")[0]
		}
		
		ua := r.UserAgent()

		ctx := context.WithValue(r.Context(), CorrelationIDKey, id)
		ctx = context.WithValue(ctx, domain.IPAddressKey, ip)
		ctx = context.WithValue(ctx, domain.UserAgentKey, ua)
		
		w.Header().Set("X-Correlation-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			// In development/test, we can be flexible with localhost/frontend origins
			if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://frontend:") {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Credentials", "false")
			}
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token, X-Org-ID")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap ResponseWriter to capture status code
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		
		next.ServeHTTP(wrapped, r)
		
		log.Printf("[%s] %s %d (%v)", r.Method, r.URL.Path, wrapped.status, time.Since(start))
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for public endpoints if they somehow end up here
		if strings.HasPrefix(r.URL.Path, "/auth/") || strings.HasPrefix(r.URL.Path, "/health/") {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			jwtSecret := os.Getenv("JWT_SECRET")
			if jwtSecret == "" {
				return nil, fmt.Errorf("JWT_SECRET not set")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid claims", http.StatusUnauthorized)
			return
		}

		userID, _ := claims["sub"].(string)
		
		// Finn Org ID: JWT vinner over header (som per eksisterende tester)
		targetOrgID, _ := claims["org_id"].(string)
		if targetOrgID == "" {
			targetOrgID = r.Header.Get("X-Org-ID")
		}
		
		// Valider UUID før bruk i DB
		isUUID := false
		if targetOrgID != "" {
			if _, err := uuid.Parse(targetOrgID); err == nil {
				isUUID = true
			}
		}

		var role string
		if isUUID && s.db != nil {
			// Sjekk hierarkisk tilgang: Har brukeren en rolle i denne orgen eller overordnede?
			query := `
				SELECT ra.role_type 
				FROM role_assignments ra
				JOIN organization_hierarchy target ON target.id = $1
				JOIN organization_hierarchy assigned ON ra.org_id = assigned.id
				WHERE ra.user_id = $2 AND assigned.path @> target.path
				LIMIT 1`
			
			err = s.db.QueryRowContext(r.Context(), query, targetOrgID, userID).Scan(&role)
			if err != nil {
				// Ingen rolle funnet i DB, sjekk om JWT har en rolle
				role, _ = claims["role"].(string)
				if role == "" {
					role = "user"
				}
			}
		} else {
			// Fallback hvis vi ikke har DB (f.eks i tester) eller ingen target org
			role, _ = claims["role"].(string)
			if role == "" {
				role = "user"
			}
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, OrgIDKey, targetOrgID)
		ctx = context.WithValue(ctx, RoleKey, role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
