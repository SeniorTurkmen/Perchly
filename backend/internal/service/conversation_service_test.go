package service

import (
	"context"
	"errors"
	"testing"

	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
)

type fakeConversationRepo struct{}

func (fakeConversationRepo) Create(_ context.Context, userID, personaID string) (model.Conversation, error) {
	return model.Conversation{ID: "conv-1", UserID: userID, PersonaID: personaID}, nil
}

func (fakeConversationRepo) GetByID(context.Context, string) (model.Conversation, error) {
	return model.Conversation{}, nil
}

func (fakeConversationRepo) ListByUserID(context.Context, string) ([]model.ConversationPreview, error) {
	return nil, nil
}

type fakeMessageLister struct {
	messages map[string]model.Message
}

func (fakeMessageLister) ListByConversation(context.Context, string) ([]model.Message, error) {
	return nil, nil
}

func (l fakeMessageLister) GetByID(_ context.Context, id string) (model.Message, error) {
	m, ok := l.messages[id]
	if !ok {
		return model.Message{}, repository.ErrMessageNotFound
	}
	return m, nil
}

func (l fakeMessageLister) SetReaction(_ context.Context, id string, emoji *string) (model.Message, error) {
	m, ok := l.messages[id]
	if !ok {
		return model.Message{}, repository.ErrMessageNotFound
	}
	m.ReactionEmoji = emoji
	l.messages[id] = m
	return m, nil
}

type fakePersonaRepoForConversation struct {
	personas map[string]model.Persona
}

func (r fakePersonaRepoForConversation) List(context.Context) ([]model.Persona, error) {
	return nil, nil
}

func (r fakePersonaRepoForConversation) GetByID(_ context.Context, id string) (model.Persona, error) {
	p, ok := r.personas[id]
	if !ok {
		return model.Persona{}, repository.ErrPersonaNotFound
	}
	return p, nil
}

type fakeOnboardingProfileRepoForConversation struct {
	profiles map[string]model.OnboardingProfile
}

func (r fakeOnboardingProfileRepoForConversation) Upsert(_ context.Context, p model.OnboardingProfile) (model.OnboardingProfile, error) {
	r.profiles[p.UserID] = p
	return p, nil
}

func (r fakeOnboardingProfileRepoForConversation) GetByUserID(_ context.Context, userID string) (model.OnboardingProfile, error) {
	p, ok := r.profiles[userID]
	if !ok {
		return model.OnboardingProfile{}, repository.ErrOnboardingProfileNotFound
	}
	return p, nil
}

func (r fakeOnboardingProfileRepoForConversation) Exists(_ context.Context, userID string) (bool, error) {
	_, ok := r.profiles[userID]
	return ok, nil
}

type fakeConversationRepoForReaction struct {
	conversation model.Conversation
}

func (r fakeConversationRepoForReaction) Create(context.Context, string, string) (model.Conversation, error) {
	return model.Conversation{}, nil
}

func (r fakeConversationRepoForReaction) GetByID(context.Context, string) (model.Conversation, error) {
	return r.conversation, nil
}

func (r fakeConversationRepoForReaction) ListByUserID(context.Context, string) ([]model.ConversationPreview, error) {
	return nil, nil
}

func TestConversationService_SetMessageReaction(t *testing.T) {
	const owner = "user-owner"
	const other = "user-other"
	const conversationID = "conv-1"
	const assistantMsgID = "msg-assistant"
	const userMsgID = "msg-user"
	const otherConvoMsgID = "msg-other-convo"

	newSvc := func() *ConversationService {
		messages := fakeMessageLister{messages: map[string]model.Message{
			assistantMsgID:  {ID: assistantMsgID, ConversationID: conversationID, Role: model.MessageRoleAssistant},
			userMsgID:       {ID: userMsgID, ConversationID: conversationID, Role: model.MessageRoleUser},
			otherConvoMsgID: {ID: otherConvoMsgID, ConversationID: "conv-2", Role: model.MessageRoleAssistant},
		}}
		conversations := fakeConversationRepoForReaction{conversation: model.Conversation{ID: conversationID, UserID: owner}}
		return NewConversationService(conversations, fakePersonaRepoForConversation{}, fakeOnboardingProfileRepoForConversation{profiles: map[string]model.OnboardingProfile{}}, messages)
	}

	heart := "❤️"
	bogus := "🐸"

	tests := []struct {
		name      string
		userID    string
		messageID string
		emoji     *string
		wantErr   error
	}{
		{"user reacts to an assistant message", owner, assistantMsgID, &heart, nil},
		{"clearing a reaction succeeds the same way", owner, assistantMsgID, nil, nil},
		{"conversation not owned by caller is forbidden", other, assistantMsgID, &heart, ErrConversationForbidden},
		{"reacting to your own side's message is rejected", owner, userMsgID, &heart, ErrCannotReactToOwnRoleMessage},
		{"an emoji outside the fixed set is rejected", owner, assistantMsgID, &bogus, ErrInvalidReactionEmoji},
		{"a message from a different conversation is forbidden", owner, otherConvoMsgID, &heart, ErrConversationForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newSvc()
			_, err := svc.SetMessageReaction(context.Background(), tt.userID, conversationID, tt.messageID, tt.emoji)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("SetMessageReaction() error = %v, want nil", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("SetMessageReaction() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestConversationService_Create_AgeGating(t *testing.T) {
	restrictedPersona := model.Persona{ID: "persona-restricted", IsMinorAppropriate: false}
	openPersona := model.Persona{ID: "persona-open", IsMinorAppropriate: true}

	personas := fakePersonaRepoForConversation{personas: map[string]model.Persona{
		restrictedPersona.ID: restrictedPersona,
		openPersona.ID:       openPersona,
	}}

	tests := []struct {
		name      string
		personaID string
		profile   *model.OnboardingProfile // nil = no onboarding profile on file
		wantErr   error
	}{
		{
			name:      "minor blocked from a non-minor-appropriate persona",
			personaID: restrictedPersona.ID,
			profile:   &model.OnboardingProfile{UserID: "user-minor", IsMinor: true},
			wantErr:   ErrPersonaNotAgeAppropriate,
		},
		{
			name:      "adult allowed on a non-minor-appropriate persona",
			personaID: restrictedPersona.ID,
			profile:   &model.OnboardingProfile{UserID: "user-adult", IsMinor: false},
			wantErr:   nil,
		},
		{
			name:      "unknown age (no onboarding profile) is permissive",
			personaID: restrictedPersona.ID,
			profile:   nil,
			wantErr:   nil,
		},
		{
			name:      "minor allowed on a minor-appropriate persona",
			personaID: openPersona.ID,
			profile:   &model.OnboardingProfile{UserID: "user-minor-2", IsMinor: true},
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profiles := fakeOnboardingProfileRepoForConversation{profiles: map[string]model.OnboardingProfile{}}
			userID := "anonymous-user"
			if tt.profile != nil {
				userID = tt.profile.UserID
				profiles.profiles[userID] = *tt.profile
			}

			svc := NewConversationService(fakeConversationRepo{}, personas, profiles, fakeMessageLister{})
			_, err := svc.Create(context.Background(), userID, tt.personaID)

			if tt.wantErr == nil && err != nil {
				t.Fatalf("Create() error = %v, want nil", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Create() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
