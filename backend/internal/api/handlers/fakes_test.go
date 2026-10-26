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

type fakeServerService struct {
	page      domain.ServerPage
	server    domain.Server
	listErr   error
	getErr    error
	createErr error
	updateErr error
	deleteErr error
}

func (f fakeServerService) List(_ context.Context, filter domain.ServerFilter) (domain.ServerPage, error) {
	return f.page, f.listErr
}

func (f fakeServerService) Get(context.Context, string) (domain.Server, error) {
	return f.server, f.getErr
}

func (f fakeServerService) Create(context.Context, service.CreateServerInput) (domain.Server, error) {
	return f.server, f.createErr
}

func (f fakeServerService) Update(context.Context, string, domain.ServerUpdate) (domain.Server, error) {
	return f.server, f.updateErr
}

func (f fakeServerService) Delete(context.Context, string) error {
	return f.deleteErr
}

type fakeServiceCatalog struct {
	page      domain.ServicePage
	service   domain.Service
	listErr   error
	getErr    error
	createErr error
	updateErr error
	deleteErr error
}

func (f fakeServiceCatalog) List(context.Context, domain.ServiceFilter) (domain.ServicePage, error) {
	return f.page, f.listErr
}

func (f fakeServiceCatalog) Get(context.Context, string) (domain.Service, error) {
	return f.service, f.getErr
}

func (f fakeServiceCatalog) Create(context.Context, service.CreateServiceInput) (domain.Service, error) {
	return f.service, f.createErr
}

func (f fakeServiceCatalog) Update(context.Context, string, domain.ServiceUpdate) (domain.Service, error) {
	return f.service, f.updateErr
}

func (f fakeServiceCatalog) Delete(context.Context, string) error {
	return f.deleteErr
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

	server := NewServer(Deps{
		DB:       fakePinger{},
		Auth:     authSvc,
		Users:    userSvc,
		Servers:  fakeServerService{},
		Services: fakeServiceCatalog{},
		Log:      log,
	})
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
