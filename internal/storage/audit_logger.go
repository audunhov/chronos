package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"register/internal/domain"

	"github.com/google/uuid"
)

type AuditLogger struct {
	db *sql.DB
}

func NewAuditLogger(db *sql.DB) *AuditLogger {
	return &AuditLogger{db: db}
}

func (l *AuditLogger) Log(ctx context.Context, action string, targetID string, detail any) {
	correlationID, _ := ctx.Value(domain.CorrelationIDKey).(string)
	if correlationID == "" {
		correlationID = uuid.New().String()
	}
	actorID, _ := ctx.Value(domain.UserIDKey).(string)
	orgID, _ := ctx.Value(domain.OrgIDKey).(string)
	ip, _ := ctx.Value(domain.IPAddressKey).(string)
	ua, _ := ctx.Value(domain.UserAgentKey).(string)

	detailJSON, _ := json.Marshal(detail)

	_, err := l.db.ExecContext(ctx, `
		INSERT INTO audit_logs (correlation_id, actor_id, org_id, action, target_id, detail, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		correlationID,
		sql.NullString{String: actorID, Valid: actorID != ""},
		sql.NullString{String: orgID, Valid: orgID != ""},
		action,
		sql.NullString{String: targetID, Valid: targetID != ""},
		detailJSON,
		ip,
		ua,
	)

	if err != nil {
		log.Printf("AuditLogger failed: %v", err)
	}
}
