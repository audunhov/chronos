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

// Event interface defines a generic event
type Event interface {
	EventType() string
}

// Member struct represents the current state of a member
type Member struct {
	ID         string         `json:"id"`
	OrgID      string         `json:"org_id"`
	Name       string         `json:"name"`
	Email      string         `json:"email"`
	Status     string         `json:"status"`
	Metadata   map[string]any `json:"metadata"`
	Balance    int            `json:"balance"`
	FeeFormula string         `json:"fee_formula"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// Constants for event types
const (
	EventTypeMemberRegistered           = "MemberRegistered"
	EventTypeMemberUpdated              = "MemberUpdated"
	EventTypeMembershipFeeFormulaDefined = "MembershipFeeFormulaDefined"
	EventTypeFeeGenerated               = "FeeGenerated"
	EventTypePaymentReceived            = "PaymentReceived"
	EventTypeMemberShredded             = "MemberShredded"
	EventTypeMemberOrgMoved             = "MemberOrgMoved"
)

// Event payloads
type MemberRegistered struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
	OrgID     string         `json:"org_id"`
	Metadata  map[string]any `json:"metadata"`
	Timestamp time.Time      `json:"timestamp"`
}
func (e MemberRegistered) EventType() string { return EventTypeMemberRegistered }

type MemberUpdated struct {
	ID            string         `json:"id"`
	UpdatedFields map[string]any `json:"updated_fields"`
	Timestamp     time.Time      `json:"timestamp"`
}
func (e MemberUpdated) EventType() string { return EventTypeMemberUpdated }

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

type MemberShredded struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}
func (e MemberShredded) EventType() string { return EventTypeMemberShredded }

type MemberOrgMoved struct {
	ID        string    `json:"id"`
	FromOrgID string    `json:"from_org_id"`
	ToOrgID   string    `json:"to_org_id"`
	Timestamp time.Time `json:"timestamp"`
}
func (e MemberOrgMoved) EventType() string { return EventTypeMemberOrgMoved }

// ApplyEvent mutates the Member state in-memory based on the event type
func ApplyEvent(m *Member, e Event) error {
	// Invariant: Et shreddet medlem kan aldri endres igjen (GDPR hard-limit)
	if m.Status == StatusShredded {
		// Vi tillater ikke engang re-registrering av samme ID hvis den er shreddet
		return fmt.Errorf("cannot modify a shredded member: %s", m.ID)
	}

	switch v := e.(type) {
	case MemberRegistered:
		m.ID = v.ID
		m.OrgID = v.OrgID
		m.Name = v.Name
		m.Email = v.Email
		m.Status = StatusActive
		m.Metadata = v.Metadata
		m.CreatedAt = v.Timestamp
		m.UpdatedAt = v.Timestamp

	case MemberUpdated:
		for key, val := range v.UpdatedFields {
			switch key {
			case "name":
				if s, ok := val.(string); ok { m.Name = s }
			case "email":
				if s, ok := val.(string); ok { m.Email = s }
			case "status":
				// Invariant: Kan ikke sette status til SHREDDED via vanlig update
				if s, ok := val.(string); ok && s != StatusShredded {
					m.Status = s
				}
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

	case MemberOrgMoved:
		m.OrgID = v.ToOrgID
		m.UpdatedAt = v.Timestamp

	case MemberShredded:
		m.Name = "REDACTED"
		m.Email = "redacted@example.com"
		m.Metadata = nil
		m.Status = StatusShredded
		m.UpdatedAt = v.Timestamp

	default:
		return fmt.Errorf("unknown event type: %T", e)
	}
	return nil
}
