package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		method         string
		expectedOrigin string
		expectedCreds  string
		expectedStatus int
	}{
		{
			name:           "Allow localhost:5173",
			origin:         "http://localhost:5173",
			method:         "GET",
			expectedOrigin: "http://localhost:5173",
			expectedCreds:  "true",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allow localhost:4173 (E2E)",
			origin:         "http://localhost:4173",
			method:         "GET",
			expectedOrigin: "http://localhost:4173",
			expectedCreds:  "true",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allow frontend:5173",
			origin:         "http://frontend:5173",
			method:         "GET",
			expectedOrigin: "http://frontend:5173",
			expectedCreds:  "true",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "OPTIONS request on localhost",
			origin:         "http://localhost:5173",
			method:         "OPTIONS",
			expectedOrigin: "http://localhost:5173",
			expectedCreds:  "true",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unknown origin defaults to *",
			origin:         "http://malicious.com",
			method:         "GET",
			expectedOrigin: "*",
			expectedCreds:  "false",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No origin defaults to *",
			origin:         "",
			method:         "GET",
			expectedOrigin: "*",
			expectedCreds:  "", // Not set if origin is empty in my new logic
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			handler := CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			gotOrigin := w.Header().Get("Access-Control-Allow-Origin")
			if gotOrigin != tt.expectedOrigin {
				t.Errorf("Expected Origin %s, got %s", tt.expectedOrigin, gotOrigin)
			}

			gotCreds := w.Header().Get("Access-Control-Allow-Credentials")
			if gotCreds != tt.expectedCreds {
				t.Errorf("Expected Credentials %s, got %s", tt.expectedCreds, gotCreds)
			}
		})
	}
}

func TestCORSWithAuthFailure(t *testing.T) {
	// Simulate the chain: CORSMiddleware -> AuthMiddleware -> Final Handler
	// But AuthMiddleware will fail (401)
	
	req := httptest.NewRequest("GET", "/api/members", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	// Missing Authorization header
	
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Should not reach inner handler")
	})
	
	server := &Server{db: nil}
	handler := CORSMiddleware(server.AuthMiddleware(innerHandler))
	
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
	
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Errorf("CORS header lost on 401: got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("CORS Credentials lost on 401: got %s", w.Header().Get("Access-Control-Allow-Credentials"))
	}
}
