package repository

import (
	"context"
	"time"

	dbgen "github.com/fabri-molinelli-w/opl/backend/internal/db/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

// NotificationRepository stores in-app notifications (CU-08).
type NotificationRepository struct {
	q *dbgen.Queries
}

func NewNotificationRepository(q *dbgen.Queries) *NotificationRepository {
	return &NotificationRepository{q: q}
}

func (r *NotificationRepository) NotifyRoles(ctx context.Context, roles []string, n domain.NewNotification) (int64, error) {
	count, err := r.q.NotifyUsersByRoles(ctx, dbgen.NotifyUsersByRolesParams{
		EventType: n.EventType, Title: n.Title, Body: n.Body, Link: n.Link, DedupeKey: n.DedupeKey, Roles: roles,
	})
	if err != nil {
		return 0, mapError(err)
	}
	return count, nil
}

func (r *NotificationRepository) NotifyUser(ctx context.Context, userID string, n domain.NewNotification) (int64, error) {
	id, err := uuidFromString(userID)
	if err != nil {
		return 0, domain.ErrNotFound
	}
	count, err := r.q.NotifyUser(ctx, dbgen.NotifyUserParams{
		EventType: n.EventType, Title: n.Title, Body: n.Body, Link: n.Link, DedupeKey: n.DedupeKey, UserID: id,
	})
	if err != nil {
		return 0, mapError(err)
	}
	return count, nil
}

func (r *NotificationRepository) List(ctx context.Context, userID string, unreadOnly bool, limit int) ([]domain.Notification, error) {
	id, err := uuidFromString(userID)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	rows, err := r.q.ListNotifications(ctx, dbgen.ListNotificationsParams{UserID: id, UnreadOnly: unreadOnly, Limit: int32(limit)})
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]domain.Notification, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.Notification{
			ID: uuidToString(row.ID), EventType: row.EventType, Title: row.Title, Body: row.Body, Link: row.Link,
			ReadAt: timePtr(row.ReadAt), CreatedAt: row.CreatedAt.Time,
		})
	}
	return out, nil
}

func (r *NotificationRepository) CountUnread(ctx context.Context, userID string) (int, error) {
	id, err := uuidFromString(userID)
	if err != nil {
		return 0, domain.ErrNotFound
	}
	n, err := r.q.CountUnreadNotifications(ctx, id)
	if err != nil {
		return 0, mapError(err)
	}
	return int(n), nil
}

// MarkRead marks one notification of the user as read. Marking an already read
// notification is idempotent; a notification of another user is 404.
func (r *NotificationRepository) MarkRead(ctx context.Context, userID, notificationID string) error {
	uid, err := uuidFromString(userID)
	if err != nil {
		return domain.ErrNotFound
	}
	nid, err := uuidFromString(notificationID)
	if err != nil {
		return domain.ErrNotFound
	}
	if _, err := r.q.MarkNotificationRead(ctx, dbgen.MarkNotificationReadParams{ID: nid, UserID: uid}); err != nil {
		return mapError(err)
	}
	exists, err := r.q.NotificationExists(ctx, dbgen.NotificationExistsParams{ID: nid, UserID: uid})
	if err != nil {
		return mapError(err)
	}
	if !exists {
		return domain.ErrNotFound
	}
	return nil
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID string) error {
	id, err := uuidFromString(userID)
	if err != nil {
		return domain.ErrNotFound
	}
	if _, err := r.q.MarkAllNotificationsRead(ctx, id); err != nil {
		return mapError(err)
	}
	return nil
}

func (r *NotificationRepository) DeleteBefore(ctx context.Context, before time.Time) (int64, error) {
	n, err := r.q.DeleteNotificationsBefore(ctx, ts(before))
	if err != nil {
		return 0, mapError(err)
	}
	return n, nil
}
