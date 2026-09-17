package handler

import (
	"context"
	"net/http"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/service"
)

type chatEnergyGetter interface {
	ChatEnergy(ctx context.Context, userID string) (service.ChatEnergy, error)
}

type ChatEnergyHandler struct {
	quotas chatEnergyGetter
}

func NewChatEnergyHandler(quotas chatEnergyGetter) *ChatEnergyHandler {
	return &ChatEnergyHandler{quotas: quotas}
}

type chatEnergyResponse struct {
	Remaining int    `json:"remaining"`
	Limit     int    `json:"limit"`
	Used      int    `json:"used"`
	MoodLabel string `json:"mood_label"`
}

// Get godoc
// @Summary Günün sohbet enerjisi
// @Description Keşfet ekranındaki tek sayı: tüm aktif personalar arasında min(remaining) olarak hesaplanır. Salt okunur; gerçek mesaj gönderme uçlarındaki 429 davranışını etkilemez.
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} chatEnergyResponse
// @Failure 401 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /users/chat-energy [get]
func (h *ChatEnergyHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "giriş gerekli")
		return
	}

	energy, err := h.quotas.ChatEnergy(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, ErrCodeChatEnergyFailed, "sohbet enerjisi getirilemedi")
		return
	}

	writeJSON(w, http.StatusOK, chatEnergyResponse{
		Remaining: energy.Remaining,
		Limit:     energy.Limit,
		Used:      energy.Used,
		MoodLabel: energy.MoodLabel,
	})
}
