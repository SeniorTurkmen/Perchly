package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"perchly-backend/internal/llm"
	"perchly-backend/internal/model"
)

type MessageRepo interface {
	Create(ctx context.Context, conversationID string, role model.MessageRole, content string) (model.Message, error)
	ListByConversation(ctx context.Context, conversationID string) ([]model.Message, error)
	SetReaction(ctx context.Context, messageID string, emoji *string) (model.Message, error)
}

// reactionTagOpen/Close delimit the tag a persona's reply can lead with to
// react to the user's message with a single emoji instead of (or before)
// replying with text — see scanForReactionTag and the prompt instruction
// built in context_builder.go.
const (
	reactionTagOpen  = "[[REACT:"
	reactionTagClose = "]]"
	// Generous upper bound on how long to keep waiting for a close
	// bracket once an open tag has appeared before giving up on it.
	// Deliberately not limited to checking position 0: real models
	// don't reliably put the tag at the exact start of the reply —
	// observed in practice preceded by a stray preamble — so the whole
	// pending tail is searched, not just its first bytes. "[[REACT:" (8)
	// + a multi-codepoint emoji (up to ~8 bytes) + "]]" (2) leaves ~46
	// bytes of room for such a preamble.
	maxReactionTagScanLen = 64
	// How much of the pending tail must be held back, even when no open
	// tag has appeared yet, because it could still be the unfinished
	// start of one arriving in a later delta — see scanForReactionTag.
	reactionTagTailSafetyLen = len(reactionTagOpen) - 1
)

// reactionScanResult is one step of scanForReactionTag: flush is safe to
// send to the client immediately (never part of a tag, whatever comes
// next), pending must be held and re-scanned once more of the reply
// arrives. Once resolved is true, scanning is over for this reply —
// flush (and, on a later call needed only to drain a final empty
// pending) is everything left to show, and emoji (if non-empty) is the
// reaction found.
type reactionScanResult struct {
	flush    string
	pending  string
	resolved bool
	emoji    string
}

// scanForReactionTag looks for a reactionTagOpen...Close tag in buffer —
// the pending tail of the reply accumulated so far that scanning hasn't
// yet resolved one way or the other. It never holds back more than it
// has to: everything provably outside of, or past, a possible tag is
// flushed immediately so a reply with no tag at all streams to the
// client exactly as it arrives, not in one lump at the end.
//
// A tag naming an emoji outside model.IsValidReactionEmoji is stripped
// but otherwise ignored (fail-closed) rather than either leaking the raw
// tag into the visible reply or rejecting the whole turn.
func scanForReactionTag(buffer string) reactionScanResult {
	if openIdx := strings.Index(buffer, reactionTagOpen); openIdx != -1 {
		if closeIdx := strings.Index(buffer[openIdx:], reactionTagClose); closeIdx != -1 {
			closeIdx += openIdx
			candidate := buffer[openIdx+len(reactionTagOpen) : closeIdx]
			before := buffer[:openIdx]
			after := buffer[closeIdx+len(reactionTagClose):]

			emoji := ""
			if model.IsValidReactionEmoji(candidate) {
				emoji = candidate
			}
			return reactionScanResult{flush: before + after, resolved: true, emoji: emoji}
		}

		// Open tag with no close yet. Everything before it is definitely
		// not part of a tag and can stream immediately; the rest waits
		// for its close bracket, bounded by maxReactionTagScanLen.
		if len(buffer) > maxReactionTagScanLen {
			return reactionScanResult{flush: buffer, resolved: true}
		}
		return reactionScanResult{flush: buffer[:openIdx], pending: buffer[openIdx:]}
	}

	// No open tag anywhere yet. Safe to flush all of buffer except a
	// small trailing window that could still turn into "[[REACT:" once
	// more arrives — anything further back is already a confirmed
	// mismatch (those bytes are fixed; they'll never match). Sliced by
	// rune, not byte, so a flush never splits a multi-byte character.
	runes := []rune(buffer)
	if len(runes) > reactionTagTailSafetyLen {
		safeLen := len(runes) - reactionTagTailSafetyLen
		return reactionScanResult{flush: string(runes[:safeLen]), pending: string(runes[safeLen:])}
	}
	return reactionScanResult{pending: buffer}
}

// ChatService orchestrates a single chat turn: persist the user's
// message, build context (persona prompt + rolling summary + relevant
// past messages + recent raw window), ask the configured LLM for a
// reply, stream it back through onDelta, then persist the full assistant
// reply. Embedding generation and rolling-summary maintenance are kicked
// off in the background and never block the reply.
//
// The persona's reply may lead with a reaction tag (see
// tryParseReactionTag) reacting to the user's just-sent message instead
// of, or before, replying with text; when it does, onUserMessageReaction
// is called with the emoji and, if the persona reacted only (no text
// left after stripping the tag), no assistant message is persisted at
// all for this turn.
//
// Conversation resolution, ownership, and quota checks all happen
// upstream (see handler.QuotaMiddleware) — SendMessage trusts the
// conversationID/personaID/userID it's given.
type ChatService struct {
	messages       MessageRepo
	personas       PersonaRepo
	personaTraits  PersonaTraitsRepo
	llmClient      llm.Client
	embeddings     *EmbeddingService
	summaries      *SummaryService
	contextBuilder *ContextBuilder
	quotas         *QuotaService
}

func NewChatService(
	messages MessageRepo,
	personas PersonaRepo,
	personaTraits PersonaTraitsRepo,
	llmClient llm.Client,
	embeddings *EmbeddingService,
	summaries *SummaryService,
	contextBuilder *ContextBuilder,
	quotas *QuotaService,
) *ChatService {
	return &ChatService{
		messages:       messages,
		personas:       personas,
		personaTraits:  personaTraits,
		llmClient:      llmClient,
		embeddings:     embeddings,
		summaries:      summaries,
		contextBuilder: contextBuilder,
		quotas:         quotas,
	}
}

// SendMessage answers one chat turn in conversationID (owned by userID,
// with persona personaID). useCredit says whether this turn should spend
// a credit instead of counting against the free daily quota, as decided
// by QuotaService.Check before this was called. onUserMessageReaction is
// called at most once, iff the persona's reply led with a reaction tag —
// see the ChatService doc comment.
//
// On success, returns the real, persisted ids of this turn's user
// message and (if one was created — see the ChatService doc comment)
// assistant message, so the caller can hand them back to a client whose
// own copies of these messages only have locally-generated placeholder
// ids so far.
func (s *ChatService) SendMessage(
	ctx context.Context,
	conversationID, personaID, userID string,
	useCredit bool,
	content string,
	onDelta func(string) error,
	onUserMessageReaction func(emoji string) error,
) (userMessageID, assistantMessageID string, err error) {
	persona, err := s.personas.GetByID(ctx, personaID)
	if err != nil {
		return "", "", fmt.Errorf("load persona: %w", err)
	}

	userMessage, err := s.messages.Create(ctx, conversationID, model.MessageRoleUser, content)
	if err != nil {
		return "", "", fmt.Errorf("save user message: %w", err)
	}
	s.embeddings.EmbedMessageAsync(conversationID, userMessage.ID, userMessage.Content)

	history, err := s.messages.ListByConversation(ctx, conversationID)
	if err != nil {
		return "", "", fmt.Errorf("load history: %w", err)
	}

	traits, _, err := s.personaTraits.GetEffective(ctx, userID, personaID)
	if err != nil {
		// A trait-lookup glitch shouldn't sink the whole turn — fall
		// back to the persona's own defaults, already in hand.
		log.Printf("chat: failed to load persona traits (user=%s persona=%s): %v, using persona defaults", userID, personaID, err)
		traits = persona.DefaultTraits
	}

	llmMessages := s.contextBuilder.Build(ctx, conversationID, persona.SystemPrompt, history, content, traits)

	var full strings.Builder
	var reactionEmoji string
	var reactionTagResolved bool
	var pending strings.Builder

	if err := s.llmClient.StreamChat(ctx, llmMessages, func(delta string) error {
		if reactionTagResolved {
			full.WriteString(delta)
			return onDelta(delta)
		}

		pending.WriteString(delta)
		result := scanForReactionTag(pending.String())
		pending.Reset()
		pending.WriteString(result.pending)
		if result.resolved {
			reactionTagResolved = true
			reactionEmoji = result.emoji
		}
		if result.flush == "" {
			return nil
		}
		full.WriteString(result.flush)
		return onDelta(result.flush)
	}); err != nil {
		return "", "", err
	}

	// The stream ended before scanning ever reached a verdict (e.g. a
	// very short reply, shorter than reactionTagTailSafetyLen) — whatever
	// is left in pending is just plain text that was never flushed.
	if !reactionTagResolved && pending.Len() > 0 {
		remainder := pending.String()
		full.WriteString(remainder)
		if err := onDelta(remainder); err != nil {
			return "", "", err
		}
	}

	if reactionEmoji != "" {
		if _, err := s.messages.SetReaction(ctx, userMessage.ID, &reactionEmoji); err != nil {
			log.Printf("chat: failed to save persona reaction (message=%s): %v", userMessage.ID, err)
		} else if err := onUserMessageReaction(reactionEmoji); err != nil {
			return "", "", err
		}
	}

	// len(history) already includes the user message we just added.
	turnMessageCount := len(history)
	if full.Len() > 0 {
		assistantMessage, err := s.messages.Create(ctx, conversationID, model.MessageRoleAssistant, full.String())
		if err != nil {
			return "", "", fmt.Errorf("save assistant message: %w", err)
		}
		s.embeddings.EmbedMessageAsync(conversationID, assistantMessage.ID, assistantMessage.Content)
		assistantMessageID = assistantMessage.ID
		turnMessageCount++
	}
	s.summaries.MaybeSummarize(conversationID, turnMessageCount)

	// The reply already succeeded and streamed to the client by this
	// point — a bookkeeping failure here shouldn't turn into an error
	// response the client can't do anything useful with.
	if err := s.quotas.RecordUsage(ctx, userID, personaID, useCredit); err != nil {
		log.Printf("quota record usage failed (user=%s persona=%s): %v", userID, personaID, err)
	}

	return userMessage.ID, assistantMessageID, nil
}
