package repository

import (
	"context"
	"fmt"

	dbgen "github.com/fabri-molinelli-w/opl/backend/internal/db/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DocumentRepository is a thin wrapper over the sqlc-generated document queries.
type DocumentRepository struct {
	q    *dbgen.Queries
	pool *pgxpool.Pool
}

func NewDocumentRepository(q *dbgen.Queries, pool *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{q: q, pool: pool}
}

// CreateDocument inserts the document and its version 1 in one transaction.
func (r *DocumentRepository) CreateDocument(ctx context.Context, doc domain.NewDocument) (string, error) {
	authorID, err := uuidFromString(doc.AuthorID)
	if err != nil {
		return "", fmt.Errorf("author id: %w", err)
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin create document transaction: %w", err)
	}
	qtx := r.q.WithTx(tx)

	id, err := qtx.CreateDocument(ctx, dbgen.CreateDocumentParams{
		Title:     doc.Title,
		ContentMd: doc.ContentMD,
		Category:  doc.Category,
		Tags:      doc.Tags,
		AuthorID:  authorID,
	})
	if err != nil {
		return "", rollbackWithError(ctx, tx, mapError(err))
	}
	if err := qtx.InsertDocumentVersion(ctx, dbgen.InsertDocumentVersionParams{
		DocumentID: id,
		Version:    1,
		Title:      doc.Title,
		ContentMd:  doc.ContentMD,
		Category:   doc.Category,
		Tags:       doc.Tags,
		EditedBy:   authorID,
	}); err != nil {
		return "", rollbackWithError(ctx, tx, fmt.Errorf("insert first version: %w", err))
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit create document transaction: %w", err)
	}
	return uuidToString(id), nil
}

func (r *DocumentRepository) GetDocumentByID(ctx context.Context, id string) (domain.Document, error) {
	uuid, err := uuidFromString(id)
	if err != nil {
		return domain.Document{}, domain.ErrNotFound
	}
	row, err := r.q.GetDocumentByID(ctx, uuid)
	if err != nil {
		return domain.Document{}, mapError(err)
	}
	return domain.Document{
		ID:            uuidToString(row.ID),
		Title:         row.Title,
		ContentMD:     row.ContentMd,
		Category:      row.Category,
		Tags:          row.Tags,
		Version:       int(row.Version),
		CreatedByID:   uuidToString(row.CreatedBy),
		CreatedByName: row.CreatedByName,
		UpdatedByName: row.UpdatedByName,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}

// UpdateDocument locks the row (SELECT ... FOR UPDATE), lets the service apply
// the change and the optimistic lock, then writes the document and the new
// version snapshot in the same transaction (RN-4).
func (r *DocumentRepository) UpdateDocument(
	ctx context.Context, id, editorID string, apply func(domain.DocumentState) (domain.DocumentState, error),
) error {
	uuid, err := uuidFromString(id)
	if err != nil {
		return domain.ErrNotFound
	}
	editor, err := uuidFromString(editorID)
	if err != nil {
		return fmt.Errorf("editor id: %w", err)
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin update document transaction: %w", err)
	}
	qtx := r.q.WithTx(tx)

	current, err := qtx.LockDocumentForUpdate(ctx, uuid)
	if err != nil {
		return rollbackWithError(ctx, tx, mapError(err))
	}
	next, err := apply(domain.DocumentState{
		Title:     current.Title,
		ContentMD: current.ContentMd,
		Category:  current.Category,
		Tags:      current.Tags,
		Version:   int(current.Version),
	})
	if err != nil {
		return rollbackWithError(ctx, tx, err)
	}

	version, err := qtx.UpdateDocument(ctx, dbgen.UpdateDocumentParams{
		ID:        uuid,
		Title:     next.Title,
		ContentMd: next.ContentMD,
		Category:  next.Category,
		Tags:      next.Tags,
		UpdatedBy: editor,
	})
	if err != nil {
		return rollbackWithError(ctx, tx, mapError(err))
	}
	if err := qtx.InsertDocumentVersion(ctx, dbgen.InsertDocumentVersionParams{
		DocumentID: uuid,
		Version:    version,
		Title:      next.Title,
		ContentMd:  next.ContentMD,
		Category:   next.Category,
		Tags:       next.Tags,
		EditedBy:   editor,
	}); err != nil {
		return rollbackWithError(ctx, tx, fmt.Errorf("insert document version: %w", err))
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit update document transaction: %w", err)
	}
	return nil
}

// sortKey maps the validated sort/order pair to the fixed keys used by the
// CASE expressions in ListDocuments. Anything unexpected falls back to the
// default order; user input never reaches the SQL text.
func sortKey(sort, order string) string {
	switch {
	case sort == domain.DocumentSortTitle && order == domain.SortAsc:
		return "title_asc"
	case sort == domain.DocumentSortTitle && order == domain.SortDesc:
		return "title_desc"
	case sort == domain.DocumentSortUpdatedAt && order == domain.SortAsc:
		return "updatedAt_asc"
	default:
		return "updatedAt_desc"
	}
}

func (r *DocumentRepository) ListDocuments(ctx context.Context, filter domain.DocumentFilter) ([]domain.DocumentSummary, error) {
	rows, err := r.q.ListDocuments(ctx, dbgen.ListDocumentsParams{
		Q:        textParam(filter.Q),
		Category: textParam(filter.Category),
		Tag:      textParam(filter.Tag),
		SortKey:  sortKey(filter.Sort, filter.Order),
		Offset:   int32((filter.Page - 1) * filter.PageSize),
		Limit:    int32(filter.PageSize),
	})
	if err != nil {
		return nil, mapError(err)
	}
	items := make([]domain.DocumentSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.DocumentSummary{
			ID:            uuidToString(row.ID),
			Title:         row.Title,
			Category:      row.Category,
			Tags:          row.Tags,
			Version:       int(row.Version),
			UpdatedByName: row.UpdatedByName,
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
		})
	}
	return items, nil
}

func (r *DocumentRepository) CountDocuments(ctx context.Context, filter domain.DocumentFilter) (int64, error) {
	count, err := r.q.CountDocuments(ctx, dbgen.CountDocumentsParams{
		Q:        textParam(filter.Q),
		Category: textParam(filter.Category),
		Tag:      textParam(filter.Tag),
	})
	if err != nil {
		return 0, mapError(err)
	}
	return count, nil
}

func (r *DocumentRepository) DeleteDocument(ctx context.Context, id string) error {
	uuid, err := uuidFromString(id)
	if err != nil {
		return domain.ErrNotFound
	}
	rows, err := r.q.SoftDeleteDocument(ctx, uuid)
	if err != nil {
		return mapError(err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *DocumentRepository) ListDocumentVersions(ctx context.Context, id string) ([]domain.DocumentVersionSummary, error) {
	uuid, err := uuidFromString(id)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	rows, err := r.q.ListDocumentVersions(ctx, uuid)
	if err != nil {
		return nil, mapError(err)
	}
	versions := make([]domain.DocumentVersionSummary, 0, len(rows))
	for _, row := range rows {
		versions = append(versions, domain.DocumentVersionSummary{
			Version:      int(row.Version),
			Title:        row.Title,
			EditedByName: row.EditedByName,
			CreatedAt:    row.CreatedAt.Time,
		})
	}
	return versions, nil
}

func (r *DocumentRepository) GetDocumentVersion(ctx context.Context, id string, version int) (domain.DocumentVersion, error) {
	uuid, err := uuidFromString(id)
	if err != nil {
		return domain.DocumentVersion{}, domain.ErrNotFound
	}
	row, err := r.q.GetDocumentVersion(ctx, dbgen.GetDocumentVersionParams{DocumentID: uuid, Version: int32(version)})
	if err != nil {
		return domain.DocumentVersion{}, mapError(err)
	}
	return domain.DocumentVersion{
		Version:      int(row.Version),
		Title:        row.Title,
		ContentMD:    row.ContentMd,
		Category:     row.Category,
		Tags:         row.Tags,
		EditedByName: row.EditedByName,
		CreatedAt:    row.CreatedAt.Time,
	}, nil
}

func (r *DocumentRepository) ListDocumentTags(ctx context.Context) ([]string, error) {
	tags, err := r.q.ListDocumentTags(ctx)
	if err != nil {
		return nil, mapError(err)
	}
	if tags == nil {
		tags = []string{}
	}
	return tags, nil
}
