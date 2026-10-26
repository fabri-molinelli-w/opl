//go:build integration

package repository_test

import (
	"context"
	"testing"

	dbgen "github.com/fabri-molinelli-w/opl/backend/internal/db/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/fabri-molinelli-w/opl/backend/internal/repository"
	"github.com/fabri-molinelli-w/opl/backend/internal/service"
)

// CU-08 CA-1..CA-3 contra PostgreSQL real.
func TestNotificationsFanOutDedupeAndRead(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, "TRUNCATE notifications"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	ids := map[string]string{}
	for _, u := range []struct{ email, name, role string }{
		{"notif-sys@example.com", "Sys Demo", "sysadmin"},
		{"notif-sop@example.com", "Soporte Demo", "soporte"},
	} {
		var id string
		if err := pool.QueryRow(ctx, `INSERT INTO users (email, name, password_hash, role) VALUES ($1, $2, 'x', $3)
			ON CONFLICT (lower(email)) DO UPDATE SET name = EXCLUDED.name, active = true RETURNING id::text`,
			u.email, u.name, u.role).Scan(&id); err != nil {
			t.Fatalf("insert user: %v", err)
		}
		ids[u.role] = id
	}

	notifications := service.NewNotificationService(repository.NewNotificationRepository(dbgen.New(pool)), discardLogger()).Synchronous()
	down := domain.CheckOutcome{
		Check:          domain.HealthCheck{ServiceID: "00000000-0000-4000-8000-000000000001", Status: domain.StatusDown, Error: "connection refused"},
		ServerID:       "00000000-0000-4000-8000-000000000002",
		ServiceName:    "lab-nginx",
		PreviousStatus: domain.StatusUp,
	}

	// CA-1: el sysadmin recibe service.down; soporte no.
	notifications.OnCheck(ctx, down)
	if n, _ := notifications.CountUnread(ctx, ids["sysadmin"]); n != 1 {
		t.Fatalf("sysadmin unread = %d, want 1", n)
	}
	if n, _ := notifications.CountUnread(ctx, ids["soporte"]); n != 0 {
		t.Fatalf("soporte unread = %d, want 0", n)
	}

	// CA-2: durante la misma caída (la anterior sin leer) no se duplica.
	notifications.OnCheck(ctx, down)
	if n, _ := notifications.CountUnread(ctx, ids["sysadmin"]); n != 1 {
		t.Fatalf("duplicate notification: unread = %d", n)
	}

	// Incidente SEV2 → todos; SEV3 → nadie.
	notifications.OnIncidentOpened(ctx, domain.Incident{ID: "00000000-0000-4000-8000-000000000003", Title: "Caída", Severity: "SEV2"})
	notifications.OnIncidentOpened(ctx, domain.Incident{ID: "00000000-0000-4000-8000-000000000004", Title: "Menor", Severity: "SEV3"})
	if n, _ := notifications.CountUnread(ctx, ids["soporte"]); n != 1 {
		t.Fatalf("soporte unread after SEV2 = %d, want 1", n)
	}

	// document.updated: solo al autor y solo si edita otra persona.
	doc := domain.Document{ID: "00000000-0000-4000-8000-000000000005", Title: "VPN", CreatedByID: ids["soporte"], UpdatedByName: "Sys Demo", Version: 2}
	notifications.OnDocumentUpdated(ctx, doc, ids["soporte"])
	if n, _ := notifications.CountUnread(ctx, ids["soporte"]); n != 1 {
		t.Fatalf("self edit must not notify: unread = %d", n)
	}
	notifications.OnDocumentUpdated(ctx, doc, ids["sysadmin"])
	list, err := notifications.List(ctx, ids["soporte"], true)
	if err != nil || len(list) != 2 || list[0].EventType != domain.EventDocumentUpdated || list[0].Link != "/documentacion/"+doc.ID {
		t.Fatalf("soporte notifications: %+v err=%v", list, err)
	}

	// Otro usuario no puede marcar una notificación ajena (404).
	if err := notifications.MarkRead(ctx, ids["sysadmin"], list[0].ID); err == nil {
		t.Fatal("marking another user's notification must fail")
	}
	if err := notifications.MarkRead(ctx, ids["soporte"], list[0].ID); err != nil {
		t.Fatalf("mark read: %v", err)
	}

	// CA-3: "marcar todas como leídas" deja el contador en 0, y después se puede volver a notificar.
	if err := notifications.MarkAllRead(ctx, ids["sysadmin"]); err != nil {
		t.Fatal(err)
	}
	if n, _ := notifications.CountUnread(ctx, ids["sysadmin"]); n != 0 {
		t.Fatalf("after read-all: unread = %d", n)
	}
	notifications.OnCheck(ctx, down)
	if n, _ := notifications.CountUnread(ctx, ids["sysadmin"]); n != 1 {
		t.Fatalf("a new outage after reading must notify again: unread = %d", n)
	}
}
