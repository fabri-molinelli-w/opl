package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/fabri-molinelli-w/opl/backend/internal/auth"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

// AuthRepository is the subset of the user repository needed by AuthService.
type AuthRepository interface {
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	GetUserByID(ctx context.Context, id string) (domain.User, error)
	CreateUser(ctx context.Context, email, name, passwordHash string, role domain.Role) (domain.User, error)
	UpdatePassword(ctx context.Context, id, passwordHash string) error
	CountAdmins(ctx context.Context) (int64, error)
}

// TokenIssuer issues a signed token for an authenticated user.
type TokenIssuer interface {
	Issue(user domain.User) (string, error)
}

type AuthService struct {
	repo   AuthRepository
	tokens TokenIssuer
	log    *slog.Logger
}

func NewAuthService(repo AuthRepository, tokens TokenIssuer, log *slog.Logger) *AuthService {
	return &AuthService{repo: repo, tokens: tokens, log: log}
}

// Login validates credentials and returns the user plus a signed token. Every
// failure returns the same error so callers cannot tell whether the email or
// the password was wrong.
func (s *AuthService) Login(ctx context.Context, email, password string) (domain.User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			auth.DummyCheck(password)
			return domain.User{}, "", domain.ErrUnauthenticated
		}
		return domain.User{}, "", fmt.Errorf("get user by email: %w", err)
	}

	if !auth.CheckPassword(user.PasswordHash, password) {
		return domain.User{}, "", domain.ErrUnauthenticated
	}
	if !user.Active {
		return domain.User{}, "", domain.ErrUnauthenticated
	}

	token, err := s.tokens.Issue(user)
	if err != nil {
		return domain.User{}, "", fmt.Errorf("issue token: %w", err)
	}
	return user, token, nil
}

// GetUser returns a user by id.
func (s *AuthService) GetUser(ctx context.Context, id string) (domain.User, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

// ChangePassword updates the caller's password after verifying the current one.
func (s *AuthService) ChangePassword(ctx context.Context, userID, current, next string) error {
	if len(next) < minPasswordLength {
		return domain.NewError(domain.ErrValidation, "la nueva contraseña debe tener al menos 8 caracteres")
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user by id: %w", err)
	}
	if !auth.CheckPassword(user.PasswordHash, current) {
		return domain.NewError(domain.ErrValidation, "la contraseña actual es incorrecta")
	}

	hash, err := auth.HashPassword(next)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := s.repo.UpdatePassword(ctx, userID, hash); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

// EnsureAdminBootstrap creates the initial admin (RN-4) when there is no active
// admin yet. It never overwrites an existing admin.
func (s *AuthService) EnsureAdminBootstrap(ctx context.Context, email, password string) error {
	count, err := s.repo.CountAdmins(ctx)
	if err != nil {
		return fmt.Errorf("count admins: %w", err)
	}
	if count > 0 {
		return nil
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || password == "" {
		return fmt.Errorf("%w: ADMIN_EMAIL y ADMIN_PASSWORD son obligatorios para crear el admin inicial", domain.ErrValidation)
	}
	if len(password) < minPasswordLength {
		return fmt.Errorf("%w: ADMIN_PASSWORD debe tener al menos 8 caracteres", domain.ErrValidation)
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	if _, err := s.repo.CreateUser(ctx, email, "Administrador", hash, domain.RoleAdmin); err != nil {
		return fmt.Errorf("create bootstrap admin: %w", err)
	}
	s.log.Info("bootstrap admin created", "email", email)
	return nil
}
