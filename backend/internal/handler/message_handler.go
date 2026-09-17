package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"perchly-backend/internal/auth"
)

type chatService interface {
	SendMessage(
		ctx context.Context,
		conversationID, personaID, userID string,
		useCredit bool,
		content string,
		onDelta func(string) error,
		onUserMessageReaction func(emoji string) error,
	) (userMessageID, assistantMessageID string, err error)
}

// MessageHandler trusts that QuotaMiddleware already resolved and
// authorized the conversation, and decided whether this turn spends a
// credit — see ConversationFromContext / UseCreditFromContext.
type MessageHandler struct {
	chat chatService
}

func NewMessageHandler(chat chatService) *MessageHandler {
	return &MessageHandler{chat: chat}
}

type createMessageRequest struct {
	Content string `json:"content"`
}

// doneEventPayload is the data carried by the final "done" SSE event —
// the real, persisted ids of this turn's messages, so the client can
// reconcile its locally-generated placeholder ids (see ChatViewModel)
// with the backend ids it needs for anything id-addressed, like setting
// a reaction. AssistantMessageID is empty when the persona reacted to
// the user's message only, with no written reply (see ChatService).
type doneEventPayload struct {
	UserMessageID      string `json:"user_message_id"`
	AssistantMessageID string `json:"assistant_message_id,omitempty"`
}

// Create handles POST /conversations/{id}/messages. On success it
// switches the response to Server-Sent Events and streams the assistant's
// reply back as one "message" event per chunk as it arrives from the LLM,
// followed by a final "done" event carrying doneEventPayload — or an
// "error" event if the stream fails partway through. If the persona
// reacted to the user's message (see ChatService.SendMessage), a single
// "reaction" event carrying the emoji is sent before "done" — with no
// "message" events at all if the persona reacted only.
func (h *MessageHandler) Create(w http.ResponseWriter, r *http.Request) {
	conversation, ok := ConversationFromContext(r.Context())
	if !ok {
		// QuotaMiddleware guarantees this on every route that mounts this
		// handler; a miss means the route is wired up wrong, not a client error.
		writeError(w, http.StatusInternalServerError, "konuşma bağlamı bulunamadı")
		return
	}
	userID, _ := auth.UserIDFromContext(r.Context())
	useCredit := UseCreditFromContext(r.Context())

	var req createMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Content) == "" {
		writeError(w, http.StatusBadRequest, "content alanı zorunludur")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming desteklenmiyor")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	sendEvent := func(event string, data any) {
		encoded, _ := json.Marshal(data)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, encoded)
		flusher.Flush()
	}

	ctx := r.Context()
	userMessageID, assistantMessageID, err := h.chat.SendMessage(ctx, conversation.ID, conversation.PersonaID, userID, useCredit, req.Content, func(delta string) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		sendEvent("message", delta)
		return nil
	}, func(emoji string) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		sendEvent("reaction", emoji)
		return nil
	})
	if err != nil {
		log.Printf("chat send message error (conversation=%s): %v", conversation.ID, err)
		sendEvent("error", "yanıt alınamadı, lütfen tekrar deneyin")
		return
	}

	sendEvent("done", doneEventPayload{UserMessageID: userMessageID, AssistantMessageID: assistantMessageID})
}
