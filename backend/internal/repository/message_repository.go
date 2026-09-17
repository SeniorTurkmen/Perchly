package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
)

// ErrMessageNotFound is returned when no message matches the given id.
var ErrMessageNotFound = errors.New("message not found")

const messageColumns = `id::text, conversation_id::text, role, content, reaction_emoji, created_at`

type MessageRepository struct {
	pool *pgxpool.Pool
}

func NewMessageRepository(pool *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{pool: pool}
}

// Create persists a single message in a conversation.
func (r *MessageRepository) Create(ctx context.Context, conversationID string, role model.MessageRole, content string) (model.Message, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO messages (conversation_id, role, content)
		VALUES ($1::uuid, $2, $3)
		RETURNING `+messageColumns, conversationID, string(role), content)

	return scanMessage(row)
}

// GetByID returns a single message by id.
func (r *MessageRepository) GetByID(ctx context.Context, id string) (model.Message, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+messageColumns+`
		FROM messages
		WHERE id = $1::uuid
	`, id)

	m, err := scanMessage(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Message{}, ErrMessageNotFound
		}
		return model.Message{}, err
	}
	return m, nil
}

// SetReaction sets (or, with a nil emoji, clears) the single reaction on
// message id. The caller is responsible for validating emoji against
// model.IsValidReactionEmoji and for any ownership/role checks — this is
// a plain, trusted write.
func (r *MessageRepository) SetReaction(ctx context.Context, id string, emoji *string) (model.Message, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE messages
		SET reaction_emoji = $2
		WHERE id = $1::uuid
		RETURNING `+messageColumns, id, emoji)

	m, err := scanMessage(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Message{}, ErrMessageNotFound
		}
		return model.Message{}, err
	}
	return m, nil
}

// ListByConversation returns every message in a conversation, oldest first.
func (r *MessageRepository) ListByConversation(ctx context.Context, conversationID string) ([]model.Message, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+messageColumns+`
		FROM messages
		WHERE conversation_id = $1::uuid
		ORDER BY created_at ASC
	`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]model.Message, 0)
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
	return messages, nil
}

func scanMessage(row rowScanner) (model.Message, error) {
	var m model.Message
	var role string
	err := row.Scan(&m.ID, &m.ConversationID, &role, &m.Content, &m.ReactionEmoji, &m.CreatedAt)
	m.Role = model.MessageRole(role)
	return m, err
}
