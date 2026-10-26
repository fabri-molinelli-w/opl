package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

const (
	defaultChecksLimit = 50
	maxChecksLimit     = 500
	checkRetention     = 7 * 24 * time.Hour
)

// MonitoringRepository is the persistence used by MonitoringService.
type MonitoringRepository interface {
	ListDueServices(ctx context.Context, limit int) ([]string, error)
	GetCheckConfig(ctx context.Context, serviceID string) (domain.CheckConfig, error)
	RecordCheck(ctx context.Context, serviceID string, result domain.CheckResult, failsToIncident int) (domain.CheckOutcome, error)
	ListHealthChecks(ctx context.Context, serviceID string, limit int) ([]domain.HealthCheck, error)
	DeleteChecksBefore(ctx context.Context, before time.Time) (int64, error)
}

// Prober performs a single health check.
type Prober interface {
	Check(ctx context.Context, cfg domain.CheckConfig) domain.CheckResult
}

// CheckObserver receives the outcome of every recorded check (CU-03 RN-7:
// internal event consumed by notifications). It must not block for long.
type CheckObserver interface {
	OnCheck(ctx context.Context, outcome domain.CheckOutcome)
}

type MonitoringService struct {
	repo            MonitoringRepository
	prober          Prober
	failsToIncident int
	observers       []CheckObserver
	log             *slog.Logger
}

func NewMonitoringService(repo MonitoringRepository, prober Prober, failsToIncident int, log *slog.Logger) *MonitoringService {
	if failsToIncident < 1 {
		failsToIncident = 1
	}
	return &MonitoringService{repo: repo, prober: prober, failsToIncident: failsToIncident, log: log}
}

// AddObserver registers a consumer of check outcomes.
func (s *MonitoringService) AddObserver(o CheckObserver) {
	s.observers = append(s.observers, o)
}

// DueServices returns the services whose next check is due.
func (s *MonitoringService) DueServices(ctx context.Context, limit int) ([]string, error) {
	ids, err := s.repo.ListDueServices(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list due services: %w", err)
	}
	return ids, nil
}

// RunCheck probes the service now and records the result (scheduler and RN-8 manual check).
func (s *MonitoringService) RunCheck(ctx context.Context, serviceID string) (domain.HealthCheck, error) {
	cfg, err := s.repo.GetCheckConfig(ctx, serviceID)
	if err != nil {
		return domain.HealthCheck{}, fmt.Errorf("get check config: %w", err)
	}
	result := s.prober.Check(ctx, cfg)
	outcome, err := s.repo.RecordCheck(ctx, serviceID, result, s.failsToIncident)
	if err != nil {
		return domain.HealthCheck{}, fmt.Errorf("record check: %w", err)
	}
	if outcome.StatusChanged() {
		s.log.Info("service status changed", "service_id", serviceID, "from", outcome.PreviousStatus, "to", outcome.Check.Status)
	}
	if outcome.IncidentOpenedID != "" {
		s.log.Warn("automatic incident opened", "service_id", serviceID, "incident_id", outcome.IncidentOpenedID)
	}
	for _, o := range s.observers {
		o.OnCheck(ctx, outcome)
	}
	return outcome.Check, nil
}

func (s *MonitoringService) ListChecks(ctx context.Context, serviceID string, limit int) ([]domain.HealthCheck, error) {
	if limit == 0 {
		limit = defaultChecksLimit
	}
	if limit < 1 || limit > maxChecksLimit {
		return nil, domain.NewFieldError(domain.ErrValidation, "El límite debe estar entre 1 y 500", "limit", "fuera de rango")
	}
	if _, err := s.repo.GetCheckConfig(ctx, serviceID); err != nil {
		return nil, fmt.Errorf("get service: %w", err)
	}
	checks, err := s.repo.ListHealthChecks(ctx, serviceID, limit)
	if err != nil {
		return nil, fmt.Errorf("list checks: %w", err)
	}
	return checks, nil
}

// CleanUp deletes checks older than the retention window (RN-9).
func (s *MonitoringService) CleanUp(ctx context.Context, now time.Time) error {
	n, err := s.repo.DeleteChecksBefore(ctx, now.Add(-checkRetention))
	if err != nil {
		return fmt.Errorf("delete old checks: %w", err)
	}
	if n > 0 {
		s.log.Info("old health checks deleted", "count", n)
	}
	return nil
}
