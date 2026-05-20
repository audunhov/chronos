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

	// Start Email Worker
	emailWorker := storage.NewEmailWorker(db)
	go emailWorker.Start(ctx)
	log.Println("Email worker started")

	// Start Reaction Worker
	reactionWorker := storage.NewReactionWorker(db)
	go reactionWorker.Start(ctx)
	log.Println("Reaction worker started")

	// Start Payment Worker
	paymentWorker := storage.NewPaymentWorker(db)
	go paymentWorker.Start(ctx)
	log.Println("Payment worker started")

	// Setup API Server
	server := api.NewServer(db, eventStore)
	
	mux := http.NewServeMux()
	
	// API routes with Auth Middleware
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/me/profile", server.GetMyProfileHandler)
	apiMux.HandleFunc("/me/memberships", server.GetMyMembershipsHandler)
	apiMux.HandleFunc("/members", server.GetMembersHandler)
	apiMux.HandleFunc("/organizations", server.GetOrganizationsHandler)
	apiMux.HandleFunc("/organizations/hierarchy", server.GetOrganizationHierarchyHandler)
	apiMux.HandleFunc("/commands/create-organization", server.CreateOrganizationHandler)
	apiMux.HandleFunc("/commands/delete-organization", server.DeleteOrganizationHandler)
	apiMux.HandleFunc("/commands/register-member", server.RegisterMemberHandler)
	apiMux.HandleFunc("/commands/update-member", server.UpdateMemberHandler)
	apiMux.HandleFunc("/commands/shred-member", server.ShredMemberHandler)
	apiMux.HandleFunc("/reports/as-of", server.GetMembersAsOfHandler)
	apiMux.HandleFunc("/reports/treasury", server.GetTreasuryReportHandler)
	apiMux.HandleFunc("/reports/stats", server.GetStatsHandler)
	apiMux.HandleFunc("/forms", server.GetFormsHandler)
	apiMux.HandleFunc("/commands/submit-form", server.SubmitFormResponseHandler)
	
	// Auth routes (Public)
	mux.HandleFunc("/api/health/", server.HealthHandler)
	mux.HandleFunc("/api/auth/signup", server.SignupHandler)
	mux.HandleFunc("/api/auth/login", server.LoginHandler)
	mux.HandleFunc("/api/auth/request-magic-link", server.RequestMagicLinkHandler)
	mux.HandleFunc("/api/auth/magic-login", server.LoginWithMagicLinkHandler)
	
	mux.HandleFunc("/swagger/", server.SwaggerHandler)
	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "openapi.yaml")
	})

	// Protected API routes
	// Note: Handle "/api/" last as it is a prefix match. 
	// Specific routes like "/api/auth/" will match their handlers first.
	mux.Handle("/api/", http.StripPrefix("/api", server.AuthMiddleware(apiMux)))

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
