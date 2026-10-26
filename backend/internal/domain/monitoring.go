package domain

import "time"

// Service statuses (UNKNOWN only exists before the first check).
const (
	StatusUp       = "UP"
	StatusDown     = "DOWN"
	StatusDegraded = "DEGRADED"
	StatusUnknown  = "UNKNOWN"
)

// CheckConfig is what the checker needs to probe a service.
type CheckConfig struct {
	ServiceID          string
	Name               string
	CheckType          string
	CheckTarget        string
	ExpectedStatus     int
	LatencyThresholdMs int
}

// CheckResult is the outcome of a single probe (CU-03 RN-1/RN-2).
type CheckResult struct {
	Status     string
	ResponseMs *int
	HTTPStatus *int
	Error      string
}

// HealthCheck is a stored check result.
type HealthCheck struct {
	ID         string
	ServiceID  string
	Status     string
	ResponseMs *int
	HTTPStatus *int
	Error      string
	CheckedAt  time.Time
}

// CheckOutcome describes the side effects of recording a check, used to emit
// internal events (notifications, CU-08).
type CheckOutcome struct {
	Check            HealthCheck
	ServerID         string
	ServiceName      string
	PreviousStatus   string
	IncidentOpenedID string
}

// StatusChanged reports whether the check changed the service status.
func (o CheckOutcome) StatusChanged() bool {
	return o.PreviousStatus != o.Check.Status
}
