package domain

import "time"

// Service is a monitored unit running on a server.
type Service struct {
	ID                 string
	ServerID           string
	Name               string
	Type               string
	Port               int
	CheckType          string
	CheckTarget        string
	ExpectedStatus     int
	IntervalSeconds    int
	LatencyThresholdMs int
	CurrentStatus      string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// ServiceUpdate describes a partial update of a service. ServerID is immutable.
type ServiceUpdate struct {
	Name               *string
	Type               *string
	Port               *int
	CheckType          *string
	CheckTarget        *string
	ExpectedStatus     *int
	IntervalSeconds    *int
	LatencyThresholdMs *int
}

// ServiceFilter carries the combinable filters of the service listing.
type ServiceFilter struct {
	ServerID    string
	Environment string
	Status      string
	Type        string
	Page        int
	PageSize    int
}

// ServicePage is a paginated service listing.
type ServicePage struct {
	Items    []Service
	Total    int64
	Page     int
	PageSize int
}

// Inventory enum values shared by the validation layer.
const (
	EnvironmentProd    = "prod"
	EnvironmentStaging = "staging"
	EnvironmentDev     = "dev"

	LocationOnPrem = "onprem"
	LocationCloud  = "cloud"

	ServiceTypeWeb   = "web"
	ServiceTypeDB    = "db"
	ServiceTypeQueue = "queue"
	ServiceTypeOther = "other"

	CheckTypeHTTP = "http"
	CheckTypeTCP  = "tcp"

	ServiceStatusUP       = "UP"
	ServiceStatusDOWN     = "DOWN"
	ServiceStatusDegraded = "DEGRADED"
	ServiceStatusUnknown  = "UNKNOWN"
)

// ValidEnvironment reports whether value is a supported environment.
func ValidEnvironment(value string) bool {
	switch value {
	case EnvironmentProd, EnvironmentStaging, EnvironmentDev:
		return true
	default:
		return false
	}
}

// ValidLocation reports whether value is a supported location.
func ValidLocation(value string) bool {
	return value == LocationOnPrem || value == LocationCloud
}

// ValidServiceType reports whether value is a supported service type.
func ValidServiceType(value string) bool {
	switch value {
	case ServiceTypeWeb, ServiceTypeDB, ServiceTypeQueue, ServiceTypeOther:
		return true
	default:
		return false
	}
}

// ValidCheckType reports whether value is a supported check type.
func ValidCheckType(value string) bool {
	return value == CheckTypeHTTP || value == CheckTypeTCP
}

// ValidServiceStatus reports whether value is a supported service status.
func ValidServiceStatus(value string) bool {
	switch value {
	case ServiceStatusUP, ServiceStatusDOWN, ServiceStatusDegraded, ServiceStatusUnknown:
		return true
	default:
		return false
	}
}
