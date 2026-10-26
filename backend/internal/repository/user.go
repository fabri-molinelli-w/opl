package repository

import (
	"context"

	dbgen "github.com/fabri-molinelli-w/opl/backend/internal/db/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

// UserRepository is a thin wrapper over the sqlc-generated queries. It maps
// driver errors to domain errors and pgtype values to plain Go types.
type UserRepository struct {
	q *dbgen.Queries
}

func NewUserRepository(q *dbgen.Queries) *UserRepository {
	return &UserRepository{q: q}
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, mapError(err)
	}
	return toDomainUser(row), nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (domain.User, error) {
	uuid, err := uuidFromString(id)
	if err != nil {
		return domain.User{}, domain.ErrNotFound
	}
	row, err := r.q.GetUserByID(ctx, uuid)
	if err != nil {
		return domain.User{}, mapError(err)
	}
	return toDomainUser(row), nil
}

func (r *UserRepository) ListUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := r.q.ListUsers(ctx)
	if err != nil {
		return nil, mapError(err)
	}
	users := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, toDomainUser(row))
	}
	return users, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, email, name, passwordHash string, role domain.Role) (domain.User, error) {
	row, err := r.q.CreateUser(ctx, dbgen.CreateUserParams{
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
		Role:         string(role),
	})
	if err != nil {
		return domain.User{}, mapError(err)
	}
	return toDomainUser(row), nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, id string, update domain.UserUpdate) (domain.User, error) {
	uuid, err := uuidFromString(id)
	if err != nil {
		return domain.User{}, domain.ErrNotFound
	}
	params := dbgen.UpdateUserParams{ID: uuid}
	if update.Name != nil {
		params.Name = pgtype.Text{String: *update.Name, Valid: true}
	}
	if update.Email != nil {
		params.Email = pgtype.Text{String: *update.Email, Valid: true}
	}
	if update.Role != nil {
		params.Role = pgtype.Text{String: string(*update.Role), Valid: true}
	}
	if update.Active != nil {
		params.Active = pgtype.Bool{Bool: *update.Active, Valid: true}
	}
	row, err := r.q.UpdateUser(ctx, params)
	if err != nil {
		return domain.User{}, mapError(err)
	}
	return toDomainUser(row), nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	uuid, err := uuidFromString(id)
	if err != nil {
		return domain.ErrNotFound
	}
	if err := r.q.UpdateUserPassword(ctx, dbgen.UpdateUserPasswordParams{ID: uuid, PasswordHash: passwordHash}); err != nil {
		return mapError(err)
	}
	return nil
}

func (r *UserRepository) CountAdmins(ctx context.Context) (int64, error) {
	count, err := r.q.CountAdmins(ctx)
	if err != nil {
		return 0, mapError(err)
	}
	return count, nil
}

func toDomainUser(row dbgen.User) domain.User {
	return domain.User{
		ID:           uuidToString(row.ID),
		Email:        row.Email,
		Name:         row.Name,
		PasswordHash: row.PasswordHash,
		Role:         domain.Role(row.Role),
		Active:       row.Active,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
}
