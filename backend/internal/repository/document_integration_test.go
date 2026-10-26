//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"

	dbgen "github.com/fabri-molinelli-w/opl/backend/internal/db/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/fabri-molinelli-w/opl/backend/internal/repository"
	"github.com/fabri-molinelli-w/opl/backend/internal/service"
)

func ptrTo[T any](v T) *T { return &v }

// CA-1, CA-2, CA-3 y bloqueo optimista contra un PostgreSQL real.
func TestDocumentsLifecycle(t *testing.T) {
	pool := integrationPool(t)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, "TRUNCATE document_versions, documents CASCADE"); err != nil {
		t.Fatalf("truncate documents: %v", err)
	}
	var authorID, editorID string
	if err := pool.QueryRow(ctx, `INSERT INTO users (email, name, password_hash, role)
		VALUES ('doc-author@example.com', 'Autora Demo', 'x', 'soporte')
		ON CONFLICT (lower(email)) DO UPDATE SET name = EXCLUDED.name RETURNING id::text`).Scan(&authorID); err != nil {
		t.Fatalf("insert author: %v", err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO users (email, name, password_hash, role)
		VALUES ('doc-editor@example.com', 'Editor Demo', 'x', 'sysadmin')
		ON CONFLICT (lower(email)) DO UPDATE SET name = EXCLUDED.name RETURNING id::text`).Scan(&editorID); err != nil {
		t.Fatalf("insert editor: %v", err)
	}

	docs := service.NewDocumentService(repository.NewDocumentRepository(dbgen.New(pool), pool), discardLogger())

	// CA-1: soporte crea con categoría y etiquetas y lo encuentra por búsqueda.
	created, err := docs.Create(ctx, service.CreateDocumentInput{
		Title: "Configurar VPN en notebook", ContentMD: "## Pasos\n\n1. Instalar el cliente WireGuard",
		Category: "guia", Tags: []string{"VPN", "notebooks"}, AuthorID: authorID,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.Version != 1 || created.CreatedByName != "Autora Demo" {
		t.Fatalf("unexpected created document: %+v", created)
	}
	if _, err := docs.Create(ctx, service.CreateDocumentInput{
		Title: "Reiniciar Nginx", ContentMD: "sudo systemctl restart nginx", Category: "runbook",
		Tags: []string{"nginx"}, AuthorID: authorID,
	}); err != nil {
		t.Fatalf("create second: %v", err)
	}

	for _, q := range []string{"wireguard", "VPN en", "notebooks"} {
		page, err := docs.List(ctx, domain.DocumentFilter{Q: q})
		if err != nil {
			t.Fatalf("list q=%q: %v", q, err)
		}
		if page.Total != 1 || page.Items[0].ID != created.ID {
			t.Errorf("q=%q: expected only the VPN document, got %+v", q, page.Items)
		}
	}
	page, err := docs.List(ctx, domain.DocumentFilter{Q: "%"})
	if err != nil {
		t.Fatalf("list q=%%: %v", err)
	}
	if page.Total != 0 {
		t.Errorf("q=%% must be literal, got %d results", page.Total)
	}
	page, err = docs.List(ctx, domain.DocumentFilter{Tag: "vpn", Category: "guia"})
	if err != nil || page.Total != 1 {
		t.Errorf("filter tag+category: total=%d err=%v", page.Total, err)
	}
	page, err = docs.List(ctx, domain.DocumentFilter{Sort: "title", Order: "asc"})
	if err != nil || len(page.Items) != 2 || page.Items[0].Title != "Configurar VPN en notebook" {
		t.Errorf("sort title asc: %+v err=%v", page.Items, err)
	}
	page, err = docs.List(ctx, domain.DocumentFilter{Sort: "title", Order: "desc"})
	if err != nil || len(page.Items) != 2 || page.Items[0].Title != "Reiniciar Nginx" {
		t.Errorf("sort title desc: %+v err=%v", page.Items, err)
	}

	// CA-3: título repetido (sin importar mayúsculas) → 409 con field title.
	_, err = docs.Create(ctx, service.CreateDocumentInput{
		Title: "configurar vpn en NOTEBOOK", ContentMD: "x", Category: "guia", AuthorID: authorID,
	})
	if !errors.Is(err, domain.ErrConflict) || domain.FieldErrors(err)[0].Field != "title" {
		t.Fatalf("expected title conflict, got %v", err)
	}

	// CA-2: editar genera versión nueva y la anterior se puede consultar.
	updated, err := docs.Update(ctx, created.ID, domain.DocumentUpdate{
		ContentMD: ptrTo("## Pasos\n\n1. Instalar el cliente WireGuard 2.0"), ExpectedVersion: ptrTo(1), EditorID: editorID,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Version != 2 || updated.UpdatedByName != "Editor Demo" || updated.CreatedByName != "Autora Demo" {
		t.Fatalf("unexpected updated document: %+v", updated)
	}
	versions, err := docs.ListVersions(ctx, created.ID)
	if err != nil || len(versions) != 2 || versions[0].Version != 2 || versions[1].EditedByName != "Autora Demo" {
		t.Fatalf("versions: %+v err=%v", versions, err)
	}
	v1, err := docs.GetVersion(ctx, created.ID, 1)
	if err != nil || v1.ContentMD != "## Pasos\n\n1. Instalar el cliente WireGuard" {
		t.Fatalf("version 1: %+v err=%v", v1, err)
	}

	// Bloqueo optimista: guardar sobre la versión 1 cuando ya hay una 2 → 409 field version.
	_, err = docs.Update(ctx, created.ID, domain.DocumentUpdate{
		Title: ptrTo("Otro título"), ExpectedVersion: ptrTo(1), EditorID: editorID,
	})
	if !errors.Is(err, domain.ErrConflict) || domain.FieldErrors(err)[0].Field != "version" {
		t.Fatalf("expected version conflict, got %v", err)
	}
	current, err := docs.Get(ctx, created.ID)
	if err != nil || current.Version != 2 || current.Title != "Configurar VPN en notebook" {
		t.Fatalf("document changed after conflict: %+v err=%v", current, err)
	}

	// Renombrar a un título existente → 409 title y no se crea versión.
	_, err = docs.Update(ctx, created.ID, domain.DocumentUpdate{Title: ptrTo("Reiniciar NGINX"), EditorID: editorID})
	if !errors.Is(err, domain.ErrConflict) || domain.FieldErrors(err)[0].Field != "title" {
		t.Fatalf("expected title conflict on update, got %v", err)
	}
	if versions, _ := docs.ListVersions(ctx, created.ID); len(versions) != 2 {
		t.Errorf("conflicting update must not create a version, got %d", len(versions))
	}

	tags, err := docs.ListTags(ctx)
	if err != nil || len(tags) != 3 || tags[0] != "nginx" {
		t.Errorf("tags: %v err=%v", tags, err)
	}

	// Baja lógica: deja de aparecer y su título queda libre.
	if err := docs.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := docs.Get(ctx, created.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected not found after delete, got %v", err)
	}
	if _, err := docs.ListVersions(ctx, created.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected not found for versions after delete, got %v", err)
	}
	if err := docs.Delete(ctx, created.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected not found on second delete, got %v", err)
	}
	if _, err := docs.Create(ctx, service.CreateDocumentInput{
		Title: "Configurar VPN en notebook", ContentMD: "nuevo", Category: "guia", AuthorID: authorID,
	}); err != nil {
		t.Errorf("title should be reusable after delete: %v", err)
	}
}
