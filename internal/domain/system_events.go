package domain

import "time"

const (
	EventTypeOrganizationCreated = "OrganizationCreated"
	EventTypeOrganizationDeleted = "OrganizationDeleted"
	EventTypeOrganCreated        = "OrganCreated"
	EventTypeOrganDeleted        = "OrganDeleted"
	EventTypeFormCreated         = "FormCreated"
	EventTypeFormUpdated         = "FormUpdated"
	EventTypeFormDeleted         = "FormDeleted"
	EventTypeFormSubmitted       = "FormResponseSubmitted"
	EventTypeReactionCreated     = "ReactionCreated"
	EventTypeReactionUpdated     = "ReactionUpdated"
	EventTypeReactionDeleted     = "ReactionDeleted"
)

type OrganizationCreated struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	ParentID  *string        `json:"parent_id"`
	Path      string         `json:"path"`
	Policy    map[string]any `json:"policy"`
	Timestamp time.Time      `json:"timestamp"`
}

func (e OrganizationCreated) EventType() string { return EventTypeOrganizationCreated }

type OrganizationDeleted struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e OrganizationDeleted) EventType() string { return EventTypeOrganizationDeleted }

type OrganCreated struct {
	ID            string    `json:"id"`
	OrgID         string    `json:"org_id"`
	Name          string    `json:"name"`
	ParentOrganID *string   `json:"parent_organ_id"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e OrganCreated) EventType() string { return EventTypeOrganCreated }

type OrganDeleted struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e OrganDeleted) EventType() string { return EventTypeOrganDeleted }

type FormCreated struct {
	ID        string         `json:"id"`
	OrgID     string         `json:"org_id"`
	Title     string         `json:"title"`
	Schema    map[string]any `json:"schema"`
	Timestamp time.Time      `json:"timestamp"`
}

func (e FormCreated) EventType() string { return EventTypeFormCreated }

type FormUpdated struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Schema    map[string]any `json:"schema"`
	Timestamp time.Time      `json:"timestamp"`
}

func (e FormUpdated) EventType() string { return EventTypeFormUpdated }

type FormDeleted struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e FormDeleted) EventType() string { return EventTypeFormDeleted }

type FormResponseSubmitted struct {
	FormID    string         `json:"form_id"`
	UserID    string         `json:"user_id"`
	Answers   map[string]any `json:"answers"`
	Timestamp time.Time      `json:"timestamp"`
}

func (e FormResponseSubmitted) EventType() string { return EventTypeFormSubmitted }

type ReactionCreated struct {
	ID                 string         `json:"id"`
	OrgID              string         `json:"org_id"`
	TriggerEvent       string         `json:"trigger_event"`
	TriggerAggregateID *string        `json:"trigger_aggregate_id"`
	ActionType         string         `json:"action_type"`
	Config             map[string]any `json:"config"`
	Timestamp          time.Time      `json:"timestamp"`
}

func (e ReactionCreated) EventType() string { return EventTypeReactionCreated }

type ReactionUpdated struct {
	ID        string         `json:"id"`
	Config    map[string]any `json:"config"`
	Timestamp time.Time      `json:"timestamp"`
}

func (e ReactionUpdated) EventType() string { return EventTypeReactionUpdated }

type ReactionDeleted struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e ReactionDeleted) EventType() string { return EventTypeReactionDeleted }
