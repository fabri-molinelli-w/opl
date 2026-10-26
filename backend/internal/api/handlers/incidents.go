package handlers

import (
	"context"

	gen "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

// MonitoringService is the CU-03 use cases consumed by the handlers.
type MonitoringService interface {
	RunCheck(ctx context.Context, serviceID string) (domain.HealthCheck, error)
	ListChecks(ctx context.Context, serviceID string, limit int) ([]domain.HealthCheck, error)
}

// IncidentService is the CU-10 use cases consumed by the handlers.
type IncidentService interface {
	List(ctx context.Context, f domain.IncidentFilter) (domain.IncidentPage, error)
	Get(ctx context.Context, id string) (domain.Incident, error)
	Create(ctx context.Context, in domain.NewIncident) (domain.Incident, error)
	Update(ctx context.Context, id string, u domain.IncidentUpdate) (domain.Incident, error)
	Transition(ctx context.Context, id, to string, rootCause *string, actorID string) (domain.Incident, error)
	AddNote(ctx context.Context, id, authorID, body string) (domain.Incident, error)
}

func (s *Server) ListServiceChecks(ctx context.Context, request gen.ListServiceChecksRequestObject) (gen.ListServiceChecksResponseObject, error) {
	checks, err := s.monitoring.ListChecks(ctx, request.Id, intValue(request.Params.Limit))
	if err != nil {
		return nil, err
	}
	items := make([]gen.HealthCheck, 0, len(checks))
	for _, c := range checks {
		items = append(items, toGenHealthCheck(c))
	}
	return gen.ListServiceChecks200JSONResponse(items), nil
}

func (s *Server) RunServiceCheck(ctx context.Context, request gen.RunServiceCheckRequestObject) (gen.RunServiceCheckResponseObject, error) {
	check, err := s.monitoring.RunCheck(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return gen.RunServiceCheck200JSONResponse(toGenHealthCheck(check)), nil
}

func (s *Server) ListIncidents(ctx context.Context, request gen.ListIncidentsRequestObject) (gen.ListIncidentsResponseObject, error) {
	p := request.Params
	page, err := s.incidents.List(ctx, domain.IncidentFilter{
		Status:     stringValue((*string)(p.Status)),
		Severity:   stringValue((*string)(p.Severity)),
		ServiceID:  stringValue(p.ServiceId),
		OnlyActive: p.Active != nil && *p.Active,
		Page:       intValue((*int)(p.Page)),
		PageSize:   intValue((*int)(p.PageSize)),
	})
	if err != nil {
		return nil, err
	}
	items := make([]gen.IncidentSummary, 0, len(page.Items))
	for _, i := range page.Items {
		items = append(items, gen.IncidentSummary{
			Id:           i.ID,
			Title:        i.Title,
			ServiceId:    optionalString(i.ServiceID),
			ServiceName:  i.ServiceName,
			Severity:     gen.Severity(i.Severity),
			Status:       gen.IncidentStatus(i.Status),
			AssigneeName: i.AssigneeName,
			AutoCreated:  i.AutoCreated,
			OpenedAt:     i.OpenedAt,
			ResolvedAt:   i.ResolvedAt,
		})
	}
	return gen.ListIncidents200JSONResponse(gen.PaginatedIncidents{
		Items: items, Total: page.Total, Page: page.Page, PageSize: page.PageSize,
	}), nil
}

func (s *Server) CreateIncident(ctx context.Context, request gen.CreateIncidentRequestObject) (gen.CreateIncidentResponseObject, error) {
	if request.Body == nil {
		return nil, domain.NewError(domain.ErrValidation, "el cuerpo de la petición es obligatorio")
	}
	actorID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	incident, err := s.incidents.Create(ctx, domain.NewIncident{
		Title:       request.Body.Title,
		Description: stringValue(request.Body.Description),
		ServiceID:   stringValue(request.Body.ServiceId),
		Severity:    string(request.Body.Severity),
		AssigneeID:  stringValue(request.Body.AssigneeId),
		AuthorID:    actorID,
	})
	if err != nil {
		return nil, err
	}
	return gen.CreateIncident201JSONResponse(toGenIncident(incident)), nil
}

func (s *Server) GetIncident(ctx context.Context, request gen.GetIncidentRequestObject) (gen.GetIncidentResponseObject, error) {
	incident, err := s.incidents.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return gen.GetIncident200JSONResponse(toGenIncident(incident)), nil
}

func (s *Server) UpdateIncident(ctx context.Context, request gen.UpdateIncidentRequestObject) (gen.UpdateIncidentResponseObject, error) {
	if request.Body == nil {
		return nil, domain.NewError(domain.ErrValidation, "el cuerpo de la petición es obligatorio")
	}
	update := domain.IncidentUpdate{
		Title:       request.Body.Title,
		Description: request.Body.Description,
		AssigneeID:  request.Body.AssigneeId,
	}
	if request.Body.Severity != nil {
		severity := string(*request.Body.Severity)
		update.Severity = &severity
	}
	incident, err := s.incidents.Update(ctx, request.Id, update)
	if err != nil {
		return nil, err
	}
	return gen.UpdateIncident200JSONResponse(toGenIncident(incident)), nil
}

func (s *Server) TransitionIncident(ctx context.Context, request gen.TransitionIncidentRequestObject) (gen.TransitionIncidentResponseObject, error) {
	if request.Body == nil {
		return nil, domain.NewError(domain.ErrValidation, "el cuerpo de la petición es obligatorio")
	}
	actorID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	incident, err := s.incidents.Transition(ctx, request.Id, string(request.Body.To), request.Body.RootCause, actorID)
	if err != nil {
		return nil, err
	}
	return gen.TransitionIncident200JSONResponse(toGenIncident(incident)), nil
}

func (s *Server) AddIncidentNote(ctx context.Context, request gen.AddIncidentNoteRequestObject) (gen.AddIncidentNoteResponseObject, error) {
	if request.Body == nil {
		return nil, domain.NewError(domain.ErrValidation, "el cuerpo de la petición es obligatorio")
	}
	actorID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	incident, err := s.incidents.AddNote(ctx, request.Id, actorID, request.Body.Body)
	if err != nil {
		return nil, err
	}
	return gen.AddIncidentNote201JSONResponse(toGenIncident(incident)), nil
}

func toGenHealthCheck(c domain.HealthCheck) gen.HealthCheck {
	return gen.HealthCheck{
		Id:         c.ID,
		ServiceId:  c.ServiceID,
		Status:     gen.HealthCheckStatus(c.Status),
		ResponseMs: c.ResponseMs,
		HttpStatus: c.HTTPStatus,
		Error:      optionalString(c.Error),
		CheckedAt:  c.CheckedAt,
	}
}

func toGenIncident(i domain.Incident) gen.Incident {
	events := make([]gen.IncidentEvent, 0, len(i.Events))
	for _, e := range i.Events {
		events = append(events, gen.IncidentEvent{
			Id: e.ID, Kind: gen.IncidentEventKind(e.Kind), Body: e.Body, AuthorName: e.AuthorName, CreatedAt: e.CreatedAt,
		})
	}
	allowed := make([]gen.IncidentStatus, 0, 2)
	for _, to := range domain.AllowedTransitions(i.Status) {
		allowed = append(allowed, gen.IncidentStatus(to))
	}
	return gen.Incident{
		Id:                 i.ID,
		Title:              i.Title,
		Description:        i.Description,
		ServiceId:          optionalString(i.ServiceID),
		ServiceName:        i.ServiceName,
		Severity:           gen.Severity(i.Severity),
		Status:             gen.IncidentStatus(i.Status),
		AssigneeId:         optionalString(i.AssigneeID),
		AssigneeName:       i.AssigneeName,
		RootCause:          i.RootCause,
		AutoCreated:        i.AutoCreated,
		OpenedAt:           i.OpenedAt,
		ResolvedAt:         i.ResolvedAt,
		ClosedAt:           i.ClosedAt,
		ResolutionMinutes:  i.ResolutionMinutes(),
		CreatedAt:          i.CreatedAt,
		UpdatedAt:          i.UpdatedAt,
		Events:             events,
		AllowedTransitions: allowed,
	}
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
