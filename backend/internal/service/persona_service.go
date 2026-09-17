package service

import (
	"context"

	"perchly-backend/internal/model"
)

// PersonaRepo is the persistence dependency PersonaService needs.
type PersonaRepo interface {
	List(ctx context.Context) ([]model.Persona, error)
	GetByID(ctx context.Context, id string) (model.Persona, error)
}

type PersonaService struct {
	repo PersonaRepo
}

func NewPersonaService(repo PersonaRepo) *PersonaService {
	return &PersonaService{repo: repo}
}

func (s *PersonaService) List(ctx context.Context) ([]model.Persona, error) {
	return s.repo.List(ctx)
}

func (s *PersonaService) GetByID(ctx context.Context, id string) (model.Persona, error) {
	return s.repo.GetByID(ctx, id)
}
