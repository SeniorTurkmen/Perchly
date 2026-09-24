package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
	"perchly-backend/internal/service"
)

type adminPersonaLister interface {
	List(ctx context.Context) ([]model.Persona, error)
}

type adminPersonaGetter interface {
	Get(ctx context.Context, id string) (model.Persona, error)
}

type adminPersonaCreator interface {
	Create(ctx context.Context, adminUserID string, p model.Persona) (model.Persona, error)
}

type adminPersonaUpdater interface {
	Update(ctx context.Context, adminUserID string, p model.Persona) (model.Persona, error)
}

// AdminPersonaHandler exposes the admin dashboard's full persona CRUD —
// unlike PersonaHandler (public, read-only, active-only, no
// system_prompt), this can see and write everything.
type AdminPersonaHandler struct {
	list   adminPersonaLister
	get    adminPersonaGetter
	create adminPersonaCreator
	update adminPersonaUpdater
}

func NewAdminPersonaHandler(personas *service.AdminPersonaService) *AdminPersonaHandler {
	return &AdminPersonaHandler{list: personas, get: personas, create: personas, update: personas}
}

// adminPersonaResponse mirrors model.Persona but with SystemPrompt
// included — that field is `json:"-"` on the model itself specifically
// so it never leaks over the public /personas API (see the model's doc
// comment); the admin dashboard is the one place it's meant to be
// readable and editable, so every admin response uses this shape
// instead of serializing model.Persona directly.
type adminPersonaResponse struct {
	ID                 string              `json:"id"`
	Slug               string              `json:"slug"`
	Name               string              `json:"name"`
	Category           string              `json:"category"`
	ShortDescription   string              `json:"short_description"`
	SystemPrompt       string              `json:"system_prompt"`
	ToneDescription    string              `json:"tone_description"`
	AvatarURL          *string             `json:"avatar_url"`
	AccentColor        string              `json:"accent_color"`
	IsMinorAppropriate bool                `json:"is_minor_appropriate"`
	IsActive           bool                `json:"is_active"`
	SortOrder          int                 `json:"sort_order"`
	DefaultTraits      model.PersonaTraits `json:"default_traits"`
	// LLMModelID is nil when this persona uses the process-wide
	// LLM_PROVIDER/LLM_MODEL default instead of a specific stored model.
	LLMModelID *string   `json:"llm_model_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func toAdminPersonaResponse(p model.Persona) adminPersonaResponse {
	return adminPersonaResponse{
		ID: p.ID, Slug: p.Slug, Name: p.Name, Category: p.Category,
		ShortDescription: p.ShortDescription, SystemPrompt: p.SystemPrompt, ToneDescription: p.ToneDescription,
		AvatarURL: p.AvatarURL, AccentColor: p.AccentColor, IsMinorAppropriate: p.IsMinorAppropriate,
		IsActive: p.IsActive, SortOrder: p.SortOrder, DefaultTraits: p.DefaultTraits,
		LLMModelID: p.LLMModelID,
		CreatedAt:  p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

// --- GET /admin/personas ---

func (h *AdminPersonaHandler) List(w http.ResponseWriter, r *http.Request) {
	personas, err := h.list.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, ErrCodeAdminPersonasListFailed, "personalar listelenemedi")
		return
	}

	response := make([]adminPersonaResponse, len(personas))
	for i, p := range personas {
		response[i] = toAdminPersonaResponse(p)
	}
	writeJSON(w, http.StatusOK, response)
}

// --- GET /admin/personas/{id} ---

func (h *AdminPersonaHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidPersonaID, "geçersiz persona id")
		return
	}

	persona, err := h.get.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrPersonaNotFound) {
			writeError(w, http.StatusNotFound, ErrCodePersonaNotFound, "persona bulunamadı")
			return
		}
		writeError(w, http.StatusInternalServerError, ErrCodePersonaFetchFailed, "persona getirilemedi")
		return
	}
	writeJSON(w, http.StatusOK, toAdminPersonaResponse(persona))
}

// adminPersonaRequest is the full editable persona shape — used for
// both create and update, matching how the admin form always submits
// the complete record.
type adminPersonaRequest struct {
	Slug               string              `json:"slug"`
	Name               string              `json:"name"`
	Category           string              `json:"category"`
	ShortDescription   string              `json:"short_description"`
	SystemPrompt       string              `json:"system_prompt"`
	ToneDescription    string              `json:"tone_description"`
	AvatarURL          *string             `json:"avatar_url"`
	AccentColor        string              `json:"accent_color"`
	IsMinorAppropriate bool                `json:"is_minor_appropriate"`
	IsActive           bool                `json:"is_active"`
	SortOrder          int                 `json:"sort_order"`
	DefaultTraits      model.PersonaTraits `json:"default_traits"`
	// LLMModelID is a *string (rather than string) so the admin form can
	// distinguish "explicitly pinned to this model" from "cleared back
	// to the default" — an empty string in the JSON body means the
	// latter and is normalized to nil in toPersona, same reasoning as
	// AvatarURL elsewhere in this struct.
	LLMModelID *string `json:"llm_model_id"`
}

func (req adminPersonaRequest) toPersona(id string) model.Persona {
	llmModelID := req.LLMModelID
	if llmModelID != nil && strings.TrimSpace(*llmModelID) == "" {
		llmModelID = nil
	}

	return model.Persona{
		ID:                 id,
		Slug:               req.Slug,
		Name:               req.Name,
		Category:           req.Category,
		ShortDescription:   req.ShortDescription,
		SystemPrompt:       req.SystemPrompt,
		ToneDescription:    req.ToneDescription,
		AvatarURL:          req.AvatarURL,
		AccentColor:        req.AccentColor,
		IsMinorAppropriate: req.IsMinorAppropriate,
		IsActive:           req.IsActive,
		SortOrder:          req.SortOrder,
		DefaultTraits:      req.DefaultTraits,
		LLMModelID:         llmModelID,
	}
}

// --- POST /admin/personas ---

func (h *AdminPersonaHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req adminPersonaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
		return
	}
	if !validPersonaLLMModelID(req.LLMModelID) {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidPersonaLLMModelID, "geçersiz model id")
		return
	}

	adminUserID, _ := auth.AdminUserIDFromContext(r.Context())

	created, err := h.create.Create(r.Context(), adminUserID, req.toPersona(""))
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, toAdminPersonaResponse(created))
	case errors.Is(err, service.ErrInvalidPersonaInput), errors.Is(err, model.ErrPersonaTraitOutOfRange):
		writeError(w, http.StatusBadRequest, ErrCodeInvalidPersonaInput, "geçersiz persona girdisi")
	case errors.Is(err, service.ErrPersonaLLMModelNotFound):
		writeError(w, http.StatusBadRequest, ErrCodePersonaLLMModelNotFound, "seçilen model bulunamadı")
	case errors.Is(err, service.ErrPersonaLLMModelInactive):
		writeError(w, http.StatusBadRequest, ErrCodePersonaLLMModelInactive, "seçilen model veya kimlik bilgisi pasif durumda")
	default:
		writeError(w, http.StatusInternalServerError, ErrCodeAdminPersonaCreateFailed, "persona oluşturulamadı")
	}
}

// validPersonaLLMModelID allows nil or an empty string (both mean "no
// override"), otherwise requires a well-formed UUID — rejected early so
// a malformed value never reaches the repository layer as a raw ::uuid
// cast, which would surface as an opaque 500 instead of a clean 400.
func validPersonaLLMModelID(id *string) bool {
	if id == nil || strings.TrimSpace(*id) == "" {
		return true
	}
	_, err := uuid.Parse(*id)
	return err == nil
}

// --- PUT /admin/personas/{id} ---

func (h *AdminPersonaHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidPersonaID, "geçersiz persona id")
		return
	}

	var req adminPersonaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
		return
	}
	if !validPersonaLLMModelID(req.LLMModelID) {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidPersonaLLMModelID, "geçersiz model id")
		return
	}

	adminUserID, _ := auth.AdminUserIDFromContext(r.Context())

	updated, err := h.update.Update(r.Context(), adminUserID, req.toPersona(id))
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, toAdminPersonaResponse(updated))
	case errors.Is(err, service.ErrInvalidPersonaInput), errors.Is(err, model.ErrPersonaTraitOutOfRange):
		writeError(w, http.StatusBadRequest, ErrCodeInvalidPersonaInput, "geçersiz persona girdisi")
	case errors.Is(err, service.ErrPersonaLLMModelNotFound):
		writeError(w, http.StatusBadRequest, ErrCodePersonaLLMModelNotFound, "seçilen model bulunamadı")
	case errors.Is(err, service.ErrPersonaLLMModelInactive):
		writeError(w, http.StatusBadRequest, ErrCodePersonaLLMModelInactive, "seçilen model veya kimlik bilgisi pasif durumda")
	case errors.Is(err, repository.ErrPersonaNotFound):
		writeError(w, http.StatusNotFound, ErrCodePersonaNotFound, "persona bulunamadı")
	default:
		writeError(w, http.StatusInternalServerError, ErrCodeAdminPersonaUpdateFailed, "persona güncellenemedi")
	}
}
