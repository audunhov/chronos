package domain

type ContextKey string

const (
	UserIDKey        ContextKey = "user_id"
	OrgIDKey         ContextKey = "org_id"
	RoleKey          ContextKey = "role"
	CorrelationIDKey ContextKey = "correlation_id"
	IPAddressKey     ContextKey = "ip_address"
	UserAgentKey     ContextKey = "user_agent"
)

// Event interface defines a generic event
type Event interface {
	EventType() string
}
