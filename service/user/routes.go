package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/hoyci/book-store-api/types"
	"github.com/hoyci/book-store-api/utils"
)

var validate = validator.New()

func passwordValidator(sl validator.StructLevel) {
	data := sl.Current().Interface().(types.CreateUserRequestPayload)
	if data.Password != data.ConfirmPassword {
		sl.ReportError(data.ConfirmPassword, "ConfirmPassword", "ConfirmPassword", "password_mismatch", "")
	}
}

type UserHandler struct {
	userStore types.UserStore
}

func NewUserHandler(userStore types.UserStore) *UserHandler {
	validate.RegisterStructValidation(passwordValidator, types.CreateUserRequestPayload{})

	return &UserHandler{userStore: userStore}
}

func (h *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var requestPayload types.CreateUserRequestPayload
	if err := utils.ParseJSON(r, &requestPayload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleCreateUser", "Body is not a valid json")
		return
	}

	if err := validate.Struct(requestPayload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is invalid: %s", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, err, "HandleCreateUser", errorMessages)
		return
	}

	hashedPassword, err := utils.HashPassword(requestPayload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleCreateUser", "An unexpected error occurred")
	}

	var databasePayload = types.CreateUserDatabasePayload{
		Username:     requestPayload.Username,
		Email:        requestPayload.Email,
		PasswordHash: hashedPassword,
	}

	_, err = h.userStore.Create(r.Context(), databasePayload)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleGetBooks", "Request canceled")
			return
		}

		if err == sql.ErrConnDone {
			utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetBooks", "An unexpected error occurred")
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, err, "HandleCreateUser", "An unexpected error occurred")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]string{"message": "User successfully created"})
}

func (h *UserHandler) HandleGetUserByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetClaimFromContext[int](r, "UserID")
	if !ok {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to retrieve userID from context"), "HandleGetUserByID", "An unexpected error occurred")
		return
	}

	user, err := h.userStore.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleGetUserByID", "Request canceled")
			return
		}

		if err == sql.ErrConnDone {
			utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetUserByID", "An unexpected error occurred")
			return
		}

		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleGetUserByID", fmt.Sprintf("No user found with ID %d", userID))
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetUserByID", "An unexpected error occurred")
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

func (h *UserHandler) HandleUpdateUserByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetClaimFromContext[int](r, "UserID")
	if !ok {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to retrieve userID from context"), "HandleGetUserByID", "An unexpected error occurred")
		return
	}

	var payload types.UpdateUserPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleUpdateUserByID", "Body is not a valid json")
		return
	}

	if err := validate.Struct(payload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, err, "HandleUpdateUserByID", errorMessages)
		return
	}

	user, err := h.userStore.UpdateByID(r.Context(), userID, payload)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleUpdateUserByID", "Request canceled")
			return
		}

		if err == sql.ErrConnDone {
			utils.WriteError(w, http.StatusInternalServerError, err, "HandleUpdateUserByID", "An unexpected error occurred")
			return
		}

		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleUpdateUserByID", fmt.Sprintf("No user found with ID %d", userID))
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUpdateUserByID", "An unexpected error occurred")
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

func (h *UserHandler) HandleDeleteUserByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetClaimFromContext[int](r, "UserID")
	if !ok {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to retrieve userID from context"), "HandleGetUserByID", "An unexpected error occurred")
		return
	}

	err := h.userStore.DeleteByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleGetBooks", "Request canceled")
			return
		}

		if err == sql.ErrConnDone {
			utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetBooks", "An unexpected error occurred")
			return
		}

		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleDeleteUserByID", fmt.Sprintf("No user found with ID %d", userID))
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleDeleteUserByID", "An unexpected error occurred")
		return
	}

	utils.WriteJSON(w, http.StatusNoContent, nil)
}
