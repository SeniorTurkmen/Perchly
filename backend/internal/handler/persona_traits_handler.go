package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
)

type personaTraitsAPI interface {
	Get(ctx context.Context, userID, personaID string) (model.PersonaTraits, bool, error)
	Set(ctx context.Context, userID, personaID string, traits model.PersonaTraits) error
	Reset(ctx context.Context, userID, personaID string) error
}

// PersonaTraitsHandler exposes a signed-in user's own personality-dial
// customization for a persona. Requires auth.Middleware — mounted
// separately from the public GET /personas, /personas/{id} routes.
type PersonaTraitsHandler struct {
	traits personaTraitsAPI
}

func NewPersonaTraitsHandler(traits personaTraitsAPI) *PersonaTraitsHandler {
	return &PersonaTraitsHandler{traits: traits}
}

type personaTraitsResponse struct {
	Traits       model.PersonaTraits `json:"traits"`
	IsCustomized bool                `json:"is_customized"`
}

// Get handles GET /personas/{id}/traits.
func (h *PersonaTraitsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "giriş gerekli")
		return
	}
	personaID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(personaID); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz persona id")
		return
	}

	traits, isCustomized, err := h.traits.Get(r.Context(), userID, personaID)
	if err != nil {
		writePersonaTraitsError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, personaTraitsResponse{Traits: traits, IsCustomized: isCustomized})
}

type setPersonaTraitsRequest struct {
	Warmth     int `json:"warmth"`
	Humor      int `json:"humor"`
	Wisdom     int `json:"wisdom"`
	Directness int `json:"directness"`
	Energy     int `json:"energy"`
}

// Set handles PUT /personas/{id}/traits. All five dials are required
// every call — the client sends the full slider panel state, not a
// partial patch.
func (h *PersonaTraitsHandler) Set(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "giriş gerekli")
		return
	}
	personaID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(personaID); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz persona id")
		return
	}

	var req setPersonaTraitsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz istek gövdesi")
		return
	}
	traits := model.PersonaTraits{
		Warmth: req.Warmth, Humor: req.Humor, Wisdom: req.Wisdom,
		Directness: req.Directness, Energy: req.Energy,
	}

	if err := h.traits.Set(r.Context(), userID, personaID, traits); err != nil {
		writePersonaTraitsError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, personaTraitsResponse{Traits: traits, IsCustomized: true})
}

// Reset handles DELETE /personas/{id}/traits — back to the persona's
// own defaults.
func (h *PersonaTraitsHandler) Reset(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "giriş gerekli")
		return
	}
	personaID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(personaID); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz persona id")
		return
	}

	if err := h.traits.Reset(r.Context(), userID, personaID); err != nil {
		writePersonaTraitsError(w, err)
		return
	}
	traits, _, err := h.traits.Get(r.Context(), userID, personaID)
	if err != nil {
		writePersonaTraitsError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, personaTraitsResponse{Traits: traits, IsCustomized: false})
}

func writePersonaTraitsError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrPersonaNotFound):
		writeError(w, http.StatusNotFound, "persona bulunamadı")
	case errors.Is(err, model.ErrPersonaTraitOutOfRange):
		writeError(w, http.StatusBadRequest, "kişilik değerleri 0-100 arasında olmalıdır")
	default:
		writeError(w, http.StatusInternalServerError, "kişilik ayarları işlenemedi")
	}
}
