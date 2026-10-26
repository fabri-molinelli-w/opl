package domain

import "time"

// Document categories (fixed list, see CU-05).
const (
	CategoryProcedimiento = "procedimiento"
	CategoryRunbook       = "runbook"
	CategoryGuia          = "guia"
	CategoryConfiguracion = "configuracion"
	CategoryReferencia    = "referencia"
)

// ValidCategory reports whether c is one of the fixed document categories.
func ValidCategory(c string) bool {
	switch c {
	case CategoryProcedimiento, CategoryRunbook, CategoryGuia, CategoryConfiguracion, CategoryReferencia:
		return true
	default:
		return false
	}
}

// Document sort fields and orders accepted by the listing. The repository maps
// the combination to a fixed key; user input never reaches the SQL text.
const (
	DocumentSortTitle     = "title"
	DocumentSortUpdatedAt = "updatedAt"
	SortAsc               = "asc"
	SortDesc              = "desc"
)

// Document is a piece of team documentation written in Markdown.
type Document struct {
	ID            string
	Title         string
	ContentMD     string
	Category      string
	Tags          []string
	Version       int
	CreatedByID   string
	CreatedByName string
	UpdatedByName string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// DocumentState is the editable part of a document, as read under lock by the
// repository during an update.
type DocumentState struct {
	Title     string
	ContentMD string
	Category  string
	Tags      []string
	Version   int
}

// DocumentSummary is a document without its content, used in listings.
type DocumentSummary struct {
	ID            string
	Title         string
	Category      string
	Tags          []string
	Version       int
	UpdatedByName string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewDocument carries the normalized data needed to create a document.
type NewDocument struct {
	Title     string
	ContentMD string
	Category  string
	Tags      []string
	AuthorID  string
}

// DocumentUpdate describes a partial update. ExpectedVersion enables the
// optimistic lock when it is not nil.
type DocumentUpdate struct {
	Title           *string
	ContentMD       *string
	Category        *string
	Tags            *[]string
	ExpectedVersion *int
	EditorID        string
}

// DocumentFilter carries the combinable filters of the document listing.
type DocumentFilter struct {
	Q        string
	Category string
	Tag      string
	Sort     string
	Order    string
	Page     int
	PageSize int
}

// DocumentPage is a paginated document listing.
type DocumentPage struct {
	Items    []DocumentSummary
	Total    int64
	Page     int
	PageSize int
}

// DocumentVersionSummary describes a version without its content.
type DocumentVersionSummary struct {
	Version      int
	Title        string
	EditedByName string
	CreatedAt    time.Time
}

// DocumentVersion is the full snapshot of a document at a given version.
type DocumentVersion struct {
	Version      int
	Title        string
	ContentMD    string
	Category     string
	Tags         []string
	EditedByName string
	CreatedAt    time.Time
}

// ErrVersionMismatch is wrapped (with ErrConflict) when the optimistic lock fails.
var ErrVersionMismatch = NewFieldError(
	ErrConflict,
	"El documento fue modificado por otra persona. Recargalo antes de guardar.",
	"version",
	"versión desactualizada",
)
