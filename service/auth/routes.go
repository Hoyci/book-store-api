package auth

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/types"
	"github.com/hoyci/book-store-api/utils"
)

var validate = validator.New()

type AuthHandler struct {
	userStore types.UserStore
	authStore types.AuthStore
}

func NewAuthHandler(userStore types.UserStore, authStore types.AuthStore) *AuthHandler {
	return &AuthHandler{
		userStore: userStore,
		authStore: authStore,
	}
}

func (h *AuthHandler) HandleUserLogin(w http.ResponseWriter, r *http.Request) {
	var requestPayload types.UserLoginPayload

	if err := utils.ParseJSON(r, &requestPayload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleUserLogin", "User sent request with an invalid JSON", "Body is not a valid json")
		return
	}

	if err := validate.Struct(requestPayload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is invalid: %s", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, err, "HandleUserLogin", "User sent a request containing JSON with information outside the permitted format", errorMessages)
		return
	}

	user, err := h.userStore.GetByEmail(r.Context(), requestPayload.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", "Failed get user by email from database", "An unexpected error occurred")
		return
	}

	accessToken, err := utils.CreateJWT(user.ID, user.Username, user.Email, config.Envs.JWTSecret, 3600)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", "An error occured during the create JWT process", "An unexpected error occurred")
	}

	refreshToken, err := utils.CreateJWT(user.ID, "", "", config.Envs.JWTSecret, config.Envs.JWTExpirationInSeconds)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", "An error occured during the create JWT process", "An unexpected error occurred")
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"access_token": accessToken, "refresh_token": refreshToken})
}

// TODO: Adicionar rotação de refresh token
func (h *AuthHandler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	var requestPayload types.RefreshTokenPayload

	if err := utils.ParseJSON(r, &requestPayload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleUserLogin", "User sent request with an invalid JSON", "Body is not a valid json")
		return
	}

	if err := validate.Struct(requestPayload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is invalid: %s", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, err, "HandleUserLogin", "User sent a request containing JSON with information outside the permitted format", errorMessages)
		return
	}

	claims, err := utils.VerifyJWT(requestPayload.AccessToken, config.Envs.JWTSecret)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err, "HandleRefreshToken", "User sent an invalid or expired refresh token", "Refresh token is invalid or has been expired")
		return
	}

	storedToken, err := h.authStore.GetRefreshTokenByUserID(r.Context(), claims.UserID)
	if err != nil || storedToken.Jti != claims.ID {
		utils.WriteError(w, http.StatusUnauthorized, err, "HandleRefreshToken", "Invalid or revoked refresh token", "Unauthorized")
		return
	}

	newAccessToken, err := utils.CreateJWT(claims.UserID, claims.Username, claims.Email, config.Envs.JWTSecret, 3600)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", "An error occured during the create JWT process", "An unexpected error occurred")
	}

	newRefreshToken, err := utils.CreateJWT(claims.UserID, "", "", config.Envs.JWTSecret, config.Envs.JWTExpirationInSeconds)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", "An error occured during the create JWT process", "An unexpected error occurred")
	}

	err = h.authStore.UpdateRefreshTokenByUserID(
		r.Context(),
		types.UpdateRefreshTokenPayload{
			UserID:    claims.UserID,
			Jti:       newRefreshToken, // Falta descobrir como eu pego o JTI do novo token
			ExpiresAt: newRefreshToken, // Falta descobrir como eu pego o expires_at do novo token
		})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleRefreshToken", "Failed to update refresh token", "Internal Server Error")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"access_token": newAccessToken, "refresh_token": newRefreshToken})
}
