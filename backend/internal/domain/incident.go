package domain

import "time"

// Incident statuses (CU-10 RN-2).
const (
	IncidentOpen       = "open"
	IncidentInProgress = "in_progress"
	IncidentResolved   = "resolved"
	IncidentClosed     = "closed"
)

// Incident event kinds.
const (
	EventStatusChange = "status_change"
	EventNote         = "note"
	EventSystem       = "system"
)

var incidentTransitions = map[string][]string{
	IncidentOpen:       {IncidentInProgress},
	IncidentInProgress: {IncidentResolved},
	IncidentResolved:   {IncidentClosed, IncidentInProgress},
	IncidentClosed:     {},
}

// CanTransition reports whether the state machine allows from → to.
func CanTransition(from, to string) bool {
	for _, allowed := range incidentTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// AllowedTransitions returns the valid next states (for the UI buttons).
func AllowedTransitions(from string) []string {
	next := incidentTransitions[from]
	out := make([]string, len(next))
	copy(out, next)
	return out
}

// ValidSeverity reports whether s is SEV1..SEV4.
func ValidSeverity(s string) bool {
	switch s {
	case "SEV1", "SEV2", "SEV3", "SEV4":
		return true
	default:
		return false
	}
}

// ValidIncidentStatus reports whether s is a known incident status.
func ValidIncidentStatus(s string) bool {
	_, ok := incidentTransitions[s]
	return ok
}

// Incident is a tracked operational problem.
type Incident struct {
	ID           string
	Title        string
	Description  string
	ServiceID    string
	ServiceName  string
	Severity     string
	Status       string
	AssigneeID   string
	AssigneeName string
	RootCause    string
	AutoCreated  bool
	OpenedAt     time.Time
	ResolvedAt   *time.Time
	ClosedAt     *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Events       []IncidentEvent
}

// ResolutionMinutes returns resolved_at - opened_at in minutes (RN-6), or nil.
func (i Incident) ResolutionMinutes() *int {
	if i.ResolvedAt == nil {
		return nil
	}
	minutes := int(i.ResolvedAt.Sub(i.OpenedAt).Round(time.Minute) / time.Minute)
	return &minutes
}

// IncidentEvent is an entry of the incident timeline.
type IncidentEvent struct {
	ID         string
	Kind       string
	Body       string
	AuthorName string
	CreatedAt  time.Time
}

// NewIncident carries the data needed to open an incident manually.
type NewIncident struct {
	Title       string
	Description string
	ServiceID   string
	Severity    string
	AssigneeID  string
	AuthorID    string
}

// IncidentUpdate is a partial update of an incident.
type IncidentUpdate struct {
	Title       *string
	Description *string
	Severity    *string
	AssigneeID  *string
}

// IncidentFilter carries the incident listing filters.
type IncidentFilter struct {
	Status     string
	Severity   string
	ServiceID  string
	OnlyActive bool
	Page       int
	PageSize   int
}

// IncidentPage is a paginated incident listing.
type IncidentPage struct {
	Items    []Incident
	Total    int64
	Page     int
	PageSize int
}
