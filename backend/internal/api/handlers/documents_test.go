package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gen "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/auth"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/fabri-molinelli-w/opl/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type fakeDocumentService struct {
	doc         domain.Document
	err         error
	createdWith *service.CreateDocumentInput
	updatedWith *domain.DocumentUpdate
	deleted     *bool
}

func (f fakeDocumentService) List(context.Context, domain.DocumentFilter) (domain.DocumentPage, error) {
	return domain.DocumentPage{Items: []domain.DocumentSummary{}, Page: 1, PageSize: 20}, f.err
}

func (f fakeDocumentService) Get(context.Context, string) (domain.Document, error) {
	return f.doc, f.err
}

func (f fakeDocumentService) Create(_ context.Context, in service.CreateDocumentInput) (domain.Document, error) {
	if f.createdWith != nil {
		*f.createdWith = in
	}
	return f.doc, f.err
}

func (f fakeDocumentService) Update(_ context.Context, _ string, update domain.DocumentUpdate) (domain.Document, error) {
	if f.updatedWith != nil {
		*f.updatedWith = update
	}
	return f.doc, f.err
}

func (f fakeDocumentService) Delete(context.Context, string) error {
	if f.deleted != nil {
		*f.deleted = true
	}
	return f.err
}

func (f fakeDocumentService) ListVersions(context.Context, string) ([]domain.DocumentVersionSummary, error) {
	return []domain.DocumentVersionSummary{}, f.err
}

func (f fakeDocumentService) GetVersion(context.Context, string, int) (domain.DocumentVersion, error) {
	return domain.DocumentVersion{}, f.err
}

func (f fakeDocumentService) ListTags(context.Context) ([]string, error) {
	return []string{}, f.err
}

func newDocumentsRouter(t *testing.T, role domain.Role, docs DocumentService) (http.Handler, string) {
	t.Helper()
	issuer := auth.NewTokenIssuer(testSecret, time.Hour)
	user := domain.User{ID: "user-1", Email: "user@example.com", Role: role, Active: true}
	token, err := issuer.Issue(user)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	server := NewServer(Deps{DB: fakePinger{}, Documents: docs, Log: discardTestLogger()})
	strict := gen.NewStrictHandlerWithOptions(
		server,
		[]gen.StrictMiddlewareFunc{auth.Authorize(auth.NewPolicy(), issuer, fakeLookup{users: map[string]domain.User{user.ID: user}})},
		NewStrictOptions(discardTestLogger()),
	)
	r := chi.NewRouter()
	gen.HandlerFromMux(strict, r)
	return r, token
}

func doRequest(router http.Handler, token, method, path, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// CA-4: un usuario soporte que intenta eliminar recibe 403.
func TestDeleteDocumentHandler_SoporteForbidden(t *testing.T) {
	deleted := false
	router, token := newDocumentsRouter(t, domain.RoleSoporte, fakeDocumentService{deleted: &deleted})

	w := doRequest(router, token, http.MethodDelete, "/api/documents/doc-1", "")

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d (%s)", w.Code, w.Body.String())
	}
	if deleted {
		t.Error("service must not be called when the role is not allowed")
	}
}

func TestDeleteDocumentHandler_SysadminNoContent(t *testing.T) {
	deleted := false
	router, token := newDocumentsRouter(t, domain.RoleSysadmin, fakeDocumentService{deleted: &deleted})

	w := doRequest(router, token, http.MethodDelete, "/api/documents/doc-1", "")

	if w.Code != http.StatusNoContent || !deleted {
		t.Fatalf("expected 204 and delete called, got %d deleted=%v", w.Code, deleted)
	}
}

// CA-1 (parte HTTP): soporte puede crear; quien crea se toma del token, no del body.
func TestCreateDocumentHandler_SoporteUsesTokenAuthor(t *testing.T) {
	var in service.CreateDocumentInput
	docs := fakeDocumentService{
		createdWith: &in,
		doc:         domain.Document{ID: "doc-1", Title: "Reiniciar Nginx", Category: "runbook", Tags: []string{}, Version: 1},
	}
	router, token := newDocumentsRouter(t, domain.RoleSoporte, docs)

	w := doRequest(router, token, http.MethodPost, "/api/documents",
		`{"title":"Reiniciar Nginx","contentMd":"# Pasos","category":"runbook","tags":["nginx"]}`)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", w.Code, w.Body.String())
	}
	if in.AuthorID != "user-1" || in.Category != "runbook" || len(in.Tags) != 1 {
		t.Errorf("unexpected service input: %+v", in)
	}
	var body gen.Document
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Version != 1 {
		t.Errorf("unexpected body %s (%v)", w.Body.String(), err)
	}
}

// CA-6: guardar con una versión desactualizada → 409 con details[].field = "version".
func TestUpdateDocumentHandler_StaleVersionConflict(t *testing.T) {
	var update domain.DocumentUpdate
	router, token := newDocumentsRouter(t, domain.RoleSoporte, fakeDocumentService{
		updatedWith: &update, err: domain.ErrVersionMismatch,
	})

	w := doRequest(router, token, http.MethodPatch, "/api/documents/doc-1", `{"contentMd":"nuevo","version":1}`)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d (%s)", w.Code, w.Body.String())
	}
	var resp gen.Error
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Code != "conflict" || resp.Details == nil || (*resp.Details)[0].Field != "version" {
		t.Errorf("unexpected error body: %s", w.Body.String())
	}
	if update.ExpectedVersion == nil || *update.ExpectedVersion != 1 || update.EditorID != "user-1" {
		t.Errorf("unexpected update passed to service: %+v", update)
	}
}

// CA-3 (parte HTTP): título repetido → 409 con field title.
func TestCreateDocumentHandler_DuplicateTitle(t *testing.T) {
	router, token := newDocumentsRouter(t, domain.RoleSoporte, fakeDocumentService{err: domain.NewFieldError(
		domain.ErrConflict, "Ya existe un documento con ese título", "title", "título duplicado")})

	w := doRequest(router, token, http.MethodPost, "/api/documents",
		`{"title":"Reiniciar Nginx","contentMd":"x","category":"runbook"}`)

	if w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), `"field":"title"`) {
		t.Fatalf("expected 409 with title detail, got %d (%s)", w.Code, w.Body.String())
	}
}
