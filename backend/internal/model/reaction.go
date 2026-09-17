package model

// OrderedAllowedReactionEmojis is the fixed, iMessage-tapback-style set of
// emoji reactions the product supports, in display order. Both the
// server-side validator below and the persona's system prompt (see
// service.reactionProtocolInstruction) derive from this single list so
// the two can never drift apart.
var OrderedAllowedReactionEmojis = []string{"❤️", "😂", "👍", "👎", "‼️", "❓"}

var allowedReactionEmojiSet = buildAllowedReactionEmojiSet()

func buildAllowedReactionEmojiSet() map[string]bool {
	set := make(map[string]bool, len(OrderedAllowedReactionEmojis))
	for _, emoji := range OrderedAllowedReactionEmojis {
		set[emoji] = true
	}
	return set
}

// IsValidReactionEmoji reports whether emoji is one of the fixed set of
// reactions the product supports. Reactions are fail-closed: anything
// outside this set — whether client-supplied or produced by a persona's
// reaction tag (see ChatService.SendMessage) — is rejected rather than
// stored as free-form text.
func IsValidReactionEmoji(emoji string) bool {
	return allowedReactionEmojiSet[emoji]
}
