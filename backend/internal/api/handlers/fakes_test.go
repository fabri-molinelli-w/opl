package handlers

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"

	gen "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/auth"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/fabri-molinelli-w/opl/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type fakeAuthService struct {
	loginUser  domain.User
	loginToken string
	loginErr   error
	changeErr  error
}

func (f fakeAuthService) Login(context.Context, string, string) (domain.User, string, error) {
	return f.loginUser, f.loginToken, f.loginErr
}

func (f fakeAuthService) GetUser(context.Context, string) (domain.User, error) {
	return f.loginUser, nil
}

func (f fakeAuthService) ChangePassword(context.Context, string, string, string) error {
	return f.changeErr
}

type fakeUserService struct {
	users     []domain.User
	created   domain.User
	updated   domain.User
	createErr error
	updateErr error
}

func (f fakeUserService) List(context.Context) ([]domain.User, error) {
	return f.users, nil
}

func (f fakeUserService) Create(context.Context, service.CreateUserInput) (domain.User, error) {
	return f.created, f.createErr
}

func (f fakeUserService) Update(context.Context, string, domain.UserUpdate, domain.User) (domain.User, error) {
	return f.updated, f.updateErr
}

type fakeLookup struct {
	users map[string]domain.User
}

func (f fakeLookup) GetUserByID(_ context.Context, id string) (domain.User, error) {
	user, ok := f.users[id]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return user, nil
}

func newAuthorizedRouter(t *testing.T, authSvc AuthService, userSvc UserService, lookup auth.UserLookup, tokens auth.TokenParser, policy auth.Policy) http.Handler {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	server := NewServer(fakePinger{}, authSvc, userSvc, log)
	middleware := auth.Authorize(policy, tokens, lookup)
	strict := gen.NewStrictHandlerWithOptions(
		server,
		[]gen.StrictMiddlewareFunc{middleware},
		NewStrictOptions(log),
	)

	r := chi.NewRouter()
	gen.HandlerFromMux(strict, r)
	return r
}
