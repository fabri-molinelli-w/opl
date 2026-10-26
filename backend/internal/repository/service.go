package repository

import (
	"context"

	dbgen "github.com/fabri-molinelli-w/opl/backend/internal/db/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

// ServiceRepository is a thin wrapper over the sqlc-generated service queries.
type ServiceRepository struct {
	q *dbgen.Queries
}

func NewServiceRepository(q *dbgen.Queries) *ServiceRepository {
	return &ServiceRepository{q: q}
}

func (r *ServiceRepository) CreateService(ctx context.Context, service domain.Service) (domain.Service, error) {
	serverID, err := uuidFromString(service.ServerID)
	if err != nil {
		return domain.Service{}, domain.ErrNotFound
	}
	row, err := r.q.CreateService(ctx, dbgen.CreateServiceParams{
		ServerID:           serverID,
		Name:               service.Name,
		Type:               service.Type,
		Port:               int32(service.Port),
		CheckType:          service.CheckType,
		CheckTarget:        service.CheckTarget,
		ExpectedStatus:     int32(service.ExpectedStatus),
		IntervalSeconds:    int32(service.IntervalSeconds),
		LatencyThresholdMs: int32(service.LatencyThresholdMs),
	})
	if err != nil {
		return domain.Service{}, mapError(err)
	}
	return toDomainService(row), nil
}

func (r *ServiceRepository) GetServiceByID(ctx context.Context, id string) (domain.Service, error) {
	uuid, err := uuidFromString(id)
	if err != nil {
		return domain.Service{}, domain.ErrNotFound
	}
	row, err := r.q.GetServiceByID(ctx, uuid)
	if err != nil {
		return domain.Service{}, mapError(err)
	}
	return toDomainService(row), nil
}

func (r *ServiceRepository) ListServicesByServerID(ctx context.Context, serverID string) ([]domain.Service, error) {
	uuid, err := uuidFromString(serverID)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	rows, err := r.q.ListServicesByServerID(ctx, uuid)
	if err != nil {
		return nil, mapError(err)
	}
	services := make([]domain.Service, 0, len(rows))
	for _, row := range rows {
		services = append(services, toDomainService(row))
	}
	return services, nil
}

func (r *ServiceRepository) ListServices(ctx context.Context, filter domain.ServiceFilter) ([]domain.Service, error) {
	params, err := listServicesParams(filter)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListServices(ctx, params)
	if err != nil {
		return nil, mapError(err)
	}
	services := make([]domain.Service, 0, len(rows))
	for _, row := range rows {
		services = append(services, toDomainService(row))
	}
	return services, nil
}

func (r *ServiceRepository) CountServices(ctx context.Context, filter domain.ServiceFilter) (int64, error) {
	serverID, err := serviceFilterServerID(filter)
	if err != nil {
		return 0, err
	}
	count, err := r.q.CountServices(ctx, dbgen.CountServicesParams{
		ServerID:    serverID,
		Environment: textParam(filter.Environment),
		Status:      textParam(filter.Status),
		Type:        textParam(filter.Type),
	})
	if err != nil {
		return 0, mapError(err)
	}
	return count, nil
}

func (r *ServiceRepository) UpdateService(ctx context.Context, id string, update domain.ServiceUpdate) (domain.Service, error) {
	uuid, err := uuidFromString(id)
	if err != nil {
		return domain.Service{}, domain.ErrNotFound
	}
	params := dbgen.UpdateServiceParams{ID: uuid}
	if update.Name != nil {
		params.Name = textParam(*update.Name)
	}
	if update.Type != nil {
		params.Type = textParam(*update.Type)
	}
	if update.Port != nil {
		params.Port = int4Param(*update.Port)
	}
	if update.CheckType != nil {
		params.CheckType = textParam(*update.CheckType)
	}
	if update.CheckTarget != nil {
		params.CheckTarget = textParam(*update.CheckTarget)
	}
	if update.ExpectedStatus != nil {
		params.ExpectedStatus = int4Param(*update.ExpectedStatus)
	}
	if update.IntervalSeconds != nil {
		params.IntervalSeconds = int4Param(*update.IntervalSeconds)
	}
	if update.LatencyThresholdMs != nil {
		params.LatencyThresholdMs = int4Param(*update.LatencyThresholdMs)
	}
	row, err := r.q.UpdateService(ctx, params)
	if err != nil {
		return domain.Service{}, mapError(err)
	}
	return toDomainService(row), nil
}

func (r *ServiceRepository) DeleteService(ctx context.Context, id string) error {
	uuid, err := uuidFromString(id)
	if err != nil {
		return domain.ErrNotFound
	}
	rows, err := r.q.SoftDeleteService(ctx, uuid)
	if err != nil {
		return mapError(err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func listServicesParams(filter domain.ServiceFilter) (dbgen.ListServicesParams, error) {
	serverID, err := serviceFilterServerID(filter)
	if err != nil {
		return dbgen.ListServicesParams{}, err
	}
	return dbgen.ListServicesParams{
		ServerID:    serverID,
		Environment: textParam(filter.Environment),
		Status:      textParam(filter.Status),
		Type:        textParam(filter.Type),
		Offset:      int32((filter.Page - 1) * filter.PageSize),
		Limit:       int32(filter.PageSize),
	}, nil
}

func serviceFilterServerID(filter domain.ServiceFilter) (pgtype.UUID, error) {
	if filter.ServerID == "" {
		return pgtype.UUID{}, nil
	}
	uuid, err := uuidFromString(filter.ServerID)
	if err != nil {
		return pgtype.UUID{}, domain.NewFieldError(
			domain.ErrValidation, "El serverId no es un UUID válido", "serverId", "UUID inválido")
	}
	return uuid, nil
}

func int4Param(value int) pgtype.Int4 {
	return pgtype.Int4{Int32: int32(value), Valid: true}
}

func toDomainService(row dbgen.Service) domain.Service {
	return domain.Service{
		ID:                 uuidToString(row.ID),
		ServerID:           uuidToString(row.ServerID),
		Name:               row.Name,
		Type:               row.Type,
		Port:               int(row.Port),
		CheckType:          row.CheckType,
		CheckTarget:        row.CheckTarget,
		ExpectedStatus:     int(row.ExpectedStatus),
		IntervalSeconds:    int(row.IntervalSeconds),
		LatencyThresholdMs: int(row.LatencyThresholdMs),
		CurrentStatus:      row.CurrentStatus,
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
	}
}
