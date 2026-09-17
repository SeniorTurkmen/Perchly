package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
)

type personaFinder interface {
	List(ctx context.Context) ([]model.Persona, error)
	GetByID(ctx context.Context, id string) (model.Persona, error)
}

type PersonaHandler struct {
	personas personaFinder
}

func NewPersonaHandler(personas personaFinder) *PersonaHandler {
	return &PersonaHandler{personas: personas}
}

// List handles GET /personas.
func (h *PersonaHandler) List(w http.ResponseWriter, r *http.Request) {
	personas, err := h.personas.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "personas listelenemedi")
		return
	}
	writeJSON(w, http.StatusOK, personas)
}

// Get handles GET /personas/{id}.
func (h *PersonaHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz persona id")
		return
	}

	persona, err := h.personas.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrPersonaNotFound) {
			writeError(w, http.StatusNotFound, "persona bulunamadı")
			return
		}
		writeError(w, http.StatusInternalServerError, "persona getirilemedi")
		return
	}
	writeJSON(w, http.StatusOK, persona)
}
