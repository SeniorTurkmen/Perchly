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

// Create godoc
// @Summary Yeni konuşma başlat
// @Description Sahibi her zaman kimlik doğrulanan çağırandır; client tarafından belirlenemez. Reşit olmayan kullanıcı is_minor_appropriate=false bir personayla konuşma açamaz.
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body createConversationRequest true "persona_id"
// @Success 201 {object} model.Conversation
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse "persona yaş grubuna uygun değil"
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /conversations [post]
func (h *ConversationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "giriş gerekli")
		return
	}

	var req createConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
		return
	}
	if _, err := uuid.Parse(req.PersonaID); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidPersonaID, "geçersiz persona_id")
		return
	}

	conversation, err := h.conversations.Create(r.Context(), userID, req.PersonaID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrPersonaNotFound):
			writeError(w, http.StatusNotFound, ErrCodePersonaNotFound, "persona bulunamadı")
		case errors.Is(err, service.ErrPersonaNotAgeAppropriate):
			writeError(w, http.StatusForbidden, ErrCodePersonaNotAgeAppropriate, "bu persona yaş grubun için uygun değil")
		default:
			writeError(w, http.StatusInternalServerError, ErrCodeConversationCreateFailed, "konuşma oluşturulamadı")
		}
		return
	}
	writeJSON(w, http.StatusCreated, conversation)
}

// List godoc
// @Summary Konuşmaları listele
// @Description Çağıranın gelen kutusu, en son etkinliğe göre sıralı.
// @Tags conversations
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.ConversationPreview
// @Failure 401 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /conversations [get]
func (h *ConversationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "giriş gerekli")
		return
	}

	previews, err := h.conversations.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, ErrCodeConversationsListFailed, "konuşmalar getirilemedi")
		return
	}
	writeJSON(w, http.StatusOK, previews)
}

// ListMessages godoc
// @Summary Konuşmadaki mesajları listele
// @Description Salt okunur; kota harcamaz. Sahiplik kontrolü sunucu tarafında yapılır.
// @Tags conversations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation ID (UUID)"
// @Success 200 {array} model.Message
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /conversations/{id}/messages [get]
func (h *ConversationHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "giriş gerekli")
		return
	}

	conversationID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(conversationID); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidConversationID, "geçersiz conversation id")
		return
	}

	messages, err := h.conversations.ListMessages(r.Context(), userID, conversationID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrConversationNotFound):
			writeError(w, http.StatusNotFound, ErrCodeConversationNotFound, "konuşma bulunamadı")
		case errors.Is(err, service.ErrConversationForbidden):
			writeError(w, http.StatusForbidden, ErrCodeConversationForbidden, "bu konuşmaya erişiminiz yok")
		default:
			writeError(w, http.StatusInternalServerError, ErrCodeMessagesListFailed, "mesajlar getirilemedi")
		}
		return
	}
	writeJSON(w, http.StatusOK, messages)
}

type setReactionRequest struct {
	Emoji string `json:"emoji"`
}

// SetReaction godoc
// @Summary Mesaja emoji tepkisi bırak
// @Description Yalnızca çağıranın sahip olduğu konuşmadaki asistan mesajlarına tepki bırakılabilir.
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation ID (UUID)"
// @Param messageID path string true "Message ID (UUID)"
// @Param body body setReactionRequest true "emoji"
// @Success 200 {object} model.Message
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /conversations/{id}/messages/{messageID}/reaction [post]
func (h *ConversationHandler) SetReaction(w http.ResponseWriter, r *http.Request) {
	var req setReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Emoji) == "" {
		writeError(w, http.StatusBadRequest, ErrCodeReactionEmojiRequired, "emoji alanı zorunludur")
		return
	}
	h.setReaction(w, r, &req.Emoji)
}

// ClearReaction godoc
// @Summary Mesajdaki emoji tepkisini kaldır
// @Tags conversations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation ID (UUID)"
// @Param messageID path string true "Message ID (UUID)"
// @Success 200 {object} model.Message
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /conversations/{id}/messages/{messageID}/reaction [delete]
func (h *ConversationHandler) ClearReaction(w http.ResponseWriter, r *http.Request) {
	h.setReaction(w, r, nil)
}

func (h *ConversationHandler) setReaction(w http.ResponseWriter, r *http.Request, emoji *string) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "giriş gerekli")
		return
	}

	conversationID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(conversationID); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidConversationID, "geçersiz conversation id")
		return
	}
	messageID := chi.URLParam(r, "messageID")
	if _, err := uuid.Parse(messageID); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidMessageID, "geçersiz message id")
		return
	}

	message, err := h.conversations.SetMessageReaction(r.Context(), userID, conversationID, messageID, emoji)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrConversationNotFound):
			writeError(w, http.StatusNotFound, ErrCodeConversationNotFound, "konuşma bulunamadı")
		case errors.Is(err, repository.ErrMessageNotFound):
			writeError(w, http.StatusNotFound, ErrCodeMessageNotFound, "mesaj bulunamadı")
		case errors.Is(err, service.ErrConversationForbidden):
			writeError(w, http.StatusForbidden, ErrCodeConversationForbidden, "bu konuşmaya erişiminiz yok")
		case errors.Is(err, service.ErrCannotReactToOwnRoleMessage):
			writeError(w, http.StatusForbidden, ErrCodeCannotReactToOwnMessage, "bu mesaja tepki bırakamazsınız")
		case errors.Is(err, service.ErrInvalidReactionEmoji):
			writeError(w, http.StatusBadRequest, ErrCodeInvalidReactionEmoji, "desteklenmeyen emoji")
		default:
			writeError(w, http.StatusInternalServerError, ErrCodeReactionFailed, "tepki kaydedilemedi")
		}
		return
	}
	writeJSON(w, http.StatusOK, message)
}
