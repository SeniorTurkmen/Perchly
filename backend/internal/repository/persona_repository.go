package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
)

// ListAll returns every persona regardless of is_active — unlike List
// (the public endpoint's method, active-only), the admin dashboard
// needs to see and re-activate a deactivated persona too.
func (r *PersonaRepository) ListAll(ctx context.Context) ([]model.Persona, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+personaColumns+`
		FROM personas
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
	return personas, rows.Err()
}

// Create inserts a new persona. p.ID/CreatedAt/UpdatedAt are ignored —
// the database assigns them.
func (r *PersonaRepository) Create(ctx context.Context, p model.Persona) (model.Persona, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO personas (
			slug, name, category, short_description, system_prompt, tone_description,
			avatar_url, accent_color, is_minor_appropriate, is_active, sort_order,
			default_warmth, default_humor, default_wisdom, default_directness, default_energy
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING `+personaColumns,
		p.Slug, p.Name, p.Category, p.ShortDescription, p.SystemPrompt, p.ToneDescription,
		p.AvatarURL, p.AccentColor, p.IsMinorAppropriate, p.IsActive, p.SortOrder,
		p.DefaultTraits.Warmth, p.DefaultTraits.Humor, p.DefaultTraits.Wisdom, p.DefaultTraits.Directness, p.DefaultTraits.Energy,
	)
	return scanPersona(row)
}

// Update overwrites every editable field of an existing persona (a full
// replace, not a partial patch — the admin form always submits the
// complete record). updated_at is bumped by the personas_set_updated_at
// trigger, not set here. Returns ErrPersonaNotFound if p.ID doesn't
// exist.
func (r *PersonaRepository) Update(ctx context.Context, p model.Persona) (model.Persona, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE personas SET
			slug = $2, name = $3, category = $4, short_description = $5, system_prompt = $6,
			tone_description = $7, avatar_url = $8, accent_color = $9, is_minor_appropriate = $10,
			is_active = $11, sort_order = $12,
			default_warmth = $13, default_humor = $14, default_wisdom = $15, default_directness = $16, default_energy = $17
		WHERE id = $1::uuid
		RETURNING `+personaColumns,
		p.ID, p.Slug, p.Name, p.Category, p.ShortDescription, p.SystemPrompt,
		p.ToneDescription, p.AvatarURL, p.AccentColor, p.IsMinorAppropriate, p.IsActive, p.SortOrder,
		p.DefaultTraits.Warmth, p.DefaultTraits.Humor, p.DefaultTraits.Wisdom, p.DefaultTraits.Directness, p.DefaultTraits.Energy,
	)
	updated, err := scanPersona(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Persona{}, ErrPersonaNotFound
		}
		return model.Persona{}, err
	}
	return updated, nil
}

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
	default_warmth, default_humor, default_wisdom, default_directness, default_energy,
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
		&p.DefaultTraits.Warmth, &p.DefaultTraits.Humor, &p.DefaultTraits.Wisdom, &p.DefaultTraits.Directness, &p.DefaultTraits.Energy,
		&p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}
