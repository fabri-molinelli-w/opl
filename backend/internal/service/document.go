package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

const (
	minTitleLength  = 3
	maxTitleLength  = 150
	maxContentBytes = 200 * 1024
	maxTags         = 20
	maxTagLength    = 40
)

// DocumentState is the editable part of a document, as read under lock by the
// repository during an update.
type DocumentState struct {
	Title     string
	ContentMD string
	Category  string
	Tags      []string
	Version   int
}

// DocumentRepository is the subset of the document repository used by
// DocumentService.
type DocumentRepository interface {
	CreateDocument(ctx context.Context, doc domain.NewDocument) (string, error)
	GetDocumentByID(ctx context.Context, id string) (domain.Document, error)
	// UpdateDocument locks the row, calls apply with the current state and
	// persists the returned state plus a new version in one transaction.
	UpdateDocument(ctx context.Context, id, editorID string, apply func(DocumentState) (DocumentState, error)) error
	ListDocuments(ctx context.Context, filter domain.DocumentFilter) ([]domain.DocumentSummary, error)
	CountDocuments(ctx context.Context, filter domain.DocumentFilter) (int64, error)
	DeleteDocument(ctx context.Context, id string) error
	ListDocumentVersions(ctx context.Context, id string) ([]domain.DocumentVersionSummary, error)
	GetDocumentVersion(ctx context.Context, id string, version int) (domain.DocumentVersion, error)
	ListDocumentTags(ctx context.Context) ([]string, error)
}

// CreateDocumentInput carries the data needed to create a document.
type CreateDocumentInput struct {
	Title     string
	ContentMD string
	Category  string
	Tags      []string
	AuthorID  string
}

type DocumentService struct {
	repo DocumentRepository
	log  *slog.Logger
}

func NewDocumentService(repo DocumentRepository, log *slog.Logger) *DocumentService {
	return &DocumentService{repo: repo, log: log}
}

func (s *DocumentService) List(ctx context.Context, filter domain.DocumentFilter) (domain.DocumentPage, error) {
	page, pageSize, err := normalizePagination(filter.Page, filter.PageSize)
	if err != nil {
		return domain.DocumentPage{}, err
	}
	filter.Page = page
	filter.PageSize = pageSize
	filter.Q = escapeLike(strings.TrimSpace(filter.Q))
	filter.Tag = strings.ToLower(strings.TrimSpace(filter.Tag))
	filter.Category = strings.TrimSpace(filter.Category)
	if filter.Category != "" && !domain.ValidCategory(filter.Category) {
		return domain.DocumentPage{}, domain.NewFieldError(
			domain.ErrValidation, "La categoría no es válida", "category", "valor no permitido")
	}
	switch filter.Sort {
	case "":
		filter.Sort = domain.DocumentSortUpdatedAt
	case domain.DocumentSortTitle, domain.DocumentSortUpdatedAt:
	default:
		return domain.DocumentPage{}, domain.NewFieldError(
			domain.ErrValidation, "El orden no es válido", "sort", "valor no permitido")
	}
	switch filter.Order {
	case "":
		filter.Order = domain.SortDesc
		if filter.Sort == domain.DocumentSortTitle {
			filter.Order = domain.SortAsc
		}
	case domain.SortAsc, domain.SortDesc:
	default:
		return domain.DocumentPage{}, domain.NewFieldError(
			domain.ErrValidation, "El sentido del orden no es válido", "order", "valor no permitido")
	}

	total, err := s.repo.CountDocuments(ctx, filter)
	if err != nil {
		return domain.DocumentPage{}, fmt.Errorf("count documents: %w", err)
	}
	items, err := s.repo.ListDocuments(ctx, filter)
	if err != nil {
		return domain.DocumentPage{}, fmt.Errorf("list documents: %w", err)
	}
	return domain.DocumentPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *DocumentService) Get(ctx context.Context, id string) (domain.Document, error) {
	doc, err := s.repo.GetDocumentByID(ctx, id)
	if err != nil {
		return domain.Document{}, fmt.Errorf("get document: %w", err)
	}
	return doc, nil
}

func (s *DocumentService) Create(ctx context.Context, in CreateDocumentInput) (domain.Document, error) {
	title, err := normalizeTitle(in.Title)
	if err != nil {
		return domain.Document{}, err
	}
	content, err := normalizeContent(in.ContentMD)
	if err != nil {
		return domain.Document{}, err
	}
	category, err := normalizeCategory(in.Category)
	if err != nil {
		return domain.Document{}, err
	}
	tags, err := NormalizeTags(in.Tags)
	if err != nil {
		return domain.Document{}, err
	}
	if in.AuthorID == "" {
		return domain.Document{}, domain.ErrUnauthenticated
	}

	id, err := s.repo.CreateDocument(ctx, domain.NewDocument{
		Title: title, ContentMD: content, Category: category, Tags: tags, AuthorID: in.AuthorID,
	})
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return domain.Document{}, duplicateTitleError()
		}
		return domain.Document{}, fmt.Errorf("create document: %w", err)
	}
	s.log.Info("document created", "document_id", id)
	return s.Get(ctx, id)
}

func (s *DocumentService) Update(ctx context.Context, id string, update domain.DocumentUpdate) (domain.Document, error) {
	if update.Title == nil && update.ContentMD == nil && update.Category == nil && update.Tags == nil {
		return domain.Document{}, domain.NewError(domain.ErrValidation, "no hay campos para actualizar")
	}
	if update.EditorID == "" {
		return domain.Document{}, domain.ErrUnauthenticated
	}

	// Validate before opening the transaction.
	var next DocumentState
	if update.Title != nil {
		title, err := normalizeTitle(*update.Title)
		if err != nil {
			return domain.Document{}, err
		}
		next.Title = title
	}
	if update.ContentMD != nil {
		content, err := normalizeContent(*update.ContentMD)
		if err != nil {
			return domain.Document{}, err
		}
		next.ContentMD = content
	}
	if update.Category != nil {
		category, err := normalizeCategory(*update.Category)
		if err != nil {
			return domain.Document{}, err
		}
		next.Category = category
	}
	var tags []string
	if update.Tags != nil {
		normalized, err := NormalizeTags(*update.Tags)
		if err != nil {
			return domain.Document{}, err
		}
		tags = normalized
	}

	err := s.repo.UpdateDocument(ctx, id, update.EditorID, func(current DocumentState) (DocumentState, error) {
		return ApplyDocumentUpdate(current, next, tags, update)
	})
	if err != nil {
		if errors.Is(err, domain.ErrConflict) && len(domain.FieldErrors(err)) == 0 {
			return domain.Document{}, duplicateTitleError()
		}
		return domain.Document{}, fmt.Errorf("update document: %w", err)
	}
	s.log.Info("document updated", "document_id", id)
	return s.Get(ctx, id)
}

// ApplyDocumentUpdate merges the validated fields over the current state and
// enforces the optimistic lock (RN: versión desactualizada → 409).
func ApplyDocumentUpdate(current, next DocumentState, tags []string, update domain.DocumentUpdate) (DocumentState, error) {
	if update.ExpectedVersion != nil && *update.ExpectedVersion != current.Version {
		return DocumentState{}, domain.ErrVersionMismatch
	}
	result := current
	if update.Title != nil {
		result.Title = next.Title
	}
	if update.ContentMD != nil {
		result.ContentMD = next.ContentMD
	}
	if update.Category != nil {
		result.Category = next.Category
	}
	if update.Tags != nil {
		result.Tags = tags
	}
	return result, nil
}

func (s *DocumentService) Delete(ctx context.Context, id string) error {
	if err := s.repo.DeleteDocument(ctx, id); err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	s.log.Info("document deleted", "document_id", id)
	return nil
}

func (s *DocumentService) ListVersions(ctx context.Context, id string) ([]domain.DocumentVersionSummary, error) {
	if _, err := s.repo.GetDocumentByID(ctx, id); err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}
	versions, err := s.repo.ListDocumentVersions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list document versions: %w", err)
	}
	return versions, nil
}

func (s *DocumentService) GetVersion(ctx context.Context, id string, version int) (domain.DocumentVersion, error) {
	if _, err := s.repo.GetDocumentByID(ctx, id); err != nil {
		return domain.DocumentVersion{}, fmt.Errorf("get document: %w", err)
	}
	v, err := s.repo.GetDocumentVersion(ctx, id, version)
	if err != nil {
		return domain.DocumentVersion{}, fmt.Errorf("get document version: %w", err)
	}
	return v, nil
}

func (s *DocumentService) ListTags(ctx context.Context) ([]string, error) {
	tags, err := s.repo.ListDocumentTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("list document tags: %w", err)
	}
	return tags, nil
}

func duplicateTitleError() error {
	return domain.NewFieldError(
		domain.ErrConflict, "Ya existe un documento con ese título", "title", "título duplicado")
}

func normalizeTitle(raw string) (string, error) {
	title := strings.TrimSpace(raw)
	n := utf8.RuneCountInString(title)
	if n < minTitleLength || n > maxTitleLength {
		return "", domain.NewFieldError(
			domain.ErrValidation, "El título debe tener entre 3 y 150 caracteres", "title", "longitud inválida")
	}
	return title, nil
}

func normalizeContent(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", domain.NewFieldError(
			domain.ErrValidation, "El contenido es obligatorio", "contentMd", "no puede estar vacío")
	}
	if len(raw) > maxContentBytes {
		return "", domain.NewFieldError(
			domain.ErrValidation, "El contenido supera el máximo de 200 KB", "contentMd", "demasiado largo")
	}
	return raw, nil
}

func normalizeCategory(raw string) (string, error) {
	category := strings.TrimSpace(raw)
	if !domain.ValidCategory(category) {
		return "", domain.NewFieldError(
			domain.ErrValidation, "La categoría no es válida", "category", "valor no permitido")
	}
	return category, nil
}

// NormalizeTags trims, lowercases and deduplicates tags preserving order.
func NormalizeTags(raw []string) ([]string, error) {
	tags := make([]string, 0, len(raw))
	seen := make(map[string]bool, len(raw))
	for _, tag := range raw {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" || seen[tag] {
			continue
		}
		if utf8.RuneCountInString(tag) > maxTagLength {
			return nil, domain.NewFieldError(
				domain.ErrValidation, "Cada etiqueta puede tener hasta 40 caracteres", "tags", "etiqueta demasiado larga")
		}
		seen[tag] = true
		tags = append(tags, tag)
	}
	if len(tags) > maxTags {
		return nil, domain.NewFieldError(
			domain.ErrValidation, "Se permiten hasta 20 etiquetas", "tags", "demasiadas etiquetas")
	}
	return tags, nil
}

// escapeLike escapes the ILIKE wildcards so the user searches literal text.
func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}
