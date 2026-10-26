package handlers

import (
	"context"

	gen "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

// NotificationService is the CU-08 use cases consumed by the handlers. Every
// operation is scoped to the authenticated user.
type NotificationService interface {
	List(ctx context.Context, userID string, unreadOnly bool) ([]domain.Notification, error)
	CountUnread(ctx context.Context, userID string) (int, error)
	MarkRead(ctx context.Context, userID, id string) error
	MarkAllRead(ctx context.Context, userID string) error
}

func (s *Server) ListMyNotifications(ctx context.Context, request gen.ListMyNotificationsRequestObject) (gen.ListMyNotificationsResponseObject, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	items, err := s.notifications.List(ctx, userID, request.Params.Unread != nil && *request.Params.Unread)
	if err != nil {
		return nil, err
	}
	out := make([]gen.Notification, 0, len(items))
	for _, n := range items {
		out = append(out, gen.Notification{
			Id: n.ID, EventType: gen.NotificationEventType(n.EventType), Title: n.Title, Body: n.Body,
			Link: n.Link, Read: n.ReadAt != nil, CreatedAt: n.CreatedAt,
		})
	}
	return gen.ListMyNotifications200JSONResponse(out), nil
}

func (s *Server) CountMyUnreadNotifications(ctx context.Context, _ gen.CountMyUnreadNotificationsRequestObject) (gen.CountMyUnreadNotificationsResponseObject, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	n, err := s.notifications.CountUnread(ctx, userID)
	if err != nil {
		return nil, err
	}
	return gen.CountMyUnreadNotifications200JSONResponse(gen.UnreadCount{Count: n}), nil
}

func (s *Server) MarkMyNotificationRead(ctx context.Context, request gen.MarkMyNotificationReadRequestObject) (gen.MarkMyNotificationReadResponseObject, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.notifications.MarkRead(ctx, userID, request.Id); err != nil {
		return nil, err
	}
	return gen.MarkMyNotificationRead204Response{}, nil
}

func (s *Server) MarkAllMyNotificationsRead(ctx context.Context, _ gen.MarkAllMyNotificationsReadRequestObject) (gen.MarkAllMyNotificationsReadResponseObject, error) {
	userID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.notifications.MarkAllRead(ctx, userID); err != nil {
		return nil, err
	}
	return gen.MarkAllMyNotificationsRead204Response{}, nil
}
