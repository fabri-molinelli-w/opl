package handlers

import (
	"context"

	gen "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/fabri-molinelli-w/opl/backend/internal/service"
)

func (s *Server) ListServers(ctx context.Context, request gen.ListServersRequestObject) (gen.ListServersResponseObject, error) {
	page, err := s.servers.List(ctx, domain.ServerFilter{
		Q:           stringValue(request.Params.Q),
		Environment: stringValue((*string)(request.Params.Environment)),
		OS:          stringValue(request.Params.Os),
		Tag:         stringValue(request.Params.Tag),
		Page:        intValue((*int)(request.Params.Page)),
		PageSize:    intValue((*int)(request.Params.PageSize)),
	})
	if err != nil {
		return nil, err
	}

	items := make([]gen.Server, 0, len(page.Items))
	for _, server := range page.Items {
		items = append(items, toGenServer(server, false))
	}
	return gen.ListServers200JSONResponse(gen.PaginatedServers{
		Items:    items,
		Total:    page.Total,
		Page:     page.Page,
		PageSize: page.PageSize,
	}), nil
}

func (s *Server) CreateServer(ctx context.Context, request gen.CreateServerRequestObject) (gen.CreateServerResponseObject, error) {
	if request.Body == nil {
		return nil, domain.NewError(domain.ErrValidation, "el cuerpo de la petición es obligatorio")
	}

	server, err := s.servers.Create(ctx, service.CreateServerInput{
		Hostname:    request.Body.Hostname,
		IP:          request.Body.Ip,
		Environment: string(request.Body.Environment),
		OS:          stringValue(request.Body.Os),
		Location:    stringValue((*string)(request.Body.Location)),
		Owner:       stringValue(request.Body.Owner),
		Tags:        tagsValue(request.Body.Tags),
	})
	if err != nil {
		return nil, err
	}
	return gen.CreateServer201JSONResponse(toGenServer(server, false)), nil
}

func (s *Server) GetServer(ctx context.Context, request gen.GetServerRequestObject) (gen.GetServerResponseObject, error) {
	server, err := s.servers.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return gen.GetServer200JSONResponse(toGenServer(server, true)), nil
}

func (s *Server) UpdateServer(ctx context.Context, request gen.UpdateServerRequestObject) (gen.UpdateServerResponseObject, error) {
	if request.Body == nil {
		return nil, domain.NewError(domain.ErrValidation, "el cuerpo de la petición es obligatorio")
	}

	update := domain.ServerUpdate{
		Hostname: request.Body.Hostname,
		IP:       request.Body.Ip,
		OS:       request.Body.Os,
		Owner:    request.Body.Owner,
		Tags:     request.Body.Tags,
	}
	if request.Body.Environment != nil {
		environment := string(*request.Body.Environment)
		update.Environment = &environment
	}
	if request.Body.Location != nil {
		location := string(*request.Body.Location)
		update.Location = &location
	}

	server, err := s.servers.Update(ctx, request.Id, update)
	if err != nil {
		return nil, err
	}
	return gen.UpdateServer200JSONResponse(toGenServer(server, false)), nil
}

func (s *Server) DeleteServer(ctx context.Context, request gen.DeleteServerRequestObject) (gen.DeleteServerResponseObject, error) {
	if err := s.servers.Delete(ctx, request.Id); err != nil {
		return nil, err
	}
	return gen.DeleteServer204Response{}, nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func tagsValue(tags *[]string) []string {
	if tags == nil {
		return nil
	}
	return *tags
}
