package service

import (
	"context"
	"errors"
	"testing"

	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
)

type fakeAdminPersonaRepoForTranslations struct {
	personas map[string]model.Persona
}

func (r *fakeAdminPersonaRepoForTranslations) ListAll(context.Context) ([]model.Persona, error) {
	return nil, nil
}

func (r *fakeAdminPersonaRepoForTranslations) GetByID(_ context.Context, id string) (model.Persona, error) {
	p, ok := r.personas[id]
	if !ok {
		return model.Persona{}, repository.ErrPersonaNotFound
	}
	return p, nil
}

func (r *fakeAdminPersonaRepoForTranslations) Create(context.Context, model.Persona) (model.Persona, error) {
	return model.Persona{}, nil
}

func (r *fakeAdminPersonaRepoForTranslations) Update(context.Context, model.Persona) (model.Persona, error) {
	return model.Persona{}, nil
}

type fakeAdminPersonaTranslationRepo struct {
	byPersonaLocale map[string]model.PersonaTranslation
	deleted         []string // "personaID:locale"
}

func newFakeAdminPersonaTranslationRepo() *fakeAdminPersonaTranslationRepo {
	return &fakeAdminPersonaTranslationRepo{byPersonaLocale: map[string]model.PersonaTranslation{}}
}

func (r *fakeAdminPersonaTranslationRepo) ListTranslations(_ context.Context, personaID string) ([]model.PersonaTranslation, error) {
	var out []model.PersonaTranslation
	for _, t := range r.byPersonaLocale {
		if t.PersonaID == personaID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (r *fakeAdminPersonaTranslationRepo) UpsertTranslation(_ context.Context, t model.PersonaTranslation) (model.PersonaTranslation, error) {
	r.byPersonaLocale[t.PersonaID+":"+t.Locale] = t
	return t, nil
}

func (r *fakeAdminPersonaTranslationRepo) DeleteTranslation(_ context.Context, personaID, locale string) error {
	delete(r.byPersonaLocale, personaID+":"+locale)
	r.deleted = append(r.deleted, personaID+":"+locale)
	return nil
}

type fakeAdminAuditLogRepo struct {
	entries []struct {
		AdminUserID, Action, TargetType string
		TargetID                        *string
	}
}

func (r *fakeAdminAuditLogRepo) Create(_ context.Context, adminUserID, action, targetType string, targetID *string, _ map[string]any) error {
	r.entries = append(r.entries, struct {
		AdminUserID, Action, TargetType string
		TargetID                        *string
	}{adminUserID, action, targetType, targetID})
	return nil
}

func newTestAdminPersonaService(personaID string) (*AdminPersonaService, *fakeAdminPersonaTranslationRepo, *fakeAdminAuditLogRepo) {
	personas := &fakeAdminPersonaRepoForTranslations{personas: map[string]model.Persona{
		personaID: {ID: personaID, Slug: "motivational-coach", Name: "Ada"},
	}}
	translations := newFakeAdminPersonaTranslationRepo()
	auditLog := &fakeAdminAuditLogRepo{}
	svc := NewAdminPersonaService(personas, translations, nil, nil, auditLog)
	return svc, translations, auditLog
}

func TestAdminPersonaService_UpsertTranslation_RejectsTurkish(t *testing.T) {
	svc, _, _ := newTestAdminPersonaService("p1")

	_, err := svc.UpsertTranslation(context.Background(), "admin1", "p1", "tr", model.PersonaTranslation{
		Name: "Ada", ShortDescription: "x", ToneDescription: "y",
	})
	if !errors.Is(err, ErrInvalidPersonaTranslationLocale) {
		t.Fatalf("err = %v, want ErrInvalidPersonaTranslationLocale", err)
	}
}

func TestAdminPersonaService_UpsertTranslation_RejectsUnsupportedLocale(t *testing.T) {
	svc, _, _ := newTestAdminPersonaService("p1")

	_, err := svc.UpsertTranslation(context.Background(), "admin1", "p1", "xx", model.PersonaTranslation{
		Name: "Ada", ShortDescription: "x", ToneDescription: "y",
	})
	if !errors.Is(err, ErrInvalidPersonaTranslationLocale) {
		t.Fatalf("err = %v, want ErrInvalidPersonaTranslationLocale", err)
	}
}

func TestAdminPersonaService_UpsertTranslation_RejectsEmptyInput(t *testing.T) {
	svc, _, _ := newTestAdminPersonaService("p1")

	_, err := svc.UpsertTranslation(context.Background(), "admin1", "p1", "en", model.PersonaTranslation{
		Name: "Ada", ShortDescription: "", ToneDescription: "y",
	})
	if !errors.Is(err, ErrInvalidPersonaTranslationInput) {
		t.Fatalf("err = %v, want ErrInvalidPersonaTranslationInput", err)
	}
}

func TestAdminPersonaService_UpsertTranslation_RejectsUnknownPersona(t *testing.T) {
	svc, _, _ := newTestAdminPersonaService("p1")

	_, err := svc.UpsertTranslation(context.Background(), "admin1", "does-not-exist", "en", model.PersonaTranslation{
		Name: "Ada", ShortDescription: "x", ToneDescription: "y",
	})
	if !errors.Is(err, repository.ErrPersonaNotFound) {
		t.Fatalf("err = %v, want repository.ErrPersonaNotFound", err)
	}
}

func TestAdminPersonaService_UpsertTranslation_SucceedsAndAudits(t *testing.T) {
	svc, translations, auditLog := newTestAdminPersonaService("p1")

	result, err := svc.UpsertTranslation(context.Background(), "admin1", "p1", "en", model.PersonaTranslation{
		Name: "Ada", ShortDescription: "A coach.", ToneDescription: "Energetic.",
	})
	if err != nil {
		t.Fatalf("UpsertTranslation() error = %v", err)
	}
	if result.PersonaID != "p1" || result.Locale != "en" {
		t.Fatalf("result = %+v, want persona_id=p1 locale=en", result)
	}
	if _, ok := translations.byPersonaLocale["p1:en"]; !ok {
		t.Fatal("expected translation to be stored in repo")
	}
	if len(auditLog.entries) != 1 || auditLog.entries[0].Action != "admin.persona.translation.upsert" {
		t.Fatalf("audit log entries = %+v, want one admin.persona.translation.upsert entry", auditLog.entries)
	}
}

func TestAdminPersonaService_DeleteTranslation_RejectsTurkishAndUnsupported(t *testing.T) {
	svc, _, _ := newTestAdminPersonaService("p1")

	if err := svc.DeleteTranslation(context.Background(), "admin1", "p1", "tr"); !errors.Is(err, ErrInvalidPersonaTranslationLocale) {
		t.Fatalf("tr: err = %v, want ErrInvalidPersonaTranslationLocale", err)
	}
	if err := svc.DeleteTranslation(context.Background(), "admin1", "p1", "xx"); !errors.Is(err, ErrInvalidPersonaTranslationLocale) {
		t.Fatalf("xx: err = %v, want ErrInvalidPersonaTranslationLocale", err)
	}
}

func TestAdminPersonaService_DeleteTranslation_SucceedsAndAudits(t *testing.T) {
	svc, translations, auditLog := newTestAdminPersonaService("p1")

	if _, err := svc.UpsertTranslation(context.Background(), "admin1", "p1", "en", model.PersonaTranslation{
		Name: "Ada", ShortDescription: "x", ToneDescription: "y",
	}); err != nil {
		t.Fatalf("seed upsert: %v", err)
	}

	if err := svc.DeleteTranslation(context.Background(), "admin1", "p1", "en"); err != nil {
		t.Fatalf("DeleteTranslation() error = %v", err)
	}
	if _, ok := translations.byPersonaLocale["p1:en"]; ok {
		t.Fatal("expected translation to be removed from repo")
	}

	found := false
	for _, e := range auditLog.entries {
		if e.Action == "admin.persona.translation.delete" {
			found = true
		}
	}
	if !found {
		t.Fatalf("audit log entries = %+v, want an admin.persona.translation.delete entry", auditLog.entries)
	}
}

func TestAdminPersonaService_ListTranslations_UnknownPersonaNotFound(t *testing.T) {
	svc, _, _ := newTestAdminPersonaService("p1")

	_, err := svc.ListTranslations(context.Background(), "does-not-exist")
	if !errors.Is(err, repository.ErrPersonaNotFound) {
		t.Fatalf("err = %v, want repository.ErrPersonaNotFound", err)
	}
}
