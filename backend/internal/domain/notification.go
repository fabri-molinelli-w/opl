package domain

import "time"

// Notification event types (CU-08).
const (
	EventServiceDown      = "service.down"
	EventIncidentOpened   = "incident.opened"
	EventIncidentAssigned = "incident.assigned"
	EventDocumentUpdated  = "document.updated"
)

// Notification is an in-app notification of a single user.
type Notification struct {
	ID        string
	EventType string
	Title     string
	Body      string
	Link      string
	ReadAt    *time.Time
	CreatedAt time.Time
}

// NewNotification is a notification to fan out to one or more users.
type NewNotification struct {
	EventType string
	Title     string
	Body      string
	Link      string
	DedupeKey string
}
