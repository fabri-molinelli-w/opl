package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"strings"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

// ServerRepository is the subset of the server repository used by
// ServerService.
type ServerRepository interface {
	CreateServer(ctx context.Context, server domain.Server) (domain.Server, error)
	GetServerByID(ctx context.Context, id string) (domain.Server, error)
	ListServers(ctx context.Context, filter domain.ServerFilter) ([]domain.Server, error)
	CountServers(ctx context.Context, filter domain.ServerFilter) (int64, error)
	UpdateServer(ctx context.Context, id string, update domain.ServerUpdate) (domain.Server, error)
	DeleteServer(ctx context.Context, id string) error
}

// InventoryServiceRepository is the subset of the service repository used by
// ServerService to build the server detail.
type InventoryServiceRepository interface {
	ListServicesByServerID(ctx context.Context, serverID string) ([]domain.Service, error)
}

// CreateServerInput carries the data needed to create a server.
type CreateServerInput struct {
	Hostname    string
	IP          string
	Environment string
	OS          string
	Location    string
	Owner       string
	Tags        []string
}

type ServerService struct {
	repo     ServerRepository
	services InventoryServiceRepository
	log      *slog.Logger
}

func NewServerService(repo ServerRepository, services InventoryServiceRepository, log *slog.Logger) *ServerService {
	return &ServerService{repo: repo, services: services, log: log}
}

func (s *ServerService) List(ctx context.Context, filter domain.ServerFilter) (domain.ServerPage, error) {
	page, pageSize, err := normalizePagination(filter.Page, filter.PageSize)
	if err != nil {
		return domain.ServerPage{}, err
	}
	filter.Page = page
	filter.PageSize = pageSize
	filter.Q = strings.TrimSpace(filter.Q)
	filter.OS = strings.TrimSpace(filter.OS)
	filter.Tag = strings.TrimSpace(filter.Tag)
	filter.Environment = strings.TrimSpace(filter.Environment)
	if filter.Environment != "" && !domain.ValidEnvironment(filter.Environment) {
		return domain.ServerPage{}, domain.NewFieldError(
			domain.ErrValidation, "El environment no es válido", "environment", "valor no permitido")
	}

	total, err := s.repo.CountServers(ctx, filter)
	if err != nil {
		return domain.ServerPage{}, fmt.Errorf("count servers: %w", err)
	}
	items, err := s.repo.ListServers(ctx, filter)
	if err != nil {
		return domain.ServerPage{}, fmt.Errorf("list servers: %w", err)
	}
	return domain.ServerPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *ServerService) Get(ctx context.Context, id string) (domain.Server, error) {
	server, err := s.repo.GetServerByID(ctx, id)
	if err != nil {
		return domain.Server{}, fmt.Errorf("get server: %w", err)
	}
	services, err := s.services.ListServicesByServerID(ctx, server.ID)
	if err != nil {
		return domain.Server{}, fmt.Errorf("list server services: %w", err)
	}
	server.Services = services
	return server, nil
}

func (s *ServerService) Create(ctx context.Context, in CreateServerInput) (domain.Server, error) {
	server, err := normalizeServer(domain.Server{
		Hostname:    in.Hostname,
		IP:          in.IP,
		Environment: in.Environment,
		OS:          in.OS,
		Location:    in.Location,
		Owner:       in.Owner,
		Tags:        in.Tags,
	})
	if err != nil {
		return domain.Server{}, err
	}

	created, err := s.repo.CreateServer(ctx, server)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return domain.Server{}, domain.NewFieldError(
				domain.ErrConflict, "Ya existe un servidor con ese hostname", "hostname", "hostname duplicado")
		}
		return domain.Server{}, fmt.Errorf("create server: %w", err)
	}
	return created, nil
}

func (s *ServerService) Update(ctx context.Context, id string, update domain.ServerUpdate) (domain.Server, error) {
	if update.Hostname == nil && update.IP == nil && update.Environment == nil &&
		update.OS == nil && update.Location == nil && update.Owner == nil && update.Tags == nil {
		return domain.Server{}, domain.NewError(domain.ErrValidation, "no hay campos para actualizar")
	}

	normalized, err := normalizeServerUpdate(update)
	if err != nil {
		return domain.Server{}, err
	}

	updated, err := s.repo.UpdateServer(ctx, id, normalized)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return domain.Server{}, domain.NewFieldError(
				domain.ErrConflict, "Ya existe un servidor con ese hostname", "hostname", "hostname duplicado")
		}
		return domain.Server{}, fmt.Errorf("update server: %w", err)
	}
	return updated, nil
}

func (s *ServerService) Delete(ctx context.Context, id string) error {
	if err := s.repo.DeleteServer(ctx, id); err != nil {
		return fmt.Errorf("delete server: %w", err)
	}
	return nil
}

func normalizeServer(server domain.Server) (domain.Server, error) {
	hostname, err := normalizeHostname(server.Hostname)
	if err != nil {
		return domain.Server{}, err
	}
	ip, err := normalizeIP(server.IP)
	if err != nil {
		return domain.Server{}, err
	}
	environment := strings.TrimSpace(server.Environment)
	if !domain.ValidEnvironment(environment) {
		return domain.Server{}, domain.NewFieldError(
			domain.ErrValidation, "El environment no es válido", "environment", "valor no permitido")
	}
	location := strings.TrimSpace(server.Location)
	if location == "" {
		location = domain.LocationOnPrem
	}
	if !domain.ValidLocation(location) {
		return domain.Server{}, domain.NewFieldError(
			domain.ErrValidation, "La location no es válida", "location", "valor no permitido")
	}
	tags := server.Tags
	if tags == nil {
		tags = []string{}
	}
	return domain.Server{
		Hostname:    hostname,
		IP:          ip,
		Environment: environment,
		OS:          strings.TrimSpace(server.OS),
		Location:    location,
		Owner:       strings.TrimSpace(server.Owner),
		Tags:        tags,
	}, nil
}

func normalizeServerUpdate(update domain.ServerUpdate) (domain.ServerUpdate, error) {
	if update.Hostname != nil {
		hostname, err := normalizeHostname(*update.Hostname)
		if err != nil {
			return domain.ServerUpdate{}, err
		}
		update.Hostname = &hostname
	}
	if update.IP != nil {
		ip, err := normalizeIP(*update.IP)
		if err != nil {
			return domain.ServerUpdate{}, err
		}
		update.IP = &ip
	}
	if update.Environment != nil {
		environment := strings.TrimSpace(*update.Environment)
		if !domain.ValidEnvironment(environment) {
			return domain.ServerUpdate{}, domain.NewFieldError(
				domain.ErrValidation, "El environment no es válido", "environment", "valor no permitido")
		}
		update.Environment = &environment
	}
	if update.Location != nil {
		location := strings.TrimSpace(*update.Location)
		if !domain.ValidLocation(location) {
			return domain.ServerUpdate{}, domain.NewFieldError(
				domain.ErrValidation, "La location no es válida", "location", "valor no permitido")
		}
		update.Location = &location
	}
	if update.OS != nil {
		os := strings.TrimSpace(*update.OS)
		update.OS = &os
	}
	if update.Owner != nil {
		owner := strings.TrimSpace(*update.Owner)
		update.Owner = &owner
	}
	return update, nil
}

func normalizeHostname(hostname string) (string, error) {
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		return "", domain.NewFieldError(
			domain.ErrValidation, "El hostname es obligatorio", "hostname", "no puede estar vacío")
	}
	return hostname, nil
}

func normalizeIP(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	addr, err := netip.ParseAddr(raw)
	if err != nil {
		return "", domain.NewFieldError(
			domain.ErrValidation, "La IP no es una dirección válida", "ip", "IPv4 o IPv6 inválida")
	}
	return addr.Unmap().String(), nil
}

func normalizePagination(page, pageSize int) (int, int, error) {
	if page == 0 {
		page = 1
	}
	if page < 1 {
		return 0, 0, domain.NewFieldError(
			domain.ErrValidation, "La página debe ser mayor o igual a 1", "page", "valor fuera de rango")
	}
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	if pageSize < 1 || pageSize > maxPageSize {
		return 0, 0, domain.NewFieldError(
			domain.ErrValidation, "El tamaño de página debe estar entre 1 y 100", "pageSize", "valor fuera de rango")
	}
	return page, pageSize, nil
}
