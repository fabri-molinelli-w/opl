package domain

import "time"

// Server is an inventoried machine. IP is stored as text in the domain and as
// inet in PostgreSQL.
type Server struct {
	ID          string
	Hostname    string
	IP          string
	Environment string
	OS          string
	Location    string
	Owner       string
	Tags        []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Services    []Service
}

// ServerUpdate describes a partial update of a server.
type ServerUpdate struct {
	Hostname    *string
	IP          *string
	Environment *string
	OS          *string
	Location    *string
	Owner       *string
	Tags        *[]string
}

// ServerFilter carries the combinable filters of the server listing.
type ServerFilter struct {
	Q           string
	Environment string
	OS          string
	Tag         string
	Page        int
	PageSize    int
}

// ServerPage is a paginated server listing.
type ServerPage struct {
	Items    []Server
	Total    int64
	Page     int
	PageSize int
}
