package domain

import (
	"database/sql"
)

type ContextKey string

const (
	UserIDKey        ContextKey = "user_id"
	OrgIDKey         ContextKey = "org_id"
	RoleKey          ContextKey = "role"
	CorrelationIDKey ContextKey = "correlation_id"
	IPAddressKey     ContextKey = "ip_address"
	UserAgentKey     ContextKey = "user_agent"
)

// DBExecutor defines a common interface for sql.DB and sql.Tx
type DBExecutor interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// Event interface defines a generic event
type Event interface {
	EventType() string
}
