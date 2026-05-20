package domain

import (
	"fmt"
	"time"
)

// Status konstanter for tydeligere tilstandsmaskin
const (
	StatusActive   = "ACTIVE"
	StatusInactive = "INACTIVE"
	StatusShredded = "SHREDDED"
)

// Membership struct represents the relationship between a User and an Organization
type Membership struct {
	ID         string         `json:"id"`
	UserID     string         `json:"user_id"`
	OrgID      string         `json:"org_id"`
	Status     string         `json:"status"`
	Role       string         `json:"role"`
	Metadata   map[string]any `json:"metadata"`
	Balance    int            `json:"balance"`
	FeeFormula string         `json:"fee_formula"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// Constants for event types
const (
	EventTypeMembershipCreated           = "MembershipCreated"
	EventTypeMembershipUpdated           = "MembershipUpdated"
	EventTypeMembershipFeeFormulaDefined = "MembershipFeeFormulaDefined"
	EventTypeFeeGenerated               = "FeeGenerated"
	EventTypePaymentReceived            = "PaymentReceived"
	EventTypeMembershipShredded          = "MembershipShredded"
	EventTypeMembershipOrgMoved          = "MembershipOrgMoved"
)

// Event payloads
type MembershipCreated struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	OrgID     string         `json:"org_id"`
	Role      string         `json:"role"`
	Metadata  map[string]any `json:"metadata"`
	Timestamp time.Time      `json:"timestamp"`
}
func (e MembershipCreated) EventType() string { return EventTypeMembershipCreated }

type MembershipUpdated struct {
	ID            string         `json:"id"`
	UpdatedFields map[string]any `json:"updated_fields"`
	Timestamp     time.Time      `json:"timestamp"`
}
func (e MembershipUpdated) EventType() string { return EventTypeMembershipUpdated }

type MembershipOrgMoved struct {
	ID        string    `json:"id"`
	FromOrgID string    `json:"from_org_id"`
	ToOrgID   string    `json:"to_org_id"`
	Timestamp time.Time `json:"timestamp"`
}
func (e MembershipOrgMoved) EventType() string { return EventTypeMembershipOrgMoved }

type MembershipShredded struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}
func (e MembershipShredded) EventType() string { return EventTypeMembershipShredded }

type MembershipFeeFormulaDefined struct {
	ID        string    `json:"id"`
	Formula   string    `json:"formula"`
	Timestamp time.Time `json:"timestamp"`
}
func (e MembershipFeeFormulaDefined) EventType() string { return EventTypeMembershipFeeFormulaDefined }

type FeeGenerated struct {
	ID        string    `json:"id"`
	Amount    int       `json:"amount"`
	Period    string    `json:"period"`
	Timestamp time.Time `json:"timestamp"`
}
func (e FeeGenerated) EventType() string { return EventTypeFeeGenerated }

type PaymentReceived struct {
	ID        string    `json:"id"`
	Amount    int       `json:"amount"`
	Timestamp time.Time `json:"timestamp"`
}
func (e PaymentReceived) EventType() string { return EventTypePaymentReceived }

func ApplyMembershipEvent(m *Membership, e Event) error {
	if m.Status == StatusShredded {
		return fmt.Errorf("cannot modify a shredded membership: %s", m.ID)
	}

	switch v := e.(type) {
	case MembershipCreated:
		m.ID = v.ID
		m.UserID = v.UserID
		m.OrgID = v.OrgID
		m.Role = v.Role
		m.Status = StatusActive
		m.Metadata = v.Metadata
		m.CreatedAt = v.Timestamp
		m.UpdatedAt = v.Timestamp

	case MembershipUpdated:
		for key, val := range v.UpdatedFields {
			switch key {
			case "status":
				if s, ok := val.(string); ok && s != StatusShredded {
					m.Status = s
				}
			case "role":
				if s, ok := val.(string); ok { m.Role = s }
			case "metadata":
				if md, ok := val.(map[string]any); ok { m.Metadata = md }
			}
		}
		m.UpdatedAt = v.Timestamp

	case MembershipFeeFormulaDefined:
		m.FeeFormula = v.Formula
		m.UpdatedAt = v.Timestamp

	case FeeGenerated:
		m.Balance -= v.Amount
		m.UpdatedAt = v.Timestamp

	case PaymentReceived:
		m.Balance += v.Amount
		m.UpdatedAt = v.Timestamp

	case MembershipOrgMoved:
		m.OrgID = v.ToOrgID
		m.UpdatedAt = v.Timestamp

	case MembershipShredded:
		m.Status = StatusShredded
		m.Metadata = nil
		m.UserID = "REDACTED" // Bryt koblingen til brukeren
		m.UpdatedAt = v.Timestamp

	default:
		return fmt.Errorf("unknown membership event type: %T", e)
	}
	return nil
}
