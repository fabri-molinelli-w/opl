package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

// IncidentRepository is the persistence used by IncidentService.
type IncidentRepository interface {
	CreateIncident(ctx context.Context, in domain.NewIncident) (string, error)
	GetIncident(ctx context.Context, id string) (domain.Incident, error)
	ListIncidents(ctx context.Context, f domain.IncidentFilter) ([]domain.Incident, error)
	CountIncidents(ctx context.Context, f domain.IncidentFilter) (int64, error)
	UpdateIncident(ctx context.Context, id string, u domain.IncidentUpdate) error
	TransitionIncident(ctx context.Context, id, to string, rootCause *string, actorID, body string, validate func(from string) error) error
	AddEvent(ctx context.Context, id, authorID, kind, body string) error
}

// IncidentObserver is notified when an incident is opened or assigned (CU-08).
type IncidentObserver interface {
	OnIncidentOpened(ctx context.Context, incident domain.Incident)
	OnIncidentAssigned(ctx context.Context, incident domain.Incident)
}

var statusLabels = map[string]string{
	domain.IncidentOpen:       "Abierto",
	domain.IncidentInProgress: "En curso",
	domain.IncidentResolved:   "Resuelto",
	domain.IncidentClosed:     "Cerrado",
}

type IncidentService struct {
	repo      IncidentRepository
	observers []IncidentObserver
	log       *slog.Logger
}

func NewIncidentService(repo IncidentRepository, log *slog.Logger) *IncidentService {
	return &IncidentService{repo: repo, log: log}
}

func (s *IncidentService) AddObserver(o IncidentObserver) {
	s.observers = append(s.observers, o)
}

func (s *IncidentService) List(ctx context.Context, f domain.IncidentFilter) (domain.IncidentPage, error) {
	page, pageSize, err := normalizePagination(f.Page, f.PageSize)
	if err != nil {
		return domain.IncidentPage{}, err
	}
	f.Page, f.PageSize = page, pageSize
	if f.Status != "" && !domain.ValidIncidentStatus(f.Status) {
		return domain.IncidentPage{}, domain.NewFieldError(domain.ErrValidation, "El estado no es válido", "status", "valor no permitido")
	}
	if f.Severity != "" && !domain.ValidSeverity(f.Severity) {
		return domain.IncidentPage{}, domain.NewFieldError(domain.ErrValidation, "La severidad no es válida", "severity", "valor no permitido")
	}
	total, err := s.repo.CountIncidents(ctx, f)
	if err != nil {
		return domain.IncidentPage{}, fmt.Errorf("count incidents: %w", err)
	}
	items, err := s.repo.ListIncidents(ctx, f)
	if err != nil {
		return domain.IncidentPage{}, fmt.Errorf("list incidents: %w", err)
	}
	return domain.IncidentPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *IncidentService) Get(ctx context.Context, id string) (domain.Incident, error) {
	incident, err := s.repo.GetIncident(ctx, id)
	if err != nil {
		return domain.Incident{}, fmt.Errorf("get incident: %w", err)
	}
	return incident, nil
}

func (s *IncidentService) Create(ctx context.Context, in domain.NewIncident) (domain.Incident, error) {
	title, err := normalizeIncidentTitle(in.Title)
	if err != nil {
		return domain.Incident{}, err
	}
	in.Title = title
	in.Description = strings.TrimSpace(in.Description)
	if !domain.ValidSeverity(in.Severity) {
		return domain.Incident{}, domain.NewFieldError(domain.ErrValidation, "La severidad no es válida", "severity", "SEV1 a SEV4")
	}
	id, err := s.repo.CreateIncident(ctx, in)
	if err != nil {
		return domain.Incident{}, fmt.Errorf("create incident: %w", err)
	}
	incident, err := s.Get(ctx, id)
	if err != nil {
		return domain.Incident{}, err
	}
	for _, o := range s.observers {
		o.OnIncidentOpened(ctx, incident)
		if incident.AssigneeID != "" {
			o.OnIncidentAssigned(ctx, incident)
		}
	}
	return incident, nil
}

func (s *IncidentService) Update(ctx context.Context, id string, u domain.IncidentUpdate) (domain.Incident, error) {
	if u.Title == nil && u.Description == nil && u.Severity == nil && u.AssigneeID == nil {
		return domain.Incident{}, domain.NewError(domain.ErrValidation, "no hay campos para actualizar")
	}
	if u.Title != nil {
		title, err := normalizeIncidentTitle(*u.Title)
		if err != nil {
			return domain.Incident{}, err
		}
		u.Title = &title
	}
	if u.Severity != nil && !domain.ValidSeverity(*u.Severity) {
		return domain.Incident{}, domain.NewFieldError(domain.ErrValidation, "La severidad no es válida", "severity", "SEV1 a SEV4")
	}
	if err := s.repo.UpdateIncident(ctx, id, u); err != nil {
		return domain.Incident{}, fmt.Errorf("update incident: %w", err)
	}
	incident, err := s.Get(ctx, id)
	if err != nil {
		return domain.Incident{}, err
	}
	if u.AssigneeID != nil {
		for _, o := range s.observers {
			o.OnIncidentAssigned(ctx, incident)
		}
	}
	return incident, nil
}

// Transition applies the state machine (RN-2). resolved requires a root cause (CA-1).
func (s *IncidentService) Transition(ctx context.Context, id, to string, rootCause *string, actorID string) (domain.Incident, error) {
	if !domain.ValidIncidentStatus(to) {
		return domain.Incident{}, domain.NewFieldError(domain.ErrValidation, "El estado destino no es válido", "to", "valor no permitido")
	}
	var rc *string
	if rootCause != nil {
		trimmed := strings.TrimSpace(*rootCause)
		if trimmed != "" {
			rc = &trimmed
		}
	}
	body := "Estado cambiado a " + statusLabels[to] + "."
	if to == domain.IncidentResolved && rc != nil {
		body += " Causa raíz: " + *rc
	}
	err := s.repo.TransitionIncident(ctx, id, to, rc, actorID, body, func(from string) error {
		if !domain.CanTransition(from, to) {
			return domain.NewFieldError(domain.ErrConflict,
				fmt.Sprintf("No se puede pasar de %s a %s", statusLabels[from], statusLabels[to]), "to", "transición inválida")
		}
		if to == domain.IncidentResolved && rc == nil {
			return domain.NewFieldError(domain.ErrConflict,
				"Para resolver el incidente hay que indicar la causa raíz", "rootCause", "obligatoria")
		}
		return nil
	})
	if err != nil {
		return domain.Incident{}, fmt.Errorf("transition incident: %w", err)
	}
	return s.Get(ctx, id)
}

func (s *IncidentService) AddNote(ctx context.Context, id, authorID, body string) (domain.Incident, error) {
	body = strings.TrimSpace(body)
	if body == "" || utf8.RuneCountInString(body) > 5000 {
		return domain.Incident{}, domain.NewFieldError(domain.ErrValidation, "La nota debe tener entre 1 y 5000 caracteres", "body", "longitud inválida")
	}
	if err := s.repo.AddEvent(ctx, id, authorID, domain.EventNote, body); err != nil {
		return domain.Incident{}, fmt.Errorf("add note: %w", err)
	}
	return s.Get(ctx, id)
}

func normalizeIncidentTitle(raw string) (string, error) {
	title := strings.TrimSpace(raw)
	if n := utf8.RuneCountInString(title); n < 3 || n > 200 {
		return "", domain.NewFieldError(domain.ErrValidation, "El título debe tener entre 3 y 200 caracteres", "title", "longitud inválida")
	}
	return title, nil
}
