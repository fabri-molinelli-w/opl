package handlers

import (
	"context"

	gen "github.com/fabri-molinelli-w/opl/backend/internal/api/gen"
	"github.com/fabri-molinelli-w/opl/backend/internal/auth"
	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

// ChatService is the OPSBot use case consumed by the handlers.
type ChatService interface {
	Ask(ctx context.Context, q domain.ChatQuestion) (domain.ChatAnswer, error)
}

func (s *Server) ChatWithOpsbot(ctx context.Context, request gen.ChatWithOpsbotRequestObject) (gen.ChatWithOpsbotResponseObject, error) {
	if request.Body == nil {
		return nil, domain.NewError(domain.ErrValidation, "el cuerpo de la petición es obligatorio")
	}
	user, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthenticated
	}
	token, _ := auth.TokenFromContext(ctx)

	answer, err := s.chat.Ask(ctx, domain.ChatQuestion{
		SessionID: request.Body.SessionId,
		Message:   request.Body.Message,
		UserID:    user.ID,
		UserToken: token,
	})
	if err != nil {
		return nil, err
	}

	sources := make([]gen.ChatSource, 0, len(answer.Sources))
	for _, src := range answer.Sources {
		sourceType := gen.ChatSourceType(src.Type)
		if !sourceType.Valid() {
			continue
		}
		sources = append(sources, gen.ChatSource{Type: sourceType, Id: src.ID, Title: src.Title})
	}
	return gen.ChatWithOpsbot200JSONResponse(gen.ChatResponse{Answer: answer.Answer, Sources: sources}), nil
}
