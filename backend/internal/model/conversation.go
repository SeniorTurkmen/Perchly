package model

import "time"

type Conversation struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	PersonaID string    `json:"persona_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ConversationPreview is a conversation as shown in the inbox: the
// persona it belongs to, plus the most recent message when one exists.
type ConversationPreview struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	PersonaID   string    `json:"persona_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Persona     Persona   `json:"persona"`
	LastMessage *Message  `json:"last_message"`
}
