package service

import (
	"context"
	"fmt"
	"sync"

	"perchly-backend/internal/crypto"
	"perchly-backend/internal/llm"
	"perchly-backend/internal/model"
)

type LLMCredentialLookup interface {
	GetByID(ctx context.Context, id string) (model.LLMCredential, error)
}

type LLMModelLookup interface {
	GetByID(ctx context.Context, id string) (model.LLMModel, error)
}

// LLMClientFactory resolves the llm.Client a persona's chat turn should
// actually use. A persona with no assigned model (Persona.LLMModelID ==
// nil) always gets fallback — the single process-wide client built once
// at startup from LLM_PROVIDER/LLM_API_KEY/LLM_MODEL, exactly the
// pre-Faz-C behavior. A persona pinned to a stored model gets a client
// built from that model's credential: decrypted, cached by model ID so
// a chat turn never pays for a DB round-trip + decrypt on every
// message, only on first use after startup or after an invalidation.
//
// Cache entries are never time-based — they're removed exactly when
// AdminLLMService mutates the credential or model backing them (see
// InvalidateAll), so a chat turn is never answered by a stale client
// built from an API key an admin already replaced or revoked.
//
// ForModel fails closed: a model or its credential being inactive is
// treated the same as it not existing, never silently falling back to
// the process-wide default — the same rule AdminPersonaService already
// enforces when a persona is assigned a model, applied again here for
// the case where a credential is deactivated after a persona was
// already pointed at it.
type LLMClientFactory struct {
	credentials LLMCredentialLookup
	models      LLMModelLookup
	box         *crypto.SecretBox
	fallback    llm.Client

	mu    sync.RWMutex
	cache map[string]llm.Client // modelID -> client
}

func NewLLMClientFactory(credentials LLMCredentialLookup, models LLMModelLookup, box *crypto.SecretBox, fallback llm.Client) *LLMClientFactory {
	return &LLMClientFactory{
		credentials: credentials,
		models:      models,
		box:         box,
		fallback:    fallback,
		cache:       make(map[string]llm.Client),
	}
}

// ForPersona resolves p's client — fallback if p.LLMModelID is nil,
// otherwise ForModel(*p.LLMModelID).
func (f *LLMClientFactory) ForPersona(ctx context.Context, p model.Persona) (llm.Client, error) {
	if p.LLMModelID == nil {
		return f.fallback, nil
	}
	return f.ForModel(ctx, *p.LLMModelID)
}

// ForModel resolves (building and caching on first use) the client for
// one stored llm_models row.
func (f *LLMClientFactory) ForModel(ctx context.Context, modelID string) (llm.Client, error) {
	if client, ok := f.cached(modelID); ok {
		return client, nil
	}

	m, err := f.models.GetByID(ctx, modelID)
	if err != nil {
		return nil, fmt.Errorf("llm client factory: load model %s: %w", modelID, err)
	}

	if !m.IsActive {
		return nil, fmt.Errorf("llm client factory: model %s is inactive", modelID)
	}

	credential, err := f.credentials.GetByID(ctx, m.CredentialID)
	if err != nil {
		return nil, fmt.Errorf("llm client factory: load credential %s: %w", m.CredentialID, err)
	}
	if !credential.IsActive {
		return nil, fmt.Errorf("llm client factory: credential %s is inactive", credential.ID)
	}

	apiKey, err := f.box.Decrypt(credential.APIKeyCiphertext, credential.APIKeyNonce)
	if err != nil {
		return nil, fmt.Errorf("llm client factory: decrypt credential %s: %w", credential.ID, err)
	}

	var baseURL string
	if credential.BaseURL != nil {
		baseURL = *credential.BaseURL
	}

	client, err := llm.New(llm.Config{
		Provider: credential.Provider,
		APIKey:   string(apiKey),
		Model:    m.ModelName,
		BaseURL:  baseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("llm client factory: build client for model %s: %w", modelID, err)
	}

	f.mu.Lock()
	f.cache[modelID] = client
	f.mu.Unlock()

	return client, nil
}

func (f *LLMClientFactory) cached(modelID string) (llm.Client, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	client, ok := f.cache[modelID]
	return client, ok
}

// InvalidateAll drops every cached client, so the next ForModel call for
// any of them rebuilds from the current DB state. Called by
// AdminLLMService after any credential or model mutation (create,
// update — including deactivate — or delete); a single flat wipe rather
// than tracking which models a given credential backs, since the cache
// is just a handful of provider clients and admin mutations are rare —
// correctness over cleverness.
func (f *LLMClientFactory) InvalidateAll() {
	f.mu.Lock()
	f.cache = make(map[string]llm.Client)
	f.mu.Unlock()
}
