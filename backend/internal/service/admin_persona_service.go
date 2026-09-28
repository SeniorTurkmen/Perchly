package service

import (
	"context"
	"errors"
	"strings"

	"perchly-backend/internal/apierror"
	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
)

var (
	ErrInvalidPersonaInput             = errors.New("invalid persona input")
	ErrPersonaLLMModelNotFound         = errors.New("persona llm model not found")
	ErrPersonaLLMModelInactive         = errors.New("persona llm model or its credential is inactive")
	ErrInvalidPersonaTranslationLocale = errors.New("invalid persona translation locale")
	ErrInvalidPersonaTranslationInput  = errors.New("invalid persona translation input")
)

type AdminPersonaRepo interface {
	ListAll(ctx context.Context) ([]model.Persona, error)
	GetByID(ctx context.Context, id string) (model.Persona, error)
	Create(ctx context.Context, p model.Persona) (model.Persona, error)
	Update(ctx context.Context, p model.Persona) (model.Persona, error)
}

// AdminPersonaTranslationRepo is the persistence dependency the
// translation CRUD methods below need.
type AdminPersonaTranslationRepo interface {
	ListTranslations(ctx context.Context, personaID string) ([]model.PersonaTranslation, error)
	UpsertTranslation(ctx context.Context, t model.PersonaTranslation) (model.PersonaTranslation, error)
	DeleteTranslation(ctx context.Context, personaID, locale string) error
}

// AdminPersonaLLMModelRepo is the narrow slice of AdminLLMModelRepo this
// service needs — just enough to validate a persona's llm_model_id
// points at a real, active model.
type AdminPersonaLLMModelRepo interface {
	GetByID(ctx context.Context, id string) (model.LLMModel, error)
}

// AdminPersonaLLMCredentialRepo is the narrow slice of
// AdminLLMCredentialRepo this service needs — validating a persona's
// model isn't enough on its own if the credential backing it has since
// been deactivated.
type AdminPersonaLLMCredentialRepo interface {
	GetByID(ctx context.Context, id string) (model.LLMCredential, error)
}

// AdminPersonaService is the admin dashboard's full CRUD over personas
// — unlike PersonaService (the public, read-only, active-only view),
// this can see inactive personas and write to every field, including
// system_prompt, which is never exposed over the public API. Every
// mutation is written to admin_audit_log.
type AdminPersonaService struct {
	personas     AdminPersonaRepo
	translations AdminPersonaTranslationRepo
	llmModels    AdminPersonaLLMModelRepo
	credentials  AdminPersonaLLMCredentialRepo
	auditLog     AdminAuditLogRepo
}

func NewAdminPersonaService(
	personas AdminPersonaRepo,
	translations AdminPersonaTranslationRepo,
	llmModels AdminPersonaLLMModelRepo,
	credentials AdminPersonaLLMCredentialRepo,
	auditLog AdminAuditLogRepo,
) *AdminPersonaService {
	return &AdminPersonaService{
		personas: personas, translations: translations,
		llmModels: llmModels, credentials: credentials, auditLog: auditLog,
	}
}

func (s *AdminPersonaService) List(ctx context.Context) ([]model.Persona, error) {
	return s.personas.ListAll(ctx)
}

func (s *AdminPersonaService) Get(ctx context.Context, id string) (model.Persona, error) {
	return s.personas.GetByID(ctx, id)
}

func (s *AdminPersonaService) Create(ctx context.Context, adminUserID string, p model.Persona) (model.Persona, error) {
	if err := validatePersonaInput(p); err != nil {
		return model.Persona{}, err
	}
	if err := s.validateLLMModel(ctx, p.LLMModelID); err != nil {
		return model.Persona{}, err
	}

	created, err := s.personas.Create(ctx, p)
	if err != nil {
		return model.Persona{}, err
	}

	_ = s.auditLog.Create(ctx, adminUserID, "admin.persona.create", "persona", &created.ID, map[string]any{
		"slug": created.Slug,
	})

	return created, nil
}

func (s *AdminPersonaService) Update(ctx context.Context, adminUserID string, p model.Persona) (model.Persona, error) {
	if err := validatePersonaInput(p); err != nil {
		return model.Persona{}, err
	}
	if err := s.validateLLMModel(ctx, p.LLMModelID); err != nil {
		return model.Persona{}, err
	}

	updated, err := s.personas.Update(ctx, p)
	if err != nil {
		return model.Persona{}, err
	}

	_ = s.auditLog.Create(ctx, adminUserID, "admin.persona.update", "persona", &updated.ID, map[string]any{
		"slug": updated.Slug,
	})

	return updated, nil
}

// validateLLMModel confirms llmModelID (when set — nil means "use the
// process-wide default", always valid) points at a model that both
// exists and is active, and whose credential is also active. Checking
// only the model wouldn't be enough: an admin can deactivate a
// credential without touching the models under it (see
// AdminLLMService.UpdateCredential), so a model can be "active" while
// unusable.
func (s *AdminPersonaService) validateLLMModel(ctx context.Context, llmModelID *string) error {
	if llmModelID == nil {
		return nil
	}

	m, err := s.llmModels.GetByID(ctx, *llmModelID)
	if err != nil {
		if errors.Is(err, repository.ErrLLMModelNotFound) {
			return ErrPersonaLLMModelNotFound
		}
		return err
	}
	if !m.IsActive {
		return ErrPersonaLLMModelInactive
	}

	credential, err := s.credentials.GetByID(ctx, m.CredentialID)
	if err != nil {
		if errors.Is(err, repository.ErrLLMCredentialNotFound) {
			return ErrPersonaLLMModelNotFound
		}
		return err
	}
	if !credential.IsActive {
		return ErrPersonaLLMModelInactive
	}

	return nil
}

func validatePersonaInput(p model.Persona) error {
	required := []string{p.Slug, p.Name, p.Category, p.ShortDescription, p.SystemPrompt, p.ToneDescription, p.AccentColor}
	for _, field := range required {
		if strings.TrimSpace(field) == "" {
			return ErrInvalidPersonaInput
		}
	}
	return p.DefaultTraits.Validate()
}

// isTranslatableLocale reports whether locale is one of Perchly's
// supported locales other than Turkish — Turkish is never stored as a
// translation row (see model.PersonaTranslation's doc comment), so a
// PUT/DELETE naming locale "tr" is rejected outright rather than
// silently doing nothing useful.
func isTranslatableLocale(locale string) bool {
	if locale == string(apierror.LocaleTR) {
		return false
	}
	return apierror.IsSupported(apierror.Locale(locale))
}

// ListTranslations returns every stored translation for personaID.
// Returns repository.ErrPersonaNotFound if personaID doesn't exist.
func (s *AdminPersonaService) ListTranslations(ctx context.Context, personaID string) ([]model.PersonaTranslation, error) {
	if _, err := s.personas.GetByID(ctx, personaID); err != nil {
		return nil, err
	}
	return s.translations.ListTranslations(ctx, personaID)
}

// UpsertTranslation creates or overwrites personaID's translation for
// locale, then records an admin_audit_log row.
func (s *AdminPersonaService) UpsertTranslation(ctx context.Context, adminUserID, personaID, locale string, t model.PersonaTranslation) (model.PersonaTranslation, error) {
	if !isTranslatableLocale(locale) {
		return model.PersonaTranslation{}, ErrInvalidPersonaTranslationLocale
	}
	if strings.TrimSpace(t.Name) == "" || strings.TrimSpace(t.ShortDescription) == "" || strings.TrimSpace(t.ToneDescription) == "" {
		return model.PersonaTranslation{}, ErrInvalidPersonaTranslationInput
	}
	if _, err := s.personas.GetByID(ctx, personaID); err != nil {
		return model.PersonaTranslation{}, err
	}

	t.PersonaID = personaID
	t.Locale = locale
	result, err := s.translations.UpsertTranslation(ctx, t)
	if err != nil {
		return model.PersonaTranslation{}, err
	}

	_ = s.auditLog.Create(ctx, adminUserID, "admin.persona.translation.upsert", "persona", &personaID, map[string]any{
		"locale": locale,
	})

	return result, nil
}

// DeleteTranslation removes personaID's translation for locale (a persona
// with no such row simply reverts to reading as Turkish — see
// ListLocalized), then records an admin_audit_log row.
func (s *AdminPersonaService) DeleteTranslation(ctx context.Context, adminUserID, personaID, locale string) error {
	if !isTranslatableLocale(locale) {
		return ErrInvalidPersonaTranslationLocale
	}
	if _, err := s.personas.GetByID(ctx, personaID); err != nil {
		return err
	}
	if err := s.translations.DeleteTranslation(ctx, personaID, locale); err != nil {
		return err
	}

	_ = s.auditLog.Create(ctx, adminUserID, "admin.persona.translation.delete", "persona", &personaID, map[string]any{
		"locale": locale,
	})

	return nil
}
