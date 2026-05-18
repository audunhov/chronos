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
	apiMux.HandleFunc("GET /members", server.GetMembersHandler)
	apiMux.HandleFunc("POST /commands/register-member", server.RegisterMemberHandler)
	apiMux.HandleFunc("POST /commands/update-member", server.UpdateMemberHandler)
	apiMux.HandleFunc("POST /commands/shred-member", server.ShredMemberHandler)
	apiMux.HandleFunc("GET /reports/as-of", server.GetMembersAsOfHandler)
	
	// Apply Auth Middleware to all /api/ routes
	mux.Handle("/api/", http.StripPrefix("/api", api.AuthMiddleware(apiMux)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
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
