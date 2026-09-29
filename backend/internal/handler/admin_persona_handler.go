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

type adminPersonaTranslationLister interface {
	ListTranslations(ctx context.Context, personaID string) ([]model.PersonaTranslation, error)
}

type adminPersonaTranslationUpserter interface {
	UpsertTranslation(ctx context.Context, adminUserID, personaID, locale string, t model.PersonaTranslation) (model.PersonaTranslation, error)
}

type adminPersonaTranslationDeleter interface {
	DeleteTranslation(ctx context.Context, adminUserID, personaID, locale string) error
}

// AdminPersonaHandler exposes the admin dashboard's full persona CRUD —
// unlike PersonaHandler (public, read-only, active-only, no
// system_prompt), this can see and write everything.
type AdminPersonaHandler struct {
	list              adminPersonaLister
	get               adminPersonaGetter
	create            adminPersonaCreator
	update            adminPersonaUpdater
	listTranslations  adminPersonaTranslationLister
	upsertTranslation adminPersonaTranslationUpserter
	deleteTranslation adminPersonaTranslationDeleter
}

func NewAdminPersonaHandler(personas *service.AdminPersonaService) *AdminPersonaHandler {
	return &AdminPersonaHandler{
		list: personas, get: personas, create: personas, update: personas,
		listTranslations: personas, upsertTranslation: personas, deleteTranslation: personas,
	}
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
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminPersonasListFailed, "personalar listelenemedi")
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
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidPersonaID, "geçersiz persona id")
		return
	}

	persona, err := h.get.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrPersonaNotFound) {
			writeError(w, r, http.StatusNotFound, ErrCodePersonaNotFound, "persona bulunamadı")
			return
		}
		writeError(w, r, http.StatusInternalServerError, ErrCodePersonaFetchFailed, "persona getirilemedi")
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
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
		return
	}
	if !validPersonaLLMModelID(req.LLMModelID) {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidPersonaLLMModelID, "geçersiz model id")
		return
	}

	adminUserID, _ := auth.AdminUserIDFromContext(r.Context())

	created, err := h.create.Create(r.Context(), adminUserID, req.toPersona(""))
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, toAdminPersonaResponse(created))
	case errors.Is(err, service.ErrInvalidPersonaInput), errors.Is(err, model.ErrPersonaTraitOutOfRange):
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidPersonaInput, "geçersiz persona girdisi")
	case errors.Is(err, service.ErrPersonaLLMModelNotFound):
		writeError(w, r, http.StatusBadRequest, ErrCodePersonaLLMModelNotFound, "seçilen model bulunamadı")
	case errors.Is(err, service.ErrPersonaLLMModelInactive):
		writeError(w, r, http.StatusBadRequest, ErrCodePersonaLLMModelInactive, "seçilen model veya kimlik bilgisi pasif durumda")
	default:
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminPersonaCreateFailed, "persona oluşturulamadı")
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
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidPersonaID, "geçersiz persona id")
		return
	}

	var req adminPersonaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
		return
	}
	if !validPersonaLLMModelID(req.LLMModelID) {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidPersonaLLMModelID, "geçersiz model id")
		return
	}

	adminUserID, _ := auth.AdminUserIDFromContext(r.Context())

	updated, err := h.update.Update(r.Context(), adminUserID, req.toPersona(id))
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, toAdminPersonaResponse(updated))
	case errors.Is(err, service.ErrInvalidPersonaInput), errors.Is(err, model.ErrPersonaTraitOutOfRange):
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidPersonaInput, "geçersiz persona girdisi")
	case errors.Is(err, service.ErrPersonaLLMModelNotFound):
		writeError(w, r, http.StatusBadRequest, ErrCodePersonaLLMModelNotFound, "seçilen model bulunamadı")
	case errors.Is(err, service.ErrPersonaLLMModelInactive):
		writeError(w, r, http.StatusBadRequest, ErrCodePersonaLLMModelInactive, "seçilen model veya kimlik bilgisi pasif durumda")
	case errors.Is(err, repository.ErrPersonaNotFound):
		writeError(w, r, http.StatusNotFound, ErrCodePersonaNotFound, "persona bulunamadı")
	default:
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminPersonaUpdateFailed, "persona güncellenemedi")
	}
}

// adminPersonaTranslationResponse mirrors model.PersonaTranslation.
type adminPersonaTranslationResponse struct {
	PersonaID        string    `json:"persona_id"`
	Locale           string    `json:"locale"`
	Name             string    `json:"name"`
	ShortDescription string    `json:"short_description"`
	ToneDescription  string    `json:"tone_description"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func toAdminPersonaTranslationResponse(t model.PersonaTranslation) adminPersonaTranslationResponse {
	return adminPersonaTranslationResponse{
		PersonaID: t.PersonaID, Locale: t.Locale, Name: t.Name,
		ShortDescription: t.ShortDescription, ToneDescription: t.ToneDescription,
		CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
}

// --- GET /admin/personas/{id}/translations ---

// ListTranslations godoc
// @Summary Persona çevirilerini listele
// @Description Bir persona için kayıtlı tüm çevirileri döner (tr hariç — Türkçe hiçbir zaman ayrı bir satır olarak saklanmaz).
// @Tags admin-personas
// @Produce json
// @Security BearerAuth
// @Param id path string true "Persona ID (UUID)"
// @Success 200 {array} adminPersonaTranslationResponse
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /admin/personas/{id}/translations [get]
func (h *AdminPersonaHandler) ListTranslations(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidPersonaID, "geçersiz persona id")
		return
	}

	translations, err := h.listTranslations.ListTranslations(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrPersonaNotFound) {
			writeError(w, r, http.StatusNotFound, ErrCodePersonaNotFound, "persona bulunamadı")
			return
		}
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminPersonaTranslationsListFailed, "persona çevirileri listelenemedi")
		return
	}

	response := make([]adminPersonaTranslationResponse, len(translations))
	for i, t := range translations {
		response[i] = toAdminPersonaTranslationResponse(t)
	}
	writeJSON(w, http.StatusOK, response)
}

// adminPersonaTranslationRequest is the editable translation shape for
// PUT /admin/personas/{id}/translations/{locale} — persona_id/locale
// come from the URL, not the body.
type adminPersonaTranslationRequest struct {
	Name             string `json:"name"`
	ShortDescription string `json:"short_description"`
	ToneDescription  string `json:"tone_description"`
}

// --- PUT /admin/personas/{id}/translations/{locale} ---

// PutTranslation godoc
// @Summary Persona çevirisini oluştur veya güncelle
// @Description name/short_description/tone_description için upsert yapar. locale "tr" olamaz — Türkçe her zaman personas tablosundaki temel sütunlardan okunur.
// @Tags admin-personas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Persona ID (UUID)"
// @Param locale path string true "Çeviri dili (tr hariç desteklenen bir dil: en, de, es, fr, ru, zh, ar)"
// @Param body body adminPersonaTranslationRequest true "name, short_description, tone_description"
// @Success 200 {object} adminPersonaTranslationResponse
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /admin/personas/{id}/translations/{locale} [put]
func (h *AdminPersonaHandler) PutTranslation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidPersonaID, "geçersiz persona id")
		return
	}
	locale := chi.URLParam(r, "locale")

	var req adminPersonaTranslationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
		return
	}

	adminUserID, _ := auth.AdminUserIDFromContext(r.Context())

	result, err := h.upsertTranslation.UpsertTranslation(r.Context(), adminUserID, id, locale, model.PersonaTranslation{
		Name: req.Name, ShortDescription: req.ShortDescription, ToneDescription: req.ToneDescription,
	})
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, toAdminPersonaTranslationResponse(result))
	case errors.Is(err, service.ErrInvalidPersonaTranslationLocale):
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidPersonaTranslationLocale, "geçersiz veya desteklenmeyen çeviri dili (tr çeviri olarak saklanamaz)")
	case errors.Is(err, service.ErrInvalidPersonaTranslationInput):
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidPersonaTranslationInput, "geçersiz çeviri girdisi")
	case errors.Is(err, repository.ErrPersonaNotFound):
		writeError(w, r, http.StatusNotFound, ErrCodePersonaNotFound, "persona bulunamadı")
	default:
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminPersonaTranslationUpsertFailed, "persona çevirisi kaydedilemedi")
	}
}

// --- DELETE /admin/personas/{id}/translations/{locale} ---

// DeleteTranslation godoc
// @Summary Persona çevirisini sil
// @Description Belirtilen dildeki çeviriyi kaldırır — persona o dilde tekrar Türkçe içeriğe döner. Çeviri zaten yoksa da başarıyla döner (idempotent).
// @Tags admin-personas
// @Produce json
// @Security BearerAuth
// @Param id path string true "Persona ID (UUID)"
// @Param locale path string true "Çeviri dili (tr hariç desteklenen bir dil)"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /admin/personas/{id}/translations/{locale} [delete]
func (h *AdminPersonaHandler) DeleteTranslation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidPersonaID, "geçersiz persona id")
		return
	}
	locale := chi.URLParam(r, "locale")

	adminUserID, _ := auth.AdminUserIDFromContext(r.Context())

	err := h.deleteTranslation.DeleteTranslation(r.Context(), adminUserID, id, locale)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	case errors.Is(err, service.ErrInvalidPersonaTranslationLocale):
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidPersonaTranslationLocale, "geçersiz veya desteklenmeyen çeviri dili (tr çeviri olarak saklanamaz)")
	case errors.Is(err, repository.ErrPersonaNotFound):
		writeError(w, r, http.StatusNotFound, ErrCodePersonaNotFound, "persona bulunamadı")
	default:
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminPersonaTranslationDeleteFailed, "persona çevirisi silinemedi")
	}
}
