package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"perchly-backend/internal/embedding"
	"perchly-backend/internal/llm"
	"perchly-backend/internal/model"
)

const (
	contextNearestNeighborCount = 5
	contextRawWindowSize        = 6
)

// MessageEmbeddingSearcher is the dependency ContextBuilder needs to find
// past messages similar to the current one.
type MessageEmbeddingSearcher interface {
	NearestByConversation(ctx context.Context, conversationID string, queryEmbedding []float32, limit int, excludeMessageIDs []string) ([]model.Message, error)
}

// SummaryRepo is the persistence dependency shared by ContextBuilder
// (reads the rolling summary) and SummaryService (writes it).
type SummaryRepo interface {
	GetByConversationID(ctx context.Context, conversationID string) (model.ConversationSummary, error)
	Upsert(ctx context.Context, conversationID, summaryText string, coveredMessageCount int) error
}

// ContextBuilder assembles the message list sent to the LLM for one chat
// turn, in a fixed order:
//  1. the persona's system prompt
//  2. the conversation's rolling summary, if it has one
//  3. the nearest contextNearestNeighborCount historical messages to the
//     new message, by embedding similarity (via pgvector), excluding
//     anything already in the raw window below
//  4. the last contextRawWindowSize messages, verbatim
//
// Retrieval is best-effort: if the embedding provider or the similarity
// search fails, step 3 is just skipped rather than failing the whole
// chat turn — a slightly less informed reply beats no reply.
type ContextBuilder struct {
	summaries         SummaryRepo
	messageEmbeddings MessageEmbeddingSearcher
	embeddingClient   embedding.Client
}

func NewContextBuilder(summaries SummaryRepo, messageEmbeddings MessageEmbeddingSearcher, embeddingClient embedding.Client) *ContextBuilder {
	return &ContextBuilder{summaries: summaries, messageEmbeddings: messageEmbeddings, embeddingClient: embeddingClient}
}

// reactionProtocolInstruction teaches the persona the leading-tag
// protocol ChatService.SendMessage parses (see tryParseReactionTag) for
// reacting to the user's message with a single emoji instead of, or
// before, a written reply. Built from model.OrderedAllowedReactionEmojis
// so the allowed set here and the server-side validator can never drift
// apart.
var reactionProtocolInstruction = fmt.Sprintf(
	"Bazen kullanıcının son mesajına sözcüklerle değil, sadece tek bir emojiyle tepki vermek daha doğal olur "+
		"- gerçek bir arkadaşın mesaja sözle değil sadece emojiyle karşılık vermesi gibi. Bunu yapmak istediğinde, "+
		"yanıtının EN BAŞINA, başka hiçbir şey yazmadan, şu formatta bir etiket koy: [[REACT:<emoji>]] - örneğin "+
		"[[REACT:%s]]. Bu etiketten sonra ayrıca yazılı bir yanıt da eklemek istiyorsan ekleyebilirsin; sadece "+
		"emojiyle yetinmek istiyorsan etiketten sonra hiçbir şey yazma. Bu etiketi SADECE şu emojilerden biriyle "+
		"kullan: %s. Bu etiketi her mesajda kullanma; yalnızca gerçekten sadece bir tepkinin yeterli ve doğal "+
		"olduğu anlarda kullan.",
	model.OrderedAllowedReactionEmojis[0],
	strings.Join(model.OrderedAllowedReactionEmojis, " "),
)

func (b *ContextBuilder) Build(ctx context.Context, conversationID, systemPrompt string, history []model.Message, newMessageContent string, traits model.PersonaTraits) []llm.Message {
	messages := make([]llm.Message, 0, contextRawWindowSize+5)
	messages = append(messages, llm.Message{Role: "system", Content: systemPrompt})
	messages = append(messages, llm.Message{Role: "system", Content: reactionProtocolInstruction})
	messages = append(messages, llm.Message{Role: "system", Content: formatTraitsInstruction(traits)})

	if summary, err := b.summaries.GetByConversationID(ctx, conversationID); err == nil && summary.SummaryText != "" {
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: "Önceki konuşmanın özeti:\n" + summary.SummaryText,
		})
	}

	rawWindow := lastN(history, contextRawWindowSize)
	rawWindowIDs := make([]string, len(rawWindow))
	for i, m := range rawWindow {
		rawWindowIDs[i] = m.ID
	}

	if nearest := b.findNearest(ctx, conversationID, newMessageContent, rawWindowIDs); len(nearest) > 0 {
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: "Bu konuşmadan, şu anki mesajla ilgili olabilecek geçmiş anlar:\n" + formatMessages(nearest),
		})
	}

	for _, m := range rawWindow {
		messages = append(messages, llm.Message{Role: string(m.Role), Content: m.Content})
	}

	return messages
}

// traitPhraseTiers maps a 0-100 dial value into one of 5 descriptive
// phrases, used to turn model.PersonaTraits into prompt guidance an LLM
// follows more reliably than a bare number would.
var (
	warmthPhraseTiers = [5]string{
		"çok soğuk ve mesafeli",
		"mesafeli, resmi bir ton",
		"orta düzeyde sıcak",
		"sıcak ve şefkatli",
		"son derece sıcak, şefkatli ve cana yakın",
	}
	humorPhraseTiers = [5]string{
		"tamamen ciddi, hiç şaka yapma",
		"nadiren ve hafif espri yapan",
		"ara sıra espri katan",
		"sık sık esprili",
		"oldukça şakacı ve eğlenceli",
	}
	wisdomPhraseTiers = [5]string{
		"yüzeysel, basit yanıtlar veren",
		"temel düzeyde düşünceli",
		"orta düzeyde derinlikli",
		"derin ve olgun",
		"çok derin, felsefi ve olgun",
	}
	directnessPhraseTiers = [5]string{
		"son derece yumuşak ve dolaylı, her şeyi nazikçe yumuşatan",
		"nazik ve yumuşatarak konuşan",
		"orta düzeyde doğrudan",
		"doğrudan ve net",
		"hiç yumuşatmadan sert ve doğrudan konuşan",
	}
	energyPhraseTiers = [5]string{
		"çok sakin ve durgun",
		"sakin",
		"orta tempolu",
		"enerjik",
		"son derece enerjik ve coşkulu",
	}
)

// traitTier buckets a 0-100 dial value into one of 5 tiers (0-4).
func traitTier(value int) int {
	switch {
	case value < 20:
		return 0
	case value < 40:
		return 1
	case value < 60:
		return 2
	case value < 80:
		return 3
	default:
		return 4
	}
}

// formatTraitsInstruction turns the user's current dial positions for
// this persona into a system-prompt instruction. Deliberately phrased
// as fine-tuning the persona's existing character, not replacing it —
// the persona's own system prompt (its identity, boundaries, platonic
// framing) always takes precedence.
func formatTraitsInstruction(t model.PersonaTraits) string {
	return fmt.Sprintf(
		"Şu anda kullanıcının senin için ayarladığı kişilik ayarlarıyla konuş (kullanıcı bunları istediği "+
			"zaman değiştirebilir, kimliğini değiştirmez, sadece tonunu ince ayarlar): "+
			"Sıcaklık %%%d (%s), Espri %%%d (%s), Bilgelik %%%d (%s), Doğrudanlık %%%d (%s), Enerji %%%d (%s).",
		t.Warmth, warmthPhraseTiers[traitTier(t.Warmth)],
		t.Humor, humorPhraseTiers[traitTier(t.Humor)],
		t.Wisdom, wisdomPhraseTiers[traitTier(t.Wisdom)],
		t.Directness, directnessPhraseTiers[traitTier(t.Directness)],
		t.Energy, energyPhraseTiers[traitTier(t.Energy)],
	)
}

func (b *ContextBuilder) findNearest(ctx context.Context, conversationID, queryText string, excludeIDs []string) []model.Message {
	queryEmbedding, err := b.embeddingClient.Embed(ctx, queryText)
	if err != nil {
		log.Printf("context: query embedding failed, skipping nearest-neighbor retrieval: %v", err)
		return nil
	}

	nearest, err := b.messageEmbeddings.NearestByConversation(ctx, conversationID, queryEmbedding, contextNearestNeighborCount, excludeIDs)
	if err != nil {
		log.Printf("context: nearest-neighbor lookup failed: %v", err)
		return nil
	}
	return nearest
}

func lastN(messages []model.Message, n int) []model.Message {
	if len(messages) <= n {
		return messages
	}
	return messages[len(messages)-n:]
}

func formatMessages(messages []model.Message) string {
	var sb strings.Builder
	for _, m := range messages {
		fmt.Fprintf(&sb, "%s: %s\n", roleLabel(m.Role), m.Content)
	}
	return sb.String()
}

func roleLabel(role model.MessageRole) string {
	if role == model.MessageRoleUser {
		return "Kullanıcı"
	}
	return "Sen"
}
