package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	dbgen "github.com/fabri-molinelli-w/opl/backend/internal/db/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MonitoringRepository persists health checks and applies their side effects
// (service status, consecutive failures, automatic incidents) atomically.
type MonitoringRepository struct {
	q    *dbgen.Queries
	pool *pgxpool.Pool
}

func NewMonitoringRepository(q *dbgen.Queries, pool *pgxpool.Pool) *MonitoringRepository {
	return &MonitoringRepository{q: q, pool: pool}
}

func (r *MonitoringRepository) ListDueServices(ctx context.Context, limit int) ([]string, error) {
	ids, err := r.q.ListDueServices(ctx, int32(limit))
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, uuidToString(id))
	}
	return out, nil
}

func (r *MonitoringRepository) GetCheckConfig(ctx context.Context, serviceID string) (domain.CheckConfig, error) {
	id, err := uuidFromString(serviceID)
	if err != nil {
		return domain.CheckConfig{}, domain.ErrNotFound
	}
	row, err := r.q.GetServiceCheckConfig(ctx, id)
	if err != nil {
		return domain.CheckConfig{}, mapError(err)
	}
	return domain.CheckConfig{
		ServiceID:          uuidToString(row.ID),
		Name:               row.Name,
		CheckType:          row.CheckType,
		CheckTarget:        row.CheckTarget,
		ExpectedStatus:     int(row.ExpectedStatus),
		LatencyThresholdMs: int(row.LatencyThresholdMs),
	}, nil
}

// RecordCheck stores the result and, in the same transaction, updates the
// service status (RN-6), opens an automatic incident after failsToIncident
// consecutive DOWN checks (RN-7b / CU-10 RN-4) and logs the recovery on the
// open incident (CU-10 RN-5).
func (r *MonitoringRepository) RecordCheck(
	ctx context.Context, serviceID string, result domain.CheckResult, failsToIncident int,
) (domain.CheckOutcome, error) {
	id, err := uuidFromString(serviceID)
	if err != nil {
		return domain.CheckOutcome{}, domain.ErrNotFound
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.CheckOutcome{}, fmt.Errorf("begin record check transaction: %w", err)
	}
	qtx := r.q.WithTx(tx)

	state, err := qtx.LockServiceCheckState(ctx, id)
	if err != nil {
		return domain.CheckOutcome{}, rollbackWithError(ctx, tx, mapError(err))
	}
	row, err := qtx.InsertHealthCheck(ctx, dbgen.InsertHealthCheckParams{
		ServiceID:  id,
		Status:     result.Status,
		ResponseMs: optionalInt4(result.ResponseMs),
		HttpStatus: optionalInt4(result.HTTPStatus),
		Error:      pgtype.Text{String: result.Error, Valid: result.Error != ""},
	})
	if err != nil {
		return domain.CheckOutcome{}, rollbackWithError(ctx, tx, fmt.Errorf("insert health check: %w", err))
	}
	failures, err := qtx.UpdateServiceCheckState(ctx, dbgen.UpdateServiceCheckStateParams{ID: id, Status: result.Status})
	if err != nil {
		return domain.CheckOutcome{}, rollbackWithError(ctx, tx, fmt.Errorf("update service status: %w", err))
	}

	outcome := domain.CheckOutcome{
		Check:          toDomainHealthCheck(row),
		ServerID:       uuidToString(state.ServerID),
		ServiceName:    state.Name,
		PreviousStatus: state.CurrentStatus,
	}

	if result.Status == domain.StatusDown && int(failures) >= failsToIncident {
		incidentID, err := qtx.CreateAutoIncident(ctx, dbgen.CreateAutoIncidentParams{
			Title:       fmt.Sprintf("Servicio %s caído", state.Name),
			Description: fmt.Sprintf("Abierto automáticamente tras %d chequeos fallidos consecutivos.", failures),
			ServiceID:   id,
		})
		switch {
		case err == nil:
			if err := qtx.InsertIncidentEvent(ctx, dbgen.InsertIncidentEventParams{
				IncidentID: incidentID,
				Kind:       domain.EventSystem,
				Body:       fmt.Sprintf("Incidente abierto automáticamente: %d chequeos DOWN seguidos. Último error: %s", failures, result.Error),
			}); err != nil {
				return domain.CheckOutcome{}, rollbackWithError(ctx, tx, fmt.Errorf("insert auto incident event: %w", err))
			}
			outcome.IncidentOpenedID = uuidToString(incidentID)
		case errors.Is(err, pgx.ErrNoRows):
			// Ya hay un incidente automático abierto para este servicio.
		default:
			return domain.CheckOutcome{}, rollbackWithError(ctx, tx, fmt.Errorf("create auto incident: %w", err))
		}
	}

	if result.Status == domain.StatusUp && state.CurrentStatus == domain.StatusDown {
		open, err := qtx.GetOpenAutoIncidentForService(ctx, id)
		switch {
		case err == nil:
			if err := qtx.InsertIncidentEvent(ctx, dbgen.InsertIncidentEventParams{
				IncidentID: open.ID,
				Kind:       domain.EventSystem,
				Body:       "El servicio volvió a responder.",
			}); err != nil {
				return domain.CheckOutcome{}, rollbackWithError(ctx, tx, fmt.Errorf("insert recovery event: %w", err))
			}
		case errors.Is(err, pgx.ErrNoRows):
		default:
			return domain.CheckOutcome{}, rollbackWithError(ctx, tx, fmt.Errorf("find open incident: %w", err))
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.CheckOutcome{}, fmt.Errorf("commit record check transaction: %w", err)
	}
	return outcome, nil
}

func (r *MonitoringRepository) ListHealthChecks(ctx context.Context, serviceID string, limit int) ([]domain.HealthCheck, error) {
	id, err := uuidFromString(serviceID)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	rows, err := r.q.ListHealthChecks(ctx, dbgen.ListHealthChecksParams{ServiceID: id, Limit: int32(limit)})
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]domain.HealthCheck, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainHealthCheck(row))
	}
	return out, nil
}

func (r *MonitoringRepository) DeleteChecksBefore(ctx context.Context, before time.Time) (int64, error) {
	n, err := r.q.DeleteHealthChecksBefore(ctx, pgtype.Timestamptz{Time: before, Valid: true})
	if err != nil {
		return 0, mapError(err)
	}
	return n, nil
}

func toDomainHealthCheck(row dbgen.HealthCheck) domain.HealthCheck {
	return domain.HealthCheck{
		ID:         uuidToString(row.ID),
		ServiceID:  uuidToString(row.ServiceID),
		Status:     row.Status,
		ResponseMs: int4Ptr(row.ResponseMs),
		HTTPStatus: int4Ptr(row.HttpStatus),
		Error:      row.Error.String,
		CheckedAt:  row.CheckedAt.Time,
	}
}

func optionalInt4(v *int) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(*v), Valid: true}
}

func int4Ptr(v pgtype.Int4) *int {
	if !v.Valid {
		return nil
	}
	n := int(v.Int32)
	return &n
}
