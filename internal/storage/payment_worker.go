package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"
)

type PaymentWorker struct {
	db *sql.DB
}

func NewPaymentWorker(db *sql.DB) *PaymentWorker {
	return &PaymentWorker{db: db}
}

func (w *PaymentWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.processInvoices(ctx); err != nil {
				log.Printf("PaymentWorker error: %v", err)
			}
		}
	}
}

func (w *PaymentWorker) processInvoices(ctx context.Context) error {
	// Finn ubetalte fakturaer som er klare for neste steg i pipelinen
	// For enkelthets skyld simulerer vi at et steg tar litt tid
	rows, err := w.db.QueryContext(ctx, `
		SELECT i.id, i.membership_id, i.amount, i.current_pipeline_step, p.sequence
		FROM invoice_view i
		JOIN membership_view m ON i.membership_id = m.id
		JOIN payment_pipelines p ON m.org_id = p.org_id
		WHERE i.status = 'UNPAID'
		FOR UPDATE SKIP LOCKED`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, membershipID string
		var amount, step int
		var sequenceJSON []byte
		if err := rows.Scan(&id, &membershipID, &amount, &step, &sequenceJSON); err != nil {
			continue
		}

		var sequence []string
		json.Unmarshal(sequenceJSON, &sequence)

		if step >= len(sequence) {
			// Ingen flere steg, merk som mislykket eller send til inkasso (simulert)
			continue
		}

		method := sequence[step]
		log.Printf("Invoice %s: Attempting step %d (%s)", id, step, method)

		// Simuler prosessering basert på metode
		success := false
		switch method {
		case "card", "vipps":
			// Simuler tilfeldig feil for å vise pipelinen i aksjon
			success = time.Now().Unix()%3 == 0 
		case "email":
			// Email fungerer "alltid" som fallback
			success = true
		}

		if success {
			w.markPaid(ctx, id, membershipID, amount)
		} else {
			w.incrementStep(ctx, id, step+1)
		}
	}

	return nil
}

func (w *PaymentWorker) markPaid(ctx context.Context, invoiceID, membershipID string, amount int) {
	tx, _ := w.db.BeginTx(ctx, nil)
	defer tx.Rollback()

	// 1. Oppdater faktura
	_, _ = tx.ExecContext(ctx, "UPDATE invoice_view SET status = 'PAID' WHERE id = $1", invoiceID)

	// 2. Registrer betaling på medlemskap (Dette ville vanligvis vært et domene-event)
	// For nå oppdaterer vi bare view-en direkte for å vise resultatet
	_, _ = tx.ExecContext(ctx, "UPDATE membership_view SET balance = balance + $1 WHERE id = $2", amount, membershipID)

	tx.Commit()
	log.Printf("Invoice %s: Payment successful", invoiceID)
}

func (w *PaymentWorker) incrementStep(ctx context.Context, id string, nextStep int) {
	_, _ = w.db.ExecContext(ctx, "UPDATE invoice_view SET current_pipeline_step = $1 WHERE id = $2", nextStep, id)
	log.Printf("Invoice %s: Moving to step %d", id, nextStep)
}
