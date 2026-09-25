package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
	"perchly-backend/internal/service"
)

type onboardingProfileSaver interface {
	SaveProfile(ctx context.Context, profile model.OnboardingProfile, preferredName *string, skipHitap *bool) (model.OnboardingProfile, error)
	GetByUserID(ctx context.Context, userID string) (model.OnboardingProfile, error)
}

type OnboardingHandler struct {
	profiles onboardingProfileSaver
}

func NewOnboardingHandler(profiles onboardingProfileSaver) *OnboardingHandler {
	return &OnboardingHandler{profiles: profiles}
}

type saveOnboardingProfileRequest struct {
	AgeRange             string  `json:"age_range"`
	MoodPreference       *string `json:"mood_preference"`
	NotificationsGranted bool    `json:"notifications_granted"`
	SelectedPersonaID    *string `json:"selected_persona_id"`
	// PreferredName/SkipHitap are pointers deliberately: nil means the
	// field was absent (an older client), which OnboardingService needs
	// to tell apart from an explicit empty string / false — see its
	// SaveProfile doc comment.
	PreferredName *string `json:"preferred_name"`
	SkipHitap     *bool   `json:"skip_hitap"`
}

// SaveProfile godoc
// @Summary Onboarding profilini kaydet
// @Description Anonim kullanıcılar için de çalışır. is_minor asla client'tan okunmaz, age_range'den türetilir. preferred_name/skip_hitap opsiyoneldir: hiçbiri gönderilmezse eski istemci davranışı değişmez; skip_hitap=true ise preferred_name her zaman NULL'a zorlanır (sunucu asla takma ad üretmez); skip_hitap=false iken preferred_name boşsa 400 döner. Bilinmeyen alanlar yok sayılır.
// @Tags onboarding
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body saveOnboardingProfileRequest true "onboarding cevapları"
// @Success 200 {object} model.OnboardingProfile
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /users/onboarding-profile [post]
func (h *OnboardingHandler) SaveProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, r, http.StatusUnauthorized, ErrCodeUnauthorized, "giriş gerekli")
		return
	}

	var req saveOnboardingProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
		return
	}
	if req.SelectedPersonaID != nil {
		if _, err := uuid.Parse(*req.SelectedPersonaID); err != nil {
			writeError(w, r, http.StatusBadRequest, ErrCodeInvalidSelectedPersonaID, "geçersiz selected_persona_id")
			return
		}
	}

	profile := model.OnboardingProfile{
		UserID:               userID,
		AgeRange:             req.AgeRange,
		MoodPreference:       req.MoodPreference,
		NotificationsGranted: req.NotificationsGranted,
		SelectedPersonaID:    req.SelectedPersonaID,
	}

	saved, err := h.profiles.SaveProfile(r.Context(), profile, req.PreferredName, req.SkipHitap)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, saved)
	case errors.Is(err, service.ErrInvalidAgeRange):
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidAgeRange, "geçersiz yaş aralığı")
	case errors.Is(err, service.ErrInvalidMoodPreference):
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidMoodPreference, "geçersiz mod tercihi")
	case errors.Is(err, service.ErrPreferredNameRequired):
		writeError(w, r, http.StatusBadRequest, ErrCodePreferredNameRequired, "preferred_name gerekli veya skip_hitap=true gönderin")
	case errors.Is(err, service.ErrPreferredNameInvalid):
		writeError(w, r, http.StatusBadRequest, ErrCodePreferredNameInvalid, "preferred_name 1-40 karakter olmalı ve kontrol karakteri içermemeli")
	default:
		writeError(w, r, http.StatusInternalServerError, ErrCodeOnboardingSaveFailed, "profil kaydedilemedi")
	}
}

// GetProfile godoc
// @Summary Onboarding profilini getir
// @Description Yeniden kurulum veya ikinci bir cihazda daha önce verilen onboarding cevaplarını senkronize etmek için kullanılır.
// @Tags onboarding
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.OnboardingProfile
// @Failure 401 {object} errorResponse
// @Failure 404 {object} errorResponse "onboarding henüz tamamlanmamış"
// @Failure 500 {object} errorResponse
// @Router /users/onboarding-profile [get]
func (h *OnboardingHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, r, http.StatusUnauthorized, ErrCodeUnauthorized, "giriş gerekli")
		return
	}

	profile, err := h.profiles.GetByUserID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrOnboardingProfileNotFound) {
			writeError(w, r, http.StatusNotFound, ErrCodeOnboardingProfileNotFound, "onboarding profili bulunamadı")
			return
		}
		writeError(w, r, http.StatusInternalServerError, ErrCodeOnboardingFetchFailed, "profil getirilemedi")
		return
	}
	writeJSON(w, http.StatusOK, profile)
}
