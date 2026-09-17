package model

import "time"

type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
)

type Message struct {
	ID             string      `json:"id"`
	ConversationID string      `json:"conversation_id"`
	Role           MessageRole `json:"role"`
	Content        string      `json:"content"`
	// ReactionEmoji is the single emoji reaction left on this message by
	// whichever side didn't send it (the user reacting to an assistant
	// message, or the persona reacting to a user message — see
	// ChatService.SendMessage), or nil if none. Always one of
	// OrderedAllowedReactionEmojis.
	ReactionEmoji *string   `json:"reaction_emoji,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
