package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/model"
	"perchly-backend/internal/service"
)

type onboardingProfileSaver interface {
	SaveProfile(ctx context.Context, profile model.OnboardingProfile) (model.OnboardingProfile, error)
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
}

// SaveProfile handles POST /users/onboarding-profile. Collected for
// anonymous users too — an anonymous session upgrading to a verified
// email later keeps the same user id, so nothing here is lost.
// IsMinor is deliberately not read from the request body — see
// OnboardingService.SaveProfile.
func (h *OnboardingHandler) SaveProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "giriş gerekli")
		return
	}

	var req saveOnboardingProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz istek gövdesi")
		return
	}
	if req.SelectedPersonaID != nil {
		if _, err := uuid.Parse(*req.SelectedPersonaID); err != nil {
			writeError(w, http.StatusBadRequest, "geçersiz selected_persona_id")
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

	saved, err := h.profiles.SaveProfile(r.Context(), profile)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, saved)
	case errors.Is(err, service.ErrInvalidAgeRange):
		writeError(w, http.StatusBadRequest, "geçersiz yaş aralığı")
	case errors.Is(err, service.ErrInvalidMoodPreference):
		writeError(w, http.StatusBadRequest, "geçersiz mod tercihi")
	default:
		writeError(w, http.StatusInternalServerError, "profil kaydedilemedi")
	}
}
