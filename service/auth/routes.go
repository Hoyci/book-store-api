package auth

import (
	"context"
	"database/sql"
	"errors"
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
	UUIDGen   types.UUIDGenerator
}

func NewAuthHandler(
	userStore types.UserStore,
	authStore types.AuthStore,
	UUIDGen types.UUIDGenerator,
) *AuthHandler {
	return &AuthHandler{
		userStore: userStore,
		authStore: authStore,
		UUIDGen:   UUIDGen,
	}
}

func (h *AuthHandler) HandleUserLogin(w http.ResponseWriter, r *http.Request) {
	var requestPayload types.UserLoginPayload
	if err := utils.ParseJSON(r, &requestPayload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleUserLogin", "Body is not a valid json")
		return
	}

	if err := validate.Struct(requestPayload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is invalid: %s", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, err, "HandleUserLogin", errorMessages)
		return
	}

	user, err := h.userStore.GetByEmail(r.Context())
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleGetBooks", "Request canceled")
			return
		}

		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleUserLogin", fmt.Sprintf("No user found with email %s", requestPayload.Email))
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", "An unexpected error occurred")
		return
	}

	accessToken, err := utils.CreateJWT(user.ID, user.Username, user.Email, config.Envs.JWTSecret, 3600, h.UUIDGen)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", "An unexpected error occurred")
	}

	refreshToken, err := utils.CreateJWT(user.ID, user.Username, user.Email, config.Envs.JWTSecret, config.Envs.JWTExpirationInSeconds, h.UUIDGen)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", "An unexpected error occurred")
	}

	refreshTokenClaims, err := utils.VerifyJWT(refreshToken, config.Envs.JWTSecret)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", "An unexpected error occurred")
		return
	}

	err = h.authStore.UpsertRefreshToken(
		r.Context(),
		types.UpdateRefreshTokenPayload{
			UserID:    refreshTokenClaims.UserID,
			Jti:       refreshTokenClaims.RegisteredClaims.ID,
			ExpiresAt: refreshTokenClaims.RegisteredClaims.ExpiresAt.Time,
		},
	)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleGetBookByID", fmt.Sprintf("No userID found with ID %d", refreshTokenClaims.UserID))
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", "Internal Server Error")
		return
	}

	utils.WriteJSON(w, http.StatusOK, types.UserLoginResponse{AccessToken: accessToken, RefreshToken: refreshToken})
}

func (h *AuthHandler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	var requestPayload types.RefreshTokenPayload
	if err := utils.ParseJSON(r, &requestPayload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleUserLogin", "Body is not a valid json")
		return
	}

	if err := validate.Struct(requestPayload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is invalid: %s", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, err, "HandleUserLogin", errorMessages)
		return
	}

	claims, err := utils.VerifyJWT(requestPayload.RefreshToken, config.Envs.JWTSecret)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err, "HandleRefreshToken", "Refresh token is invalid or has been expired")
		return
	}

	storedToken, err := h.authStore.GetRefreshTokenByUserID(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleGetBooks", "Request canceled")
			return
		}

		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleRefreshToken", fmt.Sprintf("No refresh token found with user ID %d", claims.UserID))
			return
		}
		utils.WriteError(w, http.StatusUnauthorized, err, "HandleRefreshToken", "Unauthorized")
		return
	}

	if storedToken.Jti != claims.RegisteredClaims.ID {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("stored JTI does not match the claims ID"), "HandleRefreshToken", "Unauthorized")
		return
	}

	newAccessToken, err := utils.CreateJWT(claims.UserID, claims.Username, claims.Email, config.Envs.JWTSecret, 3600, h.UUIDGen)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", "An unexpected error occurred")
	}

	newRefreshToken, err := utils.CreateJWT(claims.UserID, claims.Username, claims.Email, config.Envs.JWTSecret, config.Envs.JWTExpirationInSeconds, h.UUIDGen)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", "An unexpected error occurred")
	}

	newRefreshTokenClaims, err := utils.VerifyJWT(newRefreshToken, config.Envs.JWTSecret)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", "An unexpected error occurred")
		return
	}

	err = h.authStore.UpsertRefreshToken(
		r.Context(),
		types.UpdateRefreshTokenPayload{
			UserID:    newRefreshTokenClaims.UserID,
			Jti:       newRefreshTokenClaims.RegisteredClaims.ID,
			ExpiresAt: newRefreshTokenClaims.RegisteredClaims.ExpiresAt.Time,
		})
	if err != nil {
		// TODO: adicionar errors.Is(err, context.Canceled) pode ser uma boa
		// TODO: Adicionar sql.ErrConnDone pode ser uma boa
		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleGetBookByID", fmt.Sprintf("No userID found with ID %d", newRefreshTokenClaims.UserID))
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, err, "HandleRefreshToken", "Internal Server Error")
		return
	}

	utils.WriteJSON(w, http.StatusOK, types.UpdateRefreshTokenResponse{AccessToken: newAccessToken, RefreshToken: newRefreshToken})
}
