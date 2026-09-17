package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"perchly-backend/internal/llm"
	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
)

const (
	// summaryMessageThreshold: a conversation needs at least this many
	// messages before it gets a summary, and gets re-summarized every
	// time it accumulates this many new messages after that.
	summaryMessageThreshold = 20
	// summaryRawWindowSize must match ContextBuilder's raw window: the
	// summary only ever covers messages older than the most recent
	// summaryRawWindowSize, since those are always sent verbatim instead.
	summaryRawWindowSize = 6
)

// SummaryService keeps a rolling summary of each conversation's older
// messages, so context sent to the LLM stays bounded as a conversation
// grows instead of resending the entire transcript every turn.
type SummaryService struct {
	messages  MessageRepo
	summaries SummaryRepo
	llmClient llm.Client
}

func NewSummaryService(messages MessageRepo, summaries SummaryRepo, llmClient llm.Client) *SummaryService {
	return &SummaryService{messages: messages, summaries: summaries, llmClient: llmClient}
}

// MaybeSummarize checks whether conversationID has enough new messages to
// be worth summarizing and, if so, does it in the background.
// totalMessageCount is what the caller already knows from just saving a
// message, avoiding an extra COUNT query on every turn.
func (s *SummaryService) MaybeSummarize(conversationID string, totalMessageCount int) {
	if totalMessageCount < summaryMessageThreshold {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		if err := s.summarize(ctx, conversationID); err != nil {
			log.Printf("summarization failed (conversation=%s): %v", conversationID, err)
		}
	}()
}

func (s *SummaryService) summarize(ctx context.Context, conversationID string) error {
	// Re-fetched fresh rather than reusing history captured before this
	// goroutine started: by the time it runs, the request that triggered
	// it has already returned.
	history, err := s.messages.ListByConversation(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("load history: %w", err)
	}

	coverUpTo := len(history) - summaryRawWindowSize
	if coverUpTo <= 0 {
		return nil
	}

	existing, err := s.summaries.GetByConversationID(ctx, conversationID)
	if err != nil && !errors.Is(err, repository.ErrSummaryNotFound) {
		return fmt.Errorf("load existing summary: %w", err)
	}

	if coverUpTo <= existing.CoveredMessageCount {
		return nil // already covers everything we'd summarize now
	}

	newlyOld := history[existing.CoveredMessageCount:coverUpTo]
	if len(newlyOld) == 0 {
		return nil
	}

	summaryText, err := s.generateSummary(ctx, existing.SummaryText, newlyOld)
	if err != nil {
		return fmt.Errorf("generate summary: %w", err)
	}

	return s.summaries.Upsert(ctx, conversationID, summaryText, coverUpTo)
}

const summarizerSystemPrompt = "Bir konuşma özetleyicisisin. Sana bir konuşmanın (bazen önceki bir özetle birlikte) " +
	"bir bölümü verilecek. Önemli bilgileri, kararları, kullanıcı hakkında öğrenilenleri ve duygusal bağlamı koruyarak " +
	"kısa, üçüncü şahıs anlatımıyla güncel bir özet üret. Sadece özeti döndür, başka açıklama ekleme."

func (s *SummaryService) generateSummary(ctx context.Context, previousSummary string, messages []model.Message) (string, error) {
	var userPrompt strings.Builder
	if previousSummary != "" {
		userPrompt.WriteString("Önceki özet:\n" + previousSummary + "\n\n")
	}
	userPrompt.WriteString("Özete eklenecek yeni konuşma parçası:\n" + formatMessages(messages))

	var result strings.Builder
	err := s.llmClient.StreamChat(ctx, []llm.Message{
		{Role: "system", Content: summarizerSystemPrompt},
		{Role: "user", Content: userPrompt.String()},
	}, func(delta string) error {
		result.WriteString(delta)
		return nil
	})
	if err != nil {
		return "", err
	}
	return result.String(), nil
}
