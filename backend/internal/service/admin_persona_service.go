package service

import (
	"context"
	"errors"
	"strings"

	"perchly-backend/internal/model"
)

var ErrInvalidPersonaInput = errors.New("invalid persona input")

type AdminPersonaRepo interface {
	ListAll(ctx context.Context) ([]model.Persona, error)
	GetByID(ctx context.Context, id string) (model.Persona, error)
	Create(ctx context.Context, p model.Persona) (model.Persona, error)
	Update(ctx context.Context, p model.Persona) (model.Persona, error)
}

// AdminPersonaService is the admin dashboard's full CRUD over personas
// — unlike PersonaService (the public, read-only, active-only view),
// this can see inactive personas and write to every field, including
// system_prompt, which is never exposed over the public API. Every
// mutation is written to admin_audit_log.
type AdminPersonaService struct {
	personas AdminPersonaRepo
	auditLog AdminAuditLogRepo
}

func NewAdminPersonaService(personas AdminPersonaRepo, auditLog AdminAuditLogRepo) *AdminPersonaService {
	return &AdminPersonaService{personas: personas, auditLog: auditLog}
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

	updated, err := s.personas.Update(ctx, p)
	if err != nil {
		return model.Persona{}, err
	}

	_ = s.auditLog.Create(ctx, adminUserID, "admin.persona.update", "persona", &updated.ID, map[string]any{
		"slug": updated.Slug,
	})

	return updated, nil
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
