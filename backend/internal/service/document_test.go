package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

type fakeDocumentRepo struct {
	state       domain.DocumentState
	created     domain.NewDocument
	createErr   error
	updateErr   error
	lastFilter  domain.DocumentFilter
	updateCalls int
}

func (f *fakeDocumentRepo) CreateDocument(_ context.Context, doc domain.NewDocument) (string, error) {
	f.created = doc
	if f.createErr != nil {
		return "", f.createErr
	}
	f.state = domain.DocumentState{Title: doc.Title, ContentMD: doc.ContentMD, Category: doc.Category, Tags: doc.Tags, Version: 1}
	return "doc-1", nil
}

func (f *fakeDocumentRepo) GetDocumentByID(_ context.Context, id string) (domain.Document, error) {
	if id != "doc-1" {
		return domain.Document{}, domain.ErrNotFound
	}
	return domain.Document{
		ID: id, Title: f.state.Title, ContentMD: f.state.ContentMD, Category: f.state.Category,
		Tags: f.state.Tags, Version: f.state.Version,
	}, nil
}

func (f *fakeDocumentRepo) UpdateDocument(_ context.Context, id, _ string, apply func(domain.DocumentState) (domain.DocumentState, error)) error {
	f.updateCalls++
	if id != "doc-1" {
		return domain.ErrNotFound
	}
	if f.updateErr != nil {
		return f.updateErr
	}
	next, err := apply(f.state)
	if err != nil {
		return err
	}
	next.Version = f.state.Version + 1
	f.state = next
	return nil
}

func (f *fakeDocumentRepo) ListDocuments(_ context.Context, filter domain.DocumentFilter) ([]domain.DocumentSummary, error) {
	f.lastFilter = filter
	return []domain.DocumentSummary{}, nil
}

func (f *fakeDocumentRepo) CountDocuments(context.Context, domain.DocumentFilter) (int64, error) {
	return 0, nil
}

func (f *fakeDocumentRepo) DeleteDocument(context.Context, string) error { return nil }

func (f *fakeDocumentRepo) ListDocumentVersions(context.Context, string) ([]domain.DocumentVersionSummary, error) {
	return []domain.DocumentVersionSummary{}, nil
}

func (f *fakeDocumentRepo) GetDocumentVersion(context.Context, string, int) (domain.DocumentVersion, error) {
	return domain.DocumentVersion{}, domain.ErrNotFound
}

func (f *fakeDocumentRepo) ListDocumentTags(context.Context) ([]string, error) {
	return []string{}, nil
}

func newDocService(repo *fakeDocumentRepo) *DocumentService {
	return NewDocumentService(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func ptr[T any](v T) *T { return &v }

func validInput() CreateDocumentInput {
	return CreateDocumentInput{
		Title: "  Reiniciar Nginx ", ContentMD: "# Pasos", Category: "runbook",
		Tags: []string{"Nginx", " web ", "nginx", ""}, AuthorID: "user-1",
	}
}

func TestDocumentCreateNormalizes(t *testing.T) {
	repo := &fakeDocumentRepo{}
	doc, err := newDocService(repo).Create(context.Background(), validInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.created.Title != "Reiniciar Nginx" {
		t.Errorf("title not trimmed: %q", repo.created.Title)
	}
	if got := strings.Join(repo.created.Tags, ","); got != "nginx,web" {
		t.Errorf("tags = %q, want nginx,web", got)
	}
	if doc.Version != 1 {
		t.Errorf("version = %d, want 1", doc.Version)
	}
}

func TestDocumentCreateValidation(t *testing.T) {
	tests := []struct {
		name  string
		edit  func(*CreateDocumentInput)
		field string
	}{
		{"título corto", func(in *CreateDocumentInput) { in.Title = "ab" }, "title"},
		{"título largo", func(in *CreateDocumentInput) { in.Title = strings.Repeat("a", 151) }, "title"},
		{"contenido vacío", func(in *CreateDocumentInput) { in.ContentMD = "   " }, "contentMd"},
		{"contenido > 200KB", func(in *CreateDocumentInput) { in.ContentMD = strings.Repeat("a", 200*1024+1) }, "contentMd"},
		{"categoría inválida", func(in *CreateDocumentInput) { in.Category = "otra" }, "category"},
		{"etiqueta larga", func(in *CreateDocumentInput) { in.Tags = []string{strings.Repeat("x", 41)} }, "tags"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validInput()
			tt.edit(&in)
			_, err := newDocService(&fakeDocumentRepo{}).Create(context.Background(), in)
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected ErrValidation, got %v", err)
			}
			if fe := domain.FieldErrors(err); len(fe) != 1 || fe[0].Field != tt.field {
				t.Errorf("field errors = %+v, want field %q", fe, tt.field)
			}
		})
	}
}

// CA-3: título repetido → 409 con details[].field = "title".
func TestDocumentCreateDuplicateTitle(t *testing.T) {
	repo := &fakeDocumentRepo{createErr: domain.ErrConflict}
	_, err := newDocService(repo).Create(context.Background(), validInput())
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if fe := domain.FieldErrors(err); len(fe) != 1 || fe[0].Field != "title" {
		t.Errorf("field errors = %+v, want title", fe)
	}
}

// CA-2 (parte de service): editar incrementa la versión y conserva los campos no enviados.
func TestDocumentUpdateIncrementsVersion(t *testing.T) {
	repo := &fakeDocumentRepo{}
	svc := newDocService(repo)
	if _, err := svc.Create(context.Background(), validInput()); err != nil {
		t.Fatal(err)
	}
	doc, err := svc.Update(context.Background(), "doc-1", domain.DocumentUpdate{
		ContentMD: ptr("# Pasos nuevos"), ExpectedVersion: ptr(1), EditorID: "user-2",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Version != 2 || doc.ContentMD != "# Pasos nuevos" || doc.Title != "Reiniciar Nginx" {
		t.Errorf("unexpected document after update: %+v", doc)
	}
}

func TestDocumentUpdateStaleVersion(t *testing.T) {
	repo := &fakeDocumentRepo{}
	svc := newDocService(repo)
	if _, err := svc.Create(context.Background(), validInput()); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Update(context.Background(), "doc-1", domain.DocumentUpdate{
		Title: ptr("Otro título"), ExpectedVersion: ptr(7), EditorID: "user-2",
	})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if fe := domain.FieldErrors(err); len(fe) != 1 || fe[0].Field != "version" {
		t.Errorf("field errors = %+v, want version", fe)
	}
	if repo.state.Version != 1 {
		t.Errorf("version changed on conflict: %d", repo.state.Version)
	}
}

func TestDocumentUpdateWithoutFields(t *testing.T) {
	repo := &fakeDocumentRepo{}
	_, err := newDocService(repo).Update(context.Background(), "doc-1", domain.DocumentUpdate{EditorID: "u"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
	if repo.updateCalls != 0 {
		t.Errorf("repository should not be called")
	}
}

func TestDocumentListDefaultsAndEscaping(t *testing.T) {
	repo := &fakeDocumentRepo{}
	if _, err := newDocService(repo).List(context.Background(), domain.DocumentFilter{Q: " 100%_ok ", Tag: " VPN "}); err != nil {
		t.Fatal(err)
	}
	f := repo.lastFilter
	if f.Sort != "updatedAt" || f.Order != "desc" || f.Page != 1 || f.PageSize != 20 {
		t.Errorf("unexpected defaults: %+v", f)
	}
	if f.Q != `100\%\_ok` {
		t.Errorf("q not escaped: %q", f.Q)
	}
	if f.Tag != "vpn" {
		t.Errorf("tag not normalized: %q", f.Tag)
	}
}

func TestDocumentListTitleDefaultsAsc(t *testing.T) {
	repo := &fakeDocumentRepo{}
	if _, err := newDocService(repo).List(context.Background(), domain.DocumentFilter{Sort: "title"}); err != nil {
		t.Fatal(err)
	}
	if repo.lastFilter.Order != "asc" {
		t.Errorf("order = %q, want asc", repo.lastFilter.Order)
	}
}

func TestDocumentListRejectsInvalidSort(t *testing.T) {
	for _, f := range []domain.DocumentFilter{{Sort: "created_at; DROP TABLE users"}, {Order: "sideways"}, {Category: "x"}} {
		_, err := newDocService(&fakeDocumentRepo{}).List(context.Background(), f)
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("filter %+v: expected ErrValidation, got %v", f, err)
		}
	}
}

func TestDocumentVersionsOfMissingDocument(t *testing.T) {
	_, err := newDocService(&fakeDocumentRepo{}).ListVersions(context.Background(), "nope")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
