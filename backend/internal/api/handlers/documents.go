package handlers

import (
	"context"

	gen "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/auth"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/fabri-molinelli-w/opl/backend/internal/service"
)

// DocumentService is the documentation use cases consumed by the handlers.
type DocumentService interface {
	List(ctx context.Context, filter domain.DocumentFilter) (domain.DocumentPage, error)
	Get(ctx context.Context, id string) (domain.Document, error)
	Create(ctx context.Context, in service.CreateDocumentInput) (domain.Document, error)
	Update(ctx context.Context, id string, update domain.DocumentUpdate) (domain.Document, error)
	Delete(ctx context.Context, id string) error
	ListVersions(ctx context.Context, id string) ([]domain.DocumentVersionSummary, error)
	GetVersion(ctx context.Context, id string, version int) (domain.DocumentVersion, error)
	ListTags(ctx context.Context) ([]string, error)
}

func currentUserID(ctx context.Context) (string, error) {
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return "", domain.ErrUnauthenticated
	}
	return user.ID, nil
}

func (s *Server) ListDocuments(ctx context.Context, request gen.ListDocumentsRequestObject) (gen.ListDocumentsResponseObject, error) {
	page, err := s.documents.List(ctx, domain.DocumentFilter{
		Q:        stringValue(request.Params.Q),
		Category: stringValue((*string)(request.Params.Category)),
		Tag:      stringValue(request.Params.Tag),
		Sort:     stringValue((*string)(request.Params.Sort)),
		Order:    stringValue((*string)(request.Params.Order)),
		Page:     intValue((*int)(request.Params.Page)),
		PageSize: intValue((*int)(request.Params.PageSize)),
	})
	if err != nil {
		return nil, err
	}
	items := make([]gen.DocumentSummary, 0, len(page.Items))
	for _, doc := range page.Items {
		items = append(items, gen.DocumentSummary{
			Id:            doc.ID,
			Title:         doc.Title,
			Category:      gen.DocumentCategory(doc.Category),
			Tags:          doc.Tags,
			Version:       doc.Version,
			UpdatedByName: doc.UpdatedByName,
			CreatedAt:     doc.CreatedAt,
			UpdatedAt:     doc.UpdatedAt,
		})
	}
	return gen.ListDocuments200JSONResponse(gen.PaginatedDocuments{
		Items:    items,
		Total:    page.Total,
		Page:     page.Page,
		PageSize: page.PageSize,
	}), nil
}

func (s *Server) CreateDocument(ctx context.Context, request gen.CreateDocumentRequestObject) (gen.CreateDocumentResponseObject, error) {
	if request.Body == nil {
		return nil, domain.NewError(domain.ErrValidation, "el cuerpo de la petición es obligatorio")
	}
	authorID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	doc, err := s.documents.Create(ctx, service.CreateDocumentInput{
		Title:     request.Body.Title,
		ContentMD: request.Body.ContentMd,
		Category:  string(request.Body.Category),
		Tags:      tagsValue(request.Body.Tags),
		AuthorID:  authorID,
	})
	if err != nil {
		return nil, err
	}
	return gen.CreateDocument201JSONResponse(toGenDocument(doc)), nil
}

func (s *Server) ListDocumentTags(ctx context.Context, _ gen.ListDocumentTagsRequestObject) (gen.ListDocumentTagsResponseObject, error) {
	tags, err := s.documents.ListTags(ctx)
	if err != nil {
		return nil, err
	}
	return gen.ListDocumentTags200JSONResponse(tags), nil
}

func (s *Server) GetDocument(ctx context.Context, request gen.GetDocumentRequestObject) (gen.GetDocumentResponseObject, error) {
	doc, err := s.documents.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return gen.GetDocument200JSONResponse(toGenDocument(doc)), nil
}

func (s *Server) UpdateDocument(ctx context.Context, request gen.UpdateDocumentRequestObject) (gen.UpdateDocumentResponseObject, error) {
	if request.Body == nil {
		return nil, domain.NewError(domain.ErrValidation, "el cuerpo de la petición es obligatorio")
	}
	editorID, err := currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	update := domain.DocumentUpdate{
		Title:           request.Body.Title,
		ContentMD:       request.Body.ContentMd,
		Tags:            request.Body.Tags,
		ExpectedVersion: request.Body.Version,
		EditorID:        editorID,
	}
	if request.Body.Category != nil {
		category := string(*request.Body.Category)
		update.Category = &category
	}
	doc, err := s.documents.Update(ctx, request.Id, update)
	if err != nil {
		return nil, err
	}
	return gen.UpdateDocument200JSONResponse(toGenDocument(doc)), nil
}

func (s *Server) DeleteDocument(ctx context.Context, request gen.DeleteDocumentRequestObject) (gen.DeleteDocumentResponseObject, error) {
	if err := s.documents.Delete(ctx, request.Id); err != nil {
		return nil, err
	}
	return gen.DeleteDocument204Response{}, nil
}

func (s *Server) ListDocumentVersions(ctx context.Context, request gen.ListDocumentVersionsRequestObject) (gen.ListDocumentVersionsResponseObject, error) {
	versions, err := s.documents.ListVersions(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	items := make([]gen.DocumentVersionSummary, 0, len(versions))
	for _, v := range versions {
		items = append(items, gen.DocumentVersionSummary{
			Version:      v.Version,
			Title:        v.Title,
			EditedByName: v.EditedByName,
			CreatedAt:    v.CreatedAt,
		})
	}
	return gen.ListDocumentVersions200JSONResponse(items), nil
}

func (s *Server) GetDocumentVersion(ctx context.Context, request gen.GetDocumentVersionRequestObject) (gen.GetDocumentVersionResponseObject, error) {
	v, err := s.documents.GetVersion(ctx, request.Id, request.Version)
	if err != nil {
		return nil, err
	}
	return gen.GetDocumentVersion200JSONResponse(gen.DocumentVersion{
		Version:      v.Version,
		Title:        v.Title,
		ContentMd:    v.ContentMD,
		Category:     gen.DocumentCategory(v.Category),
		Tags:         v.Tags,
		EditedByName: v.EditedByName,
		CreatedAt:    v.CreatedAt,
	}), nil
}

func toGenDocument(doc domain.Document) gen.Document {
	return gen.Document{
		Id:            doc.ID,
		Title:         doc.Title,
		ContentMd:     doc.ContentMD,
		Category:      gen.DocumentCategory(doc.Category),
		Tags:          doc.Tags,
		Version:       doc.Version,
		CreatedByName: doc.CreatedByName,
		UpdatedByName: doc.UpdatedByName,
		CreatedAt:     doc.CreatedAt,
		UpdatedAt:     doc.UpdatedAt,
	}
}
