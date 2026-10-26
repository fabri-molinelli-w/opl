package repository

import (
	"context"
	"fmt"
	"math"
	"time"

	dbgen "github.com/fabri-molinelli-w/opl/backend/internal/db/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

// DashboardRepository runs the read-only aggregates of CU-04.
type DashboardRepository struct {
	q *dbgen.Queries
}

func NewDashboardRepository(q *dbgen.Queries) *DashboardRepository {
	return &DashboardRepository{q: q}
}

func ts(t time.Time) pgtype.Timestamptz { return pgtype.Timestamptz{Time: t, Valid: true} }

func percent(part, total int32) float64 {
	if total == 0 {
		return 0
	}
	return math.Round(float64(part)/float64(total)*1000) / 10
}

// Summary builds the dashboard. Every section is always present with zero
// values, so modules without data never break the page (RN-2, CA-4).
func (r *DashboardRepository) Summary(ctx context.Context, now time.Time) (domain.DashboardSummary, error) {
	s := domain.DashboardSummary{
		GeneratedAt:          now,
		ServersByEnvironment: map[string]int{"prod": 0, "staging": 0, "dev": 0},
		ServersByLocation:    map[string]int{"onprem": 0, "cloud": 0},
		ServicesByStatus:     map[string]int{"UP": 0, "DOWN": 0, "DEGRADED": 0, "UNKNOWN": 0},
		DocumentsByCategory:  map[string]int{},
		IncidentsBySeverity:  map[string]int{"SEV1": 0, "SEV2": 0, "SEV3": 0, "SEV4": 0},
	}

	servers, err := r.q.DashboardServerCounts(ctx)
	if err != nil {
		return s, fmt.Errorf("server counts: %w", err)
	}
	for _, row := range servers {
		s.ServersTotal += int(row.Total)
		s.ServersByEnvironment[row.Environment] += int(row.Total)
		s.ServersByLocation[row.Location] += int(row.Total)
	}

	statuses, err := r.q.DashboardServiceStatusCounts(ctx)
	if err != nil {
		return s, fmt.Errorf("service counts: %w", err)
	}
	for _, row := range statuses {
		s.ServicesTotal += int(row.Total)
		s.ServicesByStatus[row.CurrentStatus] = int(row.Total)
	}
	problems, err := r.q.DashboardProblemServices(ctx)
	if err != nil {
		return s, fmt.Errorf("problem services: %w", err)
	}
	for _, row := range problems {
		s.ProblemServices = append(s.ProblemServices, domain.DashboardService{
			ID: uuidToString(row.ID), Name: row.Name, Status: row.CurrentStatus,
			ServerID: uuidToString(row.ServerID), Hostname: row.Hostname,
		})
	}

	categories, err := r.q.DashboardDocumentCategoryCounts(ctx)
	if err != nil {
		return s, fmt.Errorf("document counts: %w", err)
	}
	for _, row := range categories {
		s.DocumentsTotal += int(row.Total)
		s.DocumentsByCategory[row.Category] = int(row.Total)
	}
	updated, err := r.q.DashboardDocumentsUpdatedSince(ctx, ts(now.Add(-7*24*time.Hour)))
	if err != nil {
		return s, fmt.Errorf("documents updated: %w", err)
	}
	s.DocumentsUpdated7d = int(updated)
	recentDocs, err := r.q.DashboardRecentDocuments(ctx)
	if err != nil {
		return s, fmt.Errorf("recent documents: %w", err)
	}
	for _, row := range recentDocs {
		s.RecentDocuments = append(s.RecentDocuments, domain.DashboardDocument{
			ID: uuidToString(row.ID), Title: row.Title, UpdatedByName: row.UpdatedByName, UpdatedAt: row.UpdatedAt.Time,
		})
	}

	severities, err := r.q.DashboardActiveIncidentCounts(ctx)
	if err != nil {
		return s, fmt.Errorf("incident counts: %w", err)
	}
	for _, row := range severities {
		s.ActiveIncidents += int(row.Total)
		s.IncidentsBySeverity[row.Severity] = int(row.Total)
	}
	recentIncidents, err := r.q.DashboardRecentActiveIncidents(ctx)
	if err != nil {
		return s, fmt.Errorf("recent incidents: %w", err)
	}
	for _, row := range recentIncidents {
		s.RecentIncidents = append(s.RecentIncidents, domain.DashboardIncident{
			ID: uuidToString(row.ID), Title: row.Title, Severity: row.Severity, Status: row.Status,
			ServiceName: row.ServiceName, OpenedAt: row.OpenedAt.Time,
		})
	}

	since := ts(now.Add(-24 * time.Hour))
	checks, err := r.q.DashboardCheckCounts(ctx, since)
	if err != nil {
		return s, fmt.Errorf("check counts: %w", err)
	}
	s.Checks24h = int(checks.Total)
	if checks.Total > 0 {
		uptime := percent(checks.Available, checks.Total)
		s.UptimePercent24h = &uptime
	}
	buckets, err := r.q.DashboardCheckBuckets(ctx, since)
	if err != nil {
		return s, fmt.Errorf("check buckets: %w", err)
	}
	for _, row := range buckets {
		s.AvailabilitySeries = append(s.AvailabilitySeries, domain.DashboardBucket{
			Start:         row.Bucket.Time,
			UptimePercent: percent(row.Available, row.Total),
			AvgResponseMs: math.Round(row.AvgResponseMs),
		})
	}
	return s, nil
}
