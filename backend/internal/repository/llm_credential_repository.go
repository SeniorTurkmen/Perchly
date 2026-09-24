package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
)

// ErrLLMCredentialNotFound is returned when no credential matches the
// given id.
var ErrLLMCredentialNotFound = errors.New("llm credential not found")

// ErrLLMCredentialInUse is returned by Delete when the credential still
// has models referencing it — a credential must have its models removed
// first, so a persona pointing at one of them is never left dangling.
var ErrLLMCredentialInUse = errors.New("llm credential still has models")

type LLMCredentialRepository struct {
	pool *pgxpool.Pool
}

func NewLLMCredentialRepository(pool *pgxpool.Pool) *LLMCredentialRepository {
	return &LLMCredentialRepository{pool: pool}
}

const llmCredentialColumns = `
	id::text, provider, label, api_key_ciphertext, api_key_nonce, api_key_last4,
	base_url, is_active, created_at, updated_at`

// ListAll returns every stored credential, newest first, regardless of
// is_active — the admin dashboard needs to see and re-activate a
// deactivated credential too.
func (r *LLMCredentialRepository) ListAll(ctx context.Context) ([]model.LLMCredential, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+llmCredentialColumns+`
		FROM llm_credentials
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	credentials := make([]model.LLMCredential, 0)
	for rows.Next() {
		c, err := scanLLMCredential(rows)
		if err != nil {
			return nil, err
		}
		credentials = append(credentials, c)
	}
	return credentials, rows.Err()
}

// GetByID returns a single credential by id, including the encrypted
// key material — used only where the key is about to be decrypted for
// an actual provider call, never for an admin API response.
func (r *LLMCredentialRepository) GetByID(ctx context.Context, id string) (model.LLMCredential, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+llmCredentialColumns+`
		FROM llm_credentials
		WHERE id = $1::uuid
	`, id)
	return scanLLMCredentialOrNotFound(row)
}

// Create inserts a new credential. c.ID/CreatedAt/UpdatedAt are ignored
// — the database assigns them. The caller (service layer) is
// responsible for having already encrypted the API key into
// APIKeyCiphertext/APIKeyNonce and computed APIKeyLast4.
func (r *LLMCredentialRepository) Create(ctx context.Context, c model.LLMCredential) (model.LLMCredential, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO llm_credentials (
			provider, label, api_key_ciphertext, api_key_nonce, api_key_last4, base_url, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING `+llmCredentialColumns,
		c.Provider, c.Label, c.APIKeyCiphertext, c.APIKeyNonce, c.APIKeyLast4, c.BaseURL, c.IsActive,
	)
	return scanLLMCredential(row)
}

// Update overwrites a credential's editable fields. When newAPIKey is
// nil, the existing encrypted key is left untouched (an admin editing
// just the label or base URL shouldn't be forced to re-enter the
// token); when non-nil, it replaces the stored ciphertext/nonce/last4.
func (r *LLMCredentialRepository) Update(ctx context.Context, c model.LLMCredential, replaceAPIKey bool) (model.LLMCredential, error) {
	var row pgx.Row
	if replaceAPIKey {
		row = r.pool.QueryRow(ctx, `
			UPDATE llm_credentials SET
				label = $2, api_key_ciphertext = $3, api_key_nonce = $4, api_key_last4 = $5,
				base_url = $6, is_active = $7
			WHERE id = $1::uuid
			RETURNING `+llmCredentialColumns,
			c.ID, c.Label, c.APIKeyCiphertext, c.APIKeyNonce, c.APIKeyLast4, c.BaseURL, c.IsActive,
		)
	} else {
		row = r.pool.QueryRow(ctx, `
			UPDATE llm_credentials SET
				label = $2, base_url = $3, is_active = $4
			WHERE id = $1::uuid
			RETURNING `+llmCredentialColumns,
			c.ID, c.Label, c.BaseURL, c.IsActive,
		)
	}

	updated, err := scanLLMCredential(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.LLMCredential{}, ErrLLMCredentialNotFound
		}
		return model.LLMCredential{}, err
	}
	return updated, nil
}

// Delete removes a credential. Returns ErrLLMCredentialInUse if any
// llm_models rows still reference it (the FK's ON DELETE CASCADE would
// silently remove them and orphan any persona pointing at one — the
// service layer requires an explicit "remove its models first" instead).
func (r *LLMCredentialRepository) Delete(ctx context.Context, id string) error {
	var modelCount int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM llm_models WHERE credential_id = $1::uuid`, id).Scan(&modelCount); err != nil {
		return err
	}
	if modelCount > 0 {
		return ErrLLMCredentialInUse
	}

	tag, err := r.pool.Exec(ctx, `DELETE FROM llm_credentials WHERE id = $1::uuid`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrLLMCredentialNotFound
	}
	return nil
}

func scanLLMCredentialOrNotFound(row rowScanner) (model.LLMCredential, error) {
	c, err := scanLLMCredential(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.LLMCredential{}, ErrLLMCredentialNotFound
		}
		return model.LLMCredential{}, err
	}
	return c, nil
}

func scanLLMCredential(row rowScanner) (model.LLMCredential, error) {
	var c model.LLMCredential
	err := row.Scan(
		&c.ID, &c.Provider, &c.Label, &c.APIKeyCiphertext, &c.APIKeyNonce, &c.APIKeyLast4,
		&c.BaseURL, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}
