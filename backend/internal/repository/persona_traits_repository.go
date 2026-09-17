package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
)

type PersonaTraitsRepository struct {
	pool *pgxpool.Pool
}

func NewPersonaTraitsRepository(pool *pgxpool.Pool) *PersonaTraitsRepository {
	return &PersonaTraitsRepository{pool: pool}
}

// GetEffective returns userID's current dial values for personaID —
// their own customized ones if they've set any, otherwise the
// persona's defaults — plus whether they've actually customized it.
// Returns ErrPersonaNotFound if personaID doesn't exist.
func (r *PersonaTraitsRepository) GetEffective(ctx context.Context, userID, personaID string) (model.PersonaTraits, bool, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(upt.warmth, p.default_warmth),
			COALESCE(upt.humor, p.default_humor),
			COALESCE(upt.wisdom, p.default_wisdom),
			COALESCE(upt.directness, p.default_directness),
			COALESCE(upt.energy, p.default_energy),
			upt.user_id IS NOT NULL
		FROM personas p
		LEFT JOIN user_persona_traits upt ON upt.persona_id = p.id AND upt.user_id = $1::uuid
		WHERE p.id = $2::uuid
	`, userID, personaID)

	var t model.PersonaTraits
	var isCustomized bool
	err := row.Scan(&t.Warmth, &t.Humor, &t.Wisdom, &t.Directness, &t.Energy, &isCustomized)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.PersonaTraits{}, false, ErrPersonaNotFound
		}
		return model.PersonaTraits{}, false, err
	}
	return t, isCustomized, nil
}

// Upsert sets userID's current dial values for personaID, replacing
// any previous customization.
func (r *PersonaTraitsRepository) Upsert(ctx context.Context, userID, personaID string, t model.PersonaTraits) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_persona_traits (user_id, persona_id, warmth, humor, wisdom, directness, energy, updated_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, now())
		ON CONFLICT (user_id, persona_id) DO UPDATE SET
			warmth = EXCLUDED.warmth,
			humor = EXCLUDED.humor,
			wisdom = EXCLUDED.wisdom,
			directness = EXCLUDED.directness,
			energy = EXCLUDED.energy,
			updated_at = now()
	`, userID, personaID, t.Warmth, t.Humor, t.Wisdom, t.Directness, t.Energy)
	return err
}

// Reset removes userID's customization for personaID, if any — the
// persona's own defaults apply again from then on.
func (r *PersonaTraitsRepository) Reset(ctx context.Context, userID, personaID string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM user_persona_traits WHERE user_id = $1::uuid AND persona_id = $2::uuid
	`, userID, personaID)
	return err
}
