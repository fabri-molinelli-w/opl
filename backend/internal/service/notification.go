package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

const (
	notificationListLimit = 20
	notificationRetention = 30 * 24 * time.Hour
	notifyTimeout         = 5 * time.Second
)

// NotificationRepository is the persistence used by NotificationService.
type NotificationRepository interface {
	NotifyRoles(ctx context.Context, roles []string, n domain.NewNotification) (int64, error)
	NotifyUser(ctx context.Context, userID string, n domain.NewNotification) (int64, error)
	List(ctx context.Context, userID string, unreadOnly bool, limit int) ([]domain.Notification, error)
	CountUnread(ctx context.Context, userID string) (int, error)
	MarkRead(ctx context.Context, userID, notificationID string) error
	MarkAllRead(ctx context.Context, userID string) error
	DeleteBefore(ctx context.Context, before time.Time) (int64, error)
}

// NotificationService exposes the user's notifications and generates them from
// internal events. Generation is asynchronous and best effort (RN-4): an error
// is logged and never fails the operation that triggered it.
type NotificationService struct {
	repo  NotificationRepository
	log   *slog.Logger
	async bool
	wg    sync.WaitGroup
}

func NewNotificationService(repo NotificationRepository, log *slog.Logger) *NotificationService {
	return &NotificationService{repo: repo, log: log, async: true}
}

// Synchronous makes event handling run inline (tests).
func (s *NotificationService) Synchronous() *NotificationService {
	s.async = false
	return s
}

// Wait blocks until pending notifications are written (graceful shutdown).
func (s *NotificationService) Wait() { s.wg.Wait() }

func (s *NotificationService) List(ctx context.Context, userID string, unreadOnly bool) ([]domain.Notification, error) {
	items, err := s.repo.List(ctx, userID, unreadOnly, notificationListLimit)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	return items, nil
}

func (s *NotificationService) CountUnread(ctx context.Context, userID string) (int, error) {
	n, err := s.repo.CountUnread(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("count notifications: %w", err)
	}
	return n, nil
}

func (s *NotificationService) MarkRead(ctx context.Context, userID, id string) error {
	if err := s.repo.MarkRead(ctx, userID, id); err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	return nil
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID string) error {
	if err := s.repo.MarkAllRead(ctx, userID); err != nil {
		return fmt.Errorf("mark all notifications read: %w", err)
	}
	return nil
}

// CleanUp deletes notifications older than 30 days (RN-5).
func (s *NotificationService) CleanUp(ctx context.Context, now time.Time) error {
	if _, err := s.repo.DeleteBefore(ctx, now.Add(-notificationRetention)); err != nil {
		return fmt.Errorf("delete old notifications: %w", err)
	}
	return nil
}

// emit runs fn detached from the request context (which may be cancelled once
// the response is written), with its own timeout.
func (s *NotificationService) emit(event string, fn func(ctx context.Context) error) {
	run := func() {
		ctx, cancel := context.WithTimeout(context.Background(), notifyTimeout)
		defer cancel()
		if err := fn(ctx); err != nil {
			s.log.Error("notification failed", "event", event, "error", err)
		}
	}
	if !s.async {
		run()
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		run()
	}()
}

// OnCheck implements CheckObserver: service.down to sysadmin/admin when a
// service goes DOWN, and incident.opened when the check opened an incident.
func (s *NotificationService) OnCheck(_ context.Context, o domain.CheckOutcome) {
	if o.Check.Status == domain.StatusDown && o.PreviousStatus != domain.StatusDown {
		s.emit(domain.EventServiceDown, func(ctx context.Context) error {
			_, err := s.repo.NotifyRoles(ctx, []string{string(domain.RoleSysadmin), string(domain.RoleAdmin)}, domain.NewNotification{
				EventType: domain.EventServiceDown,
				Title:     fmt.Sprintf("Servicio %s caído", o.ServiceName),
				Body:      o.Check.Error,
				Link:      "/inventario/" + o.ServerID,
				DedupeKey: "service.down:" + o.Check.ServiceID,
			})
			return err
		})
	}
	if o.IncidentOpenedID != "" {
		s.notifyIncidentOpened(o.IncidentOpenedID, fmt.Sprintf("Servicio %s caído", o.ServiceName), "SEV2")
	}
}

// OnIncidentOpened implements IncidentObserver: SEV1/SEV2 → todos los usuarios.
func (s *NotificationService) OnIncidentOpened(_ context.Context, incident domain.Incident) {
	s.notifyIncidentOpened(incident.ID, incident.Title, incident.Severity)
}

func (s *NotificationService) notifyIncidentOpened(id, title, severity string) {
	if severity != "SEV1" && severity != "SEV2" {
		return
	}
	s.emit(domain.EventIncidentOpened, func(ctx context.Context) error {
		_, err := s.repo.NotifyRoles(ctx, []string{
			string(domain.RoleSoporte), string(domain.RoleSysadmin), string(domain.RoleAdmin),
		}, domain.NewNotification{
			EventType: domain.EventIncidentOpened,
			Title:     fmt.Sprintf("Incidente %s abierto: %s", severity, title),
			Link:      "/incidentes/" + id,
			DedupeKey: "incident.opened:" + id,
		})
		return err
	})
}

// OnIncidentAssigned implements IncidentObserver: avisa al responsable.
func (s *NotificationService) OnIncidentAssigned(_ context.Context, incident domain.Incident) {
	if incident.AssigneeID == "" {
		return
	}
	s.emit(domain.EventIncidentAssigned, func(ctx context.Context) error {
		_, err := s.repo.NotifyUser(ctx, incident.AssigneeID, domain.NewNotification{
			EventType: domain.EventIncidentAssigned,
			Title:     "Te asignaron el incidente: " + incident.Title,
			Link:      "/incidentes/" + incident.ID,
			DedupeKey: "incident.assigned:" + incident.ID,
		})
		return err
	})
}

// OnDocumentUpdated implements DocumentObserver: avisa al autor original cuando
// otra persona edita su documento.
func (s *NotificationService) OnDocumentUpdated(_ context.Context, doc domain.Document, editorID string) {
	if doc.CreatedByID == "" || doc.CreatedByID == editorID {
		return
	}
	s.emit(domain.EventDocumentUpdated, func(ctx context.Context) error {
		_, err := s.repo.NotifyUser(ctx, doc.CreatedByID, domain.NewNotification{
			EventType: domain.EventDocumentUpdated,
			Title:     fmt.Sprintf("%s editó \"%s\"", doc.UpdatedByName, doc.Title),
			Body:      fmt.Sprintf("Versión %d", doc.Version),
			Link:      "/documentacion/" + doc.ID,
			DedupeKey: "document.updated:" + doc.ID,
		})
		return err
	})
}
