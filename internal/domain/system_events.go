package domain

import "time"

const (
	EventTypeOrganizationCreated = "OrganizationCreated"
	EventTypeOrganizationDeleted = "OrganizationDeleted"
	EventTypeOrganCreated        = "OrganCreated"
	EventTypeFormCreated         = "FormCreated"
	EventTypeFormSubmitted       = "FormResponseSubmitted"
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

type FormCreated struct {
	ID        string         `json:"id"`
	OrgID     string         `json:"org_id"`
	Title     string         `json:"title"`
	Schema    map[string]any `json:"schema"`
	Timestamp time.Time      `json:"timestamp"`
}

func (e FormCreated) EventType() string { return EventTypeFormCreated }

type FormResponseSubmitted struct {
	FormID    string         `json:"form_id"`
	UserID    string         `json:"user_id"`
	Answers   map[string]any `json:"answers"`
	Timestamp time.Time      `json:"timestamp"`
}

func (e FormResponseSubmitted) EventType() string { return EventTypeFormSubmitted }
