package auth

import (
	"database/sql"
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

// @Summary Realizar login do usuário
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body types.UserLoginPayload true "Dados para login do usuário"
// @Success 200 {object} types.UserLoginResponse "Tokens de acesso e refresh"
// @Failure 400 {object} types.BadRequestResponse "Body is not a valid json"
// @Failure 400 {object} types.BadRequestStructResponse "Validation errors for payload"
// @Failure 404 {object} types.NotFoundResponse "No user found with the given email"
// @Failure 500 {object} types.InternalServerErrorResponse "An unexpected error occurred"
// @Router /auth/login [post]
func (h *AuthHandler) HandleUserLogin(w http.ResponseWriter, r *http.Request) {
	var requestPayload types.UserLoginPayload
	if err := utils.ParseJSON(r, &requestPayload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleUserLogin", types.BadRequestResponse{Error: "Body is not a valid json"})
		return
	}

	if err := validate.Struct(requestPayload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is invalid: %s", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, err, "HandleUserLogin", types.BadRequestStructResponse{Error: errorMessages})
		return
	}

	user, err := h.userStore.GetByEmail(r.Context(), requestPayload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleUserLogin", types.NotFoundResponse{Error: fmt.Sprintf("No user found with email %s", requestPayload.Email)})
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
		return
	}

	err = utils.CheckPassword(r.Context(), user.PasswordHash, requestPayload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", types.InternalServerErrorResponse{Error: "The email or password is incorrect"})
	}

	accessToken, err := utils.CreateJWT(user.ID, user.Username, user.Email, config.Envs.JWTSecret, 3600, h.UUIDGen)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
	}

	refreshToken, err := utils.CreateJWT(user.ID, user.Username, user.Email, config.Envs.JWTSecret, config.Envs.JWTExpirationInSeconds, h.UUIDGen)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
	}

	refreshTokenClaims, err := utils.VerifyJWT(refreshToken, config.Envs.JWTSecret)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
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
			utils.WriteError(w, http.StatusNotFound, err, "HandleGetBookByID", types.NotFoundResponse{Error: fmt.Sprintf("No userID found with ID %d", refreshTokenClaims.UserID)})
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, types.UserLoginResponse{AccessToken: accessToken, RefreshToken: refreshToken})
}

// @Summary Atualizar tokens (Refresh Token)
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body types.RefreshTokenPayload true "Payload contendo o refresh token"
// @Success 200 {object} types.UpdateRefreshTokenResponse "Novos tokens de acesso e refresh"
// @Failure 400 {object} types.BadRequestResponse "Body is not a valid json"
// @Failure 400 {object} types.BadRequestStructResponse "Validation errors for payload"
// @Failure 401 {object} types.UnauthorizedResponse "Refresh token is invalid or has been expired"
// @Failure 404 {object} types.NotFoundResponse "No refresh token found with the given user ID"
// @Failure 500 {object} types.InternalServerErrorResponse "An unexpected error occurred"
// @Router /auth/refresh [post]
func (h *AuthHandler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	var requestPayload types.RefreshTokenPayload
	if err := utils.ParseJSON(r, &requestPayload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleUserLogin", types.BadRequestResponse{Error: "Body is not a valid json"})
		return
	}

	if err := validate.Struct(requestPayload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is invalid: %s", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, err, "HandleUserLogin", types.BadRequestStructResponse{Error: errorMessages})
		return
	}

	claims, err := utils.VerifyJWT(requestPayload.RefreshToken, config.Envs.JWTSecret)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, err, "HandleRefreshToken", types.UnauthorizedResponse{Error: "Refresh token is invalid or has been expired"})
		return
	}

	storedToken, err := h.authStore.GetRefreshTokenByUserID(r.Context(), claims.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleRefreshToken", types.NotFoundResponse{Error: fmt.Sprintf("No refresh token found with user ID %d", claims.UserID)})
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleRefreshToken", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
		return
	}

	if storedToken.Jti != claims.RegisteredClaims.ID {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("stored JTI does not match the claims ID"), "HandleRefreshToken", types.UnauthorizedResponse{Error: "Refresh token is invalid or has been expired"})
		return
	}

	newAccessToken, err := utils.CreateJWT(claims.UserID, claims.Username, claims.Email, config.Envs.JWTSecret, 3600, h.UUIDGen)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", types.InternalServerErrorResponse{Error: "Refresh token is invalid or has been expired"})
	}

	newRefreshToken, err := utils.CreateJWT(claims.UserID, claims.Username, claims.Email, config.Envs.JWTSecret, config.Envs.JWTExpirationInSeconds, h.UUIDGen)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", types.InternalServerErrorResponse{Error: "Refresh token is invalid or has been expired"})
	}

	newRefreshTokenClaims, err := utils.VerifyJWT(newRefreshToken, config.Envs.JWTSecret)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUserLogin", types.InternalServerErrorResponse{Error: "Refresh token is invalid or has been expired"})
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
		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleGetBookByID", types.NotFoundResponse{Error: fmt.Sprintf("No userID found with ID %d", newRefreshTokenClaims.UserID)})
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, err, "HandleRefreshToken", types.InternalServerErrorResponse{Error: "Internal Server Error"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, types.UpdateRefreshTokenResponse{AccessToken: newAccessToken, RefreshToken: newRefreshToken})
}
