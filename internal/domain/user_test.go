package domain

import (
	"testing"
	"time"
)

func TestApplyUserEvent(t *testing.T) {
	u := &User{}
	now := time.Now()

	t.Run("UserCreated", func(t *testing.T) {
		event := UserCreated{
			ID:    "u1",
			Email: "test@example.com",
			Name:  "Test User",
			Timestamp: now,
		}
		err := ApplyUserEvent(u, event)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if u.ID != "u1" {
			t.Errorf("Expected ID u1, got %s", u.ID)
		}
		if u.Email != "test@example.com" {
			t.Errorf("Expected Email test@example.com, got %s", u.Email)
		}
		if u.Name != "Test User" {
			t.Errorf("Expected Name Test User, got %s", u.Name)
		}
	})

	t.Run("UserProfileUpdated", func(t *testing.T) {
		event := UserProfileUpdated{
			ID:    "u1",
			Name:  "Updated Name",
			Email: "updated@example.com",
			Timestamp: now.Add(time.Hour),
		}
		err := ApplyUserEvent(u, event)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if u.Name != "Updated Name" {
			t.Errorf("Expected Name Updated Name, got %s", u.Name)
		}
		if u.Email != "updated@example.com" {
			t.Errorf("Expected Email updated@example.com, got %s", u.Email)
		}
	})
}
