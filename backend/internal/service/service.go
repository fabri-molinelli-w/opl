package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

// ServiceRepository is the subset of the service repository used by
// ServiceCatalog.
type ServiceRepository interface {
	CreateService(ctx context.Context, service domain.Service) (domain.Service, error)
	GetServiceByID(ctx context.Context, id string) (domain.Service, error)
	ListServicesByServerID(ctx context.Context, serverID string) ([]domain.Service, error)
	ListServices(ctx context.Context, filter domain.ServiceFilter) ([]domain.Service, error)
	CountServices(ctx context.Context, filter domain.ServiceFilter) (int64, error)
	UpdateService(ctx context.Context, id string, update domain.ServiceUpdate) (domain.Service, error)
	DeleteService(ctx context.Context, id string) error
}

// CreateServiceInput carries the data needed to create a service.
type CreateServiceInput struct {
	ServerID           string
	Name               string
	Type               string
	Port               int
	CheckType          string
	CheckTarget        string
	ExpectedStatus     *int
	IntervalSeconds    *int
	LatencyThresholdMs *int
}

// ServerLookup loads a server before creating a service.
type ServerLookup interface {
	GetServerByID(ctx context.Context, id string) (domain.Server, error)
}

type ServiceCatalog struct {
	repo    ServiceRepository
	servers ServerLookup
	log     *slog.Logger
}

func NewServiceCatalog(repo ServiceRepository, servers ServerLookup, log *slog.Logger) *ServiceCatalog {
	return &ServiceCatalog{repo: repo, servers: servers, log: log}
}

func (s *ServiceCatalog) List(ctx context.Context, filter domain.ServiceFilter) (domain.ServicePage, error) {
	page, pageSize, err := normalizePagination(filter.Page, filter.PageSize)
	if err != nil {
		return domain.ServicePage{}, err
	}
	filter.Page = page
	filter.PageSize = pageSize
	filter.Environment = strings.TrimSpace(filter.Environment)
	filter.Status = strings.TrimSpace(filter.Status)
	filter.Type = strings.TrimSpace(filter.Type)
	filter.ServerID = strings.TrimSpace(filter.ServerID)
	if filter.Environment != "" && !domain.ValidEnvironment(filter.Environment) {
		return domain.ServicePage{}, domain.NewFieldError(
			domain.ErrValidation, "El environment no es válido", "environment", "valor no permitido")
	}
	if filter.Status != "" && !domain.ValidServiceStatus(filter.Status) {
		return domain.ServicePage{}, domain.NewFieldError(
			domain.ErrValidation, "El status no es válido", "status", "valor no permitido")
	}
	if filter.Type != "" && !domain.ValidServiceType(filter.Type) {
		return domain.ServicePage{}, domain.NewFieldError(
			domain.ErrValidation, "El type no es válido", "type", "valor no permitido")
	}

	total, err := s.repo.CountServices(ctx, filter)
	if err != nil {
		return domain.ServicePage{}, fmt.Errorf("count services: %w", err)
	}
	items, err := s.repo.ListServices(ctx, filter)
	if err != nil {
		return domain.ServicePage{}, fmt.Errorf("list services: %w", err)
	}
	return domain.ServicePage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *ServiceCatalog) Get(ctx context.Context, id string) (domain.Service, error) {
	service, err := s.repo.GetServiceByID(ctx, id)
	if err != nil {
		return domain.Service{}, fmt.Errorf("get service: %w", err)
	}
	return service, nil
}

func (s *ServiceCatalog) Create(ctx context.Context, in CreateServiceInput) (domain.Service, error) {
	serverID := strings.TrimSpace(in.ServerID)
	if serverID == "" {
		return domain.Service{}, domain.NewFieldError(
			domain.ErrValidation, "El servidor es obligatorio", "serverId", "no puede estar vacío")
	}
	if _, err := s.servers.GetServerByID(ctx, serverID); err != nil {
		return domain.Service{}, err
	}

	expectedStatus := 200
	if in.ExpectedStatus != nil {
		expectedStatus = *in.ExpectedStatus
	}
	intervalSeconds := 30
	if in.IntervalSeconds != nil {
		intervalSeconds = *in.IntervalSeconds
	}
	latencyThresholdMs := 1000
	if in.LatencyThresholdMs != nil {
		latencyThresholdMs = *in.LatencyThresholdMs
	}

	service, err := normalizeService(domain.Service{
		ServerID:           serverID,
		Name:               in.Name,
		Type:               in.Type,
		Port:               in.Port,
		CheckType:          in.CheckType,
		CheckTarget:        in.CheckTarget,
		ExpectedStatus:     expectedStatus,
		IntervalSeconds:    intervalSeconds,
		LatencyThresholdMs: latencyThresholdMs,
	})
	if err != nil {
		return domain.Service{}, err
	}

	created, err := s.repo.CreateService(ctx, service)
	if err != nil {
		return domain.Service{}, fmt.Errorf("create service: %w", err)
	}
	return created, nil
}

func (s *ServiceCatalog) Update(ctx context.Context, id string, update domain.ServiceUpdate) (domain.Service, error) {
	if update.Name == nil && update.Type == nil && update.Port == nil && update.CheckType == nil &&
		update.CheckTarget == nil && update.ExpectedStatus == nil && update.IntervalSeconds == nil &&
		update.LatencyThresholdMs == nil {
		return domain.Service{}, domain.NewError(domain.ErrValidation, "no hay campos para actualizar")
	}

	current, err := s.repo.GetServiceByID(ctx, id)
	if err != nil {
		return domain.Service{}, fmt.Errorf("get service: %w", err)
	}
	next := current
	if update.Name != nil {
		next.Name = *update.Name
	}
	if update.Type != nil {
		next.Type = *update.Type
	}
	if update.Port != nil {
		next.Port = *update.Port
	}
	if update.CheckType != nil {
		next.CheckType = *update.CheckType
	}
	if update.CheckTarget != nil {
		next.CheckTarget = *update.CheckTarget
	}
	if update.ExpectedStatus != nil {
		next.ExpectedStatus = *update.ExpectedStatus
	}
	if update.IntervalSeconds != nil {
		next.IntervalSeconds = *update.IntervalSeconds
	}
	if update.LatencyThresholdMs != nil {
		next.LatencyThresholdMs = *update.LatencyThresholdMs
	}

	normalized, err := normalizeService(next)
	if err != nil {
		return domain.Service{}, err
	}
	if update.Name != nil {
		update.Name = &normalized.Name
	}
	if update.Type != nil {
		update.Type = &normalized.Type
	}
	if update.Port != nil {
		update.Port = &normalized.Port
	}
	if update.CheckType != nil {
		update.CheckType = &normalized.CheckType
	}
	if update.CheckTarget != nil {
		update.CheckTarget = &normalized.CheckTarget
	}
	if update.ExpectedStatus != nil {
		update.ExpectedStatus = &normalized.ExpectedStatus
	}
	if update.IntervalSeconds != nil {
		update.IntervalSeconds = &normalized.IntervalSeconds
	}
	if update.LatencyThresholdMs != nil {
		update.LatencyThresholdMs = &normalized.LatencyThresholdMs
	}

	updated, err := s.repo.UpdateService(ctx, id, update)
	if err != nil {
		return domain.Service{}, fmt.Errorf("update service: %w", err)
	}
	return updated, nil
}

func (s *ServiceCatalog) Delete(ctx context.Context, id string) error {
	if err := s.repo.DeleteService(ctx, id); err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	return nil
}

func normalizeService(service domain.Service) (domain.Service, error) {
	name := strings.TrimSpace(service.Name)
	if name == "" {
		return domain.Service{}, domain.NewFieldError(
			domain.ErrValidation, "El nombre es obligatorio", "name", "no puede estar vacío")
	}
	serviceType := strings.TrimSpace(service.Type)
	if !domain.ValidServiceType(serviceType) {
		return domain.Service{}, domain.NewFieldError(
			domain.ErrValidation, "El tipo de servicio no es válido", "type", "valor no permitido")
	}
	if service.Port < 1 || service.Port > 65535 {
		return domain.Service{}, domain.NewFieldError(
			domain.ErrValidation, "El puerto debe estar entre 1 y 65535", "port", "valor fuera de rango")
	}
	checkType := strings.TrimSpace(service.CheckType)
	if !domain.ValidCheckType(checkType) {
		return domain.Service{}, domain.NewFieldError(
			domain.ErrValidation, "El tipo de check no es válido", "checkType", "valor no permitido")
	}
	checkTarget := strings.TrimSpace(service.CheckTarget)
	if checkTarget == "" {
		return domain.Service{}, domain.NewFieldError(
			domain.ErrValidation, "El checkTarget es obligatorio", "checkTarget", "no puede estar vacío")
	}
	if checkType == domain.CheckTypeHTTP {
		if err := validateHTTPTarget(checkTarget); err != nil {
			return domain.Service{}, err
		}
	}
	if service.ExpectedStatus < 100 || service.ExpectedStatus > 599 {
		return domain.Service{}, domain.NewFieldError(
			domain.ErrValidation, "El expectedStatus debe estar entre 100 y 599", "expectedStatus", "valor fuera de rango")
	}
	if service.IntervalSeconds < 1 {
		return domain.Service{}, domain.NewFieldError(
			domain.ErrValidation, "El intervalSeconds debe ser mayor o igual a 1", "intervalSeconds", "valor fuera de rango")
	}
	if service.LatencyThresholdMs < 1 {
		return domain.Service{}, domain.NewFieldError(
			domain.ErrValidation, "El latencyThresholdMs debe ser mayor o igual a 1", "latencyThresholdMs", "valor fuera de rango")
	}

	return domain.Service{
		ServerID:           service.ServerID,
		Name:               name,
		Type:               serviceType,
		Port:               service.Port,
		CheckType:          checkType,
		CheckTarget:        checkTarget,
		ExpectedStatus:     service.ExpectedStatus,
		IntervalSeconds:    service.IntervalSeconds,
		LatencyThresholdMs: service.LatencyThresholdMs,
	}, nil
}

func validateHTTPTarget(target string) error {
	parsed, err := url.Parse(target)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return domain.NewFieldError(
			domain.ErrValidation, "El checkTarget debe ser una URL http o https válida", "checkTarget", "URL inválida")
	}
	return nil
}
