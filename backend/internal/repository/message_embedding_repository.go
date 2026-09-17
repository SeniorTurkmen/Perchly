package repository

import (
	"context"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"

	"perchly-backend/internal/model"
)

type MessageEmbeddingRepository struct {
	pool *pgxpool.Pool
}

func NewMessageEmbeddingRepository(pool *pgxpool.Pool) *MessageEmbeddingRepository {
	return &MessageEmbeddingRepository{pool: pool}
}

// Create stores (or replaces) the embedding for a message.
func (r *MessageEmbeddingRepository) Create(ctx context.Context, messageID, conversationID string, vector []float32) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO message_embeddings (message_id, conversation_id, embedding)
		VALUES ($1::uuid, $2::uuid, $3)
		ON CONFLICT (message_id) DO UPDATE SET embedding = EXCLUDED.embedding
	`, messageID, conversationID, pgvector.NewVector(vector))
	return err
}

// NearestByConversation returns up to limit messages from conversationID
// whose embeddings are closest to queryEmbedding by cosine distance,
// excluding any id in excludeMessageIDs (e.g. messages already shown
// verbatim elsewhere in the prompt). Results come back sorted
// chronologically, not by similarity rank, so they read naturally as
// context.
func (r *MessageEmbeddingRepository) NearestByConversation(
	ctx context.Context,
	conversationID string,
	queryEmbedding []float32,
	limit int,
	excludeMessageIDs []string,
) ([]model.Message, error) {
	if excludeMessageIDs == nil {
		// A nil slice would bind as SQL NULL, and `id = ANY(NULL)` is NULL
		// (not false) in every row, which would wrongly exclude everything.
		excludeMessageIDs = []string{}
	}

	rows, err := r.pool.Query(ctx, `
		SELECT m.id::text, m.conversation_id::text, m.role, m.content, m.created_at
		FROM message_embeddings me
		JOIN messages m ON m.id = me.message_id
		WHERE me.conversation_id = $1::uuid
		  AND NOT (m.id = ANY($2::uuid[]))
		ORDER BY me.embedding <=> $3
		LIMIT $4
	`, conversationID, excludeMessageIDs, pgvector.NewVector(queryEmbedding), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]model.Message, 0, limit)
	for rows.Next() {
		m, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt.Before(messages[j].CreatedAt)
	})

	return messages, nil
}
