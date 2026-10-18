package handlers

import (
	"context"
	"log/slog"

	gen "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/fabri-molinelli-w/opl/backend/internal/service"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Pinger is the subset of *pgxpool.Pool used by the health handler.
type Pinger interface {
	Ping(ctx context.Context) error
}

// AuthService is the auth use cases consumed by the handlers.
type AuthService interface {
	Login(ctx context.Context, email, password string) (domain.User, string, error)
	GetUser(ctx context.Context, id string) (domain.User, error)
	ChangePassword(ctx context.Context, userID, current, next string) error
}

// UserService is the user management use cases consumed by the handlers.
type UserService interface {
	List(ctx context.Context) ([]domain.User, error)
	Create(ctx context.Context, in service.CreateUserInput) (domain.User, error)
	Update(ctx context.Context, id string, update domain.UserUpdate, actor domain.User) (domain.User, error)
}

// Server implements api.StrictServerInterface. It only validates input,
// converts generated types to domain types and delegates to the services.
type Server struct {
	db    Pinger
	auth  AuthService
	users UserService
	log   *slog.Logger
}

func NewServer(db Pinger, auth AuthService, users UserService, log *slog.Logger) *Server {
	return &Server{db: db, auth: auth, users: users, log: log}
}

// toGenUser converts a domain user to the API schema. The password hash is
// never included.
func toGenUser(user domain.User) gen.User {
	return gen.User{
		Id:        user.ID,
		Email:     openapi_types.Email(user.Email),
		Name:      user.Name,
		Role:      gen.Role(user.Role),
		Active:    user.Active,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
