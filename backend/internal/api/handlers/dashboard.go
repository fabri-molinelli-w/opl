package handlers

import (
	"context"
	"sort"
	"time"

	gen "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

// DashboardService is the CU-04 use case consumed by the handlers.
type DashboardService interface {
	Summary(ctx context.Context, now time.Time) (domain.DashboardSummary, error)
}

func (s *Server) GetDashboardSummary(ctx context.Context, _ gen.GetDashboardSummaryRequestObject) (gen.GetDashboardSummaryResponseObject, error) {
	d, err := s.dashboard.Summary(ctx, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	problems := make([]gen.DashboardProblemService, 0, len(d.ProblemServices))
	for _, p := range d.ProblemServices {
		problems = append(problems, gen.DashboardProblemService{
			Id: p.ID, Name: p.Name, Status: gen.ServiceStatus(p.Status), ServerId: p.ServerID, Hostname: p.Hostname,
		})
	}
	categories := make([]gen.DashboardCategoryCount, 0, len(d.DocumentsByCategory))
	for category, count := range d.DocumentsByCategory {
		categories = append(categories, gen.DashboardCategoryCount{Category: gen.DocumentCategory(category), Count: count})
	}
	sort.Slice(categories, func(i, j int) bool { return categories[i].Count > categories[j].Count })
	recentDocs := make([]gen.DashboardRecentDocument, 0, len(d.RecentDocuments))
	for _, doc := range d.RecentDocuments {
		recentDocs = append(recentDocs, gen.DashboardRecentDocument{
			Id: doc.ID, Title: doc.Title, UpdatedByName: doc.UpdatedByName, UpdatedAt: doc.UpdatedAt,
		})
	}
	recentIncidents := make([]gen.DashboardRecentIncident, 0, len(d.RecentIncidents))
	for _, i := range d.RecentIncidents {
		recentIncidents = append(recentIncidents, gen.DashboardRecentIncident{
			Id: i.ID, Title: i.Title, Severity: gen.Severity(i.Severity), Status: gen.IncidentStatus(i.Status),
			ServiceName: i.ServiceName, OpenedAt: i.OpenedAt,
		})
	}
	series := make([]gen.DashboardBucket, 0, len(d.AvailabilitySeries))
	for _, b := range d.AvailabilitySeries {
		series = append(series, gen.DashboardBucket{Start: b.Start, UptimePercent: b.UptimePercent, AvgResponseMs: b.AvgResponseMs})
	}

	return gen.GetDashboardSummary200JSONResponse(gen.DashboardSummary{
		GeneratedAt: d.GeneratedAt,
		Servers: gen.DashboardServers{
			Total: d.ServersTotal, Prod: d.ServersByEnvironment["prod"], Staging: d.ServersByEnvironment["staging"],
			Dev: d.ServersByEnvironment["dev"], Onprem: d.ServersByLocation["onprem"], Cloud: d.ServersByLocation["cloud"],
		},
		Services: gen.DashboardServices{
			Total: d.ServicesTotal, Up: d.ServicesByStatus["UP"], Down: d.ServicesByStatus["DOWN"],
			Degraded: d.ServicesByStatus["DEGRADED"], Unknown: d.ServicesByStatus["UNKNOWN"], Problems: problems,
		},
		Documents: gen.DashboardDocuments{
			Total: d.DocumentsTotal, UpdatedLast7Days: d.DocumentsUpdated7d, ByCategory: categories, Recent: recentDocs,
		},
		Incidents: gen.DashboardIncidents{
			Active: d.ActiveIncidents, Sev1: d.IncidentsBySeverity["SEV1"], Sev2: d.IncidentsBySeverity["SEV2"],
			Sev3: d.IncidentsBySeverity["SEV3"], Sev4: d.IncidentsBySeverity["SEV4"], Recent: recentIncidents,
		},
		Availability: gen.DashboardAvailability{
			Checks24h: d.Checks24h, UptimePercent24h: d.UptimePercent24h, Series: series,
		},
	}), nil
}
