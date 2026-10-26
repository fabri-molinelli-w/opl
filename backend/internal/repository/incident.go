package repository

import (
	"context"
	"fmt"
	"time"

	dbgen "github.com/fabri-molinelli-w/opl/backend/internal/db/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IncidentRepository wraps the incident queries. Every status change writes the
// incident and its timeline event in the same transaction (CU-10 RN-3).
type IncidentRepository struct {
	q    *dbgen.Queries
	pool *pgxpool.Pool
}

func NewIncidentRepository(q *dbgen.Queries, pool *pgxpool.Pool) *IncidentRepository {
	return &IncidentRepository{q: q, pool: pool}
}

func optionalUUID(s string) (pgtype.UUID, error) {
	if s == "" {
		return pgtype.UUID{}, nil
	}
	return uuidFromString(s)
}

func (r *IncidentRepository) CreateIncident(ctx context.Context, in domain.NewIncident) (string, error) {
	serviceID, err := optionalUUID(in.ServiceID)
	if err != nil {
		return "", domain.NewFieldError(domain.ErrValidation, "El servicio no existe", "serviceId", "id inválido")
	}
	assigneeID, err := optionalUUID(in.AssigneeID)
	if err != nil {
		return "", domain.NewFieldError(domain.ErrValidation, "El responsable no existe", "assigneeId", "id inválido")
	}
	authorID, err := optionalUUID(in.AuthorID)
	if err != nil {
		return "", fmt.Errorf("author id: %w", err)
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin create incident transaction: %w", err)
	}
	qtx := r.q.WithTx(tx)
	id, err := qtx.CreateIncident(ctx, dbgen.CreateIncidentParams{
		Title: in.Title, Description: in.Description, ServiceID: serviceID, Severity: in.Severity, AssigneeID: assigneeID,
	})
	if err != nil {
		return "", rollbackWithError(ctx, tx, mapIncidentFKError(err))
	}
	if err := qtx.InsertIncidentEvent(ctx, dbgen.InsertIncidentEventParams{
		IncidentID: id, AuthorID: authorID, Kind: domain.EventStatusChange, Body: "Incidente abierto (" + in.Severity + ").",
	}); err != nil {
		return "", rollbackWithError(ctx, tx, fmt.Errorf("insert open event: %w", err))
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit create incident transaction: %w", err)
	}
	return uuidToString(id), nil
}

func (r *IncidentRepository) GetIncident(ctx context.Context, id string) (domain.Incident, error) {
	uuid, err := uuidFromString(id)
	if err != nil {
		return domain.Incident{}, domain.ErrNotFound
	}
	row, err := r.q.GetIncident(ctx, uuid)
	if err != nil {
		return domain.Incident{}, mapError(err)
	}
	events, err := r.q.ListIncidentEvents(ctx, uuid)
	if err != nil {
		return domain.Incident{}, mapError(err)
	}
	incident := domain.Incident{
		ID:           uuidToString(row.ID),
		Title:        row.Title,
		Description:  row.Description,
		ServiceID:    uuidToString(row.ServiceID),
		ServiceName:  row.ServiceName,
		Severity:     row.Severity,
		Status:       row.Status,
		AssigneeID:   uuidToString(row.AssigneeID),
		AssigneeName: row.AssigneeName,
		RootCause:    row.RootCause.String,
		AutoCreated:  row.AutoCreated,
		OpenedAt:     row.OpenedAt.Time,
		ResolvedAt:   timePtr(row.ResolvedAt),
		ClosedAt:     timePtr(row.ClosedAt),
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
		Events:       make([]domain.IncidentEvent, 0, len(events)),
	}
	for _, e := range events {
		incident.Events = append(incident.Events, domain.IncidentEvent{
			ID: uuidToString(e.ID), Kind: e.Kind, Body: e.Body, AuthorName: e.AuthorName, CreatedAt: e.CreatedAt.Time,
		})
	}
	return incident, nil
}

func incidentFilterParams(f domain.IncidentFilter) (pgtype.Text, pgtype.Text, pgtype.UUID, error) {
	serviceID, err := optionalUUID(f.ServiceID)
	if err != nil {
		return pgtype.Text{}, pgtype.Text{}, pgtype.UUID{}, domain.NewFieldError(
			domain.ErrValidation, "El servicio no es válido", "serviceId", "id inválido")
	}
	return textParam(f.Status), textParam(f.Severity), serviceID, nil
}

func (r *IncidentRepository) ListIncidents(ctx context.Context, f domain.IncidentFilter) ([]domain.Incident, error) {
	status, severity, serviceID, err := incidentFilterParams(f)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListIncidents(ctx, dbgen.ListIncidentsParams{
		Status: status, Severity: severity, ServiceID: serviceID, OnlyActive: f.OnlyActive,
		Offset: int32((f.Page - 1) * f.PageSize), Limit: int32(f.PageSize),
	})
	if err != nil {
		return nil, mapError(err)
	}
	items := make([]domain.Incident, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.Incident{
			ID:           uuidToString(row.ID),
			Title:        row.Title,
			ServiceID:    uuidToString(row.ServiceID),
			ServiceName:  row.ServiceName,
			Severity:     row.Severity,
			Status:       row.Status,
			AssigneeName: row.AssigneeName,
			AutoCreated:  row.AutoCreated,
			OpenedAt:     row.OpenedAt.Time,
			ResolvedAt:   timePtr(row.ResolvedAt),
		})
	}
	return items, nil
}

func (r *IncidentRepository) CountIncidents(ctx context.Context, f domain.IncidentFilter) (int64, error) {
	status, severity, serviceID, err := incidentFilterParams(f)
	if err != nil {
		return 0, err
	}
	n, err := r.q.CountIncidents(ctx, dbgen.CountIncidentsParams{
		Status: status, Severity: severity, ServiceID: serviceID, OnlyActive: f.OnlyActive,
	})
	if err != nil {
		return 0, mapError(err)
	}
	return n, nil
}

func (r *IncidentRepository) UpdateIncident(ctx context.Context, id string, u domain.IncidentUpdate) error {
	uuid, err := uuidFromString(id)
	if err != nil {
		return domain.ErrNotFound
	}
	params := dbgen.UpdateIncidentParams{ID: uuid}
	if u.Title != nil {
		params.Title = textValue(*u.Title)
	}
	if u.Description != nil {
		params.Description = textValue(*u.Description)
	}
	if u.Severity != nil {
		params.Severity = textValue(*u.Severity)
	}
	if u.AssigneeID != nil {
		assignee, err := uuidFromString(*u.AssigneeID)
		if err != nil {
			return domain.NewFieldError(domain.ErrValidation, "El responsable no existe", "assigneeId", "id inválido")
		}
		params.AssigneeID = assignee
	}
	n, err := r.q.UpdateIncident(ctx, params)
	if err != nil {
		return mapIncidentFKError(err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// TransitionIncident locks the incident, lets validate check the state machine
// and writes the new status plus its timeline event atomically.
func (r *IncidentRepository) TransitionIncident(
	ctx context.Context, id, to string, rootCause *string, actorID, body string, validate func(from string) error,
) error {
	uuid, err := uuidFromString(id)
	if err != nil {
		return domain.ErrNotFound
	}
	actor, err := optionalUUID(actorID)
	if err != nil {
		return fmt.Errorf("actor id: %w", err)
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transition transaction: %w", err)
	}
	qtx := r.q.WithTx(tx)
	from, err := qtx.LockIncidentStatus(ctx, uuid)
	if err != nil {
		return rollbackWithError(ctx, tx, mapError(err))
	}
	if err := validate(from); err != nil {
		return rollbackWithError(ctx, tx, err)
	}
	rc := pgtype.Text{}
	if rootCause != nil {
		rc = textValue(*rootCause)
	}
	if err := qtx.UpdateIncidentStatus(ctx, dbgen.UpdateIncidentStatusParams{ID: uuid, Status: to, RootCause: rc}); err != nil {
		return rollbackWithError(ctx, tx, mapError(err))
	}
	if err := qtx.InsertIncidentEvent(ctx, dbgen.InsertIncidentEventParams{
		IncidentID: uuid, AuthorID: actor, Kind: domain.EventStatusChange, Body: body,
	}); err != nil {
		return rollbackWithError(ctx, tx, fmt.Errorf("insert transition event: %w", err))
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transition transaction: %w", err)
	}
	return nil
}

func (r *IncidentRepository) AddEvent(ctx context.Context, id, authorID, kind, body string) error {
	uuid, err := uuidFromString(id)
	if err != nil {
		return domain.ErrNotFound
	}
	author, err := optionalUUID(authorID)
	if err != nil {
		return fmt.Errorf("author id: %w", err)
	}
	if _, err := r.q.LockIncidentStatus(ctx, uuid); err != nil {
		return mapError(err)
	}
	if err := r.q.InsertIncidentEvent(ctx, dbgen.InsertIncidentEventParams{
		IncidentID: uuid, AuthorID: author, Kind: kind, Body: body,
	}); err != nil {
		return mapError(err)
	}
	return nil
}

// mapIncidentFKError turns a foreign key violation (unknown service/assignee) into a 400.
func mapIncidentFKError(err error) error {
	if isForeignKeyViolation(err) {
		return domain.NewError(domain.ErrValidation, "El servicio o el responsable indicado no existe")
	}
	return mapError(err)
}

func timePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}
