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

// ServerService is the server inventory use cases consumed by the handlers.
type ServerService interface {
	List(ctx context.Context, filter domain.ServerFilter) (domain.ServerPage, error)
	Get(ctx context.Context, id string) (domain.Server, error)
	Create(ctx context.Context, in service.CreateServerInput) (domain.Server, error)
	Update(ctx context.Context, id string, update domain.ServerUpdate) (domain.Server, error)
	Delete(ctx context.Context, id string) error
}

// ServiceCatalog is the service inventory use cases consumed by the handlers.
type ServiceCatalog interface {
	List(ctx context.Context, filter domain.ServiceFilter) (domain.ServicePage, error)
	Get(ctx context.Context, id string) (domain.Service, error)
	Create(ctx context.Context, in service.CreateServiceInput) (domain.Service, error)
	Update(ctx context.Context, id string, update domain.ServiceUpdate) (domain.Service, error)
	Delete(ctx context.Context, id string) error
}

// Deps groups every dependency of the HTTP handlers. Adding a new use case only
// adds a field here instead of changing NewServer's signature.
type Deps struct {
	DB         Pinger
	Auth       AuthService
	Users      UserService
	Servers    ServerService
	Services   ServiceCatalog
	Documents  DocumentService
	Chat       ChatService
	Monitoring MonitoringService
	Incidents  IncidentService
	Log        *slog.Logger
}

// Server implements api.StrictServerInterface. It only validates input,
// converts generated types to domain types and delegates to the services.
type Server struct {
	db         Pinger
	auth       AuthService
	users      UserService
	servers    ServerService
	services   ServiceCatalog
	documents  DocumentService
	chat       ChatService
	monitoring MonitoringService
	incidents  IncidentService
	log        *slog.Logger
}

func NewServer(deps Deps) *Server {
	return &Server{
		db:         deps.DB,
		auth:       deps.Auth,
		users:      deps.Users,
		servers:    deps.Servers,
		services:   deps.Services,
		documents:  deps.Documents,
		chat:       deps.Chat,
		monitoring: deps.Monitoring,
		incidents:  deps.Incidents,
		log:        deps.Log,
	}
}

// toGenUser converts a domain user to the API schema. The password hash is
// never included.
func toGenServer(server domain.Server, withServices bool) gen.Server {
	result := gen.Server{
		Id:          server.ID,
		Hostname:    server.Hostname,
		Ip:          server.IP,
		Environment: gen.Environment(server.Environment),
		Os:          server.OS,
		Location:    gen.Location(server.Location),
		Owner:       server.Owner,
		Tags:        server.Tags,
		CreatedAt:   server.CreatedAt,
		UpdatedAt:   server.UpdatedAt,
	}
	if withServices {
		services := make([]gen.Service, 0, len(server.Services))
		for _, item := range server.Services {
			services = append(services, toGenService(item))
		}
		result.Services = &services
	}
	return result
}

func toGenService(service domain.Service) gen.Service {
	return gen.Service{
		Id:                 service.ID,
		ServerId:           service.ServerID,
		Name:               service.Name,
		Type:               gen.ServiceType(service.Type),
		Port:               service.Port,
		CheckType:          gen.CheckType(service.CheckType),
		CheckTarget:        service.CheckTarget,
		ExpectedStatus:     service.ExpectedStatus,
		IntervalSeconds:    service.IntervalSeconds,
		LatencyThresholdMs: service.LatencyThresholdMs,
		CurrentStatus:      gen.ServiceStatus(service.CurrentStatus),
		CreatedAt:          service.CreatedAt,
		UpdatedAt:          service.UpdatedAt,
	}
}

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
