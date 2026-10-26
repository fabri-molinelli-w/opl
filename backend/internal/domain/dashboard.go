package domain

import "time"

// DashboardSummary is the single aggregated payload of CU-04 (RN-1).
type DashboardSummary struct {
	GeneratedAt time.Time

	ServersTotal         int
	ServersByEnvironment map[string]int
	ServersByLocation    map[string]int

	ServicesTotal    int
	ServicesByStatus map[string]int
	ProblemServices  []DashboardService

	DocumentsTotal      int
	DocumentsByCategory map[string]int
	DocumentsUpdated7d  int
	RecentDocuments     []DashboardDocument
	ActiveIncidents     int
	IncidentsBySeverity map[string]int
	RecentIncidents     []DashboardIncident
	Checks24h           int
	UptimePercent24h    *float64
	AvailabilitySeries  []DashboardBucket
}

type DashboardService struct {
	ID       string
	Name     string
	Status   string
	ServerID string
	Hostname string
}

type DashboardDocument struct {
	ID            string
	Title         string
	UpdatedByName string
	UpdatedAt     time.Time
}

type DashboardIncident struct {
	ID          string
	Title       string
	Severity    string
	Status      string
	ServiceName string
	OpenedAt    time.Time
}

// DashboardBucket is a 15-minute bucket of the last 24 h.
type DashboardBucket struct {
	Start         time.Time
	UptimePercent float64
	AvgResponseMs float64
}
