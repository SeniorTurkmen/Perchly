package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
	"perchly-backend/internal/service"
)

type conversationGetter interface {
	GetByID(ctx context.Context, id string) (model.Conversation, error)
}

type quotaChecker interface {
	Check(ctx context.Context, userID, personaID string) (service.QuotaCheckResult, error)
}

type messageContextKey string

const (
	conversationContextKey messageContextKey = "conversation"
	useCreditContextKey    messageContextKey = "useCredit"
)

func withConversation(ctx context.Context, c model.Conversation) context.Context {
	return context.WithValue(ctx, conversationContextKey, c)
}

// ConversationFromContext returns the conversation QuotaMiddleware
// already resolved and authorized, so handlers never need to fetch it
// again.
func ConversationFromContext(ctx context.Context) (model.Conversation, bool) {
	c, ok := ctx.Value(conversationContextKey).(model.Conversation)
	return c, ok
}

func withUseCredit(ctx context.Context, useCredit bool) context.Context {
	return context.WithValue(ctx, useCreditContextKey, useCredit)
}

// UseCreditFromContext reports whether QuotaMiddleware decided this
// request should spend a credit instead of counting against the free
// daily quota.
func UseCreditFromContext(ctx context.Context) bool {
	useCredit, _ := ctx.Value(useCreditContextKey).(bool)
	return useCredit
}

// QuotaMiddleware resolves the {id} conversation, checks the caller owns
// it, then checks their quota — rejecting with 404/403/429 before the
// route handler ever runs. On success it stashes the conversation and
// the quota decision in the request context for the handler to use, so
// neither gets fetched or computed twice.
//
// Must run after auth.Middleware (needs the user id it puts in context)
// and only on routes with an {id} URL param that is a conversation id.
func QuotaMiddleware(conversations conversationGetter, quotas quotaChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := auth.UserIDFromContext(r.Context())
			if !ok {
				writeError(w, r, http.StatusUnauthorized, ErrCodeUnauthorized, "giriş gerekli")
				return
			}

			conversationID := chi.URLParam(r, "id")
			if _, err := uuid.Parse(conversationID); err != nil {
				writeError(w, r, http.StatusBadRequest, ErrCodeInvalidConversationID, "geçersiz conversation id")
				return
			}

			conversation, err := conversations.GetByID(r.Context(), conversationID)
			if err != nil {
				if errors.Is(err, repository.ErrConversationNotFound) {
					writeError(w, r, http.StatusNotFound, ErrCodeConversationNotFound, "konuşma bulunamadı")
					return
				}
				writeError(w, r, http.StatusInternalServerError, ErrCodeConversationFetchFailed, "konuşma getirilemedi")
				return
			}

			if conversation.UserID != userID {
				writeError(w, r, http.StatusForbidden, ErrCodeConversationForbidden, "bu konuşmaya erişiminiz yok")
				return
			}

			result, err := quotas.Check(r.Context(), userID, conversation.PersonaID)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, ErrCodeQuotaCheckFailed, "kota kontrol edilemedi")
				return
			}
			if !result.Allowed {
				writeError(w, r, http.StatusTooManyRequests, ErrCodeQuotaExceeded, "günlük mesaj hakkınız doldu")
				return
			}

			ctx := withConversation(r.Context(), conversation)
			ctx = withUseCredit(ctx, result.UseCredit)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
