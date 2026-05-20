package domain

import (
	"time"
)

type RoleAssignment struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	OrgID     string    `json:"org_id"`
	OrganID   *string   `json:"organ_id"`
	RoleType  string    `json:"role_type"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	EventTypeRoleAssigned = "RoleAssigned"
	EventTypeRoleRevoked  = "RoleRevoked"
)

type RoleAssigned struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	OrgID     string    `json:"org_id"`
	OrganID   *string   `json:"organ_id"`
	RoleType  string    `json:"role_type"`
	Timestamp time.Time `json:"timestamp"`
}
func (e RoleAssigned) EventType() string { return EventTypeRoleAssigned }

type RoleRevoked struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}
func (e RoleRevoked) EventType() string { return EventTypeRoleRevoked }
