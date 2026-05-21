package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"register/internal/domain"
	"time"
)

type ReactionWorker struct {
	db       *sql.DB
	executor *domain.PipelineExecutor
}

func NewReactionWorker(db *sql.DB) *ReactionWorker {
	w := &ReactionWorker{db: db}
	w.executor = &domain.PipelineExecutor{
		Operations: map[string]func(inputs map[string]any) (map[string]any, string, error){
			"FindOrg":        w.opFindOrg,
			"FindOrgan":      w.opFindOrgan,
			"FindRole":       w.opFindRole,
			"Template":       w.opTemplate,
			"SendEmail":      w.opSendEmail,
			"IfThen":         w.opIfThen,
			"Filter":         w.opFilter,
			"RegisterMember": w.opRegisterMember,
			"CreateForm":     w.opCreateForm,
		},
	}
	return w
}

// --- Production Operations ---

func (w *ReactionWorker) opIfThen(inputs map[string]any) (map[string]any, string, error) {
	v1 := inputs["value1"]
	v2 := inputs["value2"]
	op, _ := inputs["operator"].(string)

	if domain.EvaluateCondition(v1, op, v2) {
		return nil, "true", nil
	}
	return nil, "false", nil
}

func (w *ReactionWorker) opFilter(inputs map[string]any) (map[string]any, string, error) {
	v1 := inputs["value1"]
	v2 := inputs["value2"]
	op, _ := inputs["operator"].(string)

	if domain.EvaluateCondition(v1, op, v2) {
		return nil, "default", nil
	}
	return nil, "default", fmt.Errorf("filtered")
}

func (w *ReactionWorker) opFindOrg(inputs map[string]any) (map[string]any, string, error) {
	startID, _ := inputs["start_org_id"].(string)
	relation, _ := inputs["relation"].(string)
	
	if startID == "" { return nil, "", fmt.Errorf("missing start_org_id") }

	var query string
	if relation == "parent" {
		query = "SELECT parent_id, name FROM organization_hierarchy WHERE id = $1"
	} else {
		query = "SELECT id, name FROM organization_hierarchy WHERE id = $1"
	}

	var id, name string
	var parentID sql.NullString
	err := w.db.QueryRow(query, startID).Scan(&parentID, &name)
	if err != nil { return nil, "", err }
	
	if relation == "parent" && parentID.Valid {
		id = parentID.String
		// Hent navn for parent
		_ = w.db.QueryRow("SELECT name FROM organization_hierarchy WHERE id = $1", id).Scan(&name)
	} else if relation != "parent" {
		id = startID
	}

	return map[string]any{"org_id": id, "name": name}, "default", nil
}

func (w *ReactionWorker) opFindOrgan(inputs map[string]any) (map[string]any, string, error) {
	orgID, _ := inputs["org_id"].(string)
	name, _ := inputs["organ_name"].(string)
	
	var id string
	err := w.db.QueryRow("SELECT id FROM organs WHERE org_id = $1 AND name = $2", orgID, name).Scan(&id)
	if err != nil { return nil, "", err }

	return map[string]any{"organ_id": id}, "default", nil
}

func (w *ReactionWorker) opFindRole(inputs map[string]any) (map[string]any, string, error) {
	targetID, _ := inputs["target_id"].(string)
	roleType, _ := inputs["role_type"].(string)

	query := `
		SELECT ra.user_id, u.email, u.name 
		FROM role_assignments ra
		JOIN users u ON ra.user_id = u.id
		WHERE (ra.org_id = $1 OR ra.organ_id = $1) AND ra.role_type = $2
		LIMIT 1`
	
	var userID, email, name string
	err := w.db.QueryRow(query, targetID, roleType).Scan(&userID, &email, &name)
	if err != nil { return nil, "", err }

	return map[string]any{"user_id": userID, "email": email, "name": name}, "default", nil
}

func (w *ReactionWorker) opTemplate(inputs map[string]any) (map[string]any, string, error) {
	tmplName, _ := inputs["template_name"].(string)
	userName, _ := inputs["user_name"].(string)
	
	return map[string]any{
		"subject": "Varsel fra Chronos: " + tmplName,
		"body":    fmt.Sprintf("Hei, dette er et automatisk varsel angående %s.", userName),
	}, "default", nil
}

func (w *ReactionWorker) opSendEmail(inputs map[string]any) (map[string]any, string, error) {
	to, _ := inputs["to_email"].(string)
	subject, _ := inputs["subject"].(string)
	body, _ := inputs["body"].(string)

	if to == "" { return nil, "", fmt.Errorf("missing recipient") }

	_, err := w.db.Exec(`
		INSERT INTO email_outbox (recipient_email, subject, body_html) 
		VALUES ($1, $2, $3)`,
		to, subject, body)
	
	return map[string]any{"success": err == nil}, "default", err
}

func (w *ReactionWorker) opRegisterMember(inputs map[string]any) (map[string]any, string, error) {
	// Merk: Her kaller vi domenelaget direkte for å bevare determinisme
	// I en ekte app ville vi kanskje ha sendt en kommando over bussen
	return nil, "default", fmt.Errorf("RegisterMember node not yet implemented in worker")
}

func (w *ReactionWorker) opCreateForm(inputs map[string]any) (map[string]any, string, error) {
	return nil, "default", fmt.Errorf("CreateForm node not yet implemented in worker")
}

// --- Worker Loop ---

func (w *ReactionWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lastEventID int64
	_ = w.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(id), 0) FROM event_store").Scan(&lastEventID)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.processReactions(ctx, &lastEventID); err != nil {
				log.Printf("ReactionWorker error: %v", err)
			}
		}
	}
}

func (w *ReactionWorker) processReactions(ctx context.Context, lastID *int64) error {
	rows, err := w.db.QueryContext(ctx, `
		SELECT id, aggregate_id, event_type, payload 
		FROM event_store 
		WHERE id > $1 
		ORDER BY id ASC 
		LIMIT 100`, *lastID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var aggregateID, eventType string
		var payload []byte
		if err := rows.Scan(&id, &aggregateID, &eventType, &payload); err != nil {
			continue
		}

		if err := w.handleEvent(ctx, aggregateID, eventType, payload); err != nil {
			log.Printf("Reaction error for event %d: %v", id, err)
		}
		*lastID = id
	}

	return nil
}

func (w *ReactionWorker) handleEvent(ctx context.Context, aggregateID, eventType string, payload []byte) error {
	// Forbered trigger-data fra event payload
	var payloadMap map[string]any
	json.Unmarshal(payload, &payloadMap)

	// Spesialhåndtering for FormResponseSubmitted: 
	// Vi vil matche mot FormID (som er trigger_aggregate_id), ikke responseID (som er aggregate_id i store)
	matchID := aggregateID
	if eventType == domain.EventTypeFormSubmitted {
		if formID, ok := payloadMap["form_id"].(string); ok {
			matchID = formID
		}
	}

	// Finn alle reaksjoner for dette eventet
	// Vi matcher enten på nøyaktig aggregateID eller reaksjoner som gjelder ALLE (NULL)
	rows, err := w.db.QueryContext(ctx, `
		SELECT action_type, config 
		FROM event_reactions 
		WHERE trigger_event = $1 AND (trigger_aggregate_id = $2 OR trigger_aggregate_id IS NULL)`, 
		eventType, matchID)
	
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var actionType string
		var configJSON []byte
		if err := rows.Scan(&actionType, &configJSON); err != nil {
			continue
		}

		if actionType == "PIPELINE_DAG" {
			var config domain.PipelineConfig
			if err := json.Unmarshal(configJSON, &config); err != nil {
				continue
			}

			// Kjør DAG
			if err := w.executor.Execute(config, payloadMap); err != nil {
				log.Printf("Pipeline execution failed: %v", err)
			}
		}
	}

	return nil
}
