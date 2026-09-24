package service

import (
	"context"
	"errors"
	"strings"

	"perchly-backend/internal/crypto"
	"perchly-backend/internal/model"
)

var (
	ErrInvalidLLMCredentialInput = errors.New("invalid llm credential input")
	ErrInvalidLLMModelInput      = errors.New("invalid llm model input")
	ErrLLMCredentialInactive     = errors.New("llm credential is inactive")
)

// knownLLMProviders mirrors internal/llm.New's supported Provider
// values, minus "echo" (a local-dev-only stub that needs no stored
// credential) — kept here rather than imported from internal/llm to
// avoid this admin-only package depending on the runtime LLM client
// package for a one-line validation list.
var knownLLMProviders = map[string]bool{
	"anthropic":   true,
	"openai":      true,
	"gemini":      true,
	"huggingface": true,
	"deepseek":    true,
}

type AdminLLMCredentialRepo interface {
	ListAll(ctx context.Context) ([]model.LLMCredential, error)
	GetByID(ctx context.Context, id string) (model.LLMCredential, error)
	Create(ctx context.Context, c model.LLMCredential) (model.LLMCredential, error)
	Update(ctx context.Context, c model.LLMCredential, replaceAPIKey bool) (model.LLMCredential, error)
	Delete(ctx context.Context, id string) error
}

type AdminLLMModelRepo interface {
	ListAll(ctx context.Context) ([]model.LLMModel, error)
	ListByCredential(ctx context.Context, credentialID string) ([]model.LLMModel, error)
	GetByID(ctx context.Context, id string) (model.LLMModel, error)
	Create(ctx context.Context, m model.LLMModel) (model.LLMModel, error)
	Update(ctx context.Context, m model.LLMModel) (model.LLMModel, error)
	Delete(ctx context.Context, id string) error
	ClearDefault(ctx context.Context, credentialID string) error
}

// AdminLLMService is the admin dashboard's CRUD over LLM provider
// credentials and the models made callable through them. Every stored
// API key is encrypted with box before it ever reaches the repository
// layer, and every mutation is written to admin_audit_log — never with
// the key itself in the detail payload.
type AdminLLMService struct {
	credentials AdminLLMCredentialRepo
	models      AdminLLMModelRepo
	box         *crypto.SecretBox
	auditLog    AdminAuditLogRepo
}

func NewAdminLLMService(credentials AdminLLMCredentialRepo, models AdminLLMModelRepo, box *crypto.SecretBox, auditLog AdminAuditLogRepo) *AdminLLMService {
	return &AdminLLMService{credentials: credentials, models: models, box: box, auditLog: auditLog}
}

// --- Credentials ---

func (s *AdminLLMService) ListCredentials(ctx context.Context) ([]model.LLMCredential, error) {
	return s.credentials.ListAll(ctx)
}

func (s *AdminLLMService) GetCredential(ctx context.Context, id string) (model.LLMCredential, error) {
	return s.credentials.GetByID(ctx, id)
}

// CreateCredentialInput is the raw, plaintext form of a new credential —
// apiKey is encrypted internally and never stored or logged as-is.
type CreateCredentialInput struct {
	Provider string
	Label    string
	APIKey   string
	BaseURL  *string
	IsActive bool
}

func (s *AdminLLMService) CreateCredential(ctx context.Context, adminUserID string, in CreateCredentialInput) (model.LLMCredential, error) {
	provider := strings.ToLower(strings.TrimSpace(in.Provider))
	label := strings.TrimSpace(in.Label)
	apiKey := strings.TrimSpace(in.APIKey)

	if !knownLLMProviders[provider] || label == "" || apiKey == "" {
		return model.LLMCredential{}, ErrInvalidLLMCredentialInput
	}

	ciphertext, nonce, err := s.box.Encrypt([]byte(apiKey))
	if err != nil {
		return model.LLMCredential{}, err
	}

	created, err := s.credentials.Create(ctx, model.LLMCredential{
		Provider:         provider,
		Label:            label,
		APIKeyCiphertext: ciphertext,
		APIKeyNonce:      nonce,
		APIKeyLast4:      last4(apiKey),
		BaseURL:          in.BaseURL,
		IsActive:         in.IsActive,
	})
	if err != nil {
		return model.LLMCredential{}, err
	}

	_ = s.auditLog.Create(ctx, adminUserID, "admin.llm_credential.create", "llm_credential", &created.ID, map[string]any{
		"provider": created.Provider,
		"label":    created.Label,
	})

	return created, nil
}

// UpdateCredentialInput mirrors CreateCredentialInput but APIKey is a
// pointer: nil (or empty) means "leave the stored key untouched", so an
// admin can edit just the label/base URL/active flag without re-pasting
// the token.
type UpdateCredentialInput struct {
	Label    string
	APIKey   *string
	BaseURL  *string
	IsActive bool
}

func (s *AdminLLMService) UpdateCredential(ctx context.Context, adminUserID, id string, in UpdateCredentialInput) (model.LLMCredential, error) {
	label := strings.TrimSpace(in.Label)
	if label == "" {
		return model.LLMCredential{}, ErrInvalidLLMCredentialInput
	}

	c := model.LLMCredential{ID: id, Label: label, BaseURL: in.BaseURL, IsActive: in.IsActive}
	replaceAPIKey := false

	if in.APIKey != nil {
		apiKey := strings.TrimSpace(*in.APIKey)
		if apiKey != "" {
			ciphertext, nonce, err := s.box.Encrypt([]byte(apiKey))
			if err != nil {
				return model.LLMCredential{}, err
			}
			c.APIKeyCiphertext = ciphertext
			c.APIKeyNonce = nonce
			c.APIKeyLast4 = last4(apiKey)
			replaceAPIKey = true
		}
	}

	updated, err := s.credentials.Update(ctx, c, replaceAPIKey)
	if err != nil {
		return model.LLMCredential{}, err
	}

	_ = s.auditLog.Create(ctx, adminUserID, "admin.llm_credential.update", "llm_credential", &updated.ID, map[string]any{
		"label":            updated.Label,
		"api_key_replaced": replaceAPIKey,
	})

	return updated, nil
}

func (s *AdminLLMService) DeleteCredential(ctx context.Context, adminUserID, id string) error {
	if err := s.credentials.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.auditLog.Create(ctx, adminUserID, "admin.llm_credential.delete", "llm_credential", &id, nil)
	return nil
}

// --- Models ---

// ListModels returns every model under credentialID, or every model
// across every credential when credentialID is nil.
func (s *AdminLLMService) ListModels(ctx context.Context, credentialID *string) ([]model.LLMModel, error) {
	if credentialID == nil {
		return s.models.ListAll(ctx)
	}
	return s.models.ListByCredential(ctx, *credentialID)
}

func (s *AdminLLMService) GetModel(ctx context.Context, id string) (model.LLMModel, error) {
	return s.models.GetByID(ctx, id)
}

func (s *AdminLLMService) CreateModel(ctx context.Context, adminUserID string, m model.LLMModel) (model.LLMModel, error) {
	m.ModelName = strings.TrimSpace(m.ModelName)
	m.DisplayName = strings.TrimSpace(m.DisplayName)
	if m.CredentialID == "" || m.ModelName == "" || m.DisplayName == "" {
		return model.LLMModel{}, ErrInvalidLLMModelInput
	}

	credential, err := s.credentials.GetByID(ctx, m.CredentialID)
	if err != nil {
		return model.LLMModel{}, err
	}
	if !credential.IsActive {
		return model.LLMModel{}, ErrLLMCredentialInactive
	}

	if m.IsDefault {
		if err := s.models.ClearDefault(ctx, m.CredentialID); err != nil {
			return model.LLMModel{}, err
		}
	}

	created, err := s.models.Create(ctx, m)
	if err != nil {
		return model.LLMModel{}, err
	}

	_ = s.auditLog.Create(ctx, adminUserID, "admin.llm_model.create", "llm_model", &created.ID, map[string]any{
		"credential_id": created.CredentialID,
		"model_name":    created.ModelName,
	})

	return created, nil
}

func (s *AdminLLMService) UpdateModel(ctx context.Context, adminUserID string, m model.LLMModel) (model.LLMModel, error) {
	m.DisplayName = strings.TrimSpace(m.DisplayName)
	if m.ID == "" || m.DisplayName == "" {
		return model.LLMModel{}, ErrInvalidLLMModelInput
	}

	existing, err := s.models.GetByID(ctx, m.ID)
	if err != nil {
		return model.LLMModel{}, err
	}

	if m.IsDefault {
		if err := s.models.ClearDefault(ctx, existing.CredentialID); err != nil {
			return model.LLMModel{}, err
		}
	}

	updated, err := s.models.Update(ctx, m)
	if err != nil {
		return model.LLMModel{}, err
	}

	_ = s.auditLog.Create(ctx, adminUserID, "admin.llm_model.update", "llm_model", &updated.ID, map[string]any{
		"display_name": updated.DisplayName,
	})

	return updated, nil
}

func (s *AdminLLMService) DeleteModel(ctx context.Context, adminUserID, id string) error {
	if err := s.models.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.auditLog.Create(ctx, adminUserID, "admin.llm_model.delete", "llm_model", &id, nil)
	return nil
}

// last4 returns the last 4 characters of a secret for display purposes
// (e.g. "sk-...ab12") — never the whole thing. A key of 4 characters or
// fewer is returned unchanged rather than panicking on the slice; real
// provider API keys are always much longer than this in practice.
func last4(secret string) string {
	if len(secret) <= 4 {
		return secret
	}
	return secret[len(secret)-4:]
}
