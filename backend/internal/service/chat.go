package service

import (
	"context"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

const (
	maxChatMessage   = 4000
	maxChatSessionID = 100
)

// OpsbotClient is the ai-service client used by ChatService.
type OpsbotClient interface {
	Chat(ctx context.Context, q domain.ChatQuestion) (domain.ChatAnswer, error)
}

// ChatService validates the question and forwards it to OPSBot. A nil client
// means OPSBot is not configured (AI_SERVICE_URL / AI_INTERNAL_TOKEN unset).
type ChatService struct {
	client OpsbotClient
	log    *slog.Logger
}

func NewChatService(client OpsbotClient, log *slog.Logger) *ChatService {
	return &ChatService{client: client, log: log}
}

var errOpsbotUnavailable = domain.NewError(domain.ErrUnavailable, "OPSBot no está disponible en este momento. Probá de nuevo en unos minutos.")

func (s *ChatService) Ask(ctx context.Context, q domain.ChatQuestion) (domain.ChatAnswer, error) {
	q.Message = strings.TrimSpace(q.Message)
	q.SessionID = strings.TrimSpace(q.SessionID)
	if q.Message == "" || utf8.RuneCountInString(q.Message) > maxChatMessage {
		return domain.ChatAnswer{}, domain.NewFieldError(
			domain.ErrValidation, "La consulta debe tener entre 1 y 4000 caracteres", "message", "longitud inválida")
	}
	if q.SessionID == "" || len(q.SessionID) > maxChatSessionID {
		return domain.ChatAnswer{}, domain.NewFieldError(
			domain.ErrValidation, "La sesión de chat no es válida", "sessionId", "longitud inválida")
	}
	if q.UserID == "" || q.UserToken == "" {
		return domain.ChatAnswer{}, domain.ErrUnauthenticated
	}
	if s.client == nil {
		return domain.ChatAnswer{}, domain.NewError(domain.ErrUnavailable, "OPSBot no está configurado en este servidor.")
	}

	answer, err := s.client.Chat(ctx, q)
	if err != nil {
		// The detail is logged, never exposed (it may include internal addresses).
		s.log.Error("opsbot chat failed", "error", err, "user_id", q.UserID)
		return domain.ChatAnswer{}, errOpsbotUnavailable
	}
	return answer, nil
}
