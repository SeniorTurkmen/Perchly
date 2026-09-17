package model

import "time"

// ConversationSummary is the rolling summary of a conversation's older
// messages. CoveredMessageCount records how many of the conversation's
// messages (in creation order) are already folded into SummaryText, so
// the summarizer knows exactly which messages are new since last time.
type ConversationSummary struct {
	ConversationID      string    `json:"conversation_id"`
	SummaryText         string    `json:"summary_text"`
	CoveredMessageCount int       `json:"covered_message_count"`
	UpdatedAt           time.Time `json:"updated_at"`
}
