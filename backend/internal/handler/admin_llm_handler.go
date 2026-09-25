package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
	"perchly-backend/internal/service"
)

type adminLLMCredentialLister interface {
	ListCredentials(ctx context.Context) ([]model.LLMCredential, error)
}

type adminLLMCredentialGetter interface {
	GetCredential(ctx context.Context, id string) (model.LLMCredential, error)
}

type adminLLMCredentialCreator interface {
	CreateCredential(ctx context.Context, adminUserID string, in service.CreateCredentialInput) (model.LLMCredential, error)
}

type adminLLMCredentialUpdater interface {
	UpdateCredential(ctx context.Context, adminUserID, id string, in service.UpdateCredentialInput) (model.LLMCredential, error)
}

type adminLLMCredentialDeleter interface {
	DeleteCredential(ctx context.Context, adminUserID, id string) error
}

type adminLLMModelLister interface {
	ListModels(ctx context.Context, credentialID *string) ([]model.LLMModel, error)
}

type adminLLMModelCreator interface {
	CreateModel(ctx context.Context, adminUserID string, m model.LLMModel) (model.LLMModel, error)
}

type adminLLMModelUpdater interface {
	UpdateModel(ctx context.Context, adminUserID string, m model.LLMModel) (model.LLMModel, error)
}

type adminLLMModelDeleter interface {
	DeleteModel(ctx context.Context, adminUserID, id string) error
}

// AdminLLMHandler exposes the admin dashboard's CRUD over LLM provider
// credentials and the models made callable through them. A stored API
// key is never returned in full — every credential response carries
// only a masked preview (see toAdminLLMCredentialResponse).
type AdminLLMHandler struct {
	listCredentials  adminLLMCredentialLister
	getCredential    adminLLMCredentialGetter
	createCredential adminLLMCredentialCreator
	updateCredential adminLLMCredentialUpdater
	deleteCredential adminLLMCredentialDeleter
	listModels       adminLLMModelLister
	createModel      adminLLMModelCreator
	updateModel      adminLLMModelUpdater
	deleteModel      adminLLMModelDeleter
}

func NewAdminLLMHandler(llm *service.AdminLLMService) *AdminLLMHandler {
	return &AdminLLMHandler{
		listCredentials:  llm,
		getCredential:    llm,
		createCredential: llm,
		updateCredential: llm,
		deleteCredential: llm,
		listModels:       llm,
		createModel:      llm,
		updateModel:      llm,
		deleteModel:      llm,
	}
}

// adminLLMCredentialResponse mirrors model.LLMCredential but replaces
// the encrypted key material with a display-only masked preview built
// from APIKeyLast4 — the ciphertext/nonce never leave the server.
type adminLLMCredentialResponse struct {
	ID            string    `json:"id"`
	Provider      string    `json:"provider"`
	Label         string    `json:"label"`
	APIKeyPreview string    `json:"api_key_preview"`
	BaseURL       *string   `json:"base_url"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func toAdminLLMCredentialResponse(c model.LLMCredential) adminLLMCredentialResponse {
	return adminLLMCredentialResponse{
		ID: c.ID, Provider: c.Provider, Label: c.Label,
		APIKeyPreview: "••••" + c.APIKeyLast4,
		BaseURL:       c.BaseURL, IsActive: c.IsActive,
		CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

type adminLLMModelResponse struct {
	ID           string    `json:"id"`
	CredentialID string    `json:"credential_id"`
	ModelName    string    `json:"model_name"`
	DisplayName  string    `json:"display_name"`
	IsDefault    bool      `json:"is_default"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func toAdminLLMModelResponse(m model.LLMModel) adminLLMModelResponse {
	return adminLLMModelResponse{
		ID: m.ID, CredentialID: m.CredentialID, ModelName: m.ModelName, DisplayName: m.DisplayName,
		IsDefault: m.IsDefault, IsActive: m.IsActive, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

// --- GET /admin/llm/credentials ---

func (h *AdminLLMHandler) ListCredentials(w http.ResponseWriter, r *http.Request) {
	credentials, err := h.listCredentials.ListCredentials(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminLLMCredentialsListFailed, "sağlayıcı kimlik bilgileri listelenemedi")
		return
	}

	response := make([]adminLLMCredentialResponse, len(credentials))
	for i, c := range credentials {
		response[i] = toAdminLLMCredentialResponse(c)
	}
	writeJSON(w, http.StatusOK, response)
}

// --- GET /admin/llm/credentials/{id} ---

func (h *AdminLLMHandler) GetCredential(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeLLMCredentialNotFound, "geçersiz kimlik bilgisi id")
		return
	}

	credential, err := h.getCredential.GetCredential(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrLLMCredentialNotFound) {
			writeError(w, r, http.StatusNotFound, ErrCodeLLMCredentialNotFound, "kimlik bilgisi bulunamadı")
			return
		}
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminLLMCredentialFetchFailed, "kimlik bilgisi getirilemedi")
		return
	}
	writeJSON(w, http.StatusOK, toAdminLLMCredentialResponse(credential))
}

// --- POST /admin/llm/credentials ---

type adminCreateLLMCredentialRequest struct {
	Provider string  `json:"provider"`
	Label    string  `json:"label"`
	APIKey   string  `json:"api_key"`
	BaseURL  *string `json:"base_url"`
	IsActive bool    `json:"is_active"`
}

func (h *AdminLLMHandler) CreateCredential(w http.ResponseWriter, r *http.Request) {
	var req adminCreateLLMCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
		return
	}

	adminUserID, _ := auth.AdminUserIDFromContext(r.Context())

	created, err := h.createCredential.CreateCredential(r.Context(), adminUserID, service.CreateCredentialInput{
		Provider: req.Provider, Label: req.Label, APIKey: req.APIKey, BaseURL: req.BaseURL, IsActive: req.IsActive,
	})
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, toAdminLLMCredentialResponse(created))
	case errors.Is(err, service.ErrInvalidLLMCredentialInput):
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidLLMCredentialInput, "geçersiz kimlik bilgisi girdisi")
	default:
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminLLMCredentialCreateFailed, "kimlik bilgisi oluşturulamadı")
	}
}

// --- PUT /admin/llm/credentials/{id} ---

type adminUpdateLLMCredentialRequest struct {
	Label    string  `json:"label"`
	APIKey   *string `json:"api_key"`
	BaseURL  *string `json:"base_url"`
	IsActive bool    `json:"is_active"`
}

func (h *AdminLLMHandler) UpdateCredential(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeLLMCredentialNotFound, "geçersiz kimlik bilgisi id")
		return
	}

	var req adminUpdateLLMCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
		return
	}

	adminUserID, _ := auth.AdminUserIDFromContext(r.Context())

	updated, err := h.updateCredential.UpdateCredential(r.Context(), adminUserID, id, service.UpdateCredentialInput{
		Label: req.Label, APIKey: req.APIKey, BaseURL: req.BaseURL, IsActive: req.IsActive,
	})
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, toAdminLLMCredentialResponse(updated))
	case errors.Is(err, service.ErrInvalidLLMCredentialInput):
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidLLMCredentialInput, "geçersiz kimlik bilgisi girdisi")
	case errors.Is(err, repository.ErrLLMCredentialNotFound):
		writeError(w, r, http.StatusNotFound, ErrCodeLLMCredentialNotFound, "kimlik bilgisi bulunamadı")
	default:
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminLLMCredentialUpdateFailed, "kimlik bilgisi güncellenemedi")
	}
}

// --- DELETE /admin/llm/credentials/{id} ---

func (h *AdminLLMHandler) DeleteCredential(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeLLMCredentialNotFound, "geçersiz kimlik bilgisi id")
		return
	}

	adminUserID, _ := auth.AdminUserIDFromContext(r.Context())

	err := h.deleteCredential.DeleteCredential(r.Context(), adminUserID, id)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	case errors.Is(err, repository.ErrLLMCredentialNotFound):
		writeError(w, r, http.StatusNotFound, ErrCodeLLMCredentialNotFound, "kimlik bilgisi bulunamadı")
	case errors.Is(err, repository.ErrLLMCredentialInUse):
		writeError(w, r, http.StatusConflict, ErrCodeLLMCredentialInUse, "önce bu kimlik bilgisine bağlı modelleri silin")
	default:
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminLLMCredentialDeleteFailed, "kimlik bilgisi silinemedi")
	}
}

// --- GET /admin/llm/models ---

func (h *AdminLLMHandler) ListModels(w http.ResponseWriter, r *http.Request) {
	var credentialID *string
	if v := r.URL.Query().Get("credential_id"); v != "" {
		if _, err := uuid.Parse(v); err != nil {
			writeError(w, r, http.StatusBadRequest, ErrCodeLLMCredentialNotFound, "geçersiz kimlik bilgisi id")
			return
		}
		credentialID = &v
	}

	models, err := h.listModels.ListModels(r.Context(), credentialID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminLLMModelsListFailed, "modeller listelenemedi")
		return
	}

	response := make([]adminLLMModelResponse, len(models))
	for i, m := range models {
		response[i] = toAdminLLMModelResponse(m)
	}
	writeJSON(w, http.StatusOK, response)
}

// --- POST /admin/llm/models ---

type adminCreateLLMModelRequest struct {
	CredentialID string `json:"credential_id"`
	ModelName    string `json:"model_name"`
	DisplayName  string `json:"display_name"`
	IsDefault    bool   `json:"is_default"`
	IsActive     bool   `json:"is_active"`
}

func (h *AdminLLMHandler) CreateModel(w http.ResponseWriter, r *http.Request) {
	var req adminCreateLLMModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
		return
	}
	if _, err := uuid.Parse(req.CredentialID); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeLLMCredentialNotFound, "geçersiz kimlik bilgisi id")
		return
	}

	adminUserID, _ := auth.AdminUserIDFromContext(r.Context())

	created, err := h.createModel.CreateModel(r.Context(), adminUserID, model.LLMModel{
		CredentialID: req.CredentialID, ModelName: req.ModelName, DisplayName: req.DisplayName,
		IsDefault: req.IsDefault, IsActive: req.IsActive,
	})
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, toAdminLLMModelResponse(created))
	case errors.Is(err, service.ErrInvalidLLMModelInput):
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidLLMModelInput, "geçersiz model girdisi")
	case errors.Is(err, repository.ErrLLMCredentialNotFound):
		writeError(w, r, http.StatusNotFound, ErrCodeLLMCredentialNotFound, "kimlik bilgisi bulunamadı")
	case errors.Is(err, service.ErrLLMCredentialInactive):
		writeError(w, r, http.StatusBadRequest, ErrCodeLLMCredentialInactive, "kimlik bilgisi pasif durumda")
	default:
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminLLMModelCreateFailed, "model oluşturulamadı")
	}
}

// --- PUT /admin/llm/models/{id} ---

type adminUpdateLLMModelRequest struct {
	DisplayName string `json:"display_name"`
	IsDefault   bool   `json:"is_default"`
	IsActive    bool   `json:"is_active"`
}

func (h *AdminLLMHandler) UpdateModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeLLMModelNotFound, "geçersiz model id")
		return
	}

	var req adminUpdateLLMModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
		return
	}

	adminUserID, _ := auth.AdminUserIDFromContext(r.Context())

	updated, err := h.updateModel.UpdateModel(r.Context(), adminUserID, model.LLMModel{
		ID: id, DisplayName: req.DisplayName, IsDefault: req.IsDefault, IsActive: req.IsActive,
	})
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, toAdminLLMModelResponse(updated))
	case errors.Is(err, service.ErrInvalidLLMModelInput):
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidLLMModelInput, "geçersiz model girdisi")
	case errors.Is(err, repository.ErrLLMModelNotFound):
		writeError(w, r, http.StatusNotFound, ErrCodeLLMModelNotFound, "model bulunamadı")
	default:
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminLLMModelUpdateFailed, "model güncellenemedi")
	}
}

// --- DELETE /admin/llm/models/{id} ---

func (h *AdminLLMHandler) DeleteModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeLLMModelNotFound, "geçersiz model id")
		return
	}

	adminUserID, _ := auth.AdminUserIDFromContext(r.Context())

	err := h.deleteModel.DeleteModel(r.Context(), adminUserID, id)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	case errors.Is(err, repository.ErrLLMModelNotFound):
		writeError(w, r, http.StatusNotFound, ErrCodeLLMModelNotFound, "model bulunamadı")
	default:
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminLLMModelDeleteFailed, "model silinemedi")
	}
}
