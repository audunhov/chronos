package projection

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type Worker struct {
	db *sql.DB
}

func NewWorker(db *sql.DB) *Worker {
	return &Worker{db: db}
}

func (w *Worker) Start(ctx context.Context) {
	// Worker fungerer nå som et sikkerhetsnett eller for replays.
	// Siden vi har synkron projeksjon for hoved-flowen, er denne mindre kritisk for sanntids-UI.
	ticker := time.NewTicker(30 * time.Second) 
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// For nå deaktiverer vi pollingen for å spare ressurser, 
			// siden vi har synkron projeksjon.
			// if err := w.processNewEvents(ctx, &lastProcessedID); err != nil {
			// 	log.Printf("Projection worker error: %v", err)
			// }
			log.Println("ProjectionWorker security check heartbeat")
		}
	}
}

func (w *Worker) processNewEvents(ctx context.Context, lastID *int64) error {
	// Implementasjon her hvis vi trenger replay-funksjonalitet senere
	return nil
}
