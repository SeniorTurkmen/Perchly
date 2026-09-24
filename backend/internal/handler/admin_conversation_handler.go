package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
	"perchly-backend/internal/service"
)

type adminConversationLister interface {
	List(ctx context.Context, filter repository.AdminConversationFilter) ([]model.ConversationPreview, int, error)
}

type adminConversationGetter interface {
	Get(ctx context.Context, id string) (service.AdminConversationDetail, error)
}

type adminMessageDeleter interface {
	DeleteMessage(ctx context.Context, adminUserID, conversationID, messageID string) error
}

// AdminConversationHandler exposes the admin dashboard's conversation
// browser and its one moderation action (deleting a message) — unlike
// ConversationHandler (public, scoped to the calling user's own
// conversations), this can list and inspect any user's.
type AdminConversationHandler struct {
	list   adminConversationLister
	get    adminConversationGetter
	delete adminMessageDeleter
}

func NewAdminConversationHandler(conversations *service.AdminConversationService) *AdminConversationHandler {
	return &AdminConversationHandler{list: conversations, get: conversations, delete: conversations}
}

type adminConversationsListResponse struct {
	Conversations []model.ConversationPreview `json:"conversations"`
	Total         int                         `json:"total"`
}

// --- GET /admin/conversations ---

func (h *AdminConversationHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := repository.AdminConversationFilter{
		Search: q.Get("search"),
		Limit:  limit,
		Offset: offset,
	}
	if userID := q.Get("user_id"); userID != "" {
		if _, err := uuid.Parse(userID); err != nil {
			writeError(w, http.StatusBadRequest, ErrCodeInvalidUserID, "geçersiz kullanıcı id")
			return
		}
		filter.UserID = &userID
	}
	if personaID := q.Get("persona_id"); personaID != "" {
		if _, err := uuid.Parse(personaID); err != nil {
			writeError(w, http.StatusBadRequest, ErrCodeInvalidPersonaID, "geçersiz persona id")
			return
		}
		filter.PersonaID = &personaID
	}

	conversations, total, err := h.list.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, ErrCodeAdminConversationsListFailed, "konuşmalar listelenemedi")
		return
	}
	writeJSON(w, http.StatusOK, adminConversationsListResponse{Conversations: conversations, Total: total})
}

// --- GET /admin/conversations/{id} ---

func (h *AdminConversationHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidConversationID, "geçersiz konuşma id")
		return
	}

	detail, err := h.get.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrConversationNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeConversationNotFound, "konuşma bulunamadı")
			return
		}
		writeError(w, http.StatusInternalServerError, ErrCodeAdminConversationFetchFailed, "konuşma getirilemedi")
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// --- DELETE /admin/conversations/{id}/messages/{messageID} ---

func (h *AdminConversationHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	conversationID := chi.URLParam(r, "id")
	messageID := chi.URLParam(r, "messageID")
	if _, err := uuid.Parse(conversationID); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidConversationID, "geçersiz konuşma id")
		return
	}
	if _, err := uuid.Parse(messageID); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidMessageID, "geçersiz mesaj id")
		return
	}

	adminUserID, _ := auth.AdminUserIDFromContext(r.Context())

	err := h.delete.DeleteMessage(r.Context(), adminUserID, conversationID, messageID)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	case errors.Is(err, repository.ErrMessageNotFound), errors.Is(err, service.ErrMessageNotInConversation):
		writeError(w, http.StatusNotFound, ErrCodeMessageNotFound, "mesaj bulunamadı")
	default:
		writeError(w, http.StatusInternalServerError, ErrCodeAdminMessageDeleteFailed, "mesaj silinemedi")
	}
}
