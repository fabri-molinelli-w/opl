package repository

import (
	"context"
	"errors"
	"fmt"

	dbgen "github.com/fabri-molinelli-w/opl/backend/internal/db/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ServerRepository is a thin wrapper over the sqlc-generated server queries.
type ServerRepository struct {
	q    *dbgen.Queries
	pool *pgxpool.Pool
}

func NewServerRepository(q *dbgen.Queries, pool *pgxpool.Pool) *ServerRepository {
	return &ServerRepository{q: q, pool: pool}
}

func (r *ServerRepository) CreateServer(ctx context.Context, server domain.Server) (domain.Server, error) {
	ip, err := addrFromString(server.IP)
	if err != nil {
		return domain.Server{}, err
	}
	row, err := r.q.CreateServer(ctx, dbgen.CreateServerParams{
		Hostname:    server.Hostname,
		Column2:     ip,
		Environment: server.Environment,
		Os:          server.OS,
		Location:    server.Location,
		Owner:       server.Owner,
		Tags:        server.Tags,
	})
	if err != nil {
		return domain.Server{}, mapError(err)
	}
	return toDomainServer(row), nil
}

func (r *ServerRepository) GetServerByID(ctx context.Context, id string) (domain.Server, error) {
	uuid, err := uuidFromString(id)
	if err != nil {
		return domain.Server{}, domain.ErrNotFound
	}
	row, err := r.q.GetServerByID(ctx, uuid)
	if err != nil {
		return domain.Server{}, mapError(err)
	}
	return toDomainServer(dbgen.CreateServerRow(row)), nil
}

func (r *ServerRepository) ListServers(ctx context.Context, filter domain.ServerFilter) ([]domain.Server, error) {
	rows, err := r.q.ListServers(ctx, listServersParams(filter))
	if err != nil {
		return nil, mapError(err)
	}
	servers := make([]domain.Server, 0, len(rows))
	for _, row := range rows {
		servers = append(servers, toDomainServer(dbgen.CreateServerRow(row)))
	}
	return servers, nil
}

func (r *ServerRepository) CountServers(ctx context.Context, filter domain.ServerFilter) (int64, error) {
	count, err := r.q.CountServers(ctx, dbgen.CountServersParams{
		Q:           textParam(filter.Q),
		Environment: textParam(filter.Environment),
		Os:          textParam(filter.OS),
		Tag:         textParam(filter.Tag),
	})
	if err != nil {
		return 0, mapError(err)
	}
	return count, nil
}

func (r *ServerRepository) UpdateServer(ctx context.Context, id string, update domain.ServerUpdate) (domain.Server, error) {
	uuid, err := uuidFromString(id)
	if err != nil {
		return domain.Server{}, domain.ErrNotFound
	}
	params := dbgen.UpdateServerParams{ID: uuid}
	if update.Hostname != nil {
		params.Hostname = textParam(*update.Hostname)
	}
	if update.IP != nil {
		ip, err := addrFromString(*update.IP)
		if err != nil {
			return domain.Server{}, err
		}
		params.Ip = &ip
	}
	if update.Environment != nil {
		params.Environment = textParam(*update.Environment)
	}
	if update.OS != nil {
		params.Os = textValue(*update.OS)
	}
	if update.Location != nil {
		params.Location = textParam(*update.Location)
	}
	if update.Owner != nil {
		params.Owner = textValue(*update.Owner)
	}
	if update.Tags != nil {
		params.Tags = *update.Tags
	}
	row, err := r.q.UpdateServer(ctx, params)
	if err != nil {
		return domain.Server{}, mapError(err)
	}
	return toDomainServer(dbgen.CreateServerRow(row)), nil
}

// DeleteServer soft-deletes the server and all its services inside a single
// transaction (RN-6).
func (r *ServerRepository) DeleteServer(ctx context.Context, id string) error {
	uuid, err := uuidFromString(id)
	if err != nil {
		return domain.ErrNotFound
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin delete server transaction: %w", err)
	}
	qtx := r.q.WithTx(tx)

	rows, err := qtx.SoftDeleteServer(ctx, uuid)
	if err != nil {
		return rollbackWithError(ctx, tx, fmt.Errorf("delete server: %w", err))
	}
	if rows == 0 {
		return rollbackWithError(ctx, tx, domain.ErrNotFound)
	}
	if err := qtx.SoftDeleteServicesByServer(ctx, uuid); err != nil {
		return rollbackWithError(ctx, tx, fmt.Errorf("delete server services: %w", err))
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete server transaction: %w", err)
	}
	return nil
}

func rollbackWithError(ctx context.Context, tx pgx.Tx, cause error) error {
	if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		return fmt.Errorf("%w (rollback failed: %v)", cause, err)
	}
	return cause
}

func listServersParams(filter domain.ServerFilter) dbgen.ListServersParams {
	return dbgen.ListServersParams{
		Q:           textParam(filter.Q),
		Environment: textParam(filter.Environment),
		Os:          textParam(filter.OS),
		Tag:         textParam(filter.Tag),
		Offset:      int32((filter.Page - 1) * filter.PageSize),
		Limit:       int32(filter.PageSize),
	}
}

func toDomainServer(row dbgen.CreateServerRow) domain.Server {
	return domain.Server{
		ID:          uuidToString(row.ID),
		Hostname:    row.Hostname,
		IP:          row.Ip,
		Environment: row.Environment,
		OS:          row.Os,
		Location:    row.Location,
		Owner:       row.Owner,
		Tags:        row.Tags,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}

func textParam(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: value != ""}
}
