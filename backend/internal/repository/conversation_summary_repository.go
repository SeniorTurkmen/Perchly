package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
)

// ErrSummaryNotFound is returned when a conversation has no summary yet
// (i.e. it hasn't reached the summarization threshold).
var ErrSummaryNotFound = errors.New("conversation summary not found")

type ConversationSummaryRepository struct {
	pool *pgxpool.Pool
}

func NewConversationSummaryRepository(pool *pgxpool.Pool) *ConversationSummaryRepository {
	return &ConversationSummaryRepository{pool: pool}
}

func (r *ConversationSummaryRepository) GetByConversationID(ctx context.Context, conversationID string) (model.ConversationSummary, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT conversation_id::text, summary_text, covered_message_count, updated_at
		FROM conversation_summaries
		WHERE conversation_id = $1::uuid
	`, conversationID)

	var s model.ConversationSummary
	err := row.Scan(&s.ConversationID, &s.SummaryText, &s.CoveredMessageCount, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ConversationSummary{}, ErrSummaryNotFound
		}
		return model.ConversationSummary{}, err
	}
	return s, nil
}

// Upsert replaces the conversation's summary (there is only ever one
// rolling summary per conversation).
func (r *ConversationSummaryRepository) Upsert(ctx context.Context, conversationID, summaryText string, coveredMessageCount int) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO conversation_summaries (conversation_id, summary_text, covered_message_count, updated_at)
		VALUES ($1::uuid, $2, $3, now())
		ON CONFLICT (conversation_id) DO UPDATE
		SET summary_text = EXCLUDED.summary_text,
		    covered_message_count = EXCLUDED.covered_message_count,
		    updated_at = now()
	`, conversationID, summaryText, coveredMessageCount)
	return err
}
