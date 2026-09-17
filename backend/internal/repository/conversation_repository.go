package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
)

// ErrConversationNotFound is returned when no conversation matches the
// given id.
var ErrConversationNotFound = errors.New("conversation not found")

type ConversationRepository struct {
	pool *pgxpool.Pool
}

func NewConversationRepository(pool *pgxpool.Pool) *ConversationRepository {
	return &ConversationRepository{pool: pool}
}

const conversationColumns = `id::text, user_id::text, persona_id::text, created_at, updated_at`

// Create starts a new conversation for userID with the given persona.
// Returns ErrPersonaNotFound if personaID doesn't reference an existing
// persona.
func (r *ConversationRepository) Create(ctx context.Context, userID, personaID string) (model.Conversation, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO conversations (user_id, persona_id)
		VALUES ($1::uuid, $2::uuid)
		RETURNING `+conversationColumns, userID, personaID)

	c, err := scanConversation(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign_key_violation
			return model.Conversation{}, ErrPersonaNotFound
		}
		return model.Conversation{}, err
	}
	return c, nil
}

// GetByID returns a single conversation by id. id must already be a
// well-formed UUID string.
func (r *ConversationRepository) GetByID(ctx context.Context, id string) (model.Conversation, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+conversationColumns+`
		FROM conversations
		WHERE id = $1::uuid
	`, id)

	c, err := scanConversation(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Conversation{}, ErrConversationNotFound
		}
		return model.Conversation{}, err
	}
	return c, nil
}

func scanConversation(row rowScanner) (model.Conversation, error) {
	var c model.Conversation
	err := row.Scan(&c.ID, &c.UserID, &c.PersonaID, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

// ListByUserID returns the caller's conversations, newest activity first,
// each joined with its persona and most recent message.
func (r *ConversationRepository) ListByUserID(ctx context.Context, userID string) ([]model.ConversationPreview, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			c.id::text, c.user_id::text, c.persona_id::text, c.created_at, c.updated_at,
			`+previewPersonaColumns+`,
			m.id::text, m.conversation_id::text, m.role, m.content, m.reaction_emoji, m.created_at
		FROM conversations c
		JOIN personas p ON p.id = c.persona_id
		LEFT JOIN LATERAL (
			SELECT id, conversation_id, role, content, reaction_emoji, created_at
			FROM messages
			WHERE conversation_id = c.id
			ORDER BY created_at DESC
			LIMIT 1
		) m ON true
		WHERE c.user_id = $1::uuid
			AND m.id IS NOT NULL
		ORDER BY m.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	previews := make([]model.ConversationPreview, 0)
	for rows.Next() {
		preview, err := scanConversationPreview(rows)
		if err != nil {
			return nil, err
		}
		previews = append(previews, preview)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return previews, nil
}

// Aliased so the join's persona columns don't clash with c.id etc.
const previewPersonaColumns = `
	p.id::text, p.slug, p.name, p.category, p.short_description, p.system_prompt,
	p.tone_description, p.avatar_url, p.accent_color, p.is_minor_appropriate, p.is_active, p.sort_order,
	p.created_at, p.updated_at`

func scanConversationPreview(row rowScanner) (model.ConversationPreview, error) {
	var preview model.ConversationPreview
	var (
		lastID            *string
		lastConvID        *string
		lastRole          *string
		lastContent       *string
		lastReactionEmoji *string
		lastCreatedAt     *time.Time
	)
	err := row.Scan(
		&preview.ID, &preview.UserID, &preview.PersonaID, &preview.CreatedAt, &preview.UpdatedAt,
		&preview.Persona.ID, &preview.Persona.Slug, &preview.Persona.Name, &preview.Persona.Category,
		&preview.Persona.ShortDescription, &preview.Persona.SystemPrompt, &preview.Persona.ToneDescription,
		&preview.Persona.AvatarURL, &preview.Persona.AccentColor, &preview.Persona.IsMinorAppropriate,
		&preview.Persona.IsActive, &preview.Persona.SortOrder, &preview.Persona.CreatedAt, &preview.Persona.UpdatedAt,
		&lastID, &lastConvID, &lastRole, &lastContent, &lastReactionEmoji, &lastCreatedAt,
	)
	if err != nil {
		return model.ConversationPreview{}, err
	}
	if lastID != nil && lastConvID != nil && lastRole != nil && lastContent != nil && lastCreatedAt != nil {
		preview.LastMessage = &model.Message{
			ID:             *lastID,
			ConversationID: *lastConvID,
			Role:           model.MessageRole(*lastRole),
			Content:        *lastContent,
			ReactionEmoji:  lastReactionEmoji,
			CreatedAt:      *lastCreatedAt,
		}
	}
	return preview, nil
}
