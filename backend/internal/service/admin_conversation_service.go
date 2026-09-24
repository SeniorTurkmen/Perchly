package service

import (
	"context"
	"errors"

	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
)

// ErrMessageNotInConversation is returned when a message id is valid
// but doesn't belong to the conversation id it was addressed through —
// an admin.message.delete request that got the wrong conversation id,
// not something a legitimate caller should ever trigger.
var ErrMessageNotInConversation = errors.New("message does not belong to conversation")

type AdminConversationRepo interface {
	ListForAdmin(ctx context.Context, filter repository.AdminConversationFilter) ([]model.ConversationPreview, int, error)
	GetByID(ctx context.Context, id string) (model.Conversation, error)
}

type AdminMessageRepo interface {
	ListByConversation(ctx context.Context, conversationID string) ([]model.Message, error)
	GetByID(ctx context.Context, id string) (model.Message, error)
	Delete(ctx context.Context, id string) error
}

// AdminConversationDetail is a conversation plus its full message
// history — the admin dashboard's conversation view always needs both.
type AdminConversationDetail struct {
	Conversation model.Conversation `json:"conversation"`
	Messages     []model.Message    `json:"messages"`
}

// AdminConversationService is the admin dashboard's read access to
// every conversation (not just one user's, unlike ConversationService)
// and its moderation action — deleting a single message. Deletion is
// written to admin_audit_log, since it destroys user-visible content.
type AdminConversationService struct {
	conversations AdminConversationRepo
	messages      AdminMessageRepo
	auditLog      AdminAuditLogRepo
}

func NewAdminConversationService(conversations AdminConversationRepo, messages AdminMessageRepo, auditLog AdminAuditLogRepo) *AdminConversationService {
	return &AdminConversationService{conversations: conversations, messages: messages, auditLog: auditLog}
}

func (s *AdminConversationService) List(ctx context.Context, filter repository.AdminConversationFilter) ([]model.ConversationPreview, int, error) {
	return s.conversations.ListForAdmin(ctx, filter)
}

func (s *AdminConversationService) Get(ctx context.Context, id string) (AdminConversationDetail, error) {
	conversation, err := s.conversations.GetByID(ctx, id)
	if err != nil {
		return AdminConversationDetail{}, err
	}

	messages, err := s.messages.ListByConversation(ctx, id)
	if err != nil {
		return AdminConversationDetail{}, err
	}

	return AdminConversationDetail{Conversation: conversation, Messages: messages}, nil
}

// DeleteMessage removes one message — permanently, no undo. conversationID
// is required (not just messageID) so the caller can't delete a message
// by id alone without the UI having actually shown it in that
// conversation first; a mismatch is treated as not-found.
func (s *AdminConversationService) DeleteMessage(ctx context.Context, adminUserID, conversationID, messageID string) error {
	message, err := s.messages.GetByID(ctx, messageID)
	if err != nil {
		return err
	}
	if message.ConversationID != conversationID {
		return ErrMessageNotInConversation
	}

	if err := s.messages.Delete(ctx, messageID); err != nil {
		return err
	}

	_ = s.auditLog.Create(ctx, adminUserID, "admin.message.delete", "message", &messageID, map[string]any{
		"conversation_id": conversationID,
	})

	return nil
}
