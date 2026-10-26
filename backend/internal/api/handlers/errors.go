package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	gen "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

type httpError struct {
	status  int
	code    string
	message string
	details []domain.FieldError
}

// NewStrictOptions returns the strict options that centralize error handling:
// domain errors become the Error schema here and nowhere else.
func NewStrictOptions(log *slog.Logger) gen.StrictHTTPServerOptions {
	return gen.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, _ error) {
			writeError(w, log, httpError{
				status:  http.StatusBadRequest,
				code:    "invalid_request",
				message: "El cuerpo de la petición es inválido",
			})
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			mapped := mapDomainError(err)
			if mapped.status >= http.StatusInternalServerError {
				log.Error("request failed", "error", err, "path", r.URL.Path)
			}
			writeError(w, log, mapped)
		},
	}
}

func mapDomainError(err error) httpError {
	details := domain.FieldErrors(err)
	switch {
	case errors.Is(err, domain.ErrValidation):
		return httpError{http.StatusBadRequest, "validation_error", userMessage(err, "Datos inválidos"), details}
	case errors.Is(err, domain.ErrUnauthenticated):
		return httpError{http.StatusUnauthorized, "unauthenticated", "Credenciales inválidas o sesión expirada", details}
	case errors.Is(err, domain.ErrForbidden):
		return httpError{http.StatusForbidden, "forbidden", "No tenés permisos para realizar esta acción", details}
	case errors.Is(err, domain.ErrNotFound):
		return httpError{http.StatusNotFound, "not_found", "El recurso no existe", details}
	case errors.Is(err, domain.ErrConflict):
		return httpError{http.StatusConflict, "conflict", userMessage(err, "El recurso ya existe o la operación no es válida"), details}
	case errors.Is(err, domain.ErrUnavailable):
		return httpError{http.StatusServiceUnavailable, "unavailable", userMessage(err, "Servicio no disponible"), details}
	default:
		return httpError{http.StatusInternalServerError, "internal_error", "Error interno del servidor", details}
	}
}

func userMessage(err error, fallback string) string {
	var domainErr *domain.Error
	if errors.As(err, &domainErr) {
		return domainErr.Message
	}
	return fallback
}

func writeError(w http.ResponseWriter, log *slog.Logger, mapped httpError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(mapped.status)
	body := gen.Error{Code: mapped.code, Message: mapped.message}
	if len(mapped.details) > 0 {
		details := make([]gen.ErrorDetail, 0, len(mapped.details))
		for _, detail := range mapped.details {
			details = append(details, gen.ErrorDetail{Field: detail.Field, Issue: detail.Issue})
		}
		body.Details = &details
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Error("failed to encode error response", "error", err)
	}
}
