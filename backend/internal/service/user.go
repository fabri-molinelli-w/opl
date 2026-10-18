package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/fabri-molinelli-w/opl/backend/internal/auth"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

const minPasswordLength = 8

// UserRepository is the subset of the user repository needed by UserService.
type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (domain.User, error)
	ListUsers(ctx context.Context) ([]domain.User, error)
	CreateUser(ctx context.Context, email, name, passwordHash string, role domain.Role) (domain.User, error)
	UpdateUser(ctx context.Context, id string, update domain.UserUpdate) (domain.User, error)
}

// CreateUserInput carries the data needed to create a user.
type CreateUserInput struct {
	Email    string
	Name     string
	Password string
	Role     domain.Role
}

type UserService struct {
	repo UserRepository
	log  *slog.Logger
}

func NewUserService(repo UserRepository, log *slog.Logger) *UserService {
	return &UserService{repo: repo, log: log}
}

func (s *UserService) List(ctx context.Context) ([]domain.User, error) {
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (s *UserService) Get(ctx context.Context, id string) (domain.User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (s *UserService) Create(ctx context.Context, in CreateUserInput) (domain.User, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" {
		return domain.User{}, domain.NewError(domain.ErrValidation, "el email es obligatorio")
	}
	if strings.TrimSpace(in.Name) == "" {
		return domain.User{}, domain.NewError(domain.ErrValidation, "el nombre es obligatorio")
	}
	if len(in.Password) < minPasswordLength {
		return domain.User{}, domain.NewError(domain.ErrValidation, "la contraseña debe tener al menos 8 caracteres")
	}
	if !in.Role.Valid() {
		return domain.User{}, domain.NewError(domain.ErrValidation, "el rol es inválido")
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, email, strings.TrimSpace(in.Name), hash, in.Role)
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

// Update applies a partial update. It enforces RN-5: an admin cannot demote or
// deactivate themselves.
func (s *UserService) Update(ctx context.Context, id string, update domain.UserUpdate, actor domain.User) (domain.User, error) {
	if actor.ID == id {
		if update.Active != nil && !*update.Active {
			return domain.User{}, domain.NewError(domain.ErrConflict, "no podés desactivar tu propio usuario")
		}
		if update.Role != nil && *update.Role != domain.RoleAdmin {
			return domain.User{}, domain.NewError(domain.ErrConflict, "no podés quitarte el rol de admin")
		}
	}

	if update.Name != nil {
		trimmed := strings.TrimSpace(*update.Name)
		if trimmed == "" {
			return domain.User{}, domain.NewError(domain.ErrValidation, "el nombre no puede estar vacío")
		}
		update.Name = &trimmed
	}
	if update.Email != nil {
		normalized := strings.ToLower(strings.TrimSpace(*update.Email))
		if normalized == "" {
			return domain.User{}, domain.NewError(domain.ErrValidation, "el email no puede estar vacío")
		}
		update.Email = &normalized
	}
	if update.Role != nil && !update.Role.Valid() {
		return domain.User{}, domain.NewError(domain.ErrValidation, "el rol es inválido")
	}

	user, err := s.repo.UpdateUser(ctx, id, update)
	if err != nil {
		return domain.User{}, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}
