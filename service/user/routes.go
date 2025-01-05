package user

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/types"
	"github.com/hoyci/book-store-api/utils"
	"github.com/sirupsen/logrus"
)

var validate = validator.New()

func PasswordValidator(sl validator.StructLevel) {
	data := sl.Current().Interface().(types.CreateUserRequestPayload)
	if data.Password != data.ConfirmPassword {
		sl.ReportError(data.ConfirmPassword, "ConfirmPassword", "ConfirmPassword", "password_mismatch", "")
	}
}

type UserHandler struct {
	userStore types.UserStore
}

func NewUserHandler(userStore types.UserStore) *UserHandler {
	validate.RegisterStructValidation(PasswordValidator, types.CreateUserRequestPayload{})

	return &UserHandler{userStore: userStore}
}

func (h *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var requestPayload types.CreateUserRequestPayload
	if err := utils.ParseJSON(r, &requestPayload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleCreateUser", "User sent request with an invalid JSON", "Body is not a valid json")
		return
	}

	if err := validate.Struct(requestPayload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is invalid: %s", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, err, "HandleCreateUser", "User sent a request containing JSON with information outside the permitted format", errorMessages)
		return
	}

	hashedPassword, err := utils.HashPassword(requestPayload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleCreateUser", "An error occured during the hash password process", "An unexpected error occurred")
	}

	var databasePayload = types.CreateUserDatabasePayload{
		Username:     requestPayload.Username,
		Email:        requestPayload.Email,
		PasswordHash: hashedPassword,
	}

	newUser, err := h.userStore.Create(r.Context(), databasePayload)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleCreateUser", "Failed to insert user into database", "An unexpected error occurred")
		return
	}

	token, err := utils.CreateJWT(newUser.ID, newUser.Username, newUser.Email, config.Envs.JWTSecret, config.Envs.JWTExpirationInSeconds)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleCreateUser", "An error occured during the create JWT process", "An unexpected error occurred")
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]string{"token": token})
}

func (h *UserHandler) HandleGetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleGetUserByID", "User sent request with an invalid ID", "User ID must be a positive integer")
		return
	}

	user, err := h.userStore.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetUserByID", "Failed to get user by id from database", "An unexpected error occurred")
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

// TODO: Melhorar a validação dos campos que vem no JSON para que o usuário não possa burlar
func (h *UserHandler) HandleUpdateUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleUpdateUserByID", "User sent request with an invalid ID", "User ID must be a positive integer")
		return
	}

	var payload types.UpdateUserPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleUpdateUserByID", "User sent request with an invalid JSON", "Body is not a valid json")
		return
	}

	if err := validate.Struct(payload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, err, "HandleUpdateUserByID", "User sent a request with JSON outside the permitted format", errorMessages)
		return
	}

	user, err := h.userStore.UpdateByID(r.Context(), id, payload)
	if err != nil {
		utils.Log.WithFields(logrus.Fields{
			"error":   err.Error(),
			"context": "HandleUpdateUserByID",
		}).Error()
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUpdateUserByID", "Failed to update user by id in database", "An unexpected error occurred")
		return
	}

	if payload.Username == nil && payload.Email == nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUpdateBookByID", "User sent a request missing mandatory data", "Missing mandatory data")
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

func (h *UserHandler) HandleDeleteUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleDeleteUserByID", "User sent request with an invalid ID", "Book ID must be a positive integer")
		return
	}

	returnedID, err := h.userStore.DeleteByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleDeleteUserByID", "Failed to delete user by id from database", "An unexpected error occurred")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]int{"id": returnedID})
}
