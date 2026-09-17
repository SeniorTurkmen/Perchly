package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
)

// ErrPersonaNotFound is returned when no persona matches the given id.
var ErrPersonaNotFound = errors.New("persona not found")

type PersonaRepository struct {
	pool *pgxpool.Pool
}

func NewPersonaRepository(pool *pgxpool.Pool) *PersonaRepository {
	return &PersonaRepository{pool: pool}
}

const personaColumns = `
	id::text, slug, name, category, short_description, system_prompt,
	tone_description, avatar_url, accent_color, is_minor_appropriate, is_active, sort_order,
	created_at, updated_at`

// List returns active personas ordered for display.
func (r *PersonaRepository) List(ctx context.Context) ([]model.Persona, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+personaColumns+`
		FROM personas
		WHERE is_active = true
		ORDER BY sort_order ASC, created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	personas := make([]model.Persona, 0)
	for rows.Next() {
		p, err := scanPersona(rows)
		if err != nil {
			return nil, err
		}
		personas = append(personas, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return personas, nil
}

// GetByID returns a single persona by id. id must already be a
// well-formed UUID string; ErrPersonaNotFound is returned if it doesn't
// match any row.
func (r *PersonaRepository) GetByID(ctx context.Context, id string) (model.Persona, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+personaColumns+`
		FROM personas
		WHERE id = $1::uuid
	`, id)

	p, err := scanPersona(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Persona{}, ErrPersonaNotFound
		}
		return model.Persona{}, err
	}
	return p, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPersona(row rowScanner) (model.Persona, error) {
	var p model.Persona
	err := row.Scan(
		&p.ID, &p.Slug, &p.Name, &p.Category, &p.ShortDescription, &p.SystemPrompt,
		&p.ToneDescription, &p.AvatarURL, &p.AccentColor, &p.IsMinorAppropriate, &p.IsActive, &p.SortOrder,
		&p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}
