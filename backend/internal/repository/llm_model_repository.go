package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
)

// ErrLLMModelNotFound is returned when no model matches the given id.
var ErrLLMModelNotFound = errors.New("llm model not found")

type LLMModelRepository struct {
	pool *pgxpool.Pool
}

func NewLLMModelRepository(pool *pgxpool.Pool) *LLMModelRepository {
	return &LLMModelRepository{pool: pool}
}

const llmModelColumns = `
	id::text, credential_id::text, model_name, display_name, is_default, is_active,
	created_at, updated_at`

// ListAll returns every model across every credential, newest first —
// used to populate the persona "which model" picker.
func (r *LLMModelRepository) ListAll(ctx context.Context) ([]model.LLMModel, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+llmModelColumns+`
		FROM llm_models
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	models := make([]model.LLMModel, 0)
	for rows.Next() {
		m, err := scanLLMModel(rows)
		if err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	return models, rows.Err()
}

// ListByCredential returns every model belonging to one credential,
// newest first.
func (r *LLMModelRepository) ListByCredential(ctx context.Context, credentialID string) ([]model.LLMModel, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+llmModelColumns+`
		FROM llm_models
		WHERE credential_id = $1::uuid
		ORDER BY created_at DESC
	`, credentialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	models := make([]model.LLMModel, 0)
	for rows.Next() {
		m, err := scanLLMModel(rows)
		if err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	return models, rows.Err()
}

// GetByID returns a single model by id.
func (r *LLMModelRepository) GetByID(ctx context.Context, id string) (model.LLMModel, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+llmModelColumns+`
		FROM llm_models
		WHERE id = $1::uuid
	`, id)
	return scanLLMModelOrNotFound(row)
}

// Create inserts a new model under a credential. m.ID/CreatedAt/UpdatedAt
// are ignored — the database assigns them.
func (r *LLMModelRepository) Create(ctx context.Context, m model.LLMModel) (model.LLMModel, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO llm_models (credential_id, model_name, display_name, is_default, is_active)
		VALUES ($1::uuid, $2, $3, $4, $5)
		RETURNING `+llmModelColumns,
		m.CredentialID, m.ModelName, m.DisplayName, m.IsDefault, m.IsActive,
	)
	return scanLLMModel(row)
}

// Update overwrites a model's editable fields (model_name/credential_id
// are fixed at creation — changing which credential a model belongs to
// would silently change which API key existing personas' conversations
// authenticate with, so that's a delete+recreate, not an edit).
func (r *LLMModelRepository) Update(ctx context.Context, m model.LLMModel) (model.LLMModel, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE llm_models SET
			display_name = $2, is_default = $3, is_active = $4
		WHERE id = $1::uuid
		RETURNING `+llmModelColumns,
		m.ID, m.DisplayName, m.IsDefault, m.IsActive,
	)
	updated, err := scanLLMModel(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.LLMModel{}, ErrLLMModelNotFound
		}
		return model.LLMModel{}, err
	}
	return updated, nil
}

// ClearDefault unsets is_default on every model under credentialID —
// called before setting a new default, since exactly zero or one model
// per credential may be the default at a time.
func (r *LLMModelRepository) ClearDefault(ctx context.Context, credentialID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE llm_models SET is_default = false WHERE credential_id = $1::uuid`, credentialID)
	return err
}

// Delete removes a model. Any persona currently assigned to it falls
// back to the process-wide default automatically, via
// personas.llm_model_id's ON DELETE SET NULL.
func (r *LLMModelRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM llm_models WHERE id = $1::uuid`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrLLMModelNotFound
	}
	return nil
}

func scanLLMModelOrNotFound(row rowScanner) (model.LLMModel, error) {
	m, err := scanLLMModel(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.LLMModel{}, ErrLLMModelNotFound
		}
		return model.LLMModel{}, err
	}
	return m, nil
}

func scanLLMModel(row rowScanner) (model.LLMModel, error) {
	var m model.LLMModel
	err := row.Scan(
		&m.ID, &m.CredentialID, &m.ModelName, &m.DisplayName, &m.IsDefault, &m.IsActive,
		&m.CreatedAt, &m.UpdatedAt,
	)
	return m, err
}
