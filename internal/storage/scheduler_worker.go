package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"register/internal/domain"
	"time"
)

type SchedulerWorker struct {
	db       *sql.DB
	executor *domain.PipelineExecutor
}

func NewSchedulerWorker(db *sql.DB, executor *domain.PipelineExecutor) *SchedulerWorker {
	return &SchedulerWorker{
		db:       db,
		executor: executor,
	}
}

func (w *SchedulerWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.checkTimedTriggers(ctx); err != nil {
				log.Printf("SchedulerWorker error: %v", err)
			}
		}
	}
}

func (w *SchedulerWorker) checkTimedTriggers(ctx context.Context) error {
	// Find reactions that are timed and due to run
	// config is expected to have "interval" (daily, weekly, monthly)
	rows, err := w.db.QueryContext(ctx, `
		SELECT id, config, last_run_at 
		FROM event_reactions 
		WHERE trigger_event = 'TimedSchedule'`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var configJSON []byte
		var lastRunAt sql.NullTime
		if err := rows.Scan(&id, &configJSON, &lastRunAt); err != nil {
			continue
		}

		var rawConfig map[string]any
		if err := json.Unmarshal(configJSON, &rawConfig); err != nil {
			continue
		}

		interval, _ := rawConfig["interval"].(string)
		if interval == "" {
			interval = "daily" // Default fallback
		}

		if w.isDue(lastRunAt, interval) {
			log.Printf("Executing timed pipeline %s (Interval: %s)", id, interval)
			
			// Execute
			var dag domain.PipelineConfig
			json.Unmarshal(configJSON, &dag)

			// Update last_run_at FIRST to prevent double firing
			_, err = w.db.ExecContext(ctx, "UPDATE event_reactions SET last_run_at = NOW() WHERE id = $1", id)
			if err != nil {
				log.Printf("Failed to update last_run_at for %s: %v", id, err)
				continue
			}

			triggerData := map[string]any{
				"timestamp": time.Now().Format(time.RFC3339),
				"source":    "scheduler",
			}
			if err := w.executor.Execute(dag, triggerData); err != nil {
				log.Printf("Timed pipeline %s failed: %v", id, err)
			}
		}
	}

	return nil
}

func (w *SchedulerWorker) isDue(lastRunAt sql.NullTime, interval string) bool {
	if !lastRunAt.Valid {
		return true // Never run before
	}

	now := time.Now()
	switch interval {
	case "daily":
		return now.Sub(lastRunAt.Time) >= 24*time.Hour
	case "weekly":
		return now.Sub(lastRunAt.Time) >= 7*24*time.Hour
	case "monthly":
		return now.Sub(lastRunAt.Time) >= 30*24*time.Hour // Simplified
	}

	return false
}
