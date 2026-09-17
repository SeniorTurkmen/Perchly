package service

import (
	"context"
	"errors"
	"fmt"

	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
)

type ConversationRepo interface {
	Create(ctx context.Context, userID, personaID string) (model.Conversation, error)
	GetByID(ctx context.Context, id string) (model.Conversation, error)
	ListByUserID(ctx context.Context, userID string) ([]model.ConversationPreview, error)
}

type messageLister interface {
	ListByConversation(ctx context.Context, conversationID string) ([]model.Message, error)
	GetByID(ctx context.Context, id string) (model.Message, error)
	SetReaction(ctx context.Context, id string, emoji *string) (model.Message, error)
}

// ErrConversationForbidden is returned when a caller asks for a
// conversation (or its messages) they don't own.
var ErrConversationForbidden = errors.New("conversation not owned by this user")

// ErrInvalidReactionEmoji is returned when the requested emoji isn't one
// of model.OrderedAllowedReactionEmojis.
var ErrInvalidReactionEmoji = errors.New("emoji is not a supported reaction")

// ErrCannotReactToOwnRoleMessage is returned when a caller tries to react
// to a message sent by their own side of the conversation — only the
// user may react to an assistant message, and only the persona (via
// ChatService) may react to a user message.
var ErrCannotReactToOwnRoleMessage = errors.New("cannot react to a message from your own side of the conversation")

// ErrPersonaNotAgeAppropriate is returned when a user known to be a
// minor (via their onboarding profile) tries to start a conversation
// with a persona not marked IsMinorAppropriate.
var ErrPersonaNotAgeAppropriate = errors.New("persona not age appropriate for this user")

type ConversationService struct {
	repo               ConversationRepo
	personas           PersonaRepo
	onboardingProfiles OnboardingProfileRepo
	messages           messageLister
}

func NewConversationService(repo ConversationRepo, personas PersonaRepo, onboardingProfiles OnboardingProfileRepo, messages messageLister) *ConversationService {
	return &ConversationService{repo: repo, personas: personas, onboardingProfiles: onboardingProfiles, messages: messages}
}

// Create starts a new conversation, first checking the persona is
// age-appropriate for userID — see checkAgeAppropriate.
func (s *ConversationService) Create(ctx context.Context, userID, personaID string) (model.Conversation, error) {
	persona, err := s.personas.GetByID(ctx, personaID)
	if err != nil {
		return model.Conversation{}, err
	}

	if err := s.checkAgeAppropriate(ctx, userID, persona); err != nil {
		return model.Conversation{}, err
	}

	return s.repo.Create(ctx, userID, personaID)
}

// checkAgeAppropriate rejects a persona not marked IsMinorAppropriate
// for a user whose onboarding profile says they're a minor. A user with
// no onboarding profile yet (or any lookup error besides "not found")
// is treated permissively — their age is simply unknown, so this can't
// confirm they need restricting. That default has no practical effect
// today (every seeded persona is IsMinorAppropriate), but matters once
// a persona is added that isn't.
func (s *ConversationService) checkAgeAppropriate(ctx context.Context, userID string, persona model.Persona) error {
	if persona.IsMinorAppropriate {
		return nil
	}

	profile, err := s.onboardingProfiles.GetByUserID(ctx, userID)
	switch {
	case err == nil:
		if profile.IsMinor {
			return ErrPersonaNotAgeAppropriate
		}
		return nil
	case errors.Is(err, repository.ErrOnboardingProfileNotFound):
		return nil
	default:
		return fmt.Errorf("check age appropriateness: %w", err)
	}
}

func (s *ConversationService) GetByID(ctx context.Context, id string) (model.Conversation, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ConversationService) List(ctx context.Context, userID string) ([]model.ConversationPreview, error) {
	return s.repo.ListByUserID(ctx, userID)
}

// ListMessages returns the persisted transcript for a conversation the
// caller owns. History is never charged against quota.
func (s *ConversationService) ListMessages(ctx context.Context, userID, conversationID string) ([]model.Message, error) {
	conversation, err := s.repo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if conversation.UserID != userID {
		return nil, ErrConversationForbidden
	}
	return s.messages.ListByConversation(ctx, conversationID)
}

// SetMessageReaction sets (emoji != nil) or clears (emoji == nil) the
// user's reaction on an assistant message in a conversation they own.
// Only assistant messages can be reacted to through this path — a user
// reacting to their own message makes no sense; the persona's own
// reactions to user messages are set directly by ChatService instead,
// never through this caller-facing path.
func (s *ConversationService) SetMessageReaction(ctx context.Context, userID, conversationID, messageID string, emoji *string) (model.Message, error) {
	conversation, err := s.repo.GetByID(ctx, conversationID)
	if err != nil {
		return model.Message{}, err
	}
	if conversation.UserID != userID {
		return model.Message{}, ErrConversationForbidden
	}

	message, err := s.messages.GetByID(ctx, messageID)
	if err != nil {
		return model.Message{}, err
	}
	if message.ConversationID != conversationID {
		return model.Message{}, ErrConversationForbidden
	}
	if message.Role != model.MessageRoleAssistant {
		return model.Message{}, ErrCannotReactToOwnRoleMessage
	}
	if emoji != nil && !model.IsValidReactionEmoji(*emoji) {
		return model.Message{}, ErrInvalidReactionEmoji
	}

	return s.messages.SetReaction(ctx, messageID, emoji)
}
