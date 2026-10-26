package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gen "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/auth"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/go-chi/chi/v5"
)

func newInventoryRouter(t *testing.T, role domain.Role, servers ServerService, services ServiceCatalog) (http.Handler, string) {
	t.Helper()
	issuer := auth.NewTokenIssuer(testSecret, time.Hour)
	user := domain.User{ID: "user-1", Email: "user@example.com", Role: role, Active: true}
	token, err := issuer.Issue(user)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	server := NewServer(Deps{
		DB:       fakePinger{},
		Servers:  servers,
		Services: services,
		Log:      discardTestLogger(),
	})
	strict := gen.NewStrictHandlerWithOptions(
		server,
		[]gen.StrictMiddlewareFunc{auth.Authorize(auth.NewPolicy(), issuer, fakeLookup{users: map[string]domain.User{user.ID: user}})},
		NewStrictOptions(discardTestLogger()),
	)
	r := chi.NewRouter()
	gen.HandlerFromMux(strict, r)
	return r, token
}

func discardTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestListServersHandler_CamelCasePagination(t *testing.T) {
	servers := fakeServerService{page: domain.ServerPage{
		Items: []domain.Server{{
			ID: "srv-1", Hostname: "web-01", IP: "192.0.2.10", Environment: "prod",
		}},
		Total: 1, Page: 1, PageSize: 20,
	}}
	router, token := newInventoryRouter(t, domain.RoleSoporte, servers, fakeServiceCatalog{})

	req := httptest.NewRequest(http.MethodGet, "/api/servers?page=1&pageSize=20", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := body["pageSize"]; !ok {
		t.Errorf("expected camelCase pageSize, got %s", w.Body.String())
	}
	if _, ok := body["page_size"]; ok {
		t.Errorf("unexpected snake_case page_size: %s", w.Body.String())
	}
}

func TestCreateServerHandler_ValidationDetails(t *testing.T) {
	servers := fakeServerService{createErr: domain.NewFieldError(
		domain.ErrValidation, "La IP no es una dirección válida", "ip", "IPv4 o IPv6 inválida")}
	router, token := newInventoryRouter(t, domain.RoleSysadmin, servers, fakeServiceCatalog{})

	body := strings.NewReader(`{"hostname":"web-01","ip":"bad","environment":"prod"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/servers", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (%s)", w.Code, w.Body.String())
	}
	var resp gen.Error
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Details == nil || len(*resp.Details) != 1 {
		t.Fatalf("expected one detail, got %#v", resp.Details)
	}
	if (*resp.Details)[0].Field != "ip" {
		t.Errorf("expected ip field detail, got %#v", (*resp.Details)[0])
	}
}

func TestCreateServerHandler_DuplicateHostname(t *testing.T) {
	servers := fakeServerService{createErr: domain.NewFieldError(
		domain.ErrConflict, "Ya existe un servidor con ese hostname", "hostname", "hostname duplicado")}
	router, token := newInventoryRouter(t, domain.RoleSysadmin, servers, fakeServiceCatalog{})

	body := strings.NewReader(`{"hostname":"web-01","ip":"192.0.2.10","environment":"prod"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/servers", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d (%s)", w.Code, w.Body.String())
	}
	var resp gen.Error
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Code != "conflict" {
		t.Errorf("expected conflict code, got %q", resp.Code)
	}
	if resp.Details == nil || len(*resp.Details) != 1 || (*resp.Details)[0].Field != "hostname" {
		t.Errorf("expected hostname detail, got %#v", resp.Details)
	}
}

func TestGetServerHandler_IncludesServices(t *testing.T) {
	servers := fakeServerService{server: domain.Server{
		ID: "srv-1", Hostname: "web-01",
		Services: []domain.Service{{ID: "svc-1", Name: "API", CurrentStatus: "UP"}},
	}}
	router, token := newInventoryRouter(t, domain.RoleSoporte, servers, fakeServiceCatalog{})

	req := httptest.NewRequest(http.MethodGet, "/api/servers/srv-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
	var resp gen.Server
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode server: %v", err)
	}
	if resp.Services == nil || len(*resp.Services) != 1 {
		t.Fatalf("expected services in detail, got %#v", resp.Services)
	}
}

func TestCreateServiceHandler_PassesDefaultsToService(t *testing.T) {
	services := fakeServiceCatalog{service: domain.Service{ID: "svc-1", Name: "API"}}
	router, token := newInventoryRouter(t, domain.RoleSysadmin, fakeServerService{}, services)

	body := strings.NewReader(`{"serverId":"srv-1","name":"API","type":"web","port":8080,"checkType":"http","checkTarget":"https://api.example.com/health"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/services", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", w.Code, w.Body.String())
	}
}

var _ ServerService = fakeServerService{}
var _ ServiceCatalog = fakeServiceCatalog{}
