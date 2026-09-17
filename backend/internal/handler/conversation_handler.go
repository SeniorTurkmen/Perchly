package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
	"perchly-backend/internal/service"
)

type conversationAPI interface {
	Create(ctx context.Context, userID, personaID string) (model.Conversation, error)
	List(ctx context.Context, userID string) ([]model.ConversationPreview, error)
	ListMessages(ctx context.Context, userID, conversationID string) ([]model.Message, error)
	SetMessageReaction(ctx context.Context, userID, conversationID, messageID string, emoji *string) (model.Message, error)
}

type ConversationHandler struct {
	conversations conversationAPI
}

func NewConversationHandler(conversations conversationAPI) *ConversationHandler {
	return &ConversationHandler{conversations: conversations}
}

type createConversationRequest struct {
	PersonaID string `json:"persona_id"`
}

// Create handles POST /conversations. Requires auth.Middleware to have
// run first; the owner is always the authenticated caller, never a
// client-supplied field.
func (h *ConversationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "giriş gerekli")
		return
	}

	var req createConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz istek gövdesi")
		return
	}
	if _, err := uuid.Parse(req.PersonaID); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz persona_id")
		return
	}

	conversation, err := h.conversations.Create(r.Context(), userID, req.PersonaID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrPersonaNotFound):
			writeError(w, http.StatusNotFound, "persona bulunamadı")
		case errors.Is(err, service.ErrPersonaNotAgeAppropriate):
			writeError(w, http.StatusForbidden, "bu persona yaş grubun için uygun değil")
		default:
			writeError(w, http.StatusInternalServerError, "konuşma oluşturulamadı")
		}
		return
	}
	writeJSON(w, http.StatusCreated, conversation)
}

// List handles GET /conversations — the caller's inbox, newest activity first.
func (h *ConversationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "giriş gerekli")
		return
	}

	previews, err := h.conversations.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "konuşmalar getirilemedi")
		return
	}
	writeJSON(w, http.StatusOK, previews)
}

// ListMessages handles GET /conversations/{id}/messages. Ownership is
// enforced in the service; this is a read, so it does not spend quota.
func (h *ConversationHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "giriş gerekli")
		return
	}

	conversationID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(conversationID); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz conversation id")
		return
	}

	messages, err := h.conversations.ListMessages(r.Context(), userID, conversationID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrConversationNotFound):
			writeError(w, http.StatusNotFound, "konuşma bulunamadı")
		case errors.Is(err, service.ErrConversationForbidden):
			writeError(w, http.StatusForbidden, "bu konuşmaya erişiminiz yok")
		default:
			writeError(w, http.StatusInternalServerError, "mesajlar getirilemedi")
		}
		return
	}
	writeJSON(w, http.StatusOK, messages)
}

type setReactionRequest struct {
	Emoji string `json:"emoji"`
}

// SetReaction handles POST /conversations/{id}/messages/{messageID}/reaction.
// Only assistant messages in a conversation the caller owns can be
// reacted to — see ConversationService.SetMessageReaction.
func (h *ConversationHandler) SetReaction(w http.ResponseWriter, r *http.Request) {
	var req setReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Emoji) == "" {
		writeError(w, http.StatusBadRequest, "emoji alanı zorunludur")
		return
	}
	h.setReaction(w, r, &req.Emoji)
}

// ClearReaction handles DELETE /conversations/{id}/messages/{messageID}/reaction.
func (h *ConversationHandler) ClearReaction(w http.ResponseWriter, r *http.Request) {
	h.setReaction(w, r, nil)
}

func (h *ConversationHandler) setReaction(w http.ResponseWriter, r *http.Request, emoji *string) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "giriş gerekli")
		return
	}

	conversationID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(conversationID); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz conversation id")
		return
	}
	messageID := chi.URLParam(r, "messageID")
	if _, err := uuid.Parse(messageID); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz message id")
		return
	}

	message, err := h.conversations.SetMessageReaction(r.Context(), userID, conversationID, messageID, emoji)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrConversationNotFound):
			writeError(w, http.StatusNotFound, "konuşma bulunamadı")
		case errors.Is(err, repository.ErrMessageNotFound):
			writeError(w, http.StatusNotFound, "mesaj bulunamadı")
		case errors.Is(err, service.ErrConversationForbidden):
			writeError(w, http.StatusForbidden, "bu konuşmaya erişiminiz yok")
		case errors.Is(err, service.ErrCannotReactToOwnRoleMessage):
			writeError(w, http.StatusForbidden, "bu mesaja tepki bırakamazsınız")
		case errors.Is(err, service.ErrInvalidReactionEmoji):
			writeError(w, http.StatusBadRequest, "desteklenmeyen emoji")
		default:
			writeError(w, http.StatusInternalServerError, "tepki kaydedilemedi")
		}
		return
	}
	writeJSON(w, http.StatusOK, message)
}
