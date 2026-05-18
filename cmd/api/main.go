package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"register/internal/api"
	"register/internal/projection"
	"register/internal/storage"
	"syscall"
	"time"
)

func main() {
	db, err := storage.NewDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	eventStore := storage.NewEventStore(db)
	
	// Start Projection Worker
	worker := projection.NewWorker(db)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go worker.Start(ctx)
	log.Println("Projection worker started")

	// Setup API Server
	server := api.NewServer(db, eventStore)
	
	mux := http.NewServeMux()
	
	// API routes with Auth Middleware
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/members", server.GetMembersHandler)
	apiMux.HandleFunc("/organizations", server.GetOrganizationsHandler)
	apiMux.HandleFunc("/commands/register-member", server.RegisterMemberHandler)
	apiMux.HandleFunc("/commands/update-member", server.UpdateMemberHandler)
	apiMux.HandleFunc("/commands/shred-member", server.ShredMemberHandler)
	apiMux.HandleFunc("/reports/as-of", server.GetMembersAsOfHandler)
	
	// Auth routes (Public)
	mux.HandleFunc("/api/health/", server.HealthHandler)
	mux.HandleFunc("/swagger/", server.SwaggerHandler)
	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "openapi.yaml")
	})
	mux.HandleFunc("/api/auth/signup", server.SignupHandler)
	mux.HandleFunc("/api/auth/login", server.LoginHandler)

	// Apply Auth Middleware to all /api/ routes
	// Also apply CORS Middleware to the entire mux
	mux.Handle("/api/", http.StripPrefix("/api", api.AuthMiddleware(apiMux)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: api.CORSMiddleware(mux),
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		cancel()
		
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP shutdown error: %v", err)
		}
	}()

	log.Printf("Server listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}
