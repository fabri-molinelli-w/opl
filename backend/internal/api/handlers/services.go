package handlers

import (
	"context"

	gen "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/fabri-molinelli-w/opl/backend/internal/service"
)

func (s *Server) ListServices(ctx context.Context, request gen.ListServicesRequestObject) (gen.ListServicesResponseObject, error) {
	page, err := s.services.List(ctx, domain.ServiceFilter{
		ServerID:    stringValue(request.Params.ServerId),
		Environment: stringValue((*string)(request.Params.Environment)),
		Status:      stringValue((*string)(request.Params.Status)),
		Type:        stringValue((*string)(request.Params.Type)),
		Page:        intValue((*int)(request.Params.Page)),
		PageSize:    intValue((*int)(request.Params.PageSize)),
	})
	if err != nil {
		return nil, err
	}

	items := make([]gen.Service, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, toGenService(item))
	}
	return gen.ListServices200JSONResponse(gen.PaginatedServices{
		Items:    items,
		Total:    page.Total,
		Page:     page.Page,
		PageSize: page.PageSize,
	}), nil
}

func (s *Server) CreateService(ctx context.Context, request gen.CreateServiceRequestObject) (gen.CreateServiceResponseObject, error) {
	if request.Body == nil {
		return nil, domain.NewError(domain.ErrValidation, "el cuerpo de la petición es obligatorio")
	}

	item, err := s.services.Create(ctx, service.CreateServiceInput{
		ServerID:           request.Body.ServerId,
		Name:               request.Body.Name,
		Type:               string(request.Body.Type),
		Port:               request.Body.Port,
		CheckType:          string(request.Body.CheckType),
		CheckTarget:        request.Body.CheckTarget,
		ExpectedStatus:     request.Body.ExpectedStatus,
		IntervalSeconds:    request.Body.IntervalSeconds,
		LatencyThresholdMs: request.Body.LatencyThresholdMs,
	})
	if err != nil {
		return nil, err
	}
	return gen.CreateService201JSONResponse(toGenService(item)), nil
}

func (s *Server) GetService(ctx context.Context, request gen.GetServiceRequestObject) (gen.GetServiceResponseObject, error) {
	item, err := s.services.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return gen.GetService200JSONResponse(toGenService(item)), nil
}

func (s *Server) UpdateService(ctx context.Context, request gen.UpdateServiceRequestObject) (gen.UpdateServiceResponseObject, error) {
	if request.Body == nil {
		return nil, domain.NewError(domain.ErrValidation, "el cuerpo de la petición es obligatorio")
	}

	update := domain.ServiceUpdate{
		Name:               request.Body.Name,
		CheckTarget:        request.Body.CheckTarget,
		ExpectedStatus:     request.Body.ExpectedStatus,
		IntervalSeconds:    request.Body.IntervalSeconds,
		LatencyThresholdMs: request.Body.LatencyThresholdMs,
	}
	if request.Body.Type != nil {
		serviceType := string(*request.Body.Type)
		update.Type = &serviceType
	}
	if request.Body.Port != nil {
		port := *request.Body.Port
		update.Port = &port
	}
	if request.Body.CheckType != nil {
		checkType := string(*request.Body.CheckType)
		update.CheckType = &checkType
	}

	item, err := s.services.Update(ctx, request.Id, update)
	if err != nil {
		return nil, err
	}
	return gen.UpdateService200JSONResponse(toGenService(item)), nil
}

func (s *Server) DeleteService(ctx context.Context, request gen.DeleteServiceRequestObject) (gen.DeleteServiceResponseObject, error) {
	if err := s.services.Delete(ctx, request.Id); err != nil {
		return nil, err
	}
	return gen.DeleteService204Response{}, nil
}
