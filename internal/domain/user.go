package domain

import (
	"fmt"
	"time"
)

// User aggregate handles identity and profile
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// Vi kan legge til password_hash her hvis vi vil ha alt event-sourced,
	// men for nå holder vi oss til profil-data.
}

const (
	EventTypeUserCreated      = "UserCreated"
	EventTypeUserProfileUpdated = "UserProfileUpdated"
)

type UserCreated struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"password_hash"`
	Timestamp    time.Time `json:"timestamp"`
}
func (e UserCreated) EventType() string { return EventTypeUserCreated }

type UserProfileUpdated struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Timestamp time.Time `json:"timestamp"`
}
func (e UserProfileUpdated) EventType() string { return EventTypeUserProfileUpdated }

func ApplyUserEvent(u *User, e Event) error {
	switch v := e.(type) {
	case UserCreated:
		u.ID = v.ID
		u.Email = v.Email
		u.Name = v.Name
		u.CreatedAt = v.Timestamp
		u.UpdatedAt = v.Timestamp
	case UserProfileUpdated:
		if v.Name != "" { u.Name = v.Name }
		if v.Email != "" { u.Email = v.Email }
		u.UpdatedAt = v.Timestamp
	default:
		return fmt.Errorf("unknown user event type: %T", e)
	}
	return nil
}
