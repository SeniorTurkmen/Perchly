package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
)

type personaFinder interface {
	List(ctx context.Context) ([]model.Persona, error)
	GetByID(ctx context.Context, id string) (model.Persona, error)
	ListWithRecommendation(ctx context.Context, userID string) ([]model.PersonaRecommendation, error)
}

type PersonaHandler struct {
	personas personaFinder
}

func NewPersonaHandler(personas personaFinder) *PersonaHandler {
	return &PersonaHandler{personas: personas}
}

// List handles GET /personas and, with ?recommend=true, the
// personalized variant that adds a recommended/match_reason to each
// persona for the caller. That variant requires auth (mounted behind
// auth.OptionalMiddleware, which — unlike auth.Middleware — never
// rejects a request for lacking a token; this handler is what actually
// enforces it, only when recommend=true is asked for) — plain GET
// /personas stays exactly as public and unpersonalized as before.
// List godoc
// @Summary Personaları listele
// @Description Herkese açık liste. ?recommend=true ile (auth zorunlu) her personaya recommended/match_reason eklenmiş, kullanıcının mood_preference'ına göre kişiselleştirilmiş liste döner; tam olarak bir persona recommended:true olur, sayısal eşleşme % asla üretilmez.
// @Tags personas
// @Produce json
// @Param recommend query string false "true verilirse auth zorunlu, kişiselleştirilmiş öneri döner"
// @Success 200 {array} model.Persona
// @Failure 401 {object} errorResponse "recommend=true iken auth eksik"
// @Failure 500 {object} errorResponse
// @Router /personas [get]
func (h *PersonaHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("recommend") == "true" {
		userID, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			writeError(w, r, http.StatusUnauthorized, ErrCodeUnauthorized, "giriş gerekli")
			return
		}
		recommended, err := h.personas.ListWithRecommendation(r.Context(), userID)
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, ErrCodePersonasListFailed, "personas listelenemedi")
			return
		}
		writeJSON(w, http.StatusOK, recommended)
		return
	}

	personas, err := h.personas.List(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, ErrCodePersonasListFailed, "personas listelenemedi")
		return
	}
	writeJSON(w, http.StatusOK, personas)
}

// Get handles GET /personas/{id}.
// Get godoc
// @Summary Tek bir personayı getir
// @Tags personas
// @Produce json
// @Param id path string true "Persona ID (UUID)"
// @Success 200 {object} model.Persona
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /personas/{id} [get]
func (h *PersonaHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidPersonaID, "geçersiz persona id")
		return
	}

	persona, err := h.personas.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrPersonaNotFound) {
			writeError(w, r, http.StatusNotFound, ErrCodePersonaNotFound, "persona bulunamadı")
			return
		}
		writeError(w, r, http.StatusInternalServerError, ErrCodePersonaFetchFailed, "persona getirilemedi")
		return
	}
	writeJSON(w, http.StatusOK, persona)
}
