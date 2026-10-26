package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

type fakeServerRepo struct {
	created      domain.Server
	createErr    error
	updated      domain.Server
	updateErr    error
	getResult    domain.Server
	getErr       error
	listResult   []domain.Server
	countResult  int64
	lastFilter   domain.ServerFilter
	lastUpdate   domain.ServerUpdate
	deletedID    string
	deleteErr    error
	lastCreated  domain.Server
	lastServerID string
}

func (f *fakeServerRepo) CreateServer(_ context.Context, server domain.Server) (domain.Server, error) {
	f.lastCreated = server
	return f.created, f.createErr
}

func (f *fakeServerRepo) GetServerByID(_ context.Context, id string) (domain.Server, error) {
	f.lastServerID = id
	return f.getResult, f.getErr
}

func (f *fakeServerRepo) ListServers(_ context.Context, filter domain.ServerFilter) ([]domain.Server, error) {
	f.lastFilter = filter
	return f.listResult, nil
}

func (f *fakeServerRepo) CountServers(_ context.Context, _ domain.ServerFilter) (int64, error) {
	return f.countResult, nil
}

func (f *fakeServerRepo) UpdateServer(_ context.Context, _ string, update domain.ServerUpdate) (domain.Server, error) {
	f.lastUpdate = update
	return f.updated, f.updateErr
}

func (f *fakeServerRepo) DeleteServer(_ context.Context, id string) error {
	f.deletedID = id
	return f.deleteErr
}

type fakeInventoryServices struct {
	services []domain.Service
	err      error
}

func (f fakeInventoryServices) ListServicesByServerID(context.Context, string) ([]domain.Service, error) {
	return f.services, f.err
}

type fakeServiceRepo struct {
	created     domain.Service
	createErr   error
	getResult   domain.Service
	getErr      error
	updated     domain.Service
	updateErr   error
	lastCreated domain.Service
	lastUpdate  domain.ServiceUpdate
	lastFilter  domain.ServiceFilter
	listResult  []domain.Service
	countResult int64
	deletedID   string
	deleteErr   error
}

func (f *fakeServiceRepo) CreateService(_ context.Context, item domain.Service) (domain.Service, error) {
	f.lastCreated = item
	return f.created, f.createErr
}

func (f *fakeServiceRepo) GetServiceByID(_ context.Context, _ string) (domain.Service, error) {
	return f.getResult, f.getErr
}

func (f *fakeServiceRepo) ListServicesByServerID(context.Context, string) ([]domain.Service, error) {
	return f.listResult, nil
}

func (f *fakeServiceRepo) ListServices(_ context.Context, filter domain.ServiceFilter) ([]domain.Service, error) {
	f.lastFilter = filter
	return f.listResult, nil
}

func (f *fakeServiceRepo) CountServices(context.Context, domain.ServiceFilter) (int64, error) {
	return f.countResult, nil
}

func (f *fakeServiceRepo) UpdateService(_ context.Context, _ string, update domain.ServiceUpdate) (domain.Service, error) {
	f.lastUpdate = update
	return f.updated, f.updateErr
}

func (f *fakeServiceRepo) DeleteService(_ context.Context, id string) error {
	f.deletedID = id
	return f.deleteErr
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestCreateServer_NormalizesInput(t *testing.T) {
	repo := &fakeServerRepo{created: domain.Server{ID: "srv-1"}}
	catalog := NewServerService(repo, fakeInventoryServices{}, discardLogger())

	_, err := catalog.Create(context.Background(), CreateServerInput{
		Hostname:    "  Web-01.Example.com ",
		IP:          "2001:0db8:0000:0000:0000:0000:0000:0001",
		Environment: "prod",
		OS:          " Debian 12 ",
		Tags:        []string{"web", "prod"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastCreated.IP != "2001:db8::1" {
		t.Errorf("expected normalized IP, got %q", repo.lastCreated.IP)
	}
	if repo.lastCreated.Hostname != "Web-01.Example.com" {
		t.Errorf("expected trimmed hostname, got %q", repo.lastCreated.Hostname)
	}
	if repo.lastCreated.Location != domain.LocationOnPrem {
		t.Errorf("expected default onprem location, got %q", repo.lastCreated.Location)
	}
}

func TestCreateServer_InvalidIP(t *testing.T) {
	catalog := NewServerService(&fakeServerRepo{}, fakeInventoryServices{}, discardLogger())
	_, err := catalog.Create(context.Background(), CreateServerInput{
		Hostname: "web-01", IP: "not-an-ip", Environment: "prod",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
	details := domain.FieldErrors(err)
	if len(details) != 1 || details[0].Field != "ip" {
		t.Fatalf("expected ip detail, got %#v", details)
	}
}

func TestCreateServer_DuplicateHostnameConflict(t *testing.T) {
	repo := &fakeServerRepo{createErr: domain.ErrConflict}
	catalog := NewServerService(repo, fakeInventoryServices{}, discardLogger())
	_, err := catalog.Create(context.Background(), CreateServerInput{
		Hostname: "web-01", IP: "192.0.2.10", Environment: "prod",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestListServers_DefaultsAndInvalidEnvironment(t *testing.T) {
	repo := &fakeServerRepo{countResult: 1, listResult: []domain.Server{{ID: "srv-1"}}}
	catalog := NewServerService(repo, fakeInventoryServices{}, discardLogger())

	page, err := catalog.List(context.Background(), domain.ServerFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Page != 1 || page.PageSize != defaultPageSize {
		t.Errorf("unexpected pagination: page=%d size=%d", page.Page, page.PageSize)
	}

	_, err = catalog.List(context.Background(), domain.ServerFilter{Environment: "qa"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestGetServer_LoadsServices(t *testing.T) {
	repo := &fakeServerRepo{getResult: domain.Server{ID: "srv-1"}}
	services := fakeInventoryServices{services: []domain.Service{{ID: "svc-1"}}}
	catalog := NewServerService(repo, services, discardLogger())

	server, err := catalog.Get(context.Background(), "srv-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(server.Services) != 1 || server.Services[0].ID != "svc-1" {
		t.Fatalf("expected loaded services, got %#v", server.Services)
	}
}

func TestUpdateServer_EmptyUpdateValidation(t *testing.T) {
	catalog := NewServerService(&fakeServerRepo{}, fakeInventoryServices{}, discardLogger())
	_, err := catalog.Update(context.Background(), "srv-1", domain.ServerUpdate{})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestCreateService_DefaultsAndServerLookup(t *testing.T) {
	servers := &fakeServerRepo{getResult: domain.Server{ID: "srv-1"}}
	repo := &fakeServiceRepo{created: domain.Service{ID: "svc-1"}}
	catalog := NewServiceCatalog(repo, servers, discardLogger())

	_, err := catalog.Create(context.Background(), CreateServiceInput{
		ServerID: "srv-1", Name: " API ", Type: "web", Port: 8080,
		CheckType: "http", CheckTarget: "https://api.example.com/health",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastCreated.ExpectedStatus != 200 || repo.lastCreated.IntervalSeconds != 30 || repo.lastCreated.LatencyThresholdMs != 1000 {
		t.Errorf("unexpected defaults: %#v", repo.lastCreated)
	}
	if repo.lastCreated.Name != "API" {
		t.Errorf("expected trimmed name, got %q", repo.lastCreated.Name)
	}
}

func TestCreateService_UnknownServer(t *testing.T) {
	servers := &fakeServerRepo{getErr: domain.ErrNotFound}
	catalog := NewServiceCatalog(&fakeServiceRepo{}, servers, discardLogger())
	_, err := catalog.Create(context.Background(), CreateServiceInput{
		ServerID: "srv-x", Name: "API", Type: "web", Port: 80, CheckType: "http", CheckTarget: "http://x.example.com",
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCreateService_InvalidHTTPTarget(t *testing.T) {
	servers := &fakeServerRepo{getResult: domain.Server{ID: "srv-1"}}
	catalog := NewServiceCatalog(&fakeServiceRepo{}, servers, discardLogger())
	_, err := catalog.Create(context.Background(), CreateServiceInput{
		ServerID: "srv-1", Name: "API", Type: "web", Port: 80, CheckType: "http", CheckTarget: "ftp://x",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestUpdateService_MergesAndValidates(t *testing.T) {
	repo := &fakeServiceRepo{
		getResult: domain.Service{
			ID: "svc-1", ServerID: "srv-1", Name: "API", Type: "web", Port: 80,
			CheckType: "http", CheckTarget: "http://api.example.com", ExpectedStatus: 200,
			IntervalSeconds: 30, LatencyThresholdMs: 1000,
		},
		updated: domain.Service{ID: "svc-1"},
	}
	catalog := NewServiceCatalog(repo, &fakeServerRepo{}, discardLogger())
	name := " API v2 "
	port := 8080

	_, err := catalog.Update(context.Background(), "svc-1", domain.ServiceUpdate{Name: &name, Port: &port})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lastUpdate.Name == nil || *repo.lastUpdate.Name != "API v2" {
		t.Errorf("expected normalized name, got %#v", repo.lastUpdate.Name)
	}
	if repo.lastUpdate.Port == nil || *repo.lastUpdate.Port != 8080 {
		t.Errorf("expected port update, got %#v", repo.lastUpdate.Port)
	}
}

func TestUpdateService_InvalidPort(t *testing.T) {
	repo := &fakeServiceRepo{getResult: domain.Service{
		ID: "svc-1", Name: "API", Type: "web", Port: 80, CheckType: "tcp", CheckTarget: "api.example.com:80",
		ExpectedStatus: 200, IntervalSeconds: 30, LatencyThresholdMs: 1000,
	}}
	catalog := NewServiceCatalog(repo, &fakeServerRepo{}, discardLogger())
	port := 70000

	_, err := catalog.Update(context.Background(), "svc-1", domain.ServiceUpdate{Port: &port})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}
