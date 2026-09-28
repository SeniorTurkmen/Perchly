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
			default_warmth, default_humor, default_wisdom, default_directness, default_energy,
			llm_model_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17::uuid)
		RETURNING `+personaColumns,
		p.Slug, p.Name, p.Category, p.ShortDescription, p.SystemPrompt, p.ToneDescription,
		p.AvatarURL, p.AccentColor, p.IsMinorAppropriate, p.IsActive, p.SortOrder,
		p.DefaultTraits.Warmth, p.DefaultTraits.Humor, p.DefaultTraits.Wisdom, p.DefaultTraits.Directness, p.DefaultTraits.Energy,
		p.LLMModelID,
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
			default_warmth = $13, default_humor = $14, default_wisdom = $15, default_directness = $16, default_energy = $17,
			llm_model_id = $18::uuid
		WHERE id = $1::uuid
		RETURNING `+personaColumns,
		p.ID, p.Slug, p.Name, p.Category, p.ShortDescription, p.SystemPrompt,
		p.ToneDescription, p.AvatarURL, p.AccentColor, p.IsMinorAppropriate, p.IsActive, p.SortOrder,
		p.DefaultTraits.Warmth, p.DefaultTraits.Humor, p.DefaultTraits.Wisdom, p.DefaultTraits.Directness, p.DefaultTraits.Energy,
		p.LLMModelID,
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
	llm_model_id::text, created_at, updated_at`

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

// localizedPersonaColumns mirrors personaColumns exactly — same order,
// same set — except name/short_description/tone_description are
// COALESCEd against a joined persona_translations row. Every query using
// this must alias personas as p and LEFT JOIN persona_translations as pt,
// so scanPersona (below) can be reused unmodified.
const localizedPersonaColumns = `
	p.id::text, p.slug, COALESCE(pt.name, p.name), p.category,
	COALESCE(pt.short_description, p.short_description), p.system_prompt,
	COALESCE(pt.tone_description, p.tone_description), p.avatar_url, p.accent_color,
	p.is_minor_appropriate, p.is_active, p.sort_order,
	p.default_warmth, p.default_humor, p.default_wisdom, p.default_directness, p.default_energy,
	p.llm_model_id::text, p.created_at, p.updated_at`

// ListLocalized is List with name/short_description/tone_description
// resolved for locale — a persona with no persona_translations row for
// locale (including "tr", which is never stored there) transparently
// falls back to its base Turkish columns.
func (r *PersonaRepository) ListLocalized(ctx context.Context, locale string) ([]model.Persona, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+localizedPersonaColumns+`
		FROM personas p
		LEFT JOIN persona_translations pt ON pt.persona_id = p.id AND pt.locale = $1
		WHERE p.is_active = true
		ORDER BY p.sort_order ASC, p.created_at ASC
	`, locale)
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

// GetByIDLocalized is GetByID with the same locale-fallback behavior as
// ListLocalized.
func (r *PersonaRepository) GetByIDLocalized(ctx context.Context, id, locale string) (model.Persona, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+localizedPersonaColumns+`
		FROM personas p
		LEFT JOIN persona_translations pt ON pt.persona_id = p.id AND pt.locale = $2
		WHERE p.id = $1::uuid
	`, id, locale)

	p, err := scanPersona(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Persona{}, ErrPersonaNotFound
		}
		return model.Persona{}, err
	}
	return p, nil
}

// ListTranslations returns every stored translation for personaID, across
// whichever locales an admin has filled in — never "tr" (see
// model.PersonaTranslation's doc comment).
func (r *PersonaRepository) ListTranslations(ctx context.Context, personaID string) ([]model.PersonaTranslation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT persona_id::text, locale, name, short_description, tone_description, created_at, updated_at
		FROM persona_translations
		WHERE persona_id = $1::uuid
		ORDER BY locale ASC
	`, personaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	translations := make([]model.PersonaTranslation, 0)
	for rows.Next() {
		var t model.PersonaTranslation
		if err := rows.Scan(&t.PersonaID, &t.Locale, &t.Name, &t.ShortDescription, &t.ToneDescription, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		translations = append(translations, t)
	}
	return translations, rows.Err()
}

// UpsertTranslation creates or overwrites t.PersonaID's translation for
// t.Locale.
func (r *PersonaRepository) UpsertTranslation(ctx context.Context, t model.PersonaTranslation) (model.PersonaTranslation, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO persona_translations (persona_id, locale, name, short_description, tone_description)
		VALUES ($1::uuid, $2, $3, $4, $5)
		ON CONFLICT (persona_id, locale) DO UPDATE SET
			name = EXCLUDED.name,
			short_description = EXCLUDED.short_description,
			tone_description = EXCLUDED.tone_description
		RETURNING persona_id::text, locale, name, short_description, tone_description, created_at, updated_at
	`, t.PersonaID, t.Locale, t.Name, t.ShortDescription, t.ToneDescription)

	var result model.PersonaTranslation
	err := row.Scan(&result.PersonaID, &result.Locale, &result.Name, &result.ShortDescription, &result.ToneDescription, &result.CreatedAt, &result.UpdatedAt)
	return result, err
}

// DeleteTranslation removes personaID's translation for locale, if any —
// idempotent, since deleting an already-absent translation just means
// that locale keeps falling back to the base Turkish content, which is
// already the case.
func (r *PersonaRepository) DeleteTranslation(ctx context.Context, personaID, locale string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM persona_translations WHERE persona_id = $1::uuid AND locale = $2
	`, personaID, locale)
	return err
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
		&p.LLMModelID, &p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}
