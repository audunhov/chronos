package domain

// Event interface defines a generic event
type Event interface {
	EventType() string
}
